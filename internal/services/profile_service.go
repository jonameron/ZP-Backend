package services

import (
	"gorm.io/gorm"
	"zeitpass/internal/models"
	"zeitpass/internal/repository"
)

type ProfileService struct {
	db           *gorm.DB
	userRepo     *repository.UserRepository
	bookingRepo  *repository.BookingRepository
	activityRepo *repository.ActivityRepository
}

func NewProfileService(db *gorm.DB, ur *repository.UserRepository, br *repository.BookingRepository, ar *repository.ActivityRepository) *ProfileService {
	return &ProfileService{
		db:           db,
		userRepo:     ur,
		bookingRepo:  br,
		activityRepo: ar,
	}
}

type ProfileResponse struct {
	User     UserDTO      `json:"user"`
	Bookings []BookingDTO `json:"bookings"`
}

type UserDTO struct {
	UID               string `json:"uid"`
	FirstName         string `json:"FirstName"`
	Email             string `json:"Email"`
	Plan              string `json:"Plan"`
	TotalEvents       int    `json:"TotalEvents"`
	GuestAccess       bool   `json:"GuestAccess"`
	PlanRenewalDate   string `json:"PlanRenewalDate"`
	ProfileVisibility string `json:"ProfileVisibility"`
}

type BookingDTO struct {
	BookingRecordID    string `json:"BookingRecordId"`
	BookingID          string `json:"BookingID"`
	EventTitle         string `json:"EventTitle"`
	EventDate          string `json:"EventDate"`
	EventTime          string `json:"EventTime"`
	EventNeighbourhood string `json:"EventNeighbourhood"`
	Status             string `json:"Status"`
	RequestNotes       string `json:"RequestNotes"`
	GuestName          string `json:"GuestName"`
	FeedbackRating     *int   `json:"FeedbackRating"`
	FeedbackFreeText   string `json:"FeedbackFreeText"`
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
		UID:               user.UserID,
		FirstName:         user.FirstName,
		Email:             user.Email,
		Plan:              user.Plan,
		TotalEvents:       user.TotalEvents,
		GuestAccess:       user.GuestAccess,
		ProfileVisibility: user.ProfileVisibility,
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
			GuestName:          booking.GuestName,
			FeedbackRating:     booking.FeedbackRating,
			FeedbackFreeText:   booking.FeedbackFreeText,
		}
	}

	return &ProfileResponse{
		User:     userDTO,
		Bookings: bookingDTOs,
	}, nil
}

type DataExport struct {
	User        *models.User         `json:"user"`
	Bookings    []models.Booking     `json:"bookings"`
	Activity    []models.ActivityLog `json:"activity"`
	Reactions   []models.Reaction    `json:"reactions"`
	Vibes       []models.UserVibe    `json:"vibes"`
	Badges      []models.UserBadge   `json:"badges"`
	Connections []models.Connection  `json:"connections"`
	Invites     []models.Invite      `json:"invites"`
}

func (s *ProfileService) ExportData(uid string) (*DataExport, error) {
	user, err := s.userRepo.FindByUserID(uid)
	if err != nil {
		return nil, err
	}

	bookings, err := s.bookingRepo.FindByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	activity, err := s.activityRepo.FindByUserID(user.ID, 10000)
	if err != nil {
		return nil, err
	}

	var reactions []models.Reaction
	s.db.Where("user_id = ?", user.ID).Find(&reactions)

	var vibes []models.UserVibe
	s.db.Where("user_id = ?", user.ID).Find(&vibes)

	var badges []models.UserBadge
	s.db.Where("user_id = ?", user.ID).Find(&badges)

	var connections []models.Connection
	s.db.Where("user_ref = ? OR connected_ref = ?", user.ID, user.ID).Find(&connections)

	var invites []models.Invite
	s.db.Where("host_ref = ?", user.ID).Find(&invites)

	return &DataExport{
		User:        user,
		Bookings:    bookings,
		Activity:    activity,
		Reactions:   reactions,
		Vibes:       vibes,
		Badges:      badges,
		Connections: connections,
		Invites:     invites,
	}, nil
}

func (s *ProfileService) DeleteAccount(uid string) error {
	user, err := s.userRepo.FindByUserID(uid)
	if err != nil {
		return err
	}

	// Explicitly delete from tables that don't have ON DELETE CASCADE
	// (Connection uses gorm:"-" tags, Booking has gorm:"-" for associations,
	// Invite has no user FK cascade)
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete bookings
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.Booking{}).Error; err != nil {
			return err
		}
		// Delete connections (both directions)
		if err := tx.Where("user_ref = ? OR connected_ref = ?", user.ID, user.ID).Delete(&models.Connection{}).Error; err != nil {
			return err
		}
		// Delete invites where user is host
		if err := tx.Where("host_ref = ?", user.ID).Delete(&models.Invite{}).Error; err != nil {
			return err
		}
		// Delete the user (cascades: activity_logs, event_users, magic_link_tokens, reactions, user_vibes, user_badges)
		if err := tx.Delete(&models.User{}, user.ID).Error; err != nil {
			return err
		}
		return nil
	})
}
