package prrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dafuqqqyunglean/avito_tech/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, pr domain.CreatePRRequest) (domain.CreatePRResponse, error)
	SetMerged(ctx context.Context, prID string) (domain.MergePRResponse, error)
	Reassign(ctx context.Context, prID, userID string) (domain.ReassignPRResponse, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, pr domain.CreatePRRequest) (domain.CreatePRResponse, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)`,
		pr.PRID).Scan(&exists)
	if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to check PR existence: %w", err)
	}
	if exists {
		return domain.CreatePRResponse{}, domain.ErrPRExists
	}

	var teamName string
	err = r.db.QueryRow(ctx, `
	SELECT name FROM teams t
	JOIN team_members tm ON t.id = tm.team_id
	WHERE tm.user_id = $1;`, pr.AuthorID).Scan(&teamName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CreatePRResponse{}, domain.ErrNotFound
	} else if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to get user team: %w", err)
	}

	rows, err := r.db.Query(ctx, `
	SELECT user_id FROM team_members tm
	JOIN team t ON tm.team_id = t.id
	JOIN users u ON tm.user_id = u.id
	WHERE user_id != $1
	AND t.name = $2
	AND u.is_active = true
	LIMIT 2;`, pr.AuthorID, teamName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CreatePRResponse{}, domain.ErrNotFound
	} else if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to get user team: %w", err)
	}

	resp := domain.CreatePRResponse{
		PR: domain.PullRequest{
			ID:        pr.PRID,
			Name:      pr.PRName,
			AuthorID:  pr.AuthorID,
			Status:    "OPEN",
			Reviewers: []string{},
		},
	}

	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return domain.CreatePRResponse{}, fmt.Errorf("scan team member: %w", err)
		}

		resp.PR.Reviewers = append(resp.PR.Reviewers, userID)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to create transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
	INSERT (id, name, author_id, status) 
	INTO pull_requests
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO NOTHING;`, resp.PR.ID, resp.PR.Name, resp.PR.AuthorID, resp.PR.Status)
	if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to save pr to database: %w", err)
	}

	for _, reviewerID := range resp.PR.Reviewers {
		_, err := tx.Exec(ctx, `
        INSERT INTO pr_reviewers (pr_id, user_id) 
        VALUES ($1, $2)`,
			resp.PR.ID, reviewerID)
		if err != nil {
			return domain.CreatePRResponse{}, fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("transction uncommited: %w", err)
	}

	return resp, nil
}

func (r *repository) SetMerged(ctx context.Context, prID string) (domain.MergePRResponse, error) {
	result, err := r.db.Exec(ctx, `
	UPDATE pull_requests
	SET status = "MERGED", merged_at = $2
	WHERE id = $1 AND status != 'MERGED`, prID, time.Now())
	if err != nil {
		return domain.MergePRResponse{}, fmt.Errorf("failed to set pr status = merged %s: %w", prID, err)
	}
	if result.RowsAffected() == 0 {
		return domain.MergePRResponse{}, domain.ErrNotFound
	}

	rows, err := r.db.Query(ctx, `
	SELECT name, author_id, status, user_id, merged_at
	FROM pull_requests pr
	JOIN pr_reviewers rv ON pr.id = rv.pr_id
	WHERE pr.id = $1;`, prID)
	if err != nil {

	}
	defer rows.Close()

	resp := domain.MergePRResponse{
		PR: domain.PullRequest{
			ID:        prID,
			Reviewers: []string{},
		},
	}
	for rows.Next() {
		var name string
		var authorID string
		var status string
		var reviewerID string

		err := rows.Scan(&name, &authorID, &status, &reviewerID, &resp.MergedAt)
		if err != nil {
			return domain.MergePRResponse{}, fmt.Errorf("scan pull request error: %w", err)
		}

		resp.PR.Name = name
		resp.PR.AuthorID = authorID
		resp.PR.Status = status
		resp.PR.Reviewers = append(resp.PR.Reviewers, reviewerID)
	}

	return resp, nil
}

func (r *repository) Reassign(ctx context.Context, prID, userID string) (domain.ReassignPRResponse, error) {
	var status string
	err := r.db.QueryRow(ctx, `
	SELECT status FROM pull_requests
	WHERE id = $1`, prID).Scan(&status)
	if err == pgx.ErrNoRows {
		return domain.ReassignPRResponse{}, domain.ErrNotFound
	} else if err != nil {
		return domain.ReassignPRResponse{}, fmt.Errorf("failed to get pr status: %w", err)
	}

	if status == "MERGED" {
		return domain.ReassignPRResponse{}, domain.ErrPRMerged
	}

	var newID string
	err = r.db.QueryRow(ctx, `
	UPDATE pr_reviewers 
	SET pr_id = (
	SELECT u.id 
    FROM users u
    JOIN team_members tm ON u.id = tm.user_id
    JOIN team_members old_tm ON tm.team_id = old_tm.team_id
    WHERE old_tm.user_id = $2
    AND u.id != $2
    AND u.is_active = true
    AND u.id NOT IN (
        SELECT user_id 
        FROM pr_reviewers 
        WHERE pr_id = $1
    )
    LIMIT 1)
	WHERE pr_id = $1 AND user_id = $2`, prID, userID).Scan(&newID)
	if err != nil {
		return domain.ReassignPRResponse{}, fmt.Errorf("failed to update pr reviewer: %w", err)
	}

	rows, err := r.db.Query(ctx, `
	SELECT name, author_id, status, user_id
	FROM pull_requests pr
	JOIN pr_reviewers rv ON pr.id = rv.pr_id
	WHERE pr.id = $1`, prID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ReassignPRResponse{}, domain.ErrNotFound
	} else if err != nil {
		return domain.ReassignPRResponse{}, fmt.Errorf("team has no reviewers: %w", err)
	}

	resp := domain.ReassignPRResponse{
		PR: domain.PullRequest{
			ID: prID,
		},
		ReplacedBy: newID,
	}
	for rows.Next() {
		var name string
		var authorID string
		var status string
		var reviewerID string

		err := rows.Scan(&name, &authorID, &status, &reviewerID)
		if err != nil {
			return domain.ReassignPRResponse{}, fmt.Errorf("scan pull request error: %w", err)
		}

		resp.PR.Name = name
		resp.PR.AuthorID = authorID
		resp.PR.Status = status
		resp.PR.Reviewers = append(resp.PR.Reviewers, reviewerID)
	}

	return resp, nil
}
