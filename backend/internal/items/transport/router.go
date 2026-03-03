package transport

import (
	"barter-port/internal/libs/authkit"
	"barter-port/internal/libs/authkit/validators"
	"barter-port/internal/libs/platform/http_api"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	r.Group(func(r chi.Router) {
		r.Use(http_api.LoggerMiddleware(logger))
		r.Use(authkit.Middleware(logger, validator, nil))
		r.Route("/items", func(r chi.Router) {
			r.Post("/", h.HandleCreateItem)
			r.Get("/", h.HandleGetItems)
		})
	})

	return r
}
