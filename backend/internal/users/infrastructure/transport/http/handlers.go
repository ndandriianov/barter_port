package http

import (
	"barter-port/internal/users/application"
	"net/http"
)

type Handlers struct {
	application.UserService
}

func (h *Handlers) HandleGetUser(w http.ResponseWriter, r *http.Request) {

}

func (h *Handlers) HandleGetMe(w http.ResponseWriter, r *http.Request) {

}

func (h *Handlers) HandleUpdateMe(w http.ResponseWriter, r *http.Request) {

}
