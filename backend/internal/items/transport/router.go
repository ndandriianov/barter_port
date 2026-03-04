package transport

import (
	"barter-port/internal/libs/authkit"
	"barter-port/internal/libs/authkit/validators"
	"barter-port/internal/libs/platform/logger"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(logg *slog.Logger, validator *validators.LocalJWT, h *Handlers) http.Handler {
	if logg == nil {
		log.Fatal("logger is required")
	}
	if h == nil {
		log.Fatal("handlers are required")
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)

	r.Group(func(r chi.Router) {
		r.Use(logger.Middleware(logg))
		r.Use(authkit.Middleware(logg, validator, nil))
		r.Route("/items", func(r chi.Router) {
			r.Post("/", h.HandleCreateItem)
			r.Get("/", h.HandleGetItems)
		})
	})

	return r
}
