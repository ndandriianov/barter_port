package transport

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/ndandriianov/barter_port/backend/internal/transport/middleware/auth_jwt"
)

func NewRouter(logger *slog.Logger, h *Handlers, jwtSecret string, userGetter auth_jwt.UserGetter) http.Handler {
	r := chi.NewRouter()

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
			r.Use(auth_jwt.Middleware(logger, []byte(jwtSecret), userGetter))
			r.Get("/me", h.Me)
		})
	})

	return r
}
