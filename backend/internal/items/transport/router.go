package transport

import (
	"barter-port/internal/libs/authkit"
	"barter-port/internal/libs/authkit/validators"
	"barter-port/internal/libs/platform/logger"
	"log"
	"log/slog"
	"net/http"

	_ "barter-port/docs/items"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
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

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			logg.Error("failed to write health response", slog.String("error", err.Error()))
		}
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Group(func(r chi.Router) {
		r.Use(authkit.Middleware(logg, validator, nil))
		r.Use(logger.Middleware(logg))
		r.Route("/items", func(r chi.Router) {
			r.Post("/", h.HandleCreateItem)
			r.Get("/", h.HandleGetItems)
		})
	})

	return r
}
