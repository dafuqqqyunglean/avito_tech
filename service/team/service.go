package team

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	teamrepo "github.com/dafuqqqyunglean/avito_tech/database/team"
	"github.com/dafuqqqyunglean/avito_tech/domain"
	"github.com/dafuqqqyunglean/avito_tech/service/team/mapper"
)

type Service interface {
	CreateTeam(ctx context.Context, req domain.TeamRequest) (domain.TeamResponse, error)
	GetTeam(ctx context.Context, teamName string) (domain.TeamRequest, error)
}

type impl struct {
	repo teamrepo.Repository
}

func NewService(repo teamrepo.Repository) Service {
	return &impl{
		repo: repo,
	}
}

func (s *impl) CreateTeam(ctx context.Context, req domain.TeamRequest) (domain.TeamResponse, error) {
	err := s.validateTeam(req)
	if err != nil {
		slog.Error("team validation failed",
			"team_name", req.TeamName,
			"error", err)

		return domain.TeamResponse{}, domain.ErrBadRequest
	}

	res, err := s.repo.CreateTeam(ctx, req.TeamName, req.Members)
	if err != nil {
		slog.Error("failed to create team",
			"error", err,
			"team_name", req.TeamName)

		return domain.TeamResponse{}, err
	}

	slog.Info("team created successfully",
		"team_name", res.TeamName,
		"members_count", len(res.Members))

	return mapper.FromReqToResp(res), nil
}

func (s *impl) GetTeam(ctx context.Context, teamName string) (domain.TeamRequest, error) {
	if strings.TrimSpace(teamName) == "" {
		slog.Error("wrong team name", "team_name", teamName)
		return domain.TeamRequest{}, domain.ErrBadRequest
	}

	res, err := s.repo.GetTeam(ctx, teamName)
	if err != nil {
		slog.Error("failed to get team",
			"error", err,
			"team_name", teamName)

		return domain.TeamRequest{}, err
	}

	slog.Info("team retrieved successfully", "team_name", teamName, "members_count", len(res.Members))

	return res, nil
}

func (s *impl) validateTeam(team domain.TeamRequest) error {
	if strings.TrimSpace(team.TeamName) == "" {
		return fmt.Errorf("team name is required")
	}

	if len(team.Members) == 0 {
		return fmt.Errorf("team must have at least one member")
	}

	if len(team.TeamName) > 100 {
		return fmt.Errorf("team name too long")
	}

	seenUsers := make(map[string]bool)
	for i, member := range team.Members {
		if strings.TrimSpace(member.ID) == "" {
			return fmt.Errorf("member %d: user_id is required", i)
		}
		if strings.TrimSpace(member.Name) == "" {
			return fmt.Errorf("member %d: username is required", i)
		}
		if seenUsers[member.ID] {
			return fmt.Errorf("duplicate user_id: %s", member.ID)
		}
		seenUsers[member.ID] = true
	}

	return nil
}
