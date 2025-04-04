package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
)

type AppConfig struct {
	Port       string `json:"port"`
	Secret_key string `json:"secret_key"`
}

var cfg AppConfig

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
