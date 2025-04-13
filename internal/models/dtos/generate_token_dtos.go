package dtos

import "github.com/google/uuid"

type GenerateTokenRequest struct {
	UserId uuid.UUID `json:"user_id"`
}

type GenerateTokenResponse struct {
	Token string `json:"token"`
}
