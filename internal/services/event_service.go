package services

import (
	"zeitpass/internal/models"
	"zeitpass/internal/repository"
)

type EventService struct {
	eventRepo *repository.EventRepository
	userRepo  *repository.UserRepository
}

func NewEventService(er *repository.EventRepository, ur *repository.UserRepository) *EventService {
	return &EventService{
		eventRepo: er,
		userRepo:  ur,
	}
}

func (s *EventService) GetCuratedEvents(uid string) ([]models.Event, error) {
	user, err := s.userRepo.FindByUserID(uid)
	if err != nil {
		return nil, err
	}

	events, err := s.eventRepo.FindLiveEventsForUser(user.ID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *EventService) GetEvent(eventID string) (*models.Event, error) {
	return s.eventRepo.FindByEventID(eventID)
}
