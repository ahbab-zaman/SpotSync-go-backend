package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/spotsync/dto"
	"github.com/yourusername/spotsync/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}

	user, err := h.svc.Register(req)
	if err != nil {
		return handleServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, successResponse("User registered successfully", user))
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		return handleServiceError(c, err)
	}

	return c.JSON(http.StatusOK, successResponse("Login successful", resp))
}

func handleServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrDuplicateEmail):
		return c.JSON(http.StatusBadRequest, errorResponse("Email already registered"))
	case errors.Is(err, service.ErrInvalidCredentials):
		return c.JSON(http.StatusUnauthorized, errorResponse("Invalid credentials"))
	case errors.Is(err, service.ErrZoneNotFound):
		return c.JSON(http.StatusNotFound, errorResponse("Resource not found"))
	case errors.Is(err, service.ErrReservationNotFound):
		return c.JSON(http.StatusNotFound, errorResponse("Reservation not found"))
	case errors.Is(err, service.ErrForbidden):
		return c.JSON(http.StatusForbidden, errorResponse("Forbidden"))
	case errors.Is(err, service.ErrZoneFull):
		return c.JSON(http.StatusConflict, errorResponse("Zone is at full capacity"))
	default:
		return c.JSON(http.StatusInternalServerError, errorResponse("Internal server error"))
	}
}

func successResponse(message string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	}
}

func errorResponse(message string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"message": message,
		"errors":  nil,
	}
}
