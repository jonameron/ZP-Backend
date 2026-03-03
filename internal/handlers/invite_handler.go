package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	"github.com/gofiber/fiber/v2"
	"zeitpass/internal/models"
	"zeitpass/internal/repository"
	"zeitpass/internal/services"
)

type InviteHandler struct {
	inviteRepo   *repository.InviteRepository
	bookingRepo  *repository.BookingRepository
	userRepo     *repository.UserRepository
	emailService *services.EmailService
}

func NewInviteHandler(ir *repository.InviteRepository, br *repository.BookingRepository, ur *repository.UserRepository, es *services.EmailService) *InviteHandler {
	return &InviteHandler{inviteRepo: ir, bookingRepo: br, userRepo: ur, emailService: es}
}

func (h *InviteHandler) CreateInvite(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	bookingID := c.Params("id")
	booking, err := h.bookingRepo.FindByBookingID(bookingID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "booking not found"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	if booking.UserID != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not your booking"})
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}
	token := hex.EncodeToString(tokenBytes)

	invite := &models.Invite{
		Token:      token,
		BookingRef: booking.ID,
		EventRef:   booking.EventID,
		HostRef:    user.ID,
		Status:     "Pending",
	}

	if err := h.inviteRepo.Create(invite); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://zp.11data.ai"
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"token":  token,
		"url":    frontendURL + "/invite/" + token,
	})
}

func (h *InviteHandler) GetInvitePreview(c *fiber.Ctx) error {
	token := c.Params("token")
	invite, event, host, err := h.inviteRepo.FindByToken(token)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "invite not found"})
	}

	return c.JSON(fiber.Map{
		"eventTitle":     event.Title,
		"eventCategory":  event.Category,
		"eventTier":      event.Tier,
		"eventImage":     event.ImageURL,
		"eventDate":      event.DateText,
		"eventDay":       event.Day,
		"eventTime":      event.Time,
		"eventArea":      event.Neighbourhood,
		"eventShortDesc": event.ShortDesc,
		"hostName":       host.FirstName,
		"status":         invite.Status,
	})
}

type RSVPInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *InviteHandler) RSVP(c *fiber.Ctx) error {
	token := c.Params("token")
	invite, _, _, err := h.inviteRepo.FindByToken(token)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "invite not found"})
	}

	if invite.Status != "Pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invite already used"})
	}

	var input RSVPInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	invite.GuestName = input.Name
	invite.GuestEmail = input.Email
	invite.Status = "Accepted"

	if err := h.inviteRepo.Update(invite); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Update the booking with guest info
	if booking, err := h.bookingRepo.FindByID(invite.BookingRef); err == nil {
		booking.GuestName = input.Name
		booking.GuestEmail = input.Email
		if err := h.bookingRepo.Update(booking); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"status": "ok", "message": "RSVP confirmed"})
}
