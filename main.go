package main

import (
	"fmt"

	_ "github.com/WebChads/AuthService/docs"
	"github.com/WebChads/AuthService/internal/database"
	"github.com/WebChads/AuthService/internal/routers"
	"github.com/WebChads/AuthService/internal/services"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
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

	_, err = database.InitDatabase(&config.DbSettings)
	if err != nil {
		logger.Error("Unable to init database: " + err.Error())
		return
	}

	e := echo.New()

	authRouter := routers.NewAuthRouter(logger, tokenHandler)
	e.POST("/api/v1/auth/generate-token", authRouter.GenerateToken)
	e.POST("/api/v1/auth/register", authRouter.Register)

	echoSwagger.URL("http://localhost:" + config.Port)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Logger.Fatal(e.Start(":" + config.Port))
}
