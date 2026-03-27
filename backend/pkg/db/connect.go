package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DBUser     string `yaml:"db_user" env-required:"true"`
	DBPassword string `yaml:"db_password" env-required:"true"`
	DBHost     string `yaml:"db_host" env-required:"true"`
	DBPort     string `yaml:"db_port" env-required:"true"`
	DBName     string `yaml:"db_name" env-required:"true"`
}

func MustLoad(configPath string) *Config {
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
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
