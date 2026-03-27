package main

import (
	"barter-port/internal/auth/app"
	"barter-port/pkg/bootstrap"
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

//go:generate bash ../../scripts/generate-swagger-auth.sh

// @title Barter Port API
// @version 1.0.0
// @description API for Barter Port
// @host localhost:80
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	_ = godotenv.Load()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("auth - load config: %v", err)
	}

	err = bootstrap.RunMigrationsFromConfig(cfg)
	if err != nil {
		log.Fatalf("auth - run migrations: %v", err)
	}

	authApp, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("auth - new app: %v", err)
	}
	defer authApp.Close()

	err = authApp.Run()
	if err != nil {
		log.Fatalf("auth - run: %v", err)
	}
}

func loadConfig() (bootstrap.Config, error) {
	cfg, err := bootstrap.LoadConfig(bootstrap.ConfigOptions{
		CommonPath:  os.Getenv("CONFIG_COMMON"),
		ServicePath: resolveServiceConfigPath(),
		AppEnv:      os.Getenv("APP_ENV"),
	})
	if err != nil {
		return bootstrap.Config{}, errors.New("failed to load config: " + err.Error())
	}

	return cfg, nil
}

func resolveServiceConfigPath() string {
	if path := os.Getenv("CONFIG_SERVICE"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	const localPath = "./config/auth.yaml"
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}

	return os.Getenv("CONFIG_SERVICE")
}
