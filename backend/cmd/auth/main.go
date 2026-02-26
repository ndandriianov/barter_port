package main

import (
	"barter-port/internal/auth/repository/email_token"
	"barter-port/internal/auth/repository/refresh_token"
	"barter-port/internal/auth/repository/user"
	"barter-port/internal/auth/service"
	"barter-port/internal/auth/service/jwt"
	transport2 "barter-port/internal/auth/transport"
	"barter-port/internal/shared/database"
	"barter-port/internal/shared/logger"
	"barter-port/internal/shared/mailer"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"log"
	"net/http"
	"os"
)

//go:generate sh -c "cd ../.. && swag init -g cmd/api/main.go --parseInternal --parseDependency"
func main() {
	_ = godotenv.Load()

	DbConfigPath := getEnv("DB_CONFIG_PATH", "")
	dbConfig := database.MustLoad(DbConfigPath)
	db, err := database.NewPostgres(dbConfig)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")

	accessSecret := getEnv("ACCESS_SECRET", "")
	refreshSecret := getEnv("REFRESH_SECRET", "")
	accessTTL := getEnv("ACCESS_TTL", "")
	refreshTTL := getEnv("REFRESH_TTL", "")
	accessTTLMinutes := time.Duration(mustInt(accessTTL)) * time.Minute
	refreshTTLMinutes := time.Duration(mustInt(refreshTTL)) * time.Minute

	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	userRepo := user.NewRepository(db)
	emailTokenRepo := email_token.NewRepository(db)
	refreshTokenRepo := refresh_token.NewRepository(db)

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

	authService := service.NewService(
		userRepo,
		emailTokenRepo,
		m,
		infrastructureLogger,

		frontendURL,
		jwtManager,
		re,
	)

	handlers := transport2.NewHandlers(logg, authService, jwtManager, refreshTokenRepo)
	router := transport2.NewRouter(logg, handlers, jwtManager, userRepo)

	addr := ":8081"
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
