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
	"regexp"

	"github.com/joho/godotenv"

	"log"
	"net/http"
)

//go:generate sh -c "cd ../.. && swag init -g cmd/auth/main.go --parseInternal --parseDependency"
//go:generate sh -c "cd ../.. && npx -y openapi-to-postmanv2 -s docs/swagger.json -o docs/postman.json -p"

// @title Barter Port API
// @version 1.0.0
// @description API for Barter Port
// @BasePath /
// @schemes http https
func main() {
	_ = godotenv.Load()

	db := bootstrap.InitDatabase()
	defer db.Close()

	frontendURL := bootstrap.GetEnv("FRONTEND_URL", "http://localhost:5173")
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	userRepo := user.NewRepository(db)
	emailTokenRepo := email_token.NewRepository(db)
	refreshTokenRepo := refresh_token.NewRepository(db)

	m := bootstrap.InitMailer()

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
