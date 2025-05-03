package routers

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"

	"github.com/WebChads/AuthService/internal/database/repositories"
	"github.com/WebChads/AuthService/internal/models/dtos"
	"github.com/WebChads/AuthService/internal/models/entities"
	"github.com/WebChads/AuthService/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AuthRouter struct {
	Logger         *zap.Logger
	TokenHandler   services.TokenHandler
	UserRepository repositories.UserRepository
	KafkaProducer  services.KafkaProducer
}

func NewAuthRouter(logger *zap.Logger, tokenHandler services.TokenHandler, userRepository repositories.UserRepository, kafkaProducer services.KafkaProducer) *AuthRouter {
	authRouter := &AuthRouter{Logger: logger, TokenHandler: tokenHandler, UserRepository: userRepository, KafkaProducer: kafkaProducer}
	return authRouter
}

var possibleRoles = []string{"Player", "Trainer"}

// GenerateToken godoc
// @Title GenerateToken
// @Summary Generate a new authentication token
// @Description Generates a new JWT (or other) token for user authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.GenerateTokenRequest true "Token generation parameters"
// @Success 200 {object} dtos.TokenResponse "Successfully generated token"
// @Failure 400 {object} dtos.ErrorDto "Invalid UserId format (must be UUID)"
// @Failure 500 {object} dtos.ErrorDto "Happened internal error"
// @Router /api/v1/auth/generate-token [post]
func (authRouter *AuthRouter) GenerateToken(context echo.Context) error {
	tokenRequest := dtos.GenerateTokenRequest{}
	context.Bind(&tokenRequest)

	parsedUuid, err := uuid.Parse(tokenRequest.UserId)
	if err != nil {
		authRouter.Logger.Error(err.Error())
		return context.JSON(http.StatusBadRequest, dtos.ErrorDto{ErrorMessage: "Invalid UserId format (must be UUID)"})
	}

	token, err := authRouter.TokenHandler.GenerateToken(parsedUuid)
	if err != nil {
		authRouter.Logger.Error(err.Error())
		return context.JSON(http.StatusInternalServerError, dtos.ErrorDto{ErrorMessage: "Happened internal error"})
	}

	return context.JSON(200, dtos.TokenResponse{Token: token})
}

// Register godoc
// @Title Register
// @Summary Create user entity in database, making him ready to log in
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.RegisterRequest true "Register parameters"
// @Success 200 "Successfully created user in db"
// @Failure 400 {object} dtos.ErrorDto "Invalid phone number"
// @Failure 400 {object} dtos.ErrorDto "Invalid role"
// @Failure 500 {object} dtos.ErrorDto "Happened internal error"
// @Router /api/v1/auth/register [post]
func (authRouter *AuthRouter) Register(context echo.Context) error {
	request := dtos.RegisterRequest{}
	context.Bind(&request)

	phoneNumberRegex := `^(8|\+7)(\s|\(|-)?(\d{3})(\s|\)|-)?(\d{3})(\s|-)?(\d{2})(\s|-)?(\d{2})$`
	isPhoneNumberCorrect, err := regexp.MatchString(phoneNumberRegex, request.PhoneNumber)
	if err != nil || !isPhoneNumberCorrect {
		authRouter.Logger.Error(fmt.Errorf("user sent invalid phone number: %s", request.PhoneNumber).Error())
		return context.JSON(http.StatusBadRequest, dtos.ErrorDto{ErrorMessage: "Invalid phone number"})
	}

	if !slices.Contains(possibleRoles, request.Role) {
		authRouter.Logger.Error(fmt.Errorf("user sent invalid role: %s", request.Role).Error())
		return context.JSON(http.StatusBadRequest, dtos.ErrorDto{ErrorMessage: "Invalid role"})
	}

	err = authRouter.UserRepository.Add(&entities.User{Id: uuid.New(), PhoneNumber: request.PhoneNumber, UserRole: request.Role})
	if err != nil {
		return context.JSON(400, dtos.ErrorDto{ErrorMessage: fmt.Errorf("while adding user in db happened error: %w", err).Error()})
	}

	return context.NoContent(200)
}

// SendSmsCode godoc
// @Title SendSmsCode
// @Summary Sending sms-code to user to phone number he entered
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.SendSmsCodeRequest true "Dto with phone number"
// @Success 200 "Successfully sent code"
// @Failure 400 {object} dtos.ErrorDto "Invalid phone number"
// @Failure 500 {object} dtos.ErrorDto "Happened internal error"
// @Router /api/v1/auth/send-sms-code [post]
func (authRouter *AuthRouter) SendSmsCode(context echo.Context) error {
	request := dtos.SendSmsCodeRequest{}
	context.Bind(&request)

	phoneNumberRegex := `^(8|\+7)(\s|\(|-)?(\d{3})(\s|\)|-)?(\d{3})(\s|-)?(\d{2})(\s|-)?(\d{2})$`
	isPhoneNumberCorrect, err := regexp.MatchString(phoneNumberRegex, request.PhoneNumber)
	if err != nil || !isPhoneNumberCorrect {
		authRouter.Logger.Error(fmt.Errorf("user sent invalid phone number: %s", request.PhoneNumber).Error())
		return context.JSON(http.StatusBadRequest, dtos.ErrorDto{ErrorMessage: "Invalid phone number"})
	}

	err = authRouter.KafkaProducer.SendPhoneNumber(request.PhoneNumber)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, dtos.ErrorDto{ErrorMessage: "Happened error while sending sms to this phone number"})
	}

	return context.NoContent(200)
}
