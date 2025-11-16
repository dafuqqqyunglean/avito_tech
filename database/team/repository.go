package team

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
	CreateTeam(ctx context.Context, teamName string, members []domain.User) (domain.TeamRequest, error)
	GetTeam(ctx context.Context, teamName string) (domain.TeamRequest, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) Repository {
	return &repository{
		db: db,
	}
}

//go:embed sql/createTeam.sql
var createTeam string

//go:embed sql/createUsers.sql
var createUsers string

//go:embed sql/putUsersInTeam.sql
var putUsersInTeam string

func (r *repository) CreateTeam(ctx context.Context, teamName string, members []domain.User) (domain.TeamRequest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.TeamRequest{}, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var teamID int
	err = tx.QueryRow(ctx, createTeam, teamName).Scan(&teamID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TeamRequest{}, domain.ErrTeamExists
	}
	if err != nil {
		return domain.TeamRequest{}, fmt.Errorf("failed to create team: %w", err)
	}

	for _, v := range members {
		_, err := tx.Exec(ctx, createUsers,
			v.ID,
			v.Name,
			v.IsActive)
		if err != nil {
			return domain.TeamRequest{}, fmt.Errorf("failed to create/update user: %w", err)
		}

		_, err = tx.Exec(ctx, putUsersInTeam, teamID, v.ID)
		if err != nil {
			return domain.TeamRequest{}, fmt.Errorf("failed to add user to team: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.TeamRequest{}, fmt.Errorf("failed to commit changes: %w", err)
	}

	return domain.TeamRequest{TeamName: teamName, Members: members}, nil
}

//go:embed sql/getTeamMembers.sql
var getTeamMembers string

func (r *repository) GetTeam(ctx context.Context, teamName string) (domain.TeamRequest, error) {
	rows, err := r.db.Query(ctx, getTeamMembers, teamName)
	if err != nil {
		return domain.TeamRequest{}, fmt.Errorf("failed to query team: %w", err)
	}
	defer rows.Close()

	team := domain.TeamRequest{
		TeamName: teamName,
		Members:  []domain.User{},
	}

	for rows.Next() {
		var userID string
		var username string
		var isActive bool

		if err := rows.Scan(&userID, &username, &isActive); err != nil {
			return domain.TeamRequest{}, fmt.Errorf("scan team member: %w", err)
		}

		team.Members = append(team.Members, domain.User{
			ID:       userID,
			Name:     username,
			IsActive: isActive,
		})
	}

	if err := rows.Err(); err != nil {
		return domain.TeamRequest{}, fmt.Errorf("rows iteration error: %w", err)
	}

	if len(team.Members) == 0 {
		return domain.TeamRequest{}, domain.ErrNotFound
	}

	return team, nil
}
