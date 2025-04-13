package main

import (
	"github.com/WebChads/AuthService/internal/routers"
	"github.com/WebChads/AuthService/internal/services"
	"github.com/labstack/echo/v4"
)

func main() {
	err := services.InitializeServicesScope()
	if err != nil {
		return
	}

	services := services.GetServicesScope()

	e := echo.New()

	tokenRouter := routers.TokenRouter{Services: services}

	e.POST("/api/v1/generate-token", tokenRouter.GenerateToken)

	err = e.Start(":" + services.Configuration.Port)
	if err != nil {
		services.Logger.Error(err.Error())
		return
	}
}
