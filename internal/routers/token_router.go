package routers

import (
	"net/http"

	"github.com/WebChads/AuthService/internal/models/dtos"
	"github.com/WebChads/AuthService/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TokenRouter struct {
	Services *services.ServicesScope
}

// GenerateToken godoc
// @Title GenerateToken
// @Summary Generate a new authentication token
// @Description Generates a new JWT (or other) token for user authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.GenerateTokenRequest true "Token generation parameters"
// @Success 200 {object} dtos.GenerateTokenResponse "Successfully generated token"
// @Failure 400 {object} dtos.ErrorDto "Invalid UserId format (must be UUID)"
// @Failure 500 {object} dtos.ErrorDto "Happened internal error"
// @Router /api/v1/generate-token [post]
func (tokenRouter *TokenRouter) GenerateToken(context echo.Context) error {
	tokenRequest := dtos.GenerateTokenRequest{}
	context.Bind(&tokenRequest)

	parsedUuid, err := uuid.Parse(tokenRequest.UserId)
	if err != nil {
		tokenRouter.Services.Logger.Error(err.Error())
		return context.JSON(http.StatusBadRequest, dtos.ErrorDto{ErrorMessage: "Invalid UserId format (must be UUID)"})
	}

	token, err := tokenRouter.Services.TokenHandler.GenerateToken(parsedUuid)
	if err != nil {
		tokenRouter.Services.Logger.Error(err.Error())
		return context.JSON(http.StatusInternalServerError, dtos.ErrorDto{ErrorMessage: "Happened internal error"})
	}

	return context.JSON(200, dtos.GenerateTokenResponse{Token: token})
}
