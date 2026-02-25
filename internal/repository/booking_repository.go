package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"zeitpass/internal/models"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) FindByUserID(userID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.
		Preload("Event").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) FindByID(id uint) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.Preload("User").Preload("Event").First(&booking, id).Error
	return &booking, err
}

func (r *BookingRepository) Create(booking *models.Booking) error {
	// Generate BookingID: BKG-YYYYMMDD-XXXX
	now := time.Now()
	dateStr := now.Format("20060102")

	// Get count of bookings today for sequence number
	var count int64
	r.db.Model(&models.Booking{}).
		Where("DATE(created_at) = ?", now.Format("2006-01-02")).
		Count(&count)

	booking.BookingID = fmt.Sprintf("BKG-%s-%04d", dateStr, count+1)

	return r.db.Create(booking).Error
}

func (r *BookingRepository) Update(booking *models.Booking) error {
	return r.db.Save(booking).Error
}

func (r *BookingRepository) UpdateFeedback(bookingID string, rating int, freeText string) error {
	return r.db.Model(&models.Booking{}).
		Where("booking_id = ?", bookingID).
		Updates(map[string]interface{}{
			"feedback_rating":    rating,
			"feedback_free_text": freeText,
		}).Error
}
