package services

import (
	"time"
	"zeitpass/internal/models"
	"zeitpass/internal/repository"
)

type OnboardingService struct {
	onboardingRepo *repository.OnboardingRepository
}

func NewOnboardingService(or *repository.OnboardingRepository) *OnboardingService {
	return &OnboardingService{
		onboardingRepo: or,
	}
}

type OnboardingInput struct {
	SessionID             string `json:"session_id" validate:"required"`
	Timestamp             string `json:"timestamp" validate:"required"`
	Device                string `json:"device" validate:"required,oneof=desktop mobile tablet"`
	SourceURL             string `json:"source_url"`
	NameParam             string `json:"name_param"`
	ContactPreference     string `json:"contact_preference"`
	WeekdayVibe           string `json:"weekday_vibe"`
	WeekendFlavor         string `json:"weekend_flavor"`
	SocialStyle           string `json:"social_style"`
	EnergyLevel           string `json:"energy_level"`
	Companionship         string `json:"companionship"`
	CulturalPersonality   string `json:"cultural_personality"`
	CreativeEngagement    string `json:"creative_engagement"`
	Motto                 string `json:"motto"`
	Language              string `json:"language"`
	PilotValueExpectation string `json:"pilot_value_expectation"`
}

type OnboardingAnswers struct {
	ContactPreference     string `json:"contact_preference"`
	WeekdayVibe           string `json:"weekday_vibe"`
	WeekendFlavor         string `json:"weekend_flavor"`
	SocialStyle           string `json:"social_style"`
	EnergyLevel           string `json:"energy_level"`
	Companionship         string `json:"companionship"`
	CulturalPersonality   string `json:"cultural_personality"`
	CreativeEngagement    string `json:"creative_engagement"`
	Motto                 string `json:"motto"`
	Language              string `json:"language"`
	PilotValueExpectation string `json:"pilot_value_expectation"`
}

type OnboardingSubmissionInput struct {
	SessionID string            `json:"session_id" validate:"required"`
	Timestamp string            `json:"timestamp" validate:"required"`
	Device    string            `json:"device" validate:"required,oneof=desktop mobile tablet"`
	SourceURL string            `json:"source_url"`
	Name      string            `json:"name"`
	Email     string            `json:"email"`
	NameParam string            `json:"name_param"`
	Answers   OnboardingAnswers `json:"answers" validate:"required"`
}

func (s *OnboardingService) SubmitOnboarding(input OnboardingSubmissionInput) (*models.OnboardingSubmission, error) {
	timestamp, err := time.Parse(time.RFC3339, input.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	submission := &models.OnboardingSubmission{
		SessionID:             input.SessionID,
		Timestamp:             timestamp,
		Device:                input.Device,
		SourceURL:             input.SourceURL,
		Name:                  input.Name,
		Email:                 input.Email,
		NameParam:             input.NameParam,
		ContactPreference:     input.Answers.ContactPreference,
		WeekdayVibe:           input.Answers.WeekdayVibe,
		WeekendFlavor:         input.Answers.WeekendFlavor,
		SocialStyle:           input.Answers.SocialStyle,
		EnergyLevel:           input.Answers.EnergyLevel,
		Companionship:         input.Answers.Companionship,
		CulturalPersonality:   input.Answers.CulturalPersonality,
		CreativeEngagement:    input.Answers.CreativeEngagement,
		Motto:                 input.Answers.Motto,
		Language:              input.Answers.Language,
		PilotValueExpectation: input.Answers.PilotValueExpectation,
	}

	if err := s.onboardingRepo.Create(submission); err != nil {
		return nil, err
	}

	return submission, nil
}
