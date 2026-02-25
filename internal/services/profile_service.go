package services

import (
	"zeitpass/internal/repository"
)

type ProfileService struct {
	userRepo    *repository.UserRepository
	bookingRepo *repository.BookingRepository
}

func NewProfileService(ur *repository.UserRepository, br *repository.BookingRepository) *ProfileService {
	return &ProfileService{
		userRepo:    ur,
		bookingRepo: br,
	}
}

type ProfileResponse struct {
	User     UserDTO      `json:"user"`
	Bookings []BookingDTO `json:"bookings"`
}

type UserDTO struct {
	UID             string `json:"uid"`
	FirstName       string `json:"FirstName"`
	Plan            string `json:"Plan"`
	TotalEvents     int    `json:"TotalEvents"`
	GuestAccess     bool   `json:"GuestAccess"`
	PlanRenewalDate string `json:"PlanRenewalDate"`
}

type BookingDTO struct {
	BookingRecordID   string  `json:"BookingRecordId"`
	BookingID         string  `json:"BookingID"`
	EventTitle        string  `json:"EventTitle"`
	EventDate         string  `json:"EventDate"`
	EventTime         string  `json:"EventTime"`
	EventNeighbourhood string `json:"EventNeighbourhood"`
	Status            string  `json:"Status"`
	RequestNotes      string  `json:"RequestNotes"`
}

func (s *ProfileService) GetProfile(uid string) (*ProfileResponse, error) {
	user, err := s.userRepo.FindByUserID(uid)
	if err != nil {
		return nil, err
	}

	bookings, err := s.bookingRepo.FindByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	// Map user
	userDTO := UserDTO{
		UID:         user.UserID,
		FirstName:   user.FirstName,
		Plan:        user.Plan,
		TotalEvents: user.TotalEvents,
		GuestAccess: user.GuestAccess,
	}
	if user.PlanRenewalDate != nil {
		userDTO.PlanRenewalDate = user.PlanRenewalDate.Format("2006-01-02")
	}

	// Map bookings
	bookingDTOs := make([]BookingDTO, len(bookings))
	for i, booking := range bookings {
		eventDate := booking.EventDateConfirmed
		if eventDate == "" && booking.Event.DateText != "" {
			eventDate = booking.Event.DateText
		}

		eventTime := booking.EventTimeConfirmed
		if eventTime == "" && booking.Event.Time != "" {
			eventTime = booking.Event.Time
		}

		bookingDTOs[i] = BookingDTO{
			BookingRecordID:    booking.BookingID,
			BookingID:          booking.BookingID,
			EventTitle:         booking.Event.Title,
			EventDate:          eventDate,
			EventTime:          eventTime,
			EventNeighbourhood: booking.Event.Neighbourhood,
			Status:             booking.Status,
			RequestNotes:       booking.RequestNotes,
		}
	}

	return &ProfileResponse{
		User:     userDTO,
		Bookings: bookingDTOs,
	}, nil
}
