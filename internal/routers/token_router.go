package routers

import (
	"github.com/WebChads/AuthService/internal/services"
	"github.com/labstack/echo/v4"
)

type TokenRouter struct {
	services services.ServicesScope
}

func (tokenRouter *TokenRouter) GenerateToken(context echo.Context) error {
	return nil
}
