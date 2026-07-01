package repository

import (
	"github.com/yourusername/spotsync/models"
	"gorm.io/gorm"
)

type ZoneRepository struct {
	db *gorm.DB
}

func NewZoneRepository(db *gorm.DB) *ZoneRepository {
	return &ZoneRepository{db: db}
}

func (r *ZoneRepository) FindAll() ([]models.ParkingZone, error) {
	var zones []models.ParkingZone
	if err := r.db.Find(&zones).Error; err != nil {
		return nil, err
	}
	return zones, nil
}

func (r *ZoneRepository) FindByID(id uint) (*models.ParkingZone, error) {
	var zone models.ParkingZone
	if err := r.db.First(&zone, id).Error; err != nil {
		return nil, err
	}
	return &zone, nil
}

func (r *ZoneRepository) Create(zone *models.ParkingZone) error {
	return r.db.Create(zone).Error
}

func (r *ZoneRepository) Update(zone *models.ParkingZone) error {
	return r.db.Save(zone).Error
}

func (r *ZoneRepository) Delete(id uint) error {
	return r.db.Delete(&models.ParkingZone{}, id).Error
}

func (r *ZoneRepository) CountActiveReservations(zoneID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Reservation{}).
		Where("zone_id = ? AND status = ?", zoneID, "active").
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
