package transport

import (
	"barter-port/internal/libs/authkit"
	"barter-port/internal/libs/authkit/validators"
	"barter-port/internal/libs/jwt"
	"log"
	"log/slog"
	"net/http"

	_ "barter-port/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(logger *slog.Logger, h *Handlers, jwtManager *jwt.Manager) http.Handler {
	if logger == nil {
		log.Fatal("logger is required")
	}
	if h == nil {
		log.Fatal("handlers are required")
	}
	if jwtManager == nil {
		log.Fatal("jwt service is required")
	}

	validator := validators.NewLocalJWT(jwtManager)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/verify-email", h.VerifyEmail)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)

		r.Group(func(r chi.Router) {
			r.Use(authkit.Middleware(logger, validator, nil))
			r.Get("/me", h.Me)
		})
	})

	return r
}
