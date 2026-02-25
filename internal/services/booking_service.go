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
	}

	if err := s.bookingRepo.Create(booking); err != nil {
		return nil, err
	}

	return booking, nil
}

type FeedbackInput struct {
	BookingRecordID  string `json:"BookingRecordId"`
	FeedbackRating   int    `json:"FeedbackRating"`
	FeedbackFreeText string `json:"FeedbackFreeText"`
}

func (s *BookingService) SubmitFeedback(input FeedbackInput) error {
	return s.bookingRepo.UpdateFeedback(input.BookingRecordID, input.FeedbackRating, input.FeedbackFreeText)
}
