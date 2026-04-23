package http

import "github.com/google/uuid"

type credentialsReq struct {
	Email    string `json:"email" example:"user@email.com"`
	Password string `json:"password" example:"password"`
}

type registerResp struct {
	UserID uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
}

type verifyEmailReq struct {
	Token string `json:"token" example:"iT1VWZWO1apO2GGoXG1ahOKuHlo8WA6ESwA86WMOTiI""`
}

type loginResp struct {
	AccessToken string `json:"accessToken"`
}

type changePasswordReq struct {
	OldEmail    string `json:"oldEmail" example:"user@email.com"`
	OldPassword string `json:"oldPassword" example:"old-password"`
	NewPassword string `json:"newPassword" example:"new-password"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}
