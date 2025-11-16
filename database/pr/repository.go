package pr

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/dafuqqqyunglean/avito_tech/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, prID, prName, AuthorID string) (domain.CreatePRResponse, error)
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

//go:embed sql/checkIfPRExists.sql
var checkIfPRExists string

//go:embed sql/getTeamName.sql
var getTeamName string

//go:embed sql/selectReviewersFromTeam.sql
var selectReviewersFromTeam string

//go:embed sql/createPullRequest.sql
var createPullRequest string

//go:embed sql/admitReviewers.sql
var admitReviewers string

func (r *repository) Create(ctx context.Context, prID, prName, authorID string) (domain.CreatePRResponse, error) {
	var exists bool
	err := r.db.QueryRow(ctx, checkIfPRExists, prID).Scan(&exists)
	if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to check PR existence: %w", err)
	}
	if exists {
		return domain.CreatePRResponse{}, domain.ErrPRExists
	}

	var teamName string
	err = r.db.QueryRow(ctx, getTeamName, authorID).Scan(&teamName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CreatePRResponse{}, domain.ErrNotFound
	} else if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to get user team: %w", err)
	}

	rows, err := r.db.Query(ctx, selectReviewersFromTeam, authorID, teamName)
	if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	resp := domain.CreatePRResponse{
		PR: domain.PullRequest{
			ID:        prID,
			Name:      prName,
			AuthorID:  authorID,
			Status:    "OPEN",
			Reviewers: []string{},
		},
	}

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return domain.CreatePRResponse{}, fmt.Errorf("scan team member: %w", err)
		}

		resp.PR.Reviewers = append(resp.PR.Reviewers, userID)
	}

	if err := rows.Err(); err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("rows iteration error: %w", err)
	}

	if len(resp.PR.Reviewers) == 0 {
		return domain.CreatePRResponse{}, domain.ErrNotFound
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to create transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, createPullRequest,
		resp.PR.ID,
		resp.PR.Name,
		resp.PR.AuthorID,
		resp.PR.Status)
	if err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("failed to save pr to database: %w", err)
	}

	for _, reviewerID := range resp.PR.Reviewers {
		_, err := tx.Exec(ctx, admitReviewers,
			resp.PR.ID,
			reviewerID)
		if err != nil {
			return domain.CreatePRResponse{}, fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.CreatePRResponse{}, fmt.Errorf("transaction uncommitted: %w", err)
	}

	return resp, nil
}

//go:embed sql/getPRStatus.sql
var getPRStatus string

//go:embed sql/setMergedStatus.sql
var setMergedStatus string

//go:embed sql/getPR.sql
var getPR string

func (r *repository) SetMerged(ctx context.Context, prID string) (domain.MergePRResponse, error) {
	var status string

	err := r.db.QueryRow(ctx, getPRStatus, prID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MergePRResponse{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.MergePRResponse{}, fmt.Errorf("failed to get pr status %s: %w", prID, err)
	}

	now := time.Now()

	if status != "MERGED" {
		_, err := r.db.Exec(ctx, setMergedStatus, prID, now)
		if err != nil {
			return domain.MergePRResponse{}, fmt.Errorf("failed to set pr status = merged %s: %w", prID, err)
		}

		status = "MERGED"
	}

	rows, err := r.db.Query(ctx, getPR, prID)
	if err != nil {
		return domain.MergePRResponse{}, fmt.Errorf("failed to load merged pr %s: %w", prID, err)
	}
	defer rows.Close()

	resp := domain.MergePRResponse{
		PR: domain.PullRequest{
			ID:        prID,
			Reviewers: []string{},
		},
	}

	found := false

	for rows.Next() {
		found = true

		var (
			name       string
			authorID   string
			prStatus   string
			reviewerID string
			dbMergedAt time.Time
		)

		if err := rows.Scan(&name, &authorID, &prStatus, &reviewerID, &dbMergedAt); err != nil {
			return domain.MergePRResponse{}, fmt.Errorf("scan pull request error: %w", err)
		}

		resp.PR.Name = name
		resp.PR.AuthorID = authorID
		resp.PR.Status = prStatus
		resp.PR.Reviewers = append(resp.PR.Reviewers, reviewerID)
		resp.MergedAt = dbMergedAt
	}

	if err := rows.Err(); err != nil {
		return domain.MergePRResponse{}, fmt.Errorf("rows iteration error: %w", err)
	}

	if !found {
		return domain.MergePRResponse{}, domain.ErrNotFound
	}

	return resp, nil
}

//go:embed sql/reassignReviewer.sql
var reassignReviewer string

//go:embed sql/getReassignPR.sql
var getReassignPR string

func (r *repository) Reassign(ctx context.Context, prID, userID string) (domain.ReassignPRResponse, error) {
	var status string
	err := r.db.QueryRow(ctx, getPRStatus, prID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ReassignPRResponse{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ReassignPRResponse{}, fmt.Errorf("failed to get pr status: %w", err)
	}
	if status == "MERGED" {
		return domain.ReassignPRResponse{}, domain.ErrPRMerged
	}

	var newID string
	err = r.db.QueryRow(ctx, reassignReviewer, prID, userID).Scan(&newID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ReassignPRResponse{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ReassignPRResponse{}, fmt.Errorf("failed to update pr reviewer: %w", err)
	}

	rows, err := r.db.Query(ctx, getReassignPR, prID)
	if err != nil {
		return domain.ReassignPRResponse{}, fmt.Errorf("failed to load pr reviewers: %w", err)
	}
	defer rows.Close()

	resp := domain.ReassignPRResponse{
		PR: domain.PullRequest{
			ID:        prID,
			Reviewers: []string{},
		},
		ReplacedBy: newID,
	}

	found := false

	for rows.Next() {
		found = true

		var (
			name       string
			authorID   string
			prStatus   string
			reviewerID string
		)

		if err := rows.Scan(&name, &authorID, &prStatus, &reviewerID); err != nil {
			return domain.ReassignPRResponse{}, fmt.Errorf("scan pull request error: %w", err)
		}

		resp.PR.Name = name
		resp.PR.AuthorID = authorID
		resp.PR.Status = prStatus
		resp.PR.Reviewers = append(resp.PR.Reviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return domain.ReassignPRResponse{}, fmt.Errorf("rows iteration error: %w", err)
	}

	if !found {
		return domain.ReassignPRResponse{}, domain.ErrNotFound
	}

	return resp, nil
}
