package teamrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/dafuqqqyunglean/avito_tech/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateTeam(ctx context.Context, team domain.TeamRequest) (domain.TeamRequest, error)
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

func (r *repository) CreateTeam(ctx context.Context, team domain.TeamRequest) (domain.TeamRequest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.TeamRequest{}, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var teamID int
	err = tx.QueryRow(ctx, `
	INSERT INTO teams (name) 
	VALUES ($1) 
	ON CONFLICT (name) DO NOTHING
	RETURNING id;`, team.TeamName).Scan(&teamID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TeamRequest{}, domain.ErrTeamExists
	} else if err != nil {
		return domain.TeamRequest{}, fmt.Errorf("failed to create team: %w", err)
	}

	for _, v := range team.Members {
		_, err := tx.Exec(ctx,
			`INSERT INTO users (id, username, is_active) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (id) DO UPDATE SET 
    		username = EXCLUDED.username,
    		is_active = EXCLUDED.is_active;`,
			v.ID, v.Name, v.IsActive)
		if err != nil {
			return domain.TeamRequest{}, fmt.Errorf("failed to create/update user: %w", err)
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO team_members (team_id, user_id) 
			VALUES ($1, $2) 
			ON CONFLICT (team_id, user_id) DO NOTHING;`,
			teamID, v.ID)
		if err != nil {
			return domain.TeamRequest{}, fmt.Errorf("failed to add user to team: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.TeamRequest{}, fmt.Errorf("failed to commit changes: %w", err)
	}

	return team, nil
}

func (r *repository) GetTeam(ctx context.Context, teamName string) (domain.TeamRequest, error) {
	rows, err := r.db.Query(ctx,
		`SELECT 
            u.id as user_internal_id,
            u.username,
            u.is_active
         FROM teams t
         JOIN team_members tm ON t.id = tm.team_id
         JOIN users u ON tm.user_id = u.id
         WHERE t.name = $1
         ORDER BY u.id;`, teamName)
	if err != nil {
		return domain.TeamRequest{}, fmt.Errorf("failed to query team: %w", err)
	}
	defer rows.Close()

	team := domain.TeamRequest{
		TeamName: teamName,
		Members:  []domain.User{},
	}

	found := false

	for rows.Next() {
		found = true

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

	if !found {
		return domain.TeamRequest{}, domain.ErrNotFound
	}

	return team, nil
}
