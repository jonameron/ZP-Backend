package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"zeitpass/internal/models"
	"zeitpass/internal/repository"
	"zeitpass/internal/services"
)

// CurationHandler handles event curation endpoints
type CurationHandler struct {
	eventService *services.EventService
}

func NewCurationHandler(es *services.EventService) *CurationHandler {
	return &CurationHandler{eventService: es}
}

type EventDTO struct {
	EventID               string `json:"EventID"`
	Category              string `json:"Category"`
	Tier                  string `json:"Tier"`
	Title                 string `json:"Title"`
	Vendor                string `json:"Vendor"`
	ImageURL              string `json:"ImageURL"`
	DateText              string `json:"DateText"`
	Day                   string `json:"Day"`
	Time                  string `json:"Time"`
	Duration              string `json:"Duration"`
	Neighbourhood         string `json:"Neighbourhood"`
	Address               string `json:"Address"`
	ShortDesc             string `json:"ShortDesc"`
	LongDesc              string `json:"LongDesc"`
	Highlights            string `json:"Highlights"`
	AdditionalInformation string `json:"AdditionalInformation"`
	Status                string `json:"Status"`
	StatusQualifier       string `json:"StatusQualifier"`
	Cancellation          string `json:"Cancellation"`
	Language              string `json:"Language"`
	CurationOrder         int    `json:"CurationOrder"`
	IsMystery             bool   `json:"IsMystery"`
}

type CurationResponse struct {
	Events   []EventDTO `json:"events"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int        `json:"total"`
}

func (h *CurationHandler) GetCuratedEvents(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}

	events, err := h.eventService.GetCuratedEvents(uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	eventDTOs := make([]EventDTO, len(events))
	mysteryUsed := false
	for i, event := range events {
		dto := EventDTO{
			EventID:               event.EventID,
			Category:              event.Category,
			Tier:                  event.Tier,
			Title:                 event.Title,
			Vendor:                event.Vendor,
			ImageURL:              event.ImageURL,
			DateText:              event.DateText,
			Day:                   event.Day,
			Time:                  event.Time,
			Duration:              event.Duration,
			Neighbourhood:         event.Neighbourhood,
			Address:               event.Address,
			ShortDesc:             event.ShortDesc,
			LongDesc:              event.LongDesc,
			Highlights:            event.Highlights,
			AdditionalInformation: event.AdditionalInformation,
			Status:                event.Status,
			StatusQualifier:       event.StatusQualifier,
			Cancellation:          event.Cancellation,
			Language:              event.Language,
			CurationOrder:         event.CurationOrder,
		}

		// Mark first mystery-eligible event as mystery and redact details
		if event.MysteryEligible && !mysteryUsed {
			dto.IsMystery = true
			dto.Title = "Mystery Experience"
			dto.Vendor = ""
			dto.ImageURL = ""
			dto.ShortDesc = ""
			dto.LongDesc = ""
			dto.Highlights = ""
			dto.Address = ""
			dto.AdditionalInformation = ""
			mysteryUsed = true
		}

		eventDTOs[i] = dto
	}

	return c.JSON(CurationResponse{
		Events:   eventDTOs,
		Page:     1,
		PageSize: len(eventDTOs),
		Total:    len(eventDTOs),
	})
}

func (h *CurationHandler) GetEvent(c *fiber.Ctx) error {
	eventID := c.Params("eventId")
	if eventID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "eventId parameter is required",
		})
	}

	event, err := h.eventService.GetEvent(eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "event not found",
		})
	}

	eventDTO := EventDTO{
		EventID:               event.EventID,
		Category:              event.Category,
		Tier:                  event.Tier,
		Title:                 event.Title,
		Vendor:                event.Vendor,
		ImageURL:              event.ImageURL,
		DateText:              event.DateText,
		Day:                   event.Day,
		Time:                  event.Time,
		Duration:              event.Duration,
		Neighbourhood:         event.Neighbourhood,
		Address:               event.Address,
		ShortDesc:             event.ShortDesc,
		LongDesc:              event.LongDesc,
		Highlights:            event.Highlights,
		AdditionalInformation: event.AdditionalInformation,
		Status:                event.Status,
		StatusQualifier:       event.StatusQualifier,
		Cancellation:          event.Cancellation,
		Language:              event.Language,
		CurationOrder:         event.CurationOrder,
	}

	return c.JSON(eventDTO)
}

// BookingHandler handles booking endpoints
type BookingHandler struct {
	bookingService *services.BookingService
}

func NewBookingHandler(bs *services.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bs}
}

func (h *BookingHandler) CreateBooking(c *fiber.Ctx) error {
	var input services.CreateBookingInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Override UserID from JWT — don't trust request body
	if uid, ok := c.Locals("userID").(string); ok && uid != "" {
		input.UserID = uid
	}

	booking, err := h.bookingService.CreateBooking(input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":    "ok",
		"bookingId": booking.BookingID,
	})
}

func (h *BookingHandler) CancelBooking(c *fiber.Ctx) error {
	bookingID := c.Params("id")
	if bookingID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "booking ID required"})
	}

	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	if err := h.bookingService.CancelBooking(bookingID, uid); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok", "message": "Booking cancelled"})
}

func (h *BookingHandler) SubmitFeedback(c *fiber.Ctx) error {
	var input services.FeedbackInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.bookingService.SubmitFeedback(input); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"ok":              true,
		"message":         "Feedback recorded",
		"BookingRecordId": input.BookingRecordID,
	})
}

// ProfileHandler handles profile endpoints
type ProfileHandler struct {
	profileService *services.ProfileService
}

func NewProfileHandler(ps *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: ps}
}

func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user not authenticated"})
	}

	profile, err := h.profileService.GetProfile(uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(profile)
}

func (h *ProfileHandler) ExportData(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user not authenticated"})
	}

	data, err := h.profileService.ExportData(uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Disposition", "attachment; filename=zeitpass-data-export.json")
	return c.JSON(data)
}

func (h *ProfileHandler) DeleteAccount(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user not authenticated"})
	}

	if err := h.profileService.DeleteAccount(uid); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok", "message": "Account deleted"})
}

// ActivityHandler handles activity logging endpoints
type ActivityHandler struct {
	activityService *services.ActivityService
}

func NewActivityHandler(as *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: as}
}

func (h *ActivityHandler) LogActivity(c *fiber.Ctx) error {
	var input services.ActivityInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Inject user identity from JWT context
	if uid, ok := c.Locals("userID").(string); ok {
		input.UserID = uid
	}

	if err := h.activityService.LogActivity(input); err != nil {
		return c.JSON(fiber.Map{
			"status": "logged_with_warning",
			"error":  err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// OnboardingHandler handles onboarding quiz submission endpoints
type OnboardingHandler struct {
	onboardingService *services.OnboardingService
}

func NewOnboardingHandler(os *services.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{onboardingService: os}
}

func (h *OnboardingHandler) SubmitOnboarding(c *fiber.Ctx) error {
	var input services.OnboardingSubmissionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	submission, err := h.onboardingService.SubmitOnboarding(input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "ok",
		"id":     submission.SessionID,
	})
}

// VibeHandler handles weekly vibe poll endpoints
type VibeHandler struct {
	vibeRepo *repository.VibeRepository
	userRepo *repository.UserRepository
}

func NewVibeHandler(vr *repository.VibeRepository, ur *repository.UserRepository) *VibeHandler {
	return &VibeHandler{vibeRepo: vr, userRepo: ur}
}

type VibeInput struct {
	Vibe string `json:"vibe"`
}

func currentWeekStart() string {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	return monday.Format("2006-01-02")
}

func (h *VibeHandler) SetVibe(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	var input VibeInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	vibe := &models.UserVibe{
		UserID:    user.ID,
		Vibe:      input.Vibe,
		WeekStart: currentWeekStart(),
	}

	if err := h.vibeRepo.Upsert(vibe); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok", "vibe": input.Vibe, "weekStart": vibe.WeekStart})
}

func (h *VibeHandler) GetVibe(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	vibe, err := h.vibeRepo.FindCurrentWeek(user.ID, currentWeekStart())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if vibe == nil {
		return c.JSON(fiber.Map{"hasVibe": false})
	}

	return c.JSON(fiber.Map{"hasVibe": true, "vibe": vibe.Vibe, "weekStart": vibe.WeekStart})
}

// ReactionHandler handles event reaction endpoints
type ReactionHandler struct {
	reactionRepo *repository.ReactionRepository
	eventRepo    *repository.EventRepository
	userRepo     *repository.UserRepository
}

func NewReactionHandler(rr *repository.ReactionRepository, er *repository.EventRepository, ur *repository.UserRepository) *ReactionHandler {
	return &ReactionHandler{reactionRepo: rr, eventRepo: er, userRepo: ur}
}

type ReactionInput struct {
	EventID      string `json:"eventId"`
	ReactionType string `json:"reactionType"` // "up" or "down"
}

func (h *ReactionHandler) ToggleReaction(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	var input ReactionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if input.ReactionType != "up" && input.ReactionType != "down" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "reactionType must be 'up' or 'down'"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	event, err := h.eventRepo.FindByEventID(input.EventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "event not found"})
	}

	// Check if user already has this reaction — toggle off if same type
	existing, err := h.reactionRepo.FindByUserAndEvent(user.ID, event.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if existing != nil && existing.ReactionType == input.ReactionType {
		// Same reaction again → remove it (toggle off)
		if err := h.reactionRepo.Delete(user.ID, event.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "removed", "reactionType": nil})
	}

	// Create or update reaction
	reaction := &models.Reaction{
		UserID:       user.ID,
		EventID:      event.ID,
		ReactionType: input.ReactionType,
	}
	if err := h.reactionRepo.Upsert(reaction); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok", "reactionType": input.ReactionType})
}

func (h *ReactionHandler) GetMyReactions(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	reactions, err := h.reactionRepo.FindByUser(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Map event DB IDs to event string IDs for frontend
	result := make(map[string]string)
	for _, r := range reactions {
		event, err := h.eventRepo.FindByID(r.EventID)
		if err == nil {
			result[event.EventID] = r.ReactionType
		}
	}

	return c.JSON(fiber.Map{"reactions": result})
}

// BadgeHandler handles badge endpoints
type BadgeHandler struct {
	badgeService *services.BadgeService
	userRepo     *repository.UserRepository
}

func NewBadgeHandler(bs *services.BadgeService, ur *repository.UserRepository) *BadgeHandler {
	return &BadgeHandler{badgeService: bs, userRepo: ur}
}

func (h *BadgeHandler) GetBadges(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	badges := h.badgeService.GetUserBadges(user.ID)
	return c.JSON(fiber.Map{"badges": badges})
}

// CustomErrorHandler handles application errors
func CustomErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
