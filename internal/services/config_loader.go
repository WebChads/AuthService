package services

import (
	"errors"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Port          string `json:"port"`
	SecretKey     string `json:"secret_key"`
	IsDevelopment bool   `json:"is_development"`
}

var cfg AppConfig

func InitializeConfig() (*AppConfig, error) {
	workingDirectory, _ := os.Getwd()
	path := workingDirectory + "\\configs\\appsettings.json"

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.Port) == 0 || len(cfg.SecretKey) == 0 {
		return nil, errors.New("unsuccess try of reading config")
	}

	return &cfg, nil
}
