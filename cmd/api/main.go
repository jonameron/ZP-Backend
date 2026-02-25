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

	// Initialize services
	eventService := services.NewEventService(eventRepo, userRepo)
	bookingService := services.NewBookingService(bookingRepo, eventRepo, userRepo)
	profileService := services.NewProfileService(userRepo, bookingRepo)
	activityService := services.NewActivityService(activityRepo, userRepo)
	onboardingService := services.NewOnboardingService(onboardingRepo)

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

	// Curation & Events
	curationHandler := handlers.NewCurationHandler(eventService)
	api.Get("/curation", curationHandler.GetCuratedEvents)
	api.Get("/events/:eventId", curationHandler.GetEvent)

	// Bookings
	bookingHandler := handlers.NewBookingHandler(bookingService)
	api.Post("/bookings", bookingHandler.CreateBooking)
	api.Post("/bookings/feedback", bookingHandler.SubmitFeedback)

	// Profile
	profileHandler := handlers.NewProfileHandler(profileService)
	api.Get("/profile/:uid", profileHandler.GetProfile)

	// Activity logging
	activityHandler := handlers.NewActivityHandler(activityService)
	api.Post("/activity", activityHandler.LogActivity)

	// Onboarding
	onboardingHandler := handlers.NewOnboardingHandler(onboardingService)
	api.Post("/onboarding", onboardingHandler.SubmitOnboarding)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 ZeitPass API starting on port %s", port)
	log.Printf("📝 Environment: %s", os.Getenv("ENVIRONMENT"))
	log.Printf("🌐 Frontend URL: %s", frontendURL)

	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
