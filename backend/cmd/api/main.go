package main

import (
	"github.com/ndandriianov/barter_port/backend/internal/repository/token"
	"github.com/ndandriianov/barter_port/backend/internal/repository/user"
	"github.com/ndandriianov/barter_port/backend/internal/service/auth"
	"github.com/ndandriianov/barter_port/backend/internal/transport"

	"log"
	"net/http"
	"os"
)

func main() {
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")

	userRepo := user.NewInMemoryUserRepo()
	tokenRepo := token.NewInMemoryTokenRepo()

	authService := auth.NewService(userRepo, tokenRepo, frontendURL)

	handlers := transport.NewHandlers(authService)
	router := transport.NewRouter(handlers)

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
