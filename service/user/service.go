package userserv

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	userrepo "github.com/dafuqqqyunglean/avito_tech/database/user"
	"github.com/dafuqqqyunglean/avito_tech/domain"
)

type Service interface {
	SetActive(ctx context.Context, req domain.SetActiveRequest) (domain.SetActiveResponse, error)
	GetReview(ctx context.Context, userID string) (domain.GetReviewResponse, error)
}

type impl struct {
	repo userrepo.Repository
}

func NewService(repo userrepo.Repository) Service {
	return &impl{
		repo: repo,
	}
}

func (s *impl) SetActive(ctx context.Context, req domain.SetActiveRequest) (domain.SetActiveResponse, error) {
	if err := s.validateSetActiveRequest(req); err != nil {
		slog.Error("wrong request format", "error", err)
		return domain.SetActiveResponse{}, domain.ErrBadRequest
	}

	user, err := s.repo.SetActive(ctx, req)
	if err != nil {
		slog.Error("failed to set user active status",
			"user_id", req.UserID,
			"is_active", req.IsActive,
			"error", err)

		return domain.SetActiveResponse{}, err
	}

	slog.Info("user active status updated",
		"user_id", user.UserID,
		"username", user.Username,
		"is_active", user.IsActive,
		"team", user.TeamName)

	return user, nil
}

func (s *impl) GetReview(ctx context.Context, userID string) (domain.GetReviewResponse, error) {
	if strings.TrimSpace(userID) == "" {
		slog.Error("wrong user id", "user_id", userID)
		return domain.GetReviewResponse{}, domain.ErrBadRequest
	}

	resp, err := s.repo.GetReview(ctx, userID)
	if err != nil {
		slog.Error("failed to get user pull requests",
			"user_id", userID,
			"error", err)

		return domain.GetReviewResponse{}, err
	}

	slog.Info("user reviews retrieved",
		"user_id", userID,
		"pr_count", len(resp.PullRequests))

	return resp, nil
}

func (s *impl) validateSetActiveRequest(req domain.SetActiveRequest) error {
	if strings.TrimSpace(req.UserID) == "" {
		return fmt.Errorf("user_id is required")
	}

	if !strings.HasPrefix(req.UserID, "u") {
		return fmt.Errorf("invalid user_id format")
	}

	return nil
}
