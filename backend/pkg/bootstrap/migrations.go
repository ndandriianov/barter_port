package bootstrap

import (
	"barter-port/pkg/db"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrationsFromConfig(cfg Config) error {
	if cfg.DB.MigrationsPath == "" {
		return errors.New("db.migrations_path is not set")
	}

	dsn := db.BuildDSN(db.Config{
		DBUser:     cfg.DB.User,
		DBPassword: cfg.DB.Password,
		DBHost:     cfg.DB.Host,
		DBPort:     cfg.DB.Port,
		DBName:     cfg.DB.Name,
	}, true)

	m, err := migrate.New(cfg.DB.MigrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		sourceErr, databaseErr := m.Close()
		if sourceErr != nil || databaseErr != nil {
			return fmt.Errorf("apply migrations: %w (close source err: %v, close db err: %v)", err, sourceErr, databaseErr)
		}
		return fmt.Errorf("apply migrations: %w", err)
	}

	sourceErr, databaseErr := m.Close()
	if sourceErr != nil || databaseErr != nil {
		return fmt.Errorf("close migrate instance: source err: %v, db err: %v", sourceErr, databaseErr)
	}

	return nil
}
