package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Service is healthy",
		"data": map[string]interface{}{
			"status": "ok",
		},
	})
}
