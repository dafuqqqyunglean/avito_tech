package userrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/dafuqqqyunglean/avito_tech/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	SetActive(ctx context.Context, req domain.SetActiveRequest) (domain.SetActiveResponse, error)
	GetReview(ctx context.Context, userID string) (domain.GetReviewResponse, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) SetActive(ctx context.Context, req domain.SetActiveRequest) (domain.SetActiveResponse, error) {
	var user domain.SetActiveResponse
	err := r.db.QueryRow(ctx, `
		UPDATE users SET is_active = $2
		WHERE id = $1
		RETURNING id, username, is_active;`,
		req.UserID, req.IsActive).Scan(&user.UserID, &user.Username, &user.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SetActiveResponse{}, domain.ErrNotFound
	} else if err != nil {
		return domain.SetActiveResponse{}, fmt.Errorf("failed to set active status: %w", err)
	}

	err = r.db.QueryRow(ctx, `
        SELECT t.name 
        FROM teams t
        JOIN team_members tm ON t.id = tm.team_id
        WHERE tm.user_id = $1;`,
		req.UserID).Scan(&user.TeamName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SetActiveResponse{}, domain.ErrNotFound
	} else if err != nil {
		return domain.SetActiveResponse{}, fmt.Errorf("failed to get user team: %w", err)
	}

	return user, nil
}

func (r *repository) GetReview(ctx context.Context, userID string) (domain.GetReviewResponse, error) {
	rows, err := r.db.Query(ctx, `
	SELECT id, name, author_id, status
	FROM pull_request pr
	JOIN pr_reviewers rw ON pr.id = rw.pr_id
	WHERE rw.user_id = $1;`, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.GetReviewResponse{}, domain.ErrNotFound
	} else if err != nil {
		return domain.GetReviewResponse{}, fmt.Errorf("failed to get user reviews: %w", err)
	}
	defer rows.Close()

	response := domain.GetReviewResponse{
		UserID:       userID,
		PullRequests: []domain.CutPullRequest{},
	}

	for rows.Next() {
		var prID string
		var prName string
		var authorID string
		var status string

		err := rows.Scan(&prID, &prName, &authorID, &status)
		if err != nil {
			return domain.GetReviewResponse{}, fmt.Errorf("failed scan to pull request: %w", err)
		}

		response.PullRequests = append(response.PullRequests, domain.CutPullRequest{
			ID:       prID,
			Name:     prName,
			AuthorID: authorID,
			Status:   status,
		})
	}

	return response, nil
}
