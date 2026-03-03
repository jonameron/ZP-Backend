package handlers

import (
	"github.com/gofiber/fiber/v2"
	"zeitpass/internal/models"
	"zeitpass/internal/repository"
)

type SocialHandler struct {
	connRepo    *repository.ConnectionRepository
	userRepo    *repository.UserRepository
	bookingRepo *repository.BookingRepository
}

func NewSocialHandler(cr *repository.ConnectionRepository, ur *repository.UserRepository, br *repository.BookingRepository) *SocialHandler {
	return &SocialHandler{connRepo: cr, userRepo: ur, bookingRepo: br}
}

type ConnectInput struct {
	Email string `json:"email"`
	Type  string `json:"type"` // friend, partner
}

func (h *SocialHandler) SendRequest(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	var input ConnectInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if input.Type != "friend" && input.Type != "partner" {
		input.Type = "friend"
	}

	// Find target user by email
	target, err := h.userRepo.FindByEmail(input.Email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found with that email"})
	}

	if target.ID == user.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot connect with yourself"})
	}

	// Check existing connection
	existing, _ := h.connRepo.FindExisting(user.ID, target.ID)
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "connection already exists", "status": existing.Status})
	}

	conn := &models.Connection{
		UserRef:      user.ID,
		ConnectedRef: target.ID,
		Type:         input.Type,
		Status:       "pending",
	}

	if err := h.connRepo.Create(conn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "ok",
		"id":     conn.ID,
	})
}

type RespondInput struct {
	Action string `json:"action"` // accept, reject, block
}

func (h *SocialHandler) RespondToRequest(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	connID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid connection ID"})
	}

	conn, err := h.connRepo.FindByID(uint(connID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "connection not found"})
	}

	// Only the target user can respond
	if conn.ConnectedRef != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not your request"})
	}

	if conn.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "request already handled"})
	}

	var input RespondInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	switch input.Action {
	case "accept":
		conn.Status = "accepted"
	case "reject":
		// Delete the connection request
		if err := h.connRepo.Delete(conn.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok", "action": "rejected"})
	case "block":
		conn.Status = "blocked"
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "action must be accept, reject, or block"})
	}

	if err := h.connRepo.Update(conn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok", "action": input.Action})
}

func (h *SocialHandler) RemoveConnection(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	connID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid connection ID"})
	}

	conn, err := h.connRepo.FindByID(uint(connID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "connection not found"})
	}

	// Either party can remove
	if conn.UserRef != user.ID && conn.ConnectedRef != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not your connection"})
	}

	if err := h.connRepo.Delete(conn.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

type ConnectionDTO struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"userId"`
	FirstName string `json:"firstName"`
	Email     string `json:"email"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

func (h *SocialHandler) GetConnections(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	// Get accepted connections
	conns, err := h.connRepo.FindAcceptedConnections(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	// Get pending requests TO this user
	pending, err := h.connRepo.FindPendingForUser(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Collect all user IDs needed, then batch load
	userIDSet := make(map[uint]bool)
	for _, c := range conns {
		if c.ConnectedRef == user.ID {
			userIDSet[c.UserRef] = true
		} else {
			userIDSet[c.ConnectedRef] = true
		}
	}
	for _, c := range pending {
		userIDSet[c.UserRef] = true
	}
	var allIDs []uint
	for id := range userIDSet {
		allIDs = append(allIDs, id)
	}
	usersMap, err := h.userRepo.FindByIDs(allIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var connections []ConnectionDTO
	for _, cn := range conns {
		otherID := cn.ConnectedRef
		if cn.ConnectedRef == user.ID {
			otherID = cn.UserRef
		}
		other, ok := usersMap[otherID]
		if !ok {
			continue
		}
		connections = append(connections, ConnectionDTO{
			ID:        cn.ID,
			UserID:    other.ID,
			FirstName: other.FirstName,
			Email:     other.Email,
			Type:      cn.Type,
			Status:    "accepted",
		})
	}

	var requests []ConnectionDTO
	for _, cn := range pending {
		sender, ok := usersMap[cn.UserRef]
		if !ok {
			continue
		}
		requests = append(requests, ConnectionDTO{
			ID:        cn.ID,
			UserID:    sender.ID,
			FirstName: sender.FirstName,
			Email:     sender.Email,
			Type:      cn.Type,
			Status:    "pending",
		})
	}

	return c.JSON(fiber.Map{
		"connections": connections,
		"pending":     requests,
	})
}

type FeedItem struct {
	UserName  string `json:"userName"`
	EventTitle string `json:"eventTitle"`
	Category   string `json:"category"`
	Date       string `json:"date"`
}

func (h *SocialHandler) GetFeed(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	// Get connected user IDs
	connectedIDs, err := h.connRepo.GetConnectedUserIDs(user.ID)
	if err != nil || len(connectedIDs) == 0 {
		return c.JSON(fiber.Map{"feed": []FeedItem{}})
	}

	// Batch load connected users and filter by visibility
	usersMap, err := h.userRepo.FindByIDs(connectedIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var visibleIDs []uint
	for id, u := range usersMap {
		if u.ProfileVisibility == "friends" || u.ProfileVisibility == "public" {
			visibleIDs = append(visibleIDs, id)
		}
	}

	if len(visibleIDs) == 0 {
		return c.JSON(fiber.Map{"feed": []FeedItem{}})
	}

	// Batch load attended bookings for all visible users
	bookings, err := h.bookingRepo.FindAttendedByUserIDs(visibleIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var feed []FeedItem
	for _, b := range bookings {
		connUser, ok := usersMap[b.UserID]
		if !ok {
			continue
		}
		feed = append(feed, FeedItem{
			UserName:   connUser.FirstName,
			EventTitle: b.Event.Title,
			Category:   b.Event.Category,
			Date:       b.Event.DateText,
		})
	}

	return c.JSON(fiber.Map{"feed": feed})
}

func (h *SocialHandler) UpdateVisibility(c *fiber.Ctx) error {
	uid, _ := c.Locals("userID").(string)
	if uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	user, err := h.userRepo.FindByUserID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	var input struct {
		Visibility string `json:"visibility"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if input.Visibility != "private" && input.Visibility != "friends" && input.Visibility != "public" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "visibility must be private, friends, or public"})
	}

	user.ProfileVisibility = input.Visibility
	if err := h.userRepo.Update(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok", "visibility": input.Visibility})
}
