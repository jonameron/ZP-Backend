package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"zeitpass/internal/config"
	"zeitpass/internal/handlers"
	"zeitpass/internal/repository"
	"zeitpass/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := config.InitDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	onboardingRepo := repository.NewOnboardingRepository(db)
	magicLinkRepo := repository.NewMagicLinkRepository(db)

	// Initialize services
	eventService := services.NewEventService(eventRepo, userRepo)
	bookingService := services.NewBookingService(bookingRepo, eventRepo, userRepo)
	profileService := services.NewProfileService(db, userRepo, bookingRepo, activityRepo)
	activityService := services.NewActivityService(activityRepo, userRepo)
	onboardingService := services.NewOnboardingService(onboardingRepo)
	emailService := services.NewEmailService()
	authService := services.NewAuthService(userRepo, magicLinkRepo, emailService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: handlers.CustomErrorHandler,
		AppName:      "ZeitPass API v1.0",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))

	// CORS configuration
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "*" // Allow all in development
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: frontendURL,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "zeitpass-api",
			"version": "1.0.0",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (open)
	authHandler := handlers.NewAuthHandler(authService)
	api.Post("/auth/magic-link", authHandler.RequestMagicLink)
	api.Post("/auth/verify", authHandler.VerifyMagicLink)

	// Open routes
	onboardingHandler := handlers.NewOnboardingHandler(onboardingService)
	api.Post("/onboarding", onboardingHandler.SubmitOnboarding)

	// Protected routes (JWT required)
	protected := api.Group("", handlers.RequireAuth(authService))

	activityHandler := handlers.NewActivityHandler(activityService)
	protected.Post("/activity", activityHandler.LogActivity)

	curationHandler := handlers.NewCurationHandler(eventService)
	protected.Get("/curation", curationHandler.GetCuratedEvents)
	protected.Get("/events/:eventId", curationHandler.GetEvent)

	bookingHandler := handlers.NewBookingHandler(bookingService)
	protected.Post("/bookings", bookingHandler.CreateBooking)
	protected.Put("/bookings/:id/cancel", bookingHandler.CancelBooking)
	protected.Post("/bookings/feedback", bookingHandler.SubmitFeedback)

	vibeRepo := repository.NewVibeRepository(db)
	vibeHandler := handlers.NewVibeHandler(vibeRepo, userRepo)
	protected.Post("/vibe", vibeHandler.SetVibe)
	protected.Get("/vibe", vibeHandler.GetVibe)

	reactionRepo := repository.NewReactionRepository(db)
	reactionHandler := handlers.NewReactionHandler(reactionRepo, eventRepo, userRepo)
	protected.Post("/reactions", reactionHandler.ToggleReaction)
	protected.Get("/reactions", reactionHandler.GetMyReactions)

	profileHandler := handlers.NewProfileHandler(profileService)
	protected.Get("/profile", profileHandler.GetProfile)
	protected.Get("/profile/export", profileHandler.ExportData)
	protected.Delete("/profile", profileHandler.DeleteAccount)

	inviteRepo := repository.NewInviteRepository(db)
	inviteHandler := handlers.NewInviteHandler(inviteRepo, bookingRepo, userRepo, emailService)
	protected.Post("/bookings/:id/invite", inviteHandler.CreateInvite)

	badgeRepo := repository.NewBadgeRepository(db)
	badgeService := services.NewBadgeService(badgeRepo, bookingRepo, inviteRepo)
	badgeHandler := handlers.NewBadgeHandler(badgeService, userRepo)
	protected.Get("/badges", badgeHandler.GetBadges)

	connRepo := repository.NewConnectionRepository(db)
	socialHandler := handlers.NewSocialHandler(connRepo, userRepo, bookingRepo)
	protected.Post("/social/connect", socialHandler.SendRequest)
	protected.Put("/social/connect/:id", socialHandler.RespondToRequest)
	protected.Delete("/social/connect/:id", socialHandler.RemoveConnection)
	protected.Get("/social/connections", socialHandler.GetConnections)
	protected.Get("/social/feed", socialHandler.GetFeed)
	protected.Put("/social/visibility", socialHandler.UpdateVisibility)

	// Public invite routes (no auth)
	api.Get("/invite/:token", inviteHandler.GetInvitePreview)
	api.Post("/invite/:token/rsvp", inviteHandler.RSVP)

	// Admin routes (JWT + admin email check)
	admin := api.Group("/admin", handlers.RequireAuth(authService), handlers.RequireAdmin())
	adminHandler := handlers.NewAdminHandler(db, eventRepo, userRepo, bookingRepo, emailService, badgeService)
	admin.Get("/events", adminHandler.ListEvents)
	admin.Post("/events", adminHandler.CreateEvent)
	admin.Put("/events/:id", adminHandler.UpdateEvent)
	admin.Post("/events/:id/assign", adminHandler.AssignUsers)
	admin.Get("/users", adminHandler.ListUsers)
	admin.Get("/bookings", adminHandler.ListBookings)
	admin.Put("/bookings/:id", adminHandler.UpdateBooking)
	admin.Get("/suppliers", adminHandler.ListSuppliers)
	admin.Get("/analytics", adminHandler.GetAnalytics)
	admin.Get("/analytics/deep", adminHandler.GetAnalyticsDeep)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("ZeitPass API starting on port %s", port)
	log.Printf("Environment: %s", os.Getenv("ENVIRONMENT"))
	log.Printf("Frontend URL: %s", frontendURL)

	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
