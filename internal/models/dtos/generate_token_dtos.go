package dtos

type GenerateTokenRequest struct {
	UserId string `json:"user_id"`
	Role   string `json:"role"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
