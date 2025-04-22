package main

import (
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
	err := services.InitializeServicesScope()
	if err != nil {
		return
	}

	servicesScope := services.GetServicesScope()

	_, err = database.InitDatabase(servicesScope)
	if err != nil {
		servicesScope.Logger.Error("Unable to init database: " + err.Error())
		return
	}

	e := echo.New()

	tokenRouter := routers.TokenRouter{Services: servicesScope}

	e.POST("/api/v1/generate-token", tokenRouter.GenerateToken)

	echoSwagger.URL("http://localhost:" + servicesScope.Configuration.Port)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Logger.Fatal(e.Start(":" + servicesScope.Configuration.Port))
}
