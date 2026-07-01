package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/spotsync/dto"
	"github.com/yourusername/spotsync/service"
)

type ReservationHandler struct {
	svc *service.ReservationService
}

func NewReservationHandler(svc *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{svc: svc}
}

func (h *ReservationHandler) Create(c echo.Context) error {
	userID := c.Get("userID").(uint)

	var req dto.CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Validation failed"))
	}

	reservation, err := h.svc.Reserve(userID, req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(http.StatusCreated, successResponse("Reservation confirmed successfully", reservation))
}

func (h *ReservationHandler) GetMyReservations(c echo.Context) error {
	userID := c.Get("userID").(uint)

	reservations, err := h.svc.GetMyReservations(userID)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, successResponse("My reservations retrieved successfully", reservations))
}

func (h *ReservationHandler) Cancel(c echo.Context) error {
	userID := c.Get("userID").(uint)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse("Invalid reservation ID"))
	}

	if err := h.svc.CancelReservation(uint(id), userID); err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, successResponse("Reservation cancelled successfully", nil))
}

func (h *ReservationHandler) GetAll(c echo.Context) error {
	reservations, err := h.svc.GetAllReservations()
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, successResponse("All reservations retrieved successfully", reservations))
}
