package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"zeitpass/internal/models"
)

func InitDatabase() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// Configure logger based on environment
	var gormLogger logger.Interface
	if os.Getenv("ENVIRONMENT") == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate schemas in correct order
	// First create tables without foreign keys
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate users: %w", err)
	}
	err = db.AutoMigrate(&models.Event{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate events: %w", err)
	}
	err = db.AutoMigrate(&models.OnboardingSubmission{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate onboarding: %w", err)
	}
	// Then create tables with foreign keys
	err = db.AutoMigrate(&models.Booking{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate bookings: %w", err)
	}
	err = db.AutoMigrate(&models.ActivityLog{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	err = db.AutoMigrate(&models.MagicLinkToken{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate magic_link_tokens: %w", err)
	}
	err = db.AutoMigrate(&models.EventUser{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate event_users: %w", err)
	}
	err = db.AutoMigrate(&models.Reaction{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate reactions: %w", err)
	}
	err = db.AutoMigrate(&models.UserVibe{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate user_vibes: %w", err)
	}
	err = db.AutoMigrate(&models.Invite{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate invites: %w", err)
	}
	err = db.AutoMigrate(&models.Connection{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate connections: %w", err)
	}
	err = db.AutoMigrate(&models.UserBadge{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate user_badges: %w", err)
	}
	err = db.AutoMigrate(&models.NotificationLog{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate notification_logs: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return db, nil
}
