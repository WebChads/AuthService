package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ITokenHandler interface {
	GenerateToken(userID int) (string, error)
	ValidateToken(token string) (bool, error)
}

type TokenHandler struct {
	secretKey string
}

func InitTokenHandler(secretKey string) (*TokenHandler, error) {
	tokenHandler := TokenHandler{secretKey}
	return &tokenHandler, nil
}

func (tokenHandler *TokenHandler) GenerateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Срок действия — 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString([]byte(tokenHandler.secretKey))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return signedString, nil
}

func (tokenHandler *TokenHandler) ValidateToken(token string) (bool, error) {
	parseResult, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenHandler.secretKey), nil
	})

	if err != nil {
		return false, err
	}

	if !parseResult.Valid {
		return false, nil
	}

	return true, nil
}
