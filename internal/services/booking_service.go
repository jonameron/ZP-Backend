package services

import (
	"errors"

	"zeitpass/internal/models"
	"zeitpass/internal/repository"
)

type BookingService struct {
	bookingRepo *repository.BookingRepository
	eventRepo   *repository.EventRepository
	userRepo    *repository.UserRepository
}

func NewBookingService(br *repository.BookingRepository, er *repository.EventRepository, ur *repository.UserRepository) *BookingService {
	return &BookingService{
		bookingRepo: br,
		eventRepo:   er,
		userRepo:    ur,
	}
}

type CreateBookingInput struct {
	UserID       string `json:"UserID"`
	EventID      string `json:"EventID"`
	Status       string `json:"Status"`
	RequestNotes string `json:"RequestNotes"`
	GuestName    string `json:"GuestName"`
	GuestEmail   string `json:"GuestEmail"`
	IsMystery    bool   `json:"IsMystery"`
}

func (s *BookingService) CreateBooking(input CreateBookingInput) (*models.Booking, error) {
	user, err := s.userRepo.FindByUserID(input.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	event, err := s.eventRepo.FindByEventID(input.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	booking := &models.Booking{
		UserID:       user.ID,
		EventID:      event.ID,
		Status:       input.Status,
		RequestNotes: input.RequestNotes,
		GuestName:    input.GuestName,
		GuestEmail:   input.GuestEmail,
		IsMystery:    input.IsMystery,
	}

	if err := s.bookingRepo.Create(booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) CancelBooking(bookingID string, userID string) error {
	user, err := s.userRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	booking, err := s.bookingRepo.FindByBookingID(bookingID)
	if err != nil {
		return errors.New("booking not found")
	}

	// Verify ownership
	if booking.UserID != user.ID {
		return errors.New("not your booking")
	}

	// Only allow cancellation of Requested or Confirmed bookings
	if booking.Status != "Requested" && booking.Status != "Confirmed" {
		return errors.New("booking cannot be cancelled in its current state")
	}

	booking.Status = "Cancelled"
	return s.bookingRepo.Update(booking)
}

type FeedbackInput struct {
	BookingRecordID  string `json:"BookingRecordId"`
	FeedbackRating   int    `json:"FeedbackRating"`
	FeedbackFreeText string `json:"FeedbackFreeText"`
}

func (s *BookingService) SubmitFeedback(input FeedbackInput) error {
	return s.bookingRepo.UpdateFeedback(input.BookingRecordID, input.FeedbackRating, input.FeedbackFreeText)
}
