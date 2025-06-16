package main

import (
	"fmt"
	"net/http"

	_ "github.com/WebChads/AuthService/docs"
	"github.com/WebChads/AuthService/internal/database"
	"github.com/WebChads/AuthService/internal/database/repositories"
	"github.com/WebChads/AuthService/internal/routers"
	"github.com/WebChads/AuthService/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
		logger.Error(fmt.Sprintf("%v", config))
		logger.Error("Unable to init database: " + err.Error())
		return
	}

	userRepository := repositories.NewUserRepository(dbContext.Connection)

	kafkaProducer, err := services.NewKafkaProducer(config.KafkaConfig)
	if err != nil {
		logger.Error("Unable to init kafka: " + err.Error())
		return
	}

	kafkaConsumer, err := services.InitKafkaConsumer(config.KafkaConfig, logger)
	if err != nil {
		panic("while initing kafka consumer happened error: " + err.Error())
	}
	go kafkaConsumer.Start()

	e := echo.New()

	// Auth router
	authRouter := routers.NewAuthRouter(logger, tokenHandler, userRepository, kafkaProducer, kafkaConsumer)
	e.POST("/api/v1/auth/generate-token", authRouter.GenerateToken)
	e.POST("/api/v1/auth/validate-token", authRouter.ValidateToken)

	e.POST("/api/v1/auth/register", authRouter.Register)
	e.POST("/api/v1/auth/send-sms-code", authRouter.SendSmsCode)
	e.POST("/api/v1/auth/verify-sms-code", authRouter.VerifySmsCode)

	// Health router
	healthRouter := routers.NewHealthRouter(logger)
	e.GET("/healthz", healthRouter.HealthCheck)

	// Swagger
	echoSwagger.URL("http://localhost:" + config.Port)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// CORS handling
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch},
	}))

	e.Logger.Fatal(e.Start(":" + config.Port))
}
