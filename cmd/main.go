package main

import (
	"fmt"
	"net/http"

	"github.com/WebChads/AuthService/internal/services"
	"github.com/labstack/echo/v4"
)

func main() {
	err := services.InitializeServicesScope()
	if err != nil {
		return
	}

	services := services.GetServicesScope()

	testToken, _ := services.TokenHandler.GenerateToken(1)

	e := echo.New()

	isTokenValid, _ := services.TokenHandler.ValidateToken(testToken)
	var text string
	if isTokenValid {
		text = "Valid"
	} else {
		text = "Not valid"
	}
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Example of token: "+testToken+"\n It is even valid: "+text)
	})

	err = e.Start(":" + services.Configuration.Port)
	if err != nil {
		fmt.Println(err.Error()) // TODO: Change on logger later
		return
	}
}
