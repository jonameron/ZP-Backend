package repository

import (
	"time"

	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type BadgeRepository struct {
	db *gorm.DB
}

func NewBadgeRepository(db *gorm.DB) *BadgeRepository {
	return &BadgeRepository{db: db}
}

func (r *BadgeRepository) FindByUser(userID uint) ([]models.UserBadge, error) {
	var badges []models.UserBadge
	err := r.db.Where("user_id = ?", userID).Order("earned_at ASC").Find(&badges).Error
	return badges, err
}

func (r *BadgeRepository) HasBadge(userID uint, badgeKey string) bool {
	var count int64
	r.db.Model(&models.UserBadge{}).
		Where("user_id = ? AND badge_key = ?", userID, badgeKey).
		Count(&count)
	return count > 0
}

func (r *BadgeRepository) Award(userID uint, badgeKey string) error {
	if r.HasBadge(userID, badgeKey) {
		return nil
	}
	return r.db.Create(&models.UserBadge{
		UserID:   userID,
		BadgeKey: badgeKey,
		EarnedAt: time.Now(),
	}).Error
}
