package main

import (
	"fmt"

	_ "github.com/WebChads/AuthService/docs"
	"github.com/WebChads/AuthService/internal/database"
	"github.com/WebChads/AuthService/internal/database/repositories"
	"github.com/WebChads/AuthService/internal/routers"
	"github.com/WebChads/AuthService/internal/services"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title           AuthService API
// @version         1.0
// @description     Service for handling auth, tokens and that stuff

// @securityDefinitions.apikey  JwtBearer
// @in header
// @name Authorization
func main() {
	config, err := services.InitializeConfig()
	if err != nil {
		fmt.Println(err.Error())
	}

	logger, err := services.InitLogger(config.IsDevelopment)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	tokenHandler, err := services.InitTokenHandler(config.SecretKey)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	dbContext, err := database.InitDatabase(&config.DbSettings)
	if err != nil {
		logger.Error("Unable to init database: " + err.Error())
		return
	}

	userRepository := repositories.NewUserRepository(dbContext.Connection)

	e := echo.New()

	authRouter := routers.NewAuthRouter(logger, tokenHandler, userRepository)
	e.POST("/api/v1/auth/generate-token", authRouter.GenerateToken)
	e.POST("/api/v1/auth/register", authRouter.Register)
	e.POST("/api/v1/auth/send-sms-code", authRouter.SendSmsCode)

	echoSwagger.URL("http://localhost:" + config.Port)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Logger.Fatal(e.Start(":" + config.Port))
}
