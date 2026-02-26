package transport

import (
	"barter-port/internal/auth/service/jwt"
	"barter-port/internal/auth/transport/middleware/auth_jwt"
	"log"
	"log/slog"
	"net/http"

	_ "barter-port/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
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
			r.Use(auth_jwt.Middleware(logger, jwtManager, userGetter))
			r.Get("/me", h.Me)
		})
	})

	return r
}
