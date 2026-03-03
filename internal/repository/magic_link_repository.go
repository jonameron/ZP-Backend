package repository

import (
	"time"

	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type MagicLinkRepository struct {
	db *gorm.DB
}

func NewMagicLinkRepository(db *gorm.DB) *MagicLinkRepository {
	return &MagicLinkRepository{db: db}
}

func (r *MagicLinkRepository) Create(token *models.MagicLinkToken) error {
	return r.db.Create(token).Error
}

func (r *MagicLinkRepository) FindValidByHash(tokenHash string) (*models.MagicLinkToken, error) {
	var token models.MagicLinkToken
	err := r.db.
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&token).Error
	return &token, err
}

func (r *MagicLinkRepository) MarkUsed(token *models.MagicLinkToken) error {
	now := time.Now()
	token.UsedAt = &now
	return r.db.Save(token).Error
}

func (r *MagicLinkRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.MagicLinkToken{}).Error
}
