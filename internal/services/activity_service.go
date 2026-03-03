package services

import (
	"time"

	"zeitpass/internal/models"
	"zeitpass/internal/repository"
)

type ActivityService struct {
	activityRepo *repository.ActivityRepository
	userRepo     *repository.UserRepository
}

func NewActivityService(ar *repository.ActivityRepository, ur *repository.UserRepository) *ActivityService {
	return &ActivityService{
		activityRepo: ar,
		userRepo:     ur,
	}
}

type ActivityInput struct {
	SessionID string `json:"sessionId"`
	PageURL   string `json:"pageUrl"`
	EventID   string `json:"eventId"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Metadata  string `json:"metadata"`
	UserID    string `json:"-"` // Set from JWT context, not from request body
}

func (s *ActivityService) LogActivity(input ActivityInput) error {
	user, err := s.userRepo.FindByUserID(input.UserID)
	if err != nil {
		return err
	}

	// Parse timestamp
	timestamp := time.Now()
	if input.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, input.Timestamp)
		if err == nil {
			timestamp = parsed
		}
	}

	log := &models.ActivityLog{
		SessionID: input.SessionID,
		UserID:    user.ID,
		PageURL:   input.PageURL,
		Action:    input.Action,
		Metadata:  input.Metadata,
		Timestamp: timestamp,
	}

	return s.activityRepo.Create(log)
}
