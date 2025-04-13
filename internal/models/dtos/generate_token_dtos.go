package dtos

type GenerateTokenRequest struct {
	UserId string `json:"user_id"`
}

type GenerateTokenResponse struct {
	Token string `json:"token"`
}
