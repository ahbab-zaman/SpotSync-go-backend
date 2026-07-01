package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/yourusername/spotsync/dto"
	"github.com/yourusername/spotsync/models"
	"github.com/yourusername/spotsync/repository"
)

var (
	ErrReservationNotFound = errors.New("reservation not found")
	ErrForbidden           = errors.New("forbidden")
	ErrZoneFull            = errors.New("zone is at full capacity")
)

type ReservationService struct {
	reservationRepo *repository.ReservationRepository
	zoneRepo        *repository.ZoneRepository
}

func NewReservationService(reservationRepo *repository.ReservationRepository, zoneRepo *repository.ZoneRepository) *ReservationService {
	return &ReservationService{
		reservationRepo: reservationRepo,
		zoneRepo:        zoneRepo,
	}
}

func (s *ReservationService) Reserve(userID uint, req dto.CreateReservationRequest) (*dto.ReservationResponse, error) {
	reservation := &models.Reservation{
		UserID:       userID,
		ZoneID:       req.ZoneID,
		LicensePlate: req.LicensePlate,
		Status:       "active",
	}

	if err := s.reservationRepo.CreateWithLock(reservation, req.ZoneID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		if errors.Is(err, repository.ErrZoneFull) {
			return nil, ErrZoneFull
		}
		return nil, err
	}

	return &dto.ReservationResponse{
		ID:           reservation.ID,
		UserID:       reservation.UserID,
		ZoneID:       reservation.ZoneID,
		LicensePlate: reservation.LicensePlate,
		Status:       reservation.Status,
		CreatedAt:    reservation.CreatedAt,
		UpdatedAt:    reservation.UpdatedAt,
	}, nil
}

func (s *ReservationService) GetMyReservations(userID uint) ([]dto.MyReservationResponse, error) {
	reservations, err := s.reservationRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.MyReservationResponse, len(reservations))
	for i, r := range reservations {
		responses[i] = dto.MyReservationResponse{
			ID:           r.ID,
			LicensePlate: r.LicensePlate,
			Status:       r.Status,
			Zone: dto.ZoneInfo{
				ID:   r.Zone.ID,
				Name: r.Zone.Name,
				Type: r.Zone.Type,
			},
			CreatedAt: r.CreatedAt,
		}
	}
	return responses, nil
}

func (s *ReservationService) CancelReservation(reservationID uint, userID uint) error {
	reservation, err := s.reservationRepo.FindByID(reservationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReservationNotFound
		}
		return err
	}

	if reservation.UserID != userID {
		return ErrForbidden
	}

	return s.reservationRepo.UpdateStatus(reservationID, "cancelled")
}

func (s *ReservationService) GetAllReservations() ([]dto.AdminReservationResponse, error) {
	reservations, err := s.reservationRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.AdminReservationResponse, len(reservations))
	for i, r := range reservations {
		responses[i] = dto.AdminReservationResponse{
			ID:           r.ID,
			LicensePlate: r.LicensePlate,
			Status:       r.Status,
			User: dto.UserInfo{
				ID:    r.User.ID,
				Name:  r.User.Name,
				Email: r.User.Email,
			},
			Zone: dto.ZoneInfo{
				ID:   r.Zone.ID,
				Name: r.Zone.Name,
				Type: r.Zone.Type,
			},
			CreatedAt: r.CreatedAt,
		}
	}
	return responses, nil
}
