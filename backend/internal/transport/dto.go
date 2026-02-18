package transport

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResp struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	VerifyURL string `json:"verifyUrl"`
}
