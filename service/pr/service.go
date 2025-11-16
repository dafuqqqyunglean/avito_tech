package prserv

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	prrepo "github.com/dafuqqqyunglean/avito_tech/database/pr"
	"github.com/dafuqqqyunglean/avito_tech/domain"
)

type Service interface {
	Create(ctx context.Context, pr domain.CreatePRRequest) (domain.CreatePRResponse, error)
	SetMerged(ctx context.Context, prID string) (domain.MergePRResponse, error)
	Reassign(ctx context.Context, prID, userID string) (domain.ReassignPRResponse, error)
}

type impl struct {
	repo prrepo.Repository
}

func NewService(repo prrepo.Repository) Service {
	return &impl{
		repo: repo,
	}
}

func (s *impl) Create(ctx context.Context, pr domain.CreatePRRequest) (domain.CreatePRResponse, error) {
	if err := s.validateCreateRequest(pr); err != nil {
		slog.Error("PR creation validation failed",
			"pr_id", pr.PRID,
			"error", err)

		return domain.CreatePRResponse{}, domain.ErrBadRequest
	}

	resp, err := s.repo.Create(ctx, pr)
	if err != nil {
		slog.Error("failed to create pull request",
			"pr_id", pr.PRID,
			"pr_name", pr.PRName,
			"author_id", pr.AuthorID,
			"error", err)

		return domain.CreatePRResponse{}, err
	}

	slog.Info("PR created successfully",
		"pr_id", resp.PR.ID,
		"author", resp.PR.AuthorID,
		"reviewers_count", len(resp.PR.Reviewers),
	)

	return resp, nil
}

func (s *impl) SetMerged(ctx context.Context, prID string) (domain.MergePRResponse, error) {
	if strings.TrimSpace(prID) == "" {
		return domain.MergePRResponse{}, domain.ErrBadRequest
	}

	resp, err := s.repo.SetMerged(ctx, prID)
	if err != nil {
		slog.Error("failed to create pull request",
			"pr_id", prID,
			"error", err)

		return domain.MergePRResponse{}, err
	}

	slog.Info("PR merged successfully",
		"pr_id", prID,
		"merged_at", resp.MergedAt)

	return resp, nil
}

func (s *impl) Reassign(ctx context.Context, prID, userID string) (domain.ReassignPRResponse, error) {
	if err := s.validateReassignRequest(prID, userID); err != nil {
		slog.Error("reassign validation failed",
			"pr_id", prID,
			"old_reviewer", userID,
			"error", err)

		return domain.ReassignPRResponse{}, domain.ErrBadRequest
	}

	resp, err := s.repo.Reassign(ctx, prID, userID)
	if err != nil {
		slog.Error("failed to reassign reviewer",
			"pr_id", prID,
			"old_reviewer", userID,
			"error", err)

		return domain.ReassignPRResponse{}, err
	}

	slog.Info("reviewer reassigned successfully",
		"pr_id", prID,
		"old_reviewer", userID,
		"new_reviewer", resp.ReplacedBy)

	return resp, nil
}

func (s *impl) validateCreateRequest(req domain.CreatePRRequest) error {
	if strings.TrimSpace(req.PRID) == "" {
		return fmt.Errorf("PR ID is required")
	}
	if strings.TrimSpace(req.PRName) == "" {
		return fmt.Errorf("PR name is required")
	}
	if strings.TrimSpace(req.AuthorID) == "" {
		return fmt.Errorf("author ID is required")
	}
	if !strings.HasPrefix(req.PRID, "pr-") {
		return fmt.Errorf("PR ID must start with 'pr-'")
	}

	return nil
}

func (s *impl) validateReassignRequest(prID, userID string) error {
	if strings.TrimSpace(prID) == "" {
		return fmt.Errorf("PR ID is required")
	}
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("old reviewer ID is required")
	}
	if !strings.HasPrefix(prID, "pr-") {
		return fmt.Errorf("PR ID must start with 'pr-'")
	}
	if !strings.HasPrefix(userID, "u") {
		return fmt.Errorf("reviewer ID must start with 'u'")
	}

	return nil
}
