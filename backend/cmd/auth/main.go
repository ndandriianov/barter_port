package main

import (
	"barter-port/internal/auth/application"
	"barter-port/internal/auth/infrastructure/repository/email_token"
	"barter-port/internal/auth/infrastructure/repository/refresh_token"
	"barter-port/internal/auth/infrastructure/repository/user"
	"barter-port/internal/auth/infrastructure/transport"
	"barter-port/internal/libs/bootstrap"
	"barter-port/internal/libs/platform/logger"
	"log/slog"
	"os"
	"regexp"

	"github.com/joho/godotenv"

	"log"
	"net/http"
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

	//serviceName := bootstrap.GetEnv("SERVICE_NAME", "auth")
	serviceConfigPath := "" //fmt.Sprintf("./config/%s.yaml", serviceName)

	cfg, err := bootstrap.LoadConfig(bootstrap.ConfigOptions{
		CommonPath:  os.Getenv("CONFIG_COMMON"),
		ServicePath: serviceConfigPath,
		AppEnv:      os.Getenv("APP_ENV"),
	})
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer db.Close()

	frontendURL := cfg.Frontend.URL
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	userRepo := user.NewRepository()
	emailTokenRepo := email_token.NewRepository()
	refreshTokenRepo := refresh_token.NewRepository()

	m := bootstrap.InitMailerFromConfig(cfg)
	if err = bootstrap.ValidateMailConfig(cfg); err != nil {
		log.Fatal("failed to initialize mailer:", err)
	}

	logg := logger.NewJSONLogger(slog.LevelDebug, "auth-service", "")
	infrastructureLogger := logger.NewJSONLogger(slog.LevelDebug, "", "infrastructure")

	jwtManager, err := bootstrap.InitJWTManagerFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize JWT manager:", err)
	}

	validator, err := bootstrap.InitLocalJWTFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize JWT validator:", err)
	}

	authService := application.NewService(db, userRepo, emailTokenRepo, m, infrastructureLogger, frontendURL, re)
	handlers := transport.NewHandlers(logg, authService, jwtManager, db, refreshTokenRepo)
	router := transport.NewRouter(logg, validator, handlers)

	port := bootstrap.InitPortStringFromConfig(cfg, 8081)
	log.Println("backend listening on", port)
	log.Fatal(http.ListenAndServe(port, router))
}
