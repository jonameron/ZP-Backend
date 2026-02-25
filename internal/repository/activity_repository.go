package repository

import (
	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(log *models.ActivityLog) error {
	return r.db.Create(log).Error
}

func (r *ActivityRepository) FindByUserID(userID uint, limit int) ([]models.ActivityLog, error) {
	var logs []models.ActivityLog
	err := r.db.
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *ActivityRepository) FindBySessionID(sessionID string) ([]models.ActivityLog, error) {
	var logs []models.ActivityLog
	err := r.db.
		Where("session_id = ?", sessionID).
		Order("timestamp ASC").
		Find(&logs).Error
	return logs, err
}
