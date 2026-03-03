package main

import (
	"log"
	"math/rand"
	"sort"

	"github.com/joho/godotenv"

	"zeitpass/internal/config"
	"zeitpass/internal/models"
	"zeitpass/internal/repository"
)

// quizCategoryMap maps onboarding quiz answers to event category affinities.
// Keys are quiz field values, values are slices of categories that align.
var quizCategoryMap = map[string][]string{
	// WeekdayVibe
	"afterwork":  {"Food & Drink", "Social"},
	"cultural":   {"Theatre & Arts", "Culture"},
	"active":     {"Active & Sports", "Wellness"},
	"relaxed":    {"Wellness", "Food & Drink"},
	// WeekendFlavor
	"adventure":  {"Active & Sports", "Outdoor"},
	"creative":   {"Creative Workshop", "Theatre & Arts"},
	"social":     {"Social", "Food & Drink"},
	"chill":      {"Wellness", "Culture"},
	// EnergyLevel
	"high":       {"Active & Sports", "Outdoor"},
	"medium":     {"Creative Workshop", "Social"},
	"low":        {"Wellness", "Culture", "Food & Drink"},
	// CulturalPersonality
	"explorer":   {"Theatre & Arts", "Culture", "Creative Workshop"},
	"mainstream": {"Social", "Food & Drink"},
	"niche":      {"Creative Workshop", "Culture"},
	// CreativeEngagement
	"hands-on":   {"Creative Workshop"},
	"spectator":  {"Theatre & Arts", "Culture"},
	"both":       {"Creative Workshop", "Theatre & Arts"},
}

// tierAccess defines which tiers each plan can access.
var tierAccess = map[string]map[string]bool{
	"Solo": {
		"Essence": true,
	},
	"Duo": {
		"Essence": true,
		"Elevate": true,
	},
	"Premium": {
		"Essence": true,
		"Elevate": true,
		"Indulge": true,
	},
	// Default/empty plan gets Essence only
	"": {
		"Essence": true,
	},
}

type scoredEvent struct {
	event models.Event
	score float64
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := config.InitDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	onboardingRepo := repository.NewOnboardingRepository(db)
	vibeRepo := repository.NewVibeRepository(db)

	// Load all live events
	events, err := eventRepo.FindAll("Live")
	if err != nil {
		log.Fatal("Failed to load events:", err)
	}
	log.Printf("Loaded %d live events", len(events))

	if len(events) == 0 {
		log.Println("No live events to curate. Exiting.")
		return
	}

	// Load all users
	users, err := userRepo.FindAll()
	if err != nil {
		log.Fatal("Failed to load users:", err)
	}
	log.Printf("Loaded %d users", len(users))

	totalAssignments := 0

	for _, user := range users {
		// Get user signals
		reactions, _ := reactionRepo.FindByUser(user.ID)
		bookings, _ := bookingRepo.FindByUserID(user.ID)
		quiz, _ := onboardingRepo.FindByEmail(user.Email)
		vibe, _ := vibeRepo.FindLatest(user.ID)

		// Build reaction map (eventID → "up"/"down")
		reactionMap := make(map[uint]string)
		for _, r := range reactions {
			reactionMap[r.EventID] = r.ReactionType
		}

		// Build booked/attended event set
		bookedEvents := make(map[uint]bool)
		attendedCategories := make(map[string]int)
		for _, b := range bookings {
			if b.Status == "Confirmed" || b.Status == "Requested" || b.Status == "Attended" {
				bookedEvents[b.EventID] = true
			}
			if b.Status == "Attended" {
				attendedCategories[b.Event.Category]++
			}
		}

		// Build category affinity from quiz
		categoryAffinity := make(map[string]float64)
		if quiz != nil {
			for _, field := range []string{
				quiz.WeekdayVibe, quiz.WeekendFlavor, quiz.EnergyLevel,
				quiz.CulturalPersonality, quiz.CreativeEngagement,
			} {
				if cats, ok := quizCategoryMap[field]; ok {
					for _, cat := range cats {
						categoryAffinity[cat] += 1.0
					}
				}
			}
		}

		// Boost affinity from attended events
		for cat, count := range attendedCategories {
			categoryAffinity[cat] += float64(count) * 0.5
		}

		// Boost from current vibe
		if vibe != nil {
			vibeCategories := map[string][]string{
				"chill":    {"Wellness", "Culture", "Food & Drink"},
				"active":   {"Active & Sports", "Outdoor"},
				"creative": {"Creative Workshop", "Theatre & Arts"},
				"treat":    {"Food & Drink", "Wellness"},
				"surprise": {}, // boosts exploration
			}
			if cats, ok := vibeCategories[vibe.Vibe]; ok {
				for _, cat := range cats {
					categoryAffinity[cat] += 2.0
				}
			}
		}

		// Determine tier access
		access := tierAccess[user.Plan]
		if access == nil {
			access = tierAccess[""]
		}

		// Score each event
		var scored []scoredEvent
		for _, event := range events {
			// Hard constraint: tier access
			if !access[event.Tier] {
				continue
			}

			// Hard constraint: skip already booked/attended
			if bookedEvents[event.ID] {
				continue
			}

			// Hard constraint: skip thumbs-down
			if reactionMap[event.ID] == "down" {
				continue
			}

			var score float64

			// Tier match (higher tiers score higher as treats)
			switch event.Tier {
			case "Essence":
				score += 5
			case "Elevate":
				score += 8
			case "Indulge":
				score += 10
			}

			// Category affinity
			if aff, ok := categoryAffinity[event.Category]; ok {
				score += aff * 5
			}

			// Thumbs up boost
			if reactionMap[event.ID] == "up" {
				score += 8
			}

			// Exploration bonus (random element for serendipity)
			score += rand.Float64() * 3

			scored = append(scored, scoredEvent{event: event, score: score})
		}

		// Sort by score descending
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})

		// Take top N (max 12 events per user for the slate)
		maxSlate := 12
		if len(scored) < maxSlate {
			maxSlate = len(scored)
		}
		topEvents := scored[:maxSlate]

		// Ensure category diversity: max 3 from same category
		var diverse []scoredEvent
		catCount := make(map[string]int)
		for _, se := range topEvents {
			if catCount[se.event.Category] < 3 {
				diverse = append(diverse, se)
				catCount[se.event.Category]++
			}
		}
		// Fill remaining slots if diversity filter removed some
		if len(diverse) < maxSlate {
			for _, se := range scored[maxSlate:] {
				if len(diverse) >= maxSlate {
					break
				}
				if catCount[se.event.Category] < 3 {
					diverse = append(diverse, se)
					catCount[se.event.Category]++
				}
			}
		}

		// Clear existing assignments for this user
		db.Where("user_id = ?", user.ID).Delete(&models.EventUser{})

		// Write new assignments with curation_order
		for _, se := range diverse {
			eu := models.EventUser{
				EventID: se.event.ID,
				UserID:  user.ID,
			}
			if err := db.Create(&eu).Error; err != nil {
				log.Printf("Failed to assign event %d to user %d: %v", se.event.ID, user.ID, err)
				continue
			}

			totalAssignments++
		}

		log.Printf("User %s (%s): assigned %d events (from %d eligible)",
			user.UserID, user.Email, len(diverse), len(scored))
	}

	log.Printf("CRE complete. Total assignments: %d", totalAssignments)
}
