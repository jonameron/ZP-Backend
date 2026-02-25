package handlers

import (
	"github.com/gofiber/fiber/v2"
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
}

type CurationResponse struct {
	Events   []EventDTO `json:"events"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int        `json:"total"`
}

func (h *CurationHandler) GetCuratedEvents(c *fiber.Ctx) error {
	uid := c.Query("uid")
	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "uid parameter is required",
		})
	}

	events, err := h.eventService.GetCuratedEvents(uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	eventDTOs := make([]EventDTO, len(events))
	for i, event := range events {
		eventDTOs[i] = EventDTO{
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
	uid := c.Params("uid")
	if uid == "" {
		uid = c.Query("uid")
	}

	if uid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "uid parameter is required",
		})
	}

	profile, err := h.profileService.GetProfile(uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(profile)
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

	if err := h.activityService.LogActivity(input); err != nil {
		// Log activity errors but don't fail the request
		// This ensures activity logging doesn't break user experience
		return c.JSON(fiber.Map{
			"status": "logged_with_warning",
			"error":  err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
	})
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
