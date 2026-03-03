package repository

import (
	"time"

	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) HasBeenSent(bookingID uint, notificationType string) bool {
	var count int64
	r.db.Model(&models.NotificationLog{}).
		Where("booking_id = ? AND notification_type = ?", bookingID, notificationType).
		Count(&count)
	return count > 0
}

func (r *NotificationRepository) MarkSent(bookingID uint, notificationType string) error {
	return r.db.Create(&models.NotificationLog{
		BookingID:        bookingID,
		NotificationType: notificationType,
		SentAt:           time.Now(),
	}).Error
}
