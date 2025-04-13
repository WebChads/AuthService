package routers

import (
	"net/http"

	"github.com/WebChads/AuthService/internal/models/dtos"
	"github.com/WebChads/AuthService/internal/services"
	"github.com/labstack/echo/v4"
)

type TokenRouter struct {
	Services *services.ServicesScope
}

// GenerateToken godoc
// @Summary Generate a new authentication token
// @Description Generates a new JWT (or other) token for user authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.GenerateTokenRequest true "Token generation parameters"
// @Success 200 {object} dtos.GenerateTokenResponse "Successfully generated token"
// @Failure 400 {string} string "Invalid request data"
// @Failure 500 {string} string "Internal server error"
// @Router /api/tokens/generate [post]
func (tokenRouter *TokenRouter) GenerateToken(context echo.Context) error {
	tokenRequest := dtos.GenerateTokenRequest{}
	context.Bind(&tokenRequest)

	// TODO: validate request

	token, err := tokenRouter.Services.TokenHandler.GenerateToken(tokenRequest.UserId)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(200, dtos.GenerateTokenResponse{Token: token})
}
