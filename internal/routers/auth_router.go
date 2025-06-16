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
	KafkaConsumer  services.KafkaConsumer
}

func NewAuthRouter(logger *zap.Logger,
	tokenHandler services.TokenHandler,
	userRepository repositories.UserRepository,
	kafkaProducer services.KafkaProducer,
	kafkaConsumer services.KafkaConsumer) *AuthRouter {

	authRouter := &AuthRouter{
		Logger:         logger,
		TokenHandler:   tokenHandler,
		UserRepository: userRepository,
		KafkaProducer:  kafkaProducer,
		KafkaConsumer:  kafkaConsumer}

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

	token, err := authRouter.TokenHandler.GenerateToken(parsedUuid, tokenRequest.Role)
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

// ValidateToken godoc
// @Title ValidateToken
// @Summary Checking if authentication token is valid
// @Description It checks if token is valid and not tried to be changed
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.ValidateTokenRequest true "Dto containing token (format of JWT-token)"
// @Success 200 {object} dtos.ValidateTokenResponse "Dto with field 'is_valid' that shows if token is valid"
// @Router /api/v1/auth/validate-token [post]
func (authRouter *AuthRouter) ValidateToken(context echo.Context) error {
	tokenRequest := dtos.ValidateTokenRequest{}
	context.Bind(&tokenRequest)

	authRouter.Logger.Info(fmt.Sprintf("token sent in validate token: %s", tokenRequest.Token))

	isValid, err := authRouter.TokenHandler.ValidateToken(tokenRequest.Token)

	if err != nil {
		isValid = false
	}

	return context.JSON(200, dtos.ValidateTokenResponse{IsValid: isValid})
}

// VerifySmsCode godoc
// @Title VerifySmsCode
// @Summary Verifying SMS code if it is what was sent to user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.VerifySmsCodeRequest true "Dto with phone number and SMS code"
// @Success 200 "Valid SMS code, giving token"
// @Failure 400 {object} dtos.ErrorDto "Invalid phone number"
// @Failure 400 {object} dtos.ErrorDto "Invalid SMS code format"
// @Failure 400 {object} dtos.ErrorDto "Invalid SMS code"
// @Failure 500 {object} dtos.ErrorDto "Happened internal error"
// @Router /api/v1/auth/verify-sms-code [post]
func (authRouter *AuthRouter) VerifySmsCode(context echo.Context) error {
	request := dtos.VerifySmsCodeRequest{}
	context.Bind(&request)

	phoneNumberRegex := `^(8|\+7)(\s|\(|-)?(\d{3})(\s|\)|-)?(\d{3})(\s|-)?(\d{2})(\s|-)?(\d{2})$`
	isPhoneNumberCorrect, err := regexp.MatchString(phoneNumberRegex, request.PhoneNumber)
	if err != nil || !isPhoneNumberCorrect {
		authRouter.Logger.Error(fmt.Errorf("user sent invalid phone number: %s", request.PhoneNumber).Error())
		return context.JSON(http.StatusBadRequest, dtos.ErrorDto{ErrorMessage: "Invalid phone number"})
	}

	smsCodeRegex := `^\d{4}$`
	isSmsCodeFormatCorrect, err := regexp.MatchString(smsCodeRegex, request.SmsCode)
	if err != nil || !isSmsCodeFormatCorrect {
		authRouter.Logger.Error(fmt.Errorf("user sent invalid sms code format: %s", request.SmsCode).Error())
		return context.JSON(http.StatusBadRequest, dtos.ErrorDto{ErrorMessage: "Invalid SMS code format"})
	}

	smsCodeFromKafka, exists := authRouter.KafkaConsumer.GetSmsCode(request.PhoneNumber)
	if !exists {
		authRouter.Logger.Error(fmt.Errorf("for user with phone number %s wasn't produced any sms code", request.PhoneNumber).Error())
		return context.JSON(http.StatusBadRequest, dtos.ErrorDto{ErrorMessage: "Sms code wasn't requested"})
	}

	if smsCodeFromKafka != request.SmsCode {
		authRouter.Logger.Error(fmt.Errorf("invalid sms code for user with phone number %s. Sent: %s. Actual: %s", request.PhoneNumber, request.SmsCode, smsCodeFromKafka).Error())
		return context.JSON(http.StatusBadRequest, dtos.ErrorDto{ErrorMessage: "Invalid SMS code"})
	}

	userModel, err := authRouter.UserRepository.Get(request.PhoneNumber)
	if err != nil {
		authRouter.Logger.Error(fmt.Errorf("while retrieving user from database happened error: %w", err).Error())
		return context.JSON(http.StatusInternalServerError, dtos.ErrorDto{ErrorMessage: "Happened error while retrieving user from database"})
	}

	if userModel == nil {
		authRouter.Logger.Error(fmt.Errorf("somehow user model is nil, was requested for phone number: %s, good luck in debugging", request.PhoneNumber).Error())
		return context.JSON(http.StatusInternalServerError, dtos.ErrorDto{ErrorMessage: "Happened error while retrieving user from database"})
	}

	token, err := authRouter.TokenHandler.GenerateToken(userModel.Id, userModel.UserRole)
	if err != nil {
		authRouter.Logger.Error(fmt.Errorf("error happened while generating token for user with uuid %s: %w", userModel.Id, err).Error())
		return context.JSON(http.StatusInternalServerError, dtos.ErrorDto{ErrorMessage: "Happened error while generating token for user"})
	}

	return context.JSON(200, dtos.TokenResponse{Token: token})
}
