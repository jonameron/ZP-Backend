package repository

import (
	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type OnboardingRepository struct {
	db *gorm.DB
}

func NewOnboardingRepository(db *gorm.DB) *OnboardingRepository {
	return &OnboardingRepository{db: db}
}

func (r *OnboardingRepository) Create(submission *models.OnboardingSubmission) error {
	return r.db.Create(submission).Error
}

func (r *OnboardingRepository) FindBySessionID(sessionID string) (*models.OnboardingSubmission, error) {
	var submission models.OnboardingSubmission
	err := r.db.Where("session_id = ?", sessionID).First(&submission).Error
	return &submission, err
}
