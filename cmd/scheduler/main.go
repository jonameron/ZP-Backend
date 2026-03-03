package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"

	"zeitpass/internal/config"
	"zeitpass/internal/repository"
	"zeitpass/internal/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := config.InitDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	bookingRepo := repository.NewBookingRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	emailService := services.NewEmailService()

	today := time.Now().Truncate(24 * time.Hour)
	log.Printf("Scheduler running for date: %s", today.Format("2006-01-02"))

	bookings, err := bookingRepo.FindConfirmedWithDates()
	if err != nil {
		log.Fatal("Failed to query bookings:", err)
	}

	log.Printf("Found %d bookings with confirmed dates", len(bookings))

	var sent int
	for _, b := range bookings {
		eventDate, err := time.Parse("2006-01-02", b.EventDateConfirmed)
		if err != nil {
			log.Printf("Skipping booking %s: cannot parse date '%s'", b.BookingID, b.EventDateConfirmed)
			continue
		}

		daysUntil := int(eventDate.Sub(today).Hours() / 24)
		user := b.User
		event := b.Event

		if user.Email == "" {
			continue
		}

		timeStr := b.EventTimeConfirmed
		if timeStr == "" {
			timeStr = event.Time
		}

		// 7-day reminder
		if daysUntil == 7 && b.Status == "Confirmed" {
			if !notifRepo.HasBeenSent(b.ID, "reminder_7d") {
				if err := emailService.SendReminder(user.Email, user.FirstName, event.Title, b.EventDateConfirmed, timeStr, event.Neighbourhood, 7); err != nil {
					log.Printf("Failed to send 7-day reminder for %s: %v", b.BookingID, err)
				} else {
					notifRepo.MarkSent(b.ID, "reminder_7d")
					sent++
					log.Printf("Sent 7-day reminder: %s → %s", b.BookingID, user.Email)
				}
			}
		}

		// 1-day reminder
		if daysUntil == 1 && b.Status == "Confirmed" {
			if !notifRepo.HasBeenSent(b.ID, "reminder_1d") {
				if err := emailService.SendReminder(user.Email, user.FirstName, event.Title, b.EventDateConfirmed, timeStr, event.Neighbourhood, 1); err != nil {
					log.Printf("Failed to send 1-day reminder for %s: %v", b.BookingID, err)
				} else {
					notifRepo.MarkSent(b.ID, "reminder_1d")
					sent++
					log.Printf("Sent 1-day reminder: %s → %s", b.BookingID, user.Email)
				}
			}
		}

		// Mystery reveal: 1 day before event, reveal mystery bookings
		if daysUntil == 1 && b.Status == "Confirmed" && b.IsMystery && !b.MysteryRevealed {
			b.MysteryRevealed = true
			if err := bookingRepo.Update(&b); err != nil {
				log.Printf("Failed to reveal mystery booking %s: %v", b.BookingID, err)
			} else {
				// Send reveal email
				if !notifRepo.HasBeenSent(b.ID, "mystery_reveal") {
					if err := emailService.SendMysteryReveal(user.Email, user.FirstName, event.Title, b.EventDateConfirmed, timeStr, event.Neighbourhood); err != nil {
						log.Printf("Failed to send mystery reveal for %s: %v", b.BookingID, err)
					} else {
						notifRepo.MarkSent(b.ID, "mystery_reveal")
						sent++
						log.Printf("Mystery revealed: %s → %s (%s)", b.BookingID, user.Email, event.Title)
					}
				}
			}
		}

		// Post-event feedback request (day after event, for Attended bookings)
		if daysUntil == -1 && b.Status == "Attended" && b.FeedbackRating == nil {
			if !notifRepo.HasBeenSent(b.ID, "feedback_request") {
				if err := emailService.SendFeedbackRequest(user.Email, user.FirstName, event.Title); err != nil {
					log.Printf("Failed to send feedback request for %s: %v", b.BookingID, err)
				} else {
					notifRepo.MarkSent(b.ID, "feedback_request")
					sent++
					log.Printf("Sent feedback request: %s → %s", b.BookingID, user.Email)
				}
			}
		}
	}

	log.Printf("Scheduler complete. Sent %d notifications.", sent)
}
