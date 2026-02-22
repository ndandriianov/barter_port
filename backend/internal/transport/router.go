package transport

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ndandriianov/barter_port/backend/internal/service/auth/jwt"
	"github.com/ndandriianov/barter_port/backend/internal/transport/middleware/auth_jwt"
)

func NewRouter(logger *slog.Logger, h *Handlers, jwtManager *jwt.Manager, userGetter auth_jwt.UserGetter) http.Handler {
	if logger == nil {
		log.Fatal("logger is required")
	}
	if h == nil {
		log.Fatal("handlers are required")
	}
	if jwtManager == nil {
		log.Fatal("jwt service is required")
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/verify-email", h.VerifyEmail)
		r.Post("/login", h.Login)

		r.Group(func(r chi.Router) {
			r.Use(auth_jwt.Middleware(logger, jwtManager, userGetter))
			r.Get("/me", h.Me)
		})
	})

	return r
}
