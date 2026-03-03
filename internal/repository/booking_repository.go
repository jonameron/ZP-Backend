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

// loadRelations populates User and Event for a slice of bookings
func (r *BookingRepository) loadRelations(bookings []models.Booking, loadUser, loadEvent bool) {
	if len(bookings) == 0 {
		return
	}

	if loadUser {
		userIDs := make(map[uint]bool)
		for _, b := range bookings {
			userIDs[b.UserID] = true
		}
		ids := make([]uint, 0, len(userIDs))
		for id := range userIDs {
			ids = append(ids, id)
		}
		var users []models.User
		r.db.Where("id IN ?", ids).Find(&users)
		userMap := make(map[uint]models.User, len(users))
		for _, u := range users {
			userMap[u.ID] = u
		}
		for i := range bookings {
			if u, ok := userMap[bookings[i].UserID]; ok {
				bookings[i].User = u
			}
		}
	}

	if loadEvent {
		eventIDs := make(map[uint]bool)
		for _, b := range bookings {
			eventIDs[b.EventID] = true
		}
		ids := make([]uint, 0, len(eventIDs))
		for id := range eventIDs {
			ids = append(ids, id)
		}
		var events []models.Event
		r.db.Where("id IN ?", ids).Find(&events)
		eventMap := make(map[uint]models.Event, len(events))
		for _, e := range events {
			eventMap[e.ID] = e
		}
		for i := range bookings {
			if e, ok := eventMap[bookings[i].EventID]; ok {
				bookings[i].Event = e
			}
		}
	}
}

func (r *BookingRepository) FindByUserID(userID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bookings).Error
	if err == nil {
		r.loadRelations(bookings, false, true)
	}
	return bookings, err
}

func (r *BookingRepository) FindByID(id uint) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.First(&booking, id).Error
	if err == nil {
		bookings := []models.Booking{booking}
		r.loadRelations(bookings, true, true)
		booking = bookings[0]
	}
	return &booking, err
}

func (r *BookingRepository) FindByBookingID(bookingID string) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.Where("booking_id = ?", bookingID).First(&booking).Error
	if err == nil {
		bookings := []models.Booking{booking}
		r.loadRelations(bookings, false, true)
		booking = bookings[0]
	}
	return &booking, err
}

func (r *BookingRepository) Create(booking *models.Booking) error {
	// Generate BookingID: BKG-YYYYMMDD-XXXX with retry for uniqueness
	now := time.Now()
	dateStr := now.Format("20060102")

	for attempt := 0; attempt < 5; attempt++ {
		var count int64
		r.db.Model(&models.Booking{}).
			Where("booking_id LIKE ?", fmt.Sprintf("BKG-%s-%%", dateStr)).
			Count(&count)

		booking.BookingID = fmt.Sprintf("BKG-%s-%04d", dateStr, count+1)

		err := r.db.Create(booking).Error
		if err == nil {
			return nil
		}
		// If it's a unique constraint violation, retry with next number
		if attempt < 4 {
			continue
		}
		return err
	}
	return nil
}

func (r *BookingRepository) Update(booking *models.Booking) error {
	return r.db.Save(booking).Error
}

func (r *BookingRepository) FindAllAdmin(status string) ([]models.Booking, error) {
	var bookings []models.Booking
	q := r.db.Order("created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&bookings).Error
	if err == nil {
		r.loadRelations(bookings, true, true)
	}
	return bookings, err
}

func (r *BookingRepository) FindConfirmedWithDates() ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.
		Where("status IN ? AND event_date_confirmed != ''", []string{"Confirmed", "Attended"}).
		Find(&bookings).Error
	if err == nil {
		r.loadRelations(bookings, true, true)
	}
	return bookings, err
}

func (r *BookingRepository) FindAttendedByUserIDs(userIDs []uint) ([]models.Booking, error) {
	var bookings []models.Booking
	if len(userIDs) == 0 {
		return bookings, nil
	}
	err := r.db.
		Where("user_id IN ? AND status = ?", userIDs, "Attended").
		Order("created_at DESC").
		Find(&bookings).Error
	if err == nil {
		r.loadRelations(bookings, false, true)
	}
	return bookings, err
}

func (r *BookingRepository) UpdateFeedback(bookingID string, rating int, freeText string) error {
	return r.db.Model(&models.Booking{}).
		Where("booking_id = ?", bookingID).
		Updates(map[string]interface{}{
			"feedback_rating":    rating,
			"feedback_free_text": freeText,
		}).Error
}
