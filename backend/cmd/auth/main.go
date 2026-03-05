package main

import (
	"barter-port/internal/auth/repository/email_token"
	"barter-port/internal/auth/repository/refresh_token"
	"barter-port/internal/auth/repository/user"
	"barter-port/internal/auth/service"
	"barter-port/internal/auth/transport"
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
// @host localhost:8081
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	_ = godotenv.Load()

	cfg, err := bootstrap.LoadConfig(bootstrap.ConfigOptions{
		CommonPath:  os.Getenv("CONFIG_COMMON"),
		ServicePath: os.Getenv("CONFIG_SERVICE"),
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

	userRepo := user.NewRepository(db)
	emailTokenRepo := email_token.NewRepository(db)
	refreshTokenRepo := refresh_token.NewRepository(db)

	m, err := bootstrap.InitMailerFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize mailer:", err)
	}

	logg := logger.NewJSONLogger(slog.LevelDebug, "auth-service", "")
	infrastructureLogger := logger.NewJSONLogger(slog.LevelDebug, "", "infrastructure")

	jwtManager := bootstrap.InitJWTManager()
	validator := bootstrap.InitLocalJWT()

	authService := service.NewService(userRepo, emailTokenRepo, m, infrastructureLogger, frontendURL, re)
	handlers := transport.NewHandlers(logg, authService, jwtManager, refreshTokenRepo)
	router := transport.NewRouter(logg, validator, handlers)

	addr := ":8081"
	log.Println("backend listening on", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
