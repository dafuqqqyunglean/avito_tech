package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dafuqqqyunglean/avito_tech/api"
	"github.com/dafuqqqyunglean/avito_tech/config"
	prrepo "github.com/dafuqqqyunglean/avito_tech/database/pr"
	teamrepo "github.com/dafuqqqyunglean/avito_tech/database/team"
	userrepo "github.com/dafuqqqyunglean/avito_tech/database/user"
	prserv "github.com/dafuqqqyunglean/avito_tech/service/pr"
	teamserv "github.com/dafuqqqyunglean/avito_tech/service/team"
	userserv "github.com/dafuqqqyunglean/avito_tech/service/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		slog.Warn(".env file not found", "error", err)
		return fmt.Errorf(".env file not found: %w", err)
	}

	config, err := config.NewConfig()
	if err != nil {
		slog.Warn("failed to read config", "error", err)
		return fmt.Errorf("failed to read config: %w", err)
	}

	pool, err := a.initDatabase(config)
	if err != nil {
		slog.Warn("failed to connect to db", "error", err)
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	err = a.runMigrations(config)
	if err != nil {
		slog.Warn("failed to apply migrations", "error", err)
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	server := api.NewServer(ctx, config)

	a.initService(ctx, pool, server)

	go func() {
		if err := server.Run(); err != nil {
			slog.Error("error occured while running http server", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("server running")
	return nil
}

func (a *App) initDatabase(config config.Config) (*pgxpool.Pool, error) {
	slog.Info("connecting to DB", "conn", config.DBConnectionString)

	var pool *pgxpool.Pool
	var err error

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		pool, err = pgxpool.New(ctx, config.DBConnectionString)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				cancel()
				slog.Info("successfully connected to DB")
				return pool, nil
			} else {
				err = pingErr
			}
		}

		cancel()

		slog.Warn(fmt.Sprintf("DB not ready, attempt %d: %v\n", i+1, err))
		time.Sleep(5 * time.Second)
	}

	return pool, fmt.Errorf("failed to connect to DB after 10 attempts: %w", err)
}

func (a *App) runMigrations(config config.Config) error {
	slog.Info("running database migrations")

	db, err := sql.Open("pgx", config.DBConnectionString)
	if err != nil {
		return fmt.Errorf("failed to open sql.DB for migrations: %w", err)
	}
	defer db.Close()

	if err := goose.Up(db, "./migrations"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	slog.Info("migrations applied successfully")

	return nil
}

func (a *App) initService(ctx context.Context, db *pgxpool.Pool, server *api.Server) {
	teamService := teamserv.NewService(teamrepo.NewRepo(db))
	userService := userserv.NewService(userrepo.NewRepo(db))
	prService := prserv.NewService(prrepo.NewRepo(db))

	server.HandleRoutes(ctx, teamService, userService, prService)
}
