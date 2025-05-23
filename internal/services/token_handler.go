package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenHandler interface {
	GenerateToken(userID uuid.UUID, userRole string) (string, error)
	ValidateToken(token string) (bool, error)
}

type JwtTokenHandler struct {
	secretKey string
}

func InitTokenHandler(secretKey string) (*JwtTokenHandler, error) {
	tokenHandler := JwtTokenHandler{secretKey}
	return &tokenHandler, nil
}

func (tokenHandler *JwtTokenHandler) GenerateToken(userID uuid.UUID, userRole string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"user_role": userRole,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // Срок действия — 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString([]byte(tokenHandler.secretKey))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return signedString, nil
}

func (tokenHandler *JwtTokenHandler) ValidateToken(token string) (bool, error) {
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
