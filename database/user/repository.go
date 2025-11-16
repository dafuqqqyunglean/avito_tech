package user

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/dafuqqqyunglean/avito_tech/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	SetActive(ctx context.Context, userID string, isActive bool) (domain.SetActiveResponse, error)
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

//go:embed sql/setUserActive.sql
var setUserActive string

//go:embed sql/getTeamName.sql
var getTeamName string

func (r *repository) SetActive(ctx context.Context, userID string, isActive bool) (domain.SetActiveResponse, error) {
	var user domain.SetActiveResponse
	err := r.db.QueryRow(ctx, setUserActive,
		userID,
		isActive).Scan(&user.UserID,
		&user.Username,
		&user.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SetActiveResponse{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SetActiveResponse{}, fmt.Errorf("failed to set active status: %w", err)
	}

	err = r.db.QueryRow(ctx, getTeamName,
		userID).Scan(&user.TeamName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SetActiveResponse{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SetActiveResponse{}, fmt.Errorf("failed to get user team: %w", err)
	}

	return user, nil
}

var getPR string

func (r *repository) GetReview(ctx context.Context, userID string) (domain.GetReviewResponse, error) {
	rows, err := r.db.Query(ctx, getPR, userID)
	if err != nil {
		return domain.GetReviewResponse{}, fmt.Errorf("failed to get user reviews: %w", err)
	}
	defer rows.Close()

	response := domain.GetReviewResponse{
		UserID:       userID,
		PullRequests: []domain.CutPullRequest{},
	}

	for rows.Next() {
		var prID, prName, authorID, status string

		if err := rows.Scan(&prID, &prName, &authorID, &status); err != nil {
			return domain.GetReviewResponse{}, fmt.Errorf("failed to scan pull request: %w", err)
		}

		response.PullRequests = append(response.PullRequests, domain.CutPullRequest{
			ID:       prID,
			Name:     prName,
			AuthorID: authorID,
			Status:   status,
		})
	}

	if err := rows.Err(); err != nil {
		return domain.GetReviewResponse{}, fmt.Errorf("rows iteration error: %w", err)
	}

	if len(response.PullRequests) == 0 {
		return domain.GetReviewResponse{}, domain.ErrNotFound
	}

	return response, nil
}
