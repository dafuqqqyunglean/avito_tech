package config

import (
	"fmt"
	"os"
)

const (
	serverPortEnv = "SERVER_PORT"
	dbHostEnv     = "DB_HOST"
	dbPortEnv     = "DB_PORT"
	dbUserEnv     = "DB_USER"
	dbNameEnv     = "DB_NAME"
	dbPasswordEnv = "DB_PASSWORD"
	dbSSLModeEnv  = "DB_SSLMODE"
)

type Config struct {
	ServerPort         string
	DBConnectionString string
}

func NewConfig() (Config, error) {
	return Config{
		ServerPort: fmt.Sprintf(":%s", os.Getenv(serverPortEnv)),
		DBConnectionString: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			os.Getenv(dbUserEnv), os.Getenv(dbPasswordEnv), os.Getenv(dbHostEnv),
			os.Getenv(dbPortEnv), os.Getenv(dbNameEnv), os.Getenv(dbSSLModeEnv)),
	}, nil
}
