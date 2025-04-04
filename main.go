package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
)

type AppConfig struct {
	Port       string `json:"port"`
	Secret_key string `json:"secret_key"`
}

var cfg AppConfig

func generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Срок действия — 24 часа
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.Secret_key)
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

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Привет, Echo!")
	})

	e.Start(":" + cfg.Port)
}

func fillConfig() error {
	err := cleanenv.ReadConfig("appsettings.json", &cfg)
	if err != nil {
		return err
	}

	if len(cfg.Port) == 0 || len(cfg.Secret_key) == 0 {
		return errors.New("unsuccess try of reading config")
	}

	return nil
}
