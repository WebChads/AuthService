package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/WebChads/AuthService/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
)

type AppConfig struct {
	Port      string `json:"port"`
	SecretKey string `json:"secret_key"`
}

var cfg AppConfig

func generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Срок действия — 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.SecretKey)
}

func validateToken(token string) (bool, error) {
	return false, nil // TODO: finish it
}

func main() {
	err := fillConfig()
	if err != nil {
		fmt.Println(err.Error()) // TODO: Change on logger later
		return
	}

	tokenHandler := services.InitTokenHandler(cfg.SecretKey)
	testToken, _ := tokenHandler.GenerateToken(1)

	e := echo.New()

	isTokenValid, _ := tokenHandler.ValidateToken(testToken)
	var text string
	if isTokenValid {
		text = "Valid"
	} else {
		text = "Not valid"
	}
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Example of token: "+testToken+"\n It is even valid: "+text)
	})

	err = e.Start(":" + cfg.Port)
	if err != nil {
		fmt.Println(err.Error()) // TODO: Change on logger later
		return
	}
}

func fillConfig() error {
	err := cleanenv.ReadConfig("appsettings.json", &cfg)
	if err != nil {
		return err
	}

	if len(cfg.Port) == 0 || len(cfg.SecretKey) == 0 {
		return errors.New("unsuccess try of reading config")
	}

	return nil
}
