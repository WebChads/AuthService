package dtos

type GenerateTokenRequest struct {
	UserId string `json:"user_id"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
