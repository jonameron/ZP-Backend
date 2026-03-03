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

func (r *EventRepository) Update(event *models.Event) error {
	return r.db.Save(event).Error
}

func (r *EventRepository) FindAll(status string) ([]models.Event, error) {
	var events []models.Event
	q := r.db.Order("created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&events).Error
	return events, err
}

func (r *EventRepository) FindEventUsers(eventID uint) ([]models.EventUser, error) {
	var eus []models.EventUser
	err := r.db.Where("event_id = ?", eventID).Find(&eus).Error
	return eus, err
}

func (r *EventRepository) AssignUsers(eventID uint, userIDs []uint) error {
	// Remove existing assignments
	r.db.Where("event_id = ?", eventID).Delete(&models.EventUser{})
	// Create new ones
	for _, uid := range userIDs {
		eu := models.EventUser{EventID: eventID, UserID: uid}
		if err := r.db.Create(&eu).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *EventRepository) CountBookings(eventID uint) int64 {
	var count int64
	r.db.Model(&models.Booking{}).Where("event_id = ?", eventID).Count(&count)
	return count
}

func (r *EventRepository) CountReactions(eventID uint) (int64, int64) {
	var up, down int64
	r.db.Model(&models.Reaction{}).Where("event_id = ? AND reaction_type = ?", eventID, "up").Count(&up)
	r.db.Model(&models.Reaction{}).Where("event_id = ? AND reaction_type = ?", eventID, "down").Count(&down)
	return up, down
}
