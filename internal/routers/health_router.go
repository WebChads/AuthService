package routers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type HealthRouter struct {
	Logger *zap.Logger
}

func NewHealthRouter(logger *zap.Logger) *HealthRouter {
	return &HealthRouter{
		Logger: logger,
	}
}

// HealthCheck godoc
// @Title HealthCheck
// @Summary Health check endpoint for Kubernetes
// @Description Returns 200 if the service is healthy and ready to accept traffic
// @Tags Infrastructure
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Service is healthy"
// @Router /healthz [get]
func (h *HealthRouter) HealthCheck(context echo.Context) error {
	return context.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
