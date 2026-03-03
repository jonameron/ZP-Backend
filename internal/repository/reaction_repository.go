package repository

import (
	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type ReactionRepository struct {
	db *gorm.DB
}

func NewReactionRepository(db *gorm.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

func (r *ReactionRepository) Upsert(reaction *models.Reaction) error {
	// If user already reacted to this event, update; otherwise create
	var existing models.Reaction
	err := r.db.Where("user_id = ? AND event_id = ?", reaction.UserID, reaction.EventID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(reaction).Error
	}
	if err != nil {
		return err
	}
	existing.ReactionType = reaction.ReactionType
	return r.db.Save(&existing).Error
}

func (r *ReactionRepository) Delete(userID, eventID uint) error {
	return r.db.Where("user_id = ? AND event_id = ?", userID, eventID).Delete(&models.Reaction{}).Error
}

func (r *ReactionRepository) FindByUserAndEvent(userID, eventID uint) (*models.Reaction, error) {
	var reaction models.Reaction
	err := r.db.Where("user_id = ? AND event_id = ?", userID, eventID).First(&reaction).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &reaction, err
}

func (r *ReactionRepository) FindByUser(userID uint) ([]models.Reaction, error) {
	var reactions []models.Reaction
	err := r.db.Where("user_id = ?", userID).Find(&reactions).Error
	return reactions, err
}
