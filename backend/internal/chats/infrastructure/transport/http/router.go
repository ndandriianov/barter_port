package http

import (
	chatsdocfirst "barter-port/docs/doc-first/chats"
	"barter-port/pkg/authkit"
	"barter-port/pkg/authkit/validators"
	"barter-port/pkg/logger"
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
	openAPISpecHandler := http.StripPrefix("/swagger/", http.FileServer(http.FS(chatsdocfirst.SpecFS)))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/swagger.yaml", http.StatusPermanentRedirect)
	})
	r.Get("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/swagger.yaml", http.StatusPermanentRedirect)
	})
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		openAPISpecHandler.ServeHTTP(w, r)
	})

	r.Group(func(r chi.Router) {
		r.Use(authkit.Middleware(logg, validator, nil))
		r.Use(logger.Middleware(logg))

		r.Route("/chats", func(r chi.Router) {
			r.Get("/users", h.ListUsers)
			r.Post("/", h.CreateChat)
			r.Get("/", h.ListChats)
			r.Get("/deals/{dealId}", h.GetDealChat)
			r.Get("/{chatId}/messages", h.GetMessages)
			r.Post("/{chatId}/messages", h.SendMessage)
		})
	})

	return r
}
