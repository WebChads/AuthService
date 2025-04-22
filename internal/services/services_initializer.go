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

var servicesScope *ServicesScope

func InitializeServicesScope() error {
	servicesScope = &ServicesScope{}

	config, err := InitializeConfig()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	servicesScope.Configuration = config

	logger, err := InitLogger(servicesScope.Configuration.IsDevelopment)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	servicesScope.Logger = logger

	tokenHandler, err := InitTokenHandler(servicesScope.Configuration.SecretKey)
	if err != nil {
		servicesScope.Logger.Error(err.Error())
		return err
	}

	servicesScope.TokenHandler = tokenHandler

	return nil
}

func GetServicesScope() *ServicesScope {
	return servicesScope
}
