package services

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type ITokenHandler interface {
	GenerateToken(userID int) (string, error)
	ValidateToken(token string) (bool, error)
}

type TokenHandler struct {
	secretKey string
}

func InitTokenHandler(secretKey string) *TokenHandler {
	tokenHandler := TokenHandler{secretKey}
	return &tokenHandler
}

func (tokenHandler *TokenHandler) GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Срок действия — 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tokenHandler.secretKey)
}

func (tokenHandler *TokenHandler) ValidateToken(token string) (bool, error) {
	return false, nil // TODO: finish it
}
