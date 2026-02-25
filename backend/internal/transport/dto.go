package transport

import "github.com/google/uuid"

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResp struct {
	UserID uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
}

type verifyEmailReq struct {
	Token string `json:"token"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResp struct {
	AccessToken string `json:"accessToken"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}
