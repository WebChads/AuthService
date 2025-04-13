package services

import (
	"fmt"

	"go.uber.org/zap"
)

type ServicesScope struct {
	Logger        *zap.Logger
	TokenHandler  *TokenHandler
	Configuration *AppConfig
}

var services *ServicesScope

func InitializeServicesScope() error {
	services = &ServicesScope{}

	config, err := InitializeConfig()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	services.Configuration = config

	logger, err := InitLogger(services.Configuration.IsDevelopment)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	services.Logger = logger

	tokenHandler, err := InitTokenHandler(services.Configuration.SecretKey)
	if err != nil {
		services.Logger.Error(err.Error())
		return err
	}

	services.TokenHandler = tokenHandler

	return nil
}

func GetServicesScope() *ServicesScope {
	return services
}
