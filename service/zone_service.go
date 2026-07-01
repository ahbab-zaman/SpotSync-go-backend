package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/yourusername/spotsync/dto"
	"github.com/yourusername/spotsync/models"
	"github.com/yourusername/spotsync/repository"
)

var (
	ErrZoneNotFound = errors.New("zone not found")
)

type ZoneService struct {
	repo *repository.ZoneRepository
}

func NewZoneService(repo *repository.ZoneRepository) *ZoneService {
	return &ZoneService{repo: repo}
}

func (s *ZoneService) GetAll() ([]dto.ZoneResponse, error) {
	zones, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ZoneResponse, len(zones))
	for i, zone := range zones {
		responses[i] = s.toZoneResponse(&zone)
	}
	return responses, nil
}

func (s *ZoneService) GetByID(id uint) (*dto.ZoneResponse, error) {
	zone, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}
	resp := s.toZoneResponse(zone)
	return &resp, nil
}

func (s *ZoneService) Create(req dto.CreateZoneRequest) (*dto.ZoneResponse, error) {
	zone := &models.ParkingZone{
		Name:          req.Name,
		Type:          req.Type,
		TotalCapacity: req.TotalCapacity,
		PricePerHour:  req.PricePerHour,
	}
	if err := s.repo.Create(zone); err != nil {
		return nil, err
	}
	return &dto.ZoneResponse{
		ID:             zone.ID,
		Name:           zone.Name,
		Type:           zone.Type,
		TotalCapacity:  zone.TotalCapacity,
		AvailableSpots: zone.TotalCapacity,
		PricePerHour:   zone.PricePerHour,
		CreatedAt:      zone.CreatedAt,
		UpdatedAt:      zone.UpdatedAt,
	}, nil
}

func (s *ZoneService) Update(id uint, req dto.UpdateZoneRequest) (*dto.ZoneResponse, error) {
	zone, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		zone.Name = *req.Name
	}
	if req.Type != nil {
		zone.Type = *req.Type
	}
	if req.TotalCapacity != nil {
		zone.TotalCapacity = *req.TotalCapacity
	}
	if req.PricePerHour != nil {
		zone.PricePerHour = *req.PricePerHour
	}

	if err := s.repo.Update(zone); err != nil {
		return nil, err
	}
	resp := s.toZoneResponse(zone)
	return &resp, nil
}

func (s *ZoneService) Delete(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrZoneNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *ZoneService) toZoneResponse(zone *models.ParkingZone) dto.ZoneResponse {
	activeCount, err := s.repo.CountActiveReservations(zone.ID)
	if err != nil {
		activeCount = 0
	}
	return dto.ZoneResponse{
		ID:             zone.ID,
		Name:           zone.Name,
		Type:           zone.Type,
		TotalCapacity:  zone.TotalCapacity,
		AvailableSpots: zone.TotalCapacity - int(activeCount),
		PricePerHour:   zone.PricePerHour,
		CreatedAt:      zone.CreatedAt,
		UpdatedAt:      zone.UpdatedAt,
	}
}
