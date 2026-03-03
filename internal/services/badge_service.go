package services

import (
	"log"

	"zeitpass/internal/models"
	"zeitpass/internal/repository"
)

// BadgeDef defines a badge with its display info and check logic.
type BadgeDef struct {
	Key   string
	Name  string
	Emoji string
}

var AllBadges = []BadgeDef{
	// Milestone badges
	{Key: "first_timer", Name: "First Timer", Emoji: "🎉"},
	{Key: "regular", Name: "Regular", Emoji: "⭐"},
	{Key: "explorer", Name: "Explorer", Emoji: "🧭"},
	{Key: "connoisseur", Name: "Connoisseur", Emoji: "🏆"},
	{Key: "trailblazer", Name: "Trailblazer", Emoji: "🔥"},
	// Category badges
	{Key: "culture_vulture", Name: "Culture Vulture", Emoji: "🦉"},
	{Key: "action_hero", Name: "The Action Hero", Emoji: "💪"},
	{Key: "culinary_explorer", Name: "Culinary Explorer", Emoji: "🍷"},
	{Key: "creative_soul", Name: "Creative Soul", Emoji: "🎨"},
	{Key: "night_owl", Name: "Night Owl", Emoji: "🌙"},
	// Social badges (invite-based)
	{Key: "social_catalyst", Name: "Social Catalyst", Emoji: "🦋"},
}

type BadgeService struct {
	badgeRepo   *repository.BadgeRepository
	bookingRepo *repository.BookingRepository
	inviteRepo  *repository.InviteRepository
}

func NewBadgeService(br *repository.BadgeRepository, bookR *repository.BookingRepository, ir *repository.InviteRepository) *BadgeService {
	return &BadgeService{badgeRepo: br, bookingRepo: bookR, inviteRepo: ir}
}

// CheckAndAward evaluates all badge criteria for a user and awards any newly earned badges.
// Returns the list of newly awarded badge keys.
func (s *BadgeService) CheckAndAward(userID uint) []string {
	bookings, err := s.bookingRepo.FindByUserID(userID)
	if err != nil {
		log.Printf("Badge check failed for user %d: %v", userID, err)
		return nil
	}

	// Count attended bookings and categories
	var attendedCount int
	categoryCount := make(map[string]int)
	categories := make(map[string]bool)
	eveningCount := 0

	for _, b := range bookings {
		if b.Status != "Attended" {
			continue
		}
		attendedCount++
		cat := b.Event.Category
		categoryCount[cat]++
		categories[cat] = true

		// Check if evening event (time contains 18:, 19:, 20:, 21:, 22:)
		t := b.Event.Time
		if len(t) >= 2 {
			hour := t[:2]
			if hour >= "18" && hour <= "23" {
				eveningCount++
			}
		}
	}

	var newBadges []string

	// Milestone badges
	if attendedCount >= 1 {
		if s.award(userID, "first_timer") {
			newBadges = append(newBadges, "first_timer")
		}
	}
	if attendedCount >= 5 {
		if s.award(userID, "regular") {
			newBadges = append(newBadges, "regular")
		}
	}
	if attendedCount >= 10 {
		if s.award(userID, "explorer") {
			newBadges = append(newBadges, "explorer")
		}
	}
	if attendedCount >= 25 {
		if s.award(userID, "connoisseur") {
			newBadges = append(newBadges, "connoisseur")
		}
	}
	if len(categories) >= 3 {
		if s.award(userID, "trailblazer") {
			newBadges = append(newBadges, "trailblazer")
		}
	}

	// Category badges (3+ in category)
	categoryBadges := map[string]string{
		"Theatre & Arts":    "culture_vulture",
		"Culture":           "culture_vulture",
		"Active & Sports":   "action_hero",
		"Food & Drink":      "culinary_explorer",
		"Creative Workshop": "creative_soul",
	}
	for cat, badge := range categoryBadges {
		if categoryCount[cat] >= 3 {
			if s.award(userID, badge) {
				newBadges = append(newBadges, badge)
			}
		}
	}

	// Night Owl: 3+ evening events
	if eveningCount >= 3 {
		if s.award(userID, "night_owl") {
			newBadges = append(newBadges, "night_owl")
		}
	}

	// Social Catalyst: 3+ accepted invites
	if s.inviteRepo != nil {
		accepted := s.inviteRepo.CountAcceptedByHost(userID)
		if accepted >= 3 {
			if s.award(userID, "social_catalyst") {
				newBadges = append(newBadges, "social_catalyst")
			}
		}
	}

	return newBadges
}

// award returns true if the badge was newly awarded (didn't exist before).
func (s *BadgeService) award(userID uint, badgeKey string) bool {
	if s.badgeRepo.HasBadge(userID, badgeKey) {
		return false
	}
	if err := s.badgeRepo.Award(userID, badgeKey); err != nil {
		log.Printf("Failed to award badge %s to user %d: %v", badgeKey, userID, err)
		return false
	}
	return true
}

// GetUserBadges returns all badges for a user with their display info.
func (s *BadgeService) GetUserBadges(userID uint) []map[string]interface{} {
	earned, _ := s.badgeRepo.FindByUser(userID)

	earnedMap := make(map[string]models.UserBadge)
	for _, b := range earned {
		earnedMap[b.BadgeKey] = b
	}

	var result []map[string]interface{}
	for _, def := range AllBadges {
		entry := map[string]interface{}{
			"key":    def.Key,
			"name":   def.Name,
			"emoji":  def.Emoji,
			"earned": false,
		}
		if ub, ok := earnedMap[def.Key]; ok {
			entry["earned"] = true
			entry["earnedAt"] = ub.EarnedAt
		}
		result = append(result, entry)
	}
	return result
}
