package main

import (
	"barter-port/internal/users/app"
	"barter-port/pkg/bootstrap"
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("users - load config: %v", err)
	}

	err = bootstrap.RunMigrationsFromConfig(cfg)
	if err != nil {
		log.Fatalf("users - run migrations: %v", err)
	}

	usersApp, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("users - new app: %v", err)
	}

	err = usersApp.Run()
	if err != nil {
		log.Fatalf("users - run: %v", err)
	}
}

func loadConfig() (bootstrap.Config, error) {
	cfg, err := bootstrap.LoadConfig(bootstrap.ConfigOptions{
		CommonPath:  os.Getenv("CONFIG_COMMON"),
		ServicePath: os.Getenv("CONFIG_SERVICE"),
		AppEnv:      os.Getenv("APP_ENV"),
	})
	if err != nil {
		return bootstrap.Config{}, errors.New("failed to load config: " + err.Error())
	}

	return cfg, nil
}
