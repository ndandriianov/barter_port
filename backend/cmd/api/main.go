package main

import (
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/logger"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/mailer"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/repository/token"
	"github.com/ndandriianov/barter_port/backend/internal/infrastructure/repository/user"
	"github.com/ndandriianov/barter_port/backend/internal/service/auth"
	"github.com/ndandriianov/barter_port/backend/internal/transport"

	"log"
	"net/http"
	"os"
)

func main() {
	_ = godotenv.Load()

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")

	jwtSecret := getEnv("JWT_SECRET", "")
	jwtTTL := getEnv("JWT_TTL", "")
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	userRepo := user.NewInMemoryUserRepo()
	tokenRepo := token.NewInMemoryTokenRepo()

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

	jwtService := auth.NewJWTService([]byte(jwtSecret), time.Duration(mustInt(jwtTTL))*time.Minute)

	authService := auth.NewService(
		userRepo,
		tokenRepo,
		m,
		infrastructureLogger,

		frontendURL,
		jwtService,
		re,
	)

	handlers := transport.NewHandlers(logg, authService)
	router := transport.NewRouter(logg, handlers, jwtService, userRepo)

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
