package transport

import (
	"barter-port/internal/libs/authkit"
	"barter-port/internal/libs/authkit/validators"
	"log"
	"log/slog"
	"net/http"

	_ "barter-port/docs/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(logger *slog.Logger, validator *validators.LocalJWT, h *Handlers) http.Handler {
	if logger == nil {
		log.Fatal("logger is required")
	}
	if h == nil {
		log.Fatal("handlers are required")
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			logger.Error("failed to write health response", slog.String("error", err.Error()))
		}
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
