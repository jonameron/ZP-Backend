package repository

import (
	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) FindLiveEventsForUser(userID uint) ([]models.Event, error) {
	var events []models.Event

	err := r.db.
		Joins("JOIN event_users ON event_users.event_id = events.id").
		Where("event_users.user_id = ? AND events.status = ?", userID, "Live").
		Order("events.curation_order ASC, events.date_text ASC").
		Find(&events).Error

	return events, err
}

func (r *EventRepository) FindByEventID(eventID string) (*models.Event, error) {
	var event models.Event
	err := r.db.Where("event_id = ?", eventID).First(&event).Error
	return &event, err
}

func (r *EventRepository) FindByID(id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.First(&event, id).Error
	return &event, err
}

func (r *EventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}
