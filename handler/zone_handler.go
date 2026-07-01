package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/spotsync/dto"
	"github.com/yourusername/spotsync/service"
)

type ZoneHandler struct {
	svc *service.ZoneService
}

func NewZoneHandler(svc *service.ZoneService) *ZoneHandler {
	return &ZoneHandler{svc: svc}
}

func (h *ZoneHandler) GetAll(c echo.Context) error {
	zones, err := h.svc.GetAll()
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, successResponse("Parking zones retrieved successfully", zones))
}

func (h *ZoneHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid zone ID"))
	}

	zone, err := h.svc.GetByID(uint(id))
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, successResponse("Parking zone retrieved successfully", zone))
}

func (h *ZoneHandler) Create(c echo.Context) error {
	var req dto.CreateZoneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}

	zone, err := h.svc.Create(req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(http.StatusCreated, successResponse("Parking zone created successfully", zone))
}

func (h *ZoneHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid zone ID"))
	}

	var req dto.UpdateZoneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}

	zone, err := h.svc.Update(uint(id), req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, successResponse("Parking zone updated successfully", zone))
}

func (h *ZoneHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid zone ID"))
	}

	if err := h.svc.Delete(uint(id)); err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, successResponse("Parking zone deleted successfully", nil))
}
