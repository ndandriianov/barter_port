package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DBUser     string `yaml:"db_user" env-required:"true"`
	DBPassword string `yaml:"db_password" env-required:"true"`
	DBHost     string `yaml:"db_host" env-required:"true"`
	DBPort     string `yaml:"db_port" env-required:"true"`
	DBName     string `yaml:"db_name" env-required:"true"`
}

func NewPostgres(config Config) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), BuildDSN(config, false))
}

func BuildDSN(config Config, sslModeDisabled bool) string {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

	if sslModeDisabled {
		dsn += "?sslmode=disable"
	}

	return dsn
}
