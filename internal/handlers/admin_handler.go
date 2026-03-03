package handlers

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"zeitpass/internal/models"
	"zeitpass/internal/repository"
	"zeitpass/internal/services"
)

type AdminHandler struct {
	db           *gorm.DB
	eventRepo    *repository.EventRepository
	userRepo     *repository.UserRepository
	bookingRepo  *repository.BookingRepository
	emailService *services.EmailService
	badgeService *services.BadgeService
}

func NewAdminHandler(db *gorm.DB, er *repository.EventRepository, ur *repository.UserRepository, br *repository.BookingRepository, es *services.EmailService, bs *services.BadgeService) *AdminHandler {
	return &AdminHandler{db: db, eventRepo: er, userRepo: ur, bookingRepo: br, emailService: es, badgeService: bs}
}

// ---- Events ----

type AdminEventDTO struct {
	ID            uint   `json:"id"`
	EventID       string `json:"eventId"`
	Title         string `json:"title"`
	Category      string `json:"category"`
	Tier          string `json:"tier"`
	Vendor        string `json:"vendor"`
	Status        string `json:"status"`
	Day           string `json:"day"`
	DateText      string `json:"dateText"`
	Neighbourhood string `json:"neighbourhood"`
	SourceName    string `json:"sourceName"`
	VibeTags      string `json:"vibeTags"`
	Bookings      int64  `json:"bookings"`
	ThumbsUp      int64  `json:"thumbsUp"`
	ThumbsDown    int64  `json:"thumbsDown"`
	AssignedUsers []uint `json:"assignedUsers"`
}

func (h *AdminHandler) ListEvents(c *fiber.Ctx) error {
	status := c.Query("status")
	events, err := h.eventRepo.FindAll(status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	dtos := make([]AdminEventDTO, len(events))
	for i, e := range events {
		bookings := h.eventRepo.CountBookings(e.ID)
		up, down := h.eventRepo.CountReactions(e.ID)

		eus, _ := h.eventRepo.FindEventUsers(e.ID)
		userIDs := make([]uint, len(eus))
		for j, eu := range eus {
			userIDs[j] = eu.UserID
		}

		dtos[i] = AdminEventDTO{
			ID:            e.ID,
			EventID:       e.EventID,
			Title:         e.Title,
			Category:      e.Category,
			Tier:          e.Tier,
			Vendor:        e.Vendor,
			Status:        e.Status,
			Day:           e.Day,
			DateText:      e.DateText,
			Neighbourhood: e.Neighbourhood,
			SourceName:    e.SourceName,
			VibeTags:      e.VibeTags,
			Bookings:      bookings,
			ThumbsUp:      up,
			ThumbsDown:    down,
			AssignedUsers: userIDs,
		}
	}

	return c.JSON(fiber.Map{"events": dtos, "total": len(dtos)})
}

type CreateEventInput struct {
	Category              string   `json:"category"`
	Tier                  string   `json:"tier"`
	Title                 string   `json:"title"`
	Vendor                string   `json:"vendor"`
	ImageURL              string   `json:"imageUrl"`
	DateText              string   `json:"dateText"`
	Day                   string   `json:"day"`
	Time                  string   `json:"time"`
	Duration              string   `json:"duration"`
	Neighbourhood         string   `json:"neighbourhood"`
	Address               string   `json:"address"`
	ShortDesc             string   `json:"shortDesc"`
	LongDesc              string   `json:"longDesc"`
	Highlights            string   `json:"highlights"`
	AdditionalInformation string   `json:"additionalInformation"`
	Status                string   `json:"status"`
	Cancellation          string   `json:"cancellation"`
	Language              string   `json:"language"`
	RetailPrice           *float64 `json:"retailPrice"`
	MysteryEligible       *bool    `json:"mysteryEligible"`
}

func (h *AdminHandler) CreateEvent(c *fiber.Ctx) error {
	var input CreateEventInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Generate EventID
	eventID := fmt.Sprintf("EVT-%s-%04d", time.Now().Format("20060102"), time.Now().UnixMilli()%10000)

	event := &models.Event{
		EventID:               eventID,
		Category:              input.Category,
		Tier:                  input.Tier,
		Title:                 input.Title,
		Vendor:                input.Vendor,
		ImageURL:              input.ImageURL,
		DateText:              input.DateText,
		Day:                   input.Day,
		Time:                  input.Time,
		Duration:              input.Duration,
		Neighbourhood:         input.Neighbourhood,
		Address:               input.Address,
		ShortDesc:             input.ShortDesc,
		LongDesc:              input.LongDesc,
		Highlights:            input.Highlights,
		AdditionalInformation: input.AdditionalInformation,
		Status:                input.Status,
		Cancellation:          input.Cancellation,
		Language:              input.Language,
		RetailPrice:           input.RetailPrice,
	}

	if input.MysteryEligible != nil {
		event.MysteryEligible = *input.MysteryEligible
	}

	if event.Status == "" {
		event.Status = "Draft"
	}

	if err := h.eventRepo.Create(event); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "ok", "eventId": event.EventID, "id": event.ID})
}

func (h *AdminHandler) UpdateEvent(c *fiber.Ctx) error {
	eventID := c.Params("id")
	event, err := h.eventRepo.FindByEventID(eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "event not found"})
	}

	var input CreateEventInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Update fields (only non-empty values)
	if input.Title != "" {
		event.Title = input.Title
	}
	if input.Category != "" {
		event.Category = input.Category
	}
	if input.Tier != "" {
		event.Tier = input.Tier
	}
	if input.Vendor != "" {
		event.Vendor = input.Vendor
	}
	if input.ImageURL != "" {
		event.ImageURL = input.ImageURL
	}
	if input.DateText != "" {
		event.DateText = input.DateText
	}
	if input.Day != "" {
		event.Day = input.Day
	}
	if input.Time != "" {
		event.Time = input.Time
	}
	if input.Duration != "" {
		event.Duration = input.Duration
	}
	if input.Neighbourhood != "" {
		event.Neighbourhood = input.Neighbourhood
	}
	if input.Address != "" {
		event.Address = input.Address
	}
	if input.ShortDesc != "" {
		event.ShortDesc = input.ShortDesc
	}
	if input.LongDesc != "" {
		event.LongDesc = input.LongDesc
	}
	if input.Highlights != "" {
		event.Highlights = input.Highlights
	}
	if input.AdditionalInformation != "" {
		event.AdditionalInformation = input.AdditionalInformation
	}
	if input.Status != "" {
		event.Status = input.Status
	}
	if input.Cancellation != "" {
		event.Cancellation = input.Cancellation
	}
	if input.Language != "" {
		event.Language = input.Language
	}
	if input.RetailPrice != nil {
		event.RetailPrice = input.RetailPrice
	}
	if input.MysteryEligible != nil {
		event.MysteryEligible = *input.MysteryEligible
	}

	if err := h.eventRepo.Update(event); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

type AssignUsersInput struct {
	UserIDs []uint `json:"userIds"`
}

func (h *AdminHandler) AssignUsers(c *fiber.Ctx) error {
	eventID := c.Params("id")
	event, err := h.eventRepo.FindByEventID(eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "event not found"})
	}

	var input AssignUsersInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.eventRepo.AssignUsers(event.ID, input.UserIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok", "assigned": len(input.UserIDs)})
}

// ---- Users (for admin lookups) ----

type AdminUserDTO struct {
	ID        uint   `json:"id"`
	UserID    string `json:"userId"`
	FirstName string `json:"firstName"`
	Email     string `json:"email"`
	Plan      string `json:"plan"`
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.userRepo.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	dtos := make([]AdminUserDTO, len(users))
	for i, u := range users {
		dtos[i] = AdminUserDTO{
			ID:        u.ID,
			UserID:    u.UserID,
			FirstName: u.FirstName,
			Email:     u.Email,
			Plan:      u.Plan,
		}
	}

	return c.JSON(fiber.Map{"users": dtos})
}

// ---- Bookings (admin) ----

type AdminBookingDTO struct {
	ID            uint    `json:"id"`
	BookingID     string  `json:"bookingId"`
	UserName      string  `json:"userName"`
	UserEmail     string  `json:"userEmail"`
	EventTitle    string  `json:"eventTitle"`
	EventID       string  `json:"eventId"`
	Status        string  `json:"status"`
	RequestNotes  string  `json:"requestNotes"`
	GuestName     string  `json:"guestName"`
	BookingPrice  *float64 `json:"bookingPrice"`
	DateConfirmed string  `json:"dateConfirmed"`
	TimeConfirmed string  `json:"timeConfirmed"`
	CreatedAt     string  `json:"createdAt"`
}

func (h *AdminHandler) ListBookings(c *fiber.Ctx) error {
	status := c.Query("status")
	bookings, err := h.bookingRepo.FindAllAdmin(status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	dtos := make([]AdminBookingDTO, len(bookings))
	for i, b := range bookings {
		dtos[i] = AdminBookingDTO{
			ID:            b.ID,
			BookingID:     b.BookingID,
			UserName:      strings.TrimSpace(b.User.FirstName + " " + b.User.LastName),
			UserEmail:     b.User.Email,
			EventTitle:    b.Event.Title,
			EventID:       b.Event.EventID,
			Status:        b.Status,
			RequestNotes:  b.RequestNotes,
			GuestName:     b.GuestName,
			BookingPrice:  b.BookingPrice,
			DateConfirmed: b.EventDateConfirmed,
			TimeConfirmed: b.EventTimeConfirmed,
			CreatedAt:     b.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	return c.JSON(fiber.Map{"bookings": dtos, "total": len(dtos)})
}

// ---- Suppliers (vendor analytics) ----

type SupplierDTO struct {
	Vendor       string  `json:"vendor"`
	Experiences  int64   `json:"experiences"`
	LiveEvents   int64   `json:"liveEvents"`
	Bookings     int64   `json:"bookings"`
	Attended     int64   `json:"attended"`
	Revenue      float64 `json:"revenue"`
	AvgRating    float64 `json:"avgRating"`
	RatingCount  int64   `json:"ratingCount"`
}

func (h *AdminHandler) ListSuppliers(c *fiber.Ctx) error {
	events, err := h.eventRepo.FindAll("")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Group events by vendor
	type vendorStats struct {
		experiences int64
		liveEvents  int64
		eventIDs    []uint
	}
	vendorMap := make(map[string]*vendorStats)
	for _, e := range events {
		vendor := e.Vendor
		if vendor == "" {
			vendor = "(unknown)"
		}
		vs, ok := vendorMap[vendor]
		if !ok {
			vs = &vendorStats{}
			vendorMap[vendor] = vs
		}
		vs.experiences++
		if e.Status == "Live" {
			vs.liveEvents++
		}
		vs.eventIDs = append(vs.eventIDs, e.ID)
	}

	// Get all bookings to aggregate per vendor
	allBookings, err := h.bookingRepo.FindAllAdmin("")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Build event→vendor lookup
	eventVendor := make(map[uint]string)
	for _, e := range events {
		v := e.Vendor
		if v == "" {
			v = "(unknown)"
		}
		eventVendor[e.ID] = v
	}

	type bookingAgg struct {
		bookings    int64
		attended    int64
		revenue     float64
		ratingSum   int64
		ratingCount int64
	}
	bookingMap := make(map[string]*bookingAgg)
	for _, b := range allBookings {
		vendor, ok := eventVendor[b.EventID]
		if !ok {
			continue
		}
		ba, ok := bookingMap[vendor]
		if !ok {
			ba = &bookingAgg{}
			bookingMap[vendor] = ba
		}
		ba.bookings++
		if b.Status == "Attended" {
			ba.attended++
		}
		if b.BookingPrice != nil {
			ba.revenue += *b.BookingPrice
		}
		if b.FeedbackRating != nil && *b.FeedbackRating > 0 {
			ba.ratingSum += int64(*b.FeedbackRating)
			ba.ratingCount++
		}
	}

	var suppliers []SupplierDTO
	for vendor, vs := range vendorMap {
		s := SupplierDTO{
			Vendor:      vendor,
			Experiences: vs.experiences,
			LiveEvents:  vs.liveEvents,
		}
		if ba, ok := bookingMap[vendor]; ok {
			s.Bookings = ba.bookings
			s.Attended = ba.attended
			s.Revenue = ba.revenue
			if ba.ratingCount > 0 {
				s.AvgRating = float64(ba.ratingSum) / float64(ba.ratingCount)
				s.RatingCount = ba.ratingCount
			}
		}
		suppliers = append(suppliers, s)
	}

	return c.JSON(fiber.Map{"suppliers": suppliers, "total": len(suppliers)})
}

type UpdateBookingInput struct {
	Status        string   `json:"status"`
	BookingPrice  *float64 `json:"bookingPrice"`
	DateConfirmed string   `json:"dateConfirmed"`
	TimeConfirmed string   `json:"timeConfirmed"`
}

func (h *AdminHandler) UpdateBooking(c *fiber.Ctx) error {
	bookingID := c.Params("id")
	booking, err := h.bookingRepo.FindByBookingID(bookingID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "booking not found"})
	}

	var input UpdateBookingInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	oldStatus := booking.Status

	if input.BookingPrice != nil {
		booking.BookingPrice = input.BookingPrice
	}
	if input.DateConfirmed != "" {
		booking.EventDateConfirmed = input.DateConfirmed
	}
	if input.TimeConfirmed != "" {
		booking.EventTimeConfirmed = input.TimeConfirmed
	}
	if input.Status != "" {
		booking.Status = input.Status
	}
	if input.Status == "Confirmed" {
		now := time.Now()
		booking.BookedAt = &now
	}

	if err := h.bookingRepo.Update(booking); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Send email notifications on status change
	if input.Status != "" && input.Status != oldStatus {
		// Load user and event for email
		user, _ := h.userRepo.FindByID(booking.UserID)
		event, _ := h.eventRepo.FindByID(booking.EventID)
		if user != nil && event != nil && user.Email != "" {
			switch input.Status {
			case "Confirmed":
				date := booking.EventDateConfirmed
				if date == "" {
					date = event.DateText
				}
				t := booking.EventTimeConfirmed
				if t == "" {
					t = event.Time
				}
				if err := h.emailService.SendBookingConfirmed(user.Email, user.FirstName, event.Title, date, t); err != nil {
					log.Printf("Failed to send confirmation email: %v", err)
				}
			case "Cancelled":
				if err := h.emailService.SendBookingCancelled(user.Email, user.FirstName, event.Title); err != nil {
					log.Printf("Failed to send cancellation email: %v", err)
				}
			case "Attended":
				// Check and award badges
				if h.badgeService != nil {
					newBadges := h.badgeService.CheckAndAward(booking.UserID)
					if len(newBadges) > 0 {
						log.Printf("User %d earned badges: %v", booking.UserID, newBadges)
					}
				}
			}
		}
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// ---- Analytics ----

func (h *AdminHandler) GetAnalytics(c *fiber.Ctx) error {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)
	thirtyDaysAgo := today.AddDate(0, 0, -30)

	// --- KPIs ---
	var dau, wau, mau int64
	h.db.Model(&models.ActivityLog{}).Where("created_at >= ?", today).Distinct("user_id").Count(&dau)
	h.db.Model(&models.ActivityLog{}).Where("created_at >= ?", weekAgo).Distinct("user_id").Count(&wau)
	h.db.Model(&models.ActivityLog{}).Where("created_at >= ?", monthAgo).Distinct("user_id").Count(&mau)

	var totalUsers int64
	h.db.Model(&models.User{}).Count(&totalUsers)

	var totalBookings int64
	h.db.Model(&models.Booking{}).Count(&totalBookings)

	var totalRevenue struct{ Sum *float64 }
	h.db.Model(&models.Booking{}).Select("SUM(booking_price) as sum").Where("booking_price IS NOT NULL").Scan(&totalRevenue)
	revenue := 0.0
	if totalRevenue.Sum != nil {
		revenue = *totalRevenue.Sum
	}

	var avgRating struct{ Avg *float64 }
	h.db.Model(&models.Booking{}).Select("AVG(feedback_rating) as avg").Where("feedback_rating IS NOT NULL AND feedback_rating > 0").Scan(&avgRating)
	rating := 0.0
	if avgRating.Avg != nil {
		rating = *avgRating.Avg
	}

	var ratingCount int64
	h.db.Model(&models.Booking{}).Where("feedback_rating IS NOT NULL AND feedback_rating > 0").Count(&ratingCount)

	// --- Conversion funnel ---
	funnelActions := []string{"session_start", "page_view_home", "page_view_event", "click_book", "booking_request"}
	funnel := make([]fiber.Map, 0, len(funnelActions)+2)
	for _, action := range funnelActions {
		var count int64
		h.db.Model(&models.ActivityLog{}).Where("action = ?", action).Count(&count)
		funnel = append(funnel, fiber.Map{"step": action, "count": count})
	}
	var confirmedCount, attendedCount int64
	h.db.Model(&models.Booking{}).Where("status = ?", "Confirmed").Count(&confirmedCount)
	h.db.Model(&models.Booking{}).Where("status = ?", "Attended").Count(&attendedCount)
	funnel = append(funnel, fiber.Map{"step": "confirmed", "count": confirmedCount})
	funnel = append(funnel, fiber.Map{"step": "attended", "count": attendedCount})

	// --- Action breakdown ---
	type actionCount struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	var actions []actionCount
	h.db.Model(&models.ActivityLog{}).Select("action, COUNT(*) as count").Group("action").Order("count DESC").Scan(&actions)

	// --- Activity timeline (last 30 days) ---
	type dayCount struct {
		Day   string `json:"day"`
		Count int64  `json:"count"`
	}
	var activityTimeline []dayCount
	h.db.Model(&models.ActivityLog{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as day, COUNT(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("day").Order("day").Scan(&activityTimeline)

	// --- Booking timeline (last 30 days) ---
	var bookingTimeline []dayCount
	h.db.Model(&models.Booking{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as day, COUNT(*) as count").
		Where("created_at >= ?", thirtyDaysAgo).
		Group("day").Order("day").Scan(&bookingTimeline)

	// --- Top events by views ---
	type topEvent struct {
		EventID string `json:"eventId"`
		Views   int64  `json:"views"`
		Title   string `json:"title"`
	}
	type eventViewRow struct {
		Meta  string
		Count int64
	}
	var eventViews []eventViewRow
	h.db.Model(&models.ActivityLog{}).
		Select("meta, COUNT(*) as count").
		Where("action = ?", "page_view_event").
		Group("meta").Order("count DESC").Limit(10).Scan(&eventViews)

	topEvents := make([]topEvent, 0, len(eventViews))
	for _, ev := range eventViews {
		te := topEvent{EventID: ev.Meta, Views: ev.Count}
		event, err := h.eventRepo.FindByEventID(ev.Meta)
		if err == nil && event != nil {
			te.Title = event.Title
		}
		topEvents = append(topEvents, te)
	}

	// --- Booking status breakdown ---
	type statusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var bookingStatuses []statusCount
	h.db.Model(&models.Booking{}).Select("status, COUNT(*) as count").Group("status").Scan(&bookingStatuses)

	return c.JSON(fiber.Map{
		"kpis": fiber.Map{
			"dau":           dau,
			"wau":           wau,
			"mau":           mau,
			"totalUsers":    totalUsers,
			"totalBookings": totalBookings,
			"totalRevenue":  revenue,
			"avgRating":     rating,
			"ratingCount":   ratingCount,
		},
		"funnel":           funnel,
		"actions":          actions,
		"activityTimeline": activityTimeline,
		"bookingTimeline":  bookingTimeline,
		"topEvents":        topEvents,
		"bookingStatuses":  bookingStatuses,
	})
}

// ---- Deep Analytics (per-user, cohorts, activation) ----

func isoWeekMonday(year, week int) time.Time {
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	wd := int(jan4.Weekday())
	if wd == 0 {
		wd = 7
	}
	w1Monday := jan4.AddDate(0, 0, -(wd - 1))
	return w1Monday.AddDate(0, 0, (week-1)*7)
}

func (h *AdminHandler) GetAnalyticsDeep(c *fiber.Ctx) error {
	now := time.Now()

	// --- Per-user summary ---
	var users []models.User
	h.db.Order("created_at ASC").Find(&users)

	type UserSummary struct {
		ID              uint    `json:"id"`
		UserID          string  `json:"userId"`
		FirstName       string  `json:"firstName"`
		Email           string  `json:"email"`
		Plan            string  `json:"plan"`
		SignupDate      string  `json:"signupDate"`
		FirstBooking    *string `json:"firstBooking"`
		DaysToActivate  *int    `json:"daysToActivate"`
		TotalBookings   int64   `json:"totalBookings"`
		LastActive      *string `json:"lastActive"`
		DaysSinceActive int     `json:"daysSinceActive"`
		TotalRevenue    float64 `json:"totalRevenue"`
		SlateSize       int64   `json:"slateSize"`
		EventViews      int64   `json:"eventViews"`
	}

	summaries := make([]UserSummary, 0, len(users))
	for _, u := range users {
		s := UserSummary{
			ID:        u.ID,
			UserID:    u.UserID,
			FirstName: u.FirstName,
			Email:     u.Email,
			Plan:      u.Plan,
		}
		if !u.CreatedAt.IsZero() {
			s.SignupDate = u.CreatedAt.Format("2006-01-02")
		}

		// First booking
		var firstBk struct{ Min *time.Time }
		h.db.Model(&models.Booking{}).Select("MIN(created_at) as min").Where("user_id = ?", u.ID).Scan(&firstBk)
		if firstBk.Min != nil {
			fb := firstBk.Min.Format("2006-01-02")
			s.FirstBooking = &fb
			days := int(firstBk.Min.Sub(u.CreatedAt).Hours() / 24)
			s.DaysToActivate = &days
		}

		// Total bookings
		h.db.Model(&models.Booking{}).Where("user_id = ?", u.ID).Count(&s.TotalBookings)

		// Last active
		var lastAct struct{ Max *time.Time }
		h.db.Model(&models.ActivityLog{}).Select("MAX(created_at) as max").Where("user_id = ?", u.ID).Scan(&lastAct)
		if lastAct.Max != nil {
			la := lastAct.Max.Format("2006-01-02")
			s.LastActive = &la
			s.DaysSinceActive = int(now.Sub(*lastAct.Max).Hours() / 24)
		} else {
			s.DaysSinceActive = -1 // never active
		}

		// Revenue
		var rev struct{ Sum *float64 }
		h.db.Model(&models.Booking{}).Select("COALESCE(SUM(booking_price), 0) as sum").Where("user_id = ? AND booking_price IS NOT NULL", u.ID).Scan(&rev)
		if rev.Sum != nil {
			s.TotalRevenue = *rev.Sum
		}

		// Slate size (events assigned to user)
		h.db.Table("event_users").Where("user_id = ?", u.ID).Count(&s.SlateSize)

		// Event views
		h.db.Model(&models.ActivityLog{}).Where("user_id = ? AND action = ?", u.ID, "page_view_event").Count(&s.EventViews)

		summaries = append(summaries, s)
	}

	// --- Booking distribution ---
	type bookingBucket struct {
		Bookings int64 `json:"bookings"`
		Users    int64 `json:"users"`
	}
	bookingCountMap := map[uint]int64{}
	var allBookings []models.Booking
	h.db.Select("user_id").Find(&allBookings)
	for _, b := range allBookings {
		bookingCountMap[b.UserID]++
	}
	buckets := map[int64]int64{}
	for _, cnt := range bookingCountMap {
		buckets[cnt]++
	}
	zeroBookers := int64(len(users)) - int64(len(bookingCountMap))
	distribution := []bookingBucket{{Bookings: 0, Users: zeroBookers}}
	for b := int64(1); b <= 10; b++ {
		if cnt, ok := buckets[b]; ok {
			distribution = append(distribution, bookingBucket{Bookings: b, Users: cnt})
		}
	}

	// --- Weekly cohort retention ---
	type CohortRow struct {
		Week      string `json:"week"`
		Size      int    `json:"size"`
		Retention []int  `json:"retention"`
	}

	cohortMap := make(map[string][]uint)
	cohortOrder := []string{}
	for _, u := range users {
		year, week := u.CreatedAt.ISOWeek()
		key := fmt.Sprintf("%d-W%02d", year, week)
		if _, exists := cohortMap[key]; !exists {
			cohortOrder = append(cohortOrder, key)
		}
		cohortMap[key] = append(cohortMap[key], u.ID)
	}
	sort.Strings(cohortOrder)

	var cohorts []CohortRow
	for _, weekKey := range cohortOrder {
		userIDs := cohortMap[weekKey]
		var year, week int
		fmt.Sscanf(weekKey, "%d-W%d", &year, &week)
		cohortStart := isoWeekMonday(year, week)
		weeksElapsed := int(now.Sub(cohortStart).Hours() / (24 * 7))
		if weeksElapsed < 0 {
			weeksElapsed = 0
		}
		if weeksElapsed > 26 {
			weeksElapsed = 26 // cap at 6 months
		}

		retention := make([]int, weeksElapsed+1)
		for w := 0; w <= weeksElapsed; w++ {
			weekStart := cohortStart.AddDate(0, 0, w*7)
			weekEnd := weekStart.AddDate(0, 0, 7)
			var activeCount int64
			h.db.Model(&models.ActivityLog{}).
				Where("user_id IN ? AND created_at >= ? AND created_at < ?", userIDs, weekStart, weekEnd).
				Distinct("user_id").Count(&activeCount)
			retention[w] = int(activeCount)
		}

		cohorts = append(cohorts, CohortRow{
			Week:      weekKey,
			Size:      len(userIDs),
			Retention: retention,
		})
	}

	// --- Activation metrics ---
	var activatedUsers int64
	h.db.Model(&models.Booking{}).Distinct("user_id").Count(&activatedUsers)
	activationRate := 0.0
	paidUsers := 0
	for _, u := range users {
		if u.Plan != "" {
			paidUsers++
		}
	}
	if len(users) > 0 {
		activationRate = float64(activatedUsers) / float64(len(users)) * 100
	}

	// Session depth
	var totalSessions, totalEventViews int64
	h.db.Model(&models.ActivityLog{}).Where("action = ?", "session_start").Count(&totalSessions)
	h.db.Model(&models.ActivityLog{}).Where("action = ?", "page_view_event").Count(&totalEventViews)
	sessionDepth := 0.0
	if totalSessions > 0 {
		sessionDepth = float64(totalEventViews) / float64(totalSessions)
	}

	// Dormancy (14+ days inactive)
	var dormantCount int
	for _, s := range summaries {
		if s.LastActive == nil || s.DaysSinceActive >= 14 {
			dormantCount++
		}
	}
	dormancyRate := 0.0
	if len(users) > 0 {
		dormancyRate = float64(dormantCount) / float64(len(users)) * 100
	}

	// Booking rate (bookings per paid user per month)
	var firstBkTime struct{ Min *time.Time }
	h.db.Model(&models.Booking{}).Select("MIN(created_at) as min").Scan(&firstBkTime)
	bookingRate := 0.0
	if paidUsers > 0 && firstBkTime.Min != nil {
		monthsElapsed := now.Sub(*firstBkTime.Min).Hours() / (24 * 30)
		if monthsElapsed < 1 {
			monthsElapsed = 1
		}
		var totalBk int64
		h.db.Model(&models.Booking{}).Count(&totalBk)
		bookingRate = float64(totalBk) / float64(paidUsers) / monthsElapsed
	}

	// Avg time to first booking
	var avgActivation float64
	var activationCount int
	for _, s := range summaries {
		if s.DaysToActivate != nil {
			avgActivation += float64(*s.DaysToActivate)
			activationCount++
		}
	}
	if activationCount > 0 {
		avgActivation /= float64(activationCount)
	}

	return c.JSON(fiber.Map{
		"users":        summaries,
		"distribution": distribution,
		"cohorts":      cohorts,
		"activation": fiber.Map{
			"rate":               activationRate,
			"activatedUsers":     activatedUsers,
			"totalUsers":         len(users),
			"paidUsers":          paidUsers,
			"sessionDepth":       sessionDepth,
			"dormancyRate":       dormancyRate,
			"dormantUsers":       dormantCount,
			"bookingRate":        bookingRate,
			"avgDaysToActivate":  avgActivation,
			"activationCount":    activationCount,
		},
	})
}
