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
	SessionID string `json:"SessionID"`
	PageURL   string `json:"PageURL"`
	EventID   string `json:"EventID"`
	Action    string `json:"Action"`
	Timestamp string `json:"Timestamp"`
	Metadata  string `json:"Metadata"`
}

func (s *ActivityService) LogActivity(input ActivityInput) error {
	// Extract UserID from SessionID (first 9 characters)
	userIDStr := input.SessionID
	if len(userIDStr) > 9 {
		userIDStr = userIDStr[:9]
	}

	user, err := s.userRepo.FindByUserID(userIDStr)
	if err != nil {
		// If user not found, still log activity but without user linkage
		// For now, return error
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
