package main

import (
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/logger"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/mailer"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/repository/email_token"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/repository/refresh_token"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/repository/user"
	"github.com/ndandriianov/barter_port/backend/internal/service/auth"
	"github.com/ndandriianov/barter_port/backend/internal/service/auth/jwt"
	"github.com/ndandriianov/barter_port/backend/internal/transport"

	"log"
	"net/http"
	"os"
)

//go:generate sh -c "cd ../.. && swag init -g cmd/api/main.go --parseInternal --parseDependency"
func main() {
	_ = godotenv.Load()

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")

	accessSecret := getEnv("ACCESS_SECRET", "")
	refreshSecret := getEnv("REFRESH_SECRET", "")
	accessTTL := getEnv("ACCESS_TTL", "")
	refreshTTL := getEnv("REFRESH_TTL", "")
	accessTTLMinutes := time.Duration(mustInt(accessTTL)) * time.Minute
	refreshTTLMinutes := time.Duration(mustInt(refreshTTL)) * time.Minute

	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	userRepo := user.NewInMemoryUserRepo()
	emailTokenRepo := email_token.NewInMemoryTokenRepo()
	refreshTokenRepo := refresh_token.NewInMemoryRefreshRepo()

	smtpHost := getEnv("SMTP_HOST", "")
	smtpPort := mustInt(getEnv("SMTP_PORT", ""))
	smtpUser := getEnv("SMTP_USER", "")
	smtpPass := getEnv("SMTP_PASS", "")
	smtpFrom := getEnv("SMTP_FROM", smtpUser)

	if smtpHost == "" {
		log.Fatal("SMTP_HOST is required")
	}

	m := mailer.NewSMTPMailer(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)

	logg := logger.NewJSONLogger(slog.LevelDebug, "auth-service", "")
	infrastructureLogger := logger.NewJSONLogger(slog.LevelDebug, "", "infrastructure")

	jwtManager := jwt.NewManager(jwt.Config{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		AccessTTL:     accessTTLMinutes,
		RefreshTTL:    refreshTTLMinutes,
	})

	authService := auth.NewService(
		userRepo,
		emailTokenRepo,
		m,
		infrastructureLogger,

		frontendURL,
		jwtManager,
		re,
	)

	handlers := transport.NewHandlers(logg, authService, jwtManager, refreshTokenRepo)
	router := transport.NewRouter(logg, handlers, jwtManager, userRepo)

	addr := ":8080"
	log.Println("backend listening on", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func mustInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("invalid integer value: %s", s)
	}
	return v
}
