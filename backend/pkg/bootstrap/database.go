package bootstrap

import (
	"barter-port/pkg/db"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDatabaseFromConfig(cfg Config) (*pgxpool.Pool, error) {
	dbConfig := db.Config{
		DBUser:     cfg.DB.User,
		DBPassword: cfg.DB.Password,
		DBHost:     cfg.DB.Host,
		DBPort:     cfg.DB.Port,
		DBName:     cfg.DB.Name,
	}
	db, err := db.NewPostgres(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
