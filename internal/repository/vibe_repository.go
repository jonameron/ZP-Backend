package repository

import (
	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type VibeRepository struct {
	db *gorm.DB
}

func NewVibeRepository(db *gorm.DB) *VibeRepository {
	return &VibeRepository{db: db}
}

func (r *VibeRepository) Upsert(vibe *models.UserVibe) error {
	var existing models.UserVibe
	err := r.db.Where("user_id = ? AND week_start = ?", vibe.UserID, vibe.WeekStart).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(vibe).Error
	}
	if err != nil {
		return err
	}
	existing.Vibe = vibe.Vibe
	return r.db.Save(&existing).Error
}

func (r *VibeRepository) FindCurrentWeek(userID uint, weekStart string) (*models.UserVibe, error) {
	var vibe models.UserVibe
	err := r.db.Where("user_id = ? AND week_start = ?", userID, weekStart).First(&vibe).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &vibe, err
}

func (r *VibeRepository) FindLatest(userID uint) (*models.UserVibe, error) {
	var vibe models.UserVibe
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&vibe).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &vibe, err
}
