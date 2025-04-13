package services

import (
	"errors"

	"go.uber.org/zap"
)

var globalLogger *zap.Logger

func InitLogger(isDevelopment bool) (*zap.Logger, error) {
	var err error

	if globalLogger != nil {
		return nil, errors.New("Logger already initialized")
	}

	if isDevelopment {
		globalLogger, err = zap.NewDevelopment()
	} else {
		globalLogger, err = zap.NewProduction()
	}

	return globalLogger, err
}
