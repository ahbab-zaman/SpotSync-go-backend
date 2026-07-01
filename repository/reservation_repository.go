package repository

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yourusername/spotsync/models"
)

var ErrZoneFull = errors.New("zone is at full capacity")

type ReservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) CreateWithLock(reservation *models.Reservation, zoneID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var zone models.ParkingZone
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&zone, zoneID).Error; err != nil {
			return err
		}

		var activeCount int64
		if err := tx.Model(&models.Reservation{}).
			Where("zone_id = ? AND status = ?", zoneID, "active").
			Count(&activeCount).Error; err != nil {
			return err
		}

		if activeCount >= int64(zone.TotalCapacity) {
			return ErrZoneFull
		}

		return tx.Create(reservation).Error
	})
}

func (r *ReservationRepository) FindByUserID(userID uint) ([]models.Reservation, error) {
	var reservations []models.Reservation
	if err := r.db.Where("user_id = ?", userID).
		Preload("Zone").
		Find(&reservations).Error; err != nil {
		return nil, err
	}
	return reservations, nil
}

func (r *ReservationRepository) FindByID(id uint) (*models.Reservation, error) {
	var reservation models.Reservation
	if err := r.db.First(&reservation, id).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Reservation{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *ReservationRepository) FindAll() ([]models.Reservation, error) {
	var reservations []models.Reservation
	if err := r.db.Preload("User").Preload("Zone").
		Find(&reservations).Error; err != nil {
		return nil, err
	}
	return reservations, nil
}
