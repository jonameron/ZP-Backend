package models

import (
	"time"
)

// User represents a ZeitPass user
type User struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          string     `gorm:"uniqueIndex;not null;size:50" json:"userId"`
	FirstName       string     `gorm:"size:100" json:"firstName"`
	LastName        string     `gorm:"size:100" json:"lastName"`
	Email           string     `gorm:"uniqueIndex;size:255" json:"email"`
	Phone           string     `gorm:"size:50" json:"phone"`
	Plan            string     `gorm:"size:50" json:"plan"`
	PlanDuration    string     `gorm:"size:50" json:"planDuration"`
	PlanRenewalDate *time.Time `json:"planRenewalDate"`
	TotalEvents     int        `gorm:"default:0" json:"totalEvents"`
	GuestAccess     bool       `gorm:"default:false" json:"guestAccess"`
	EventsUpcoming  int        `gorm:"default:0" json:"eventsUpcoming"`
	EventsAttended  int        `gorm:"default:0" json:"eventsAttended"`
	ProfileVisibility string    `gorm:"size:20;default:'private'" json:"profileVisibility"` // private, friends, public
	Notes             string    `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// Connection represents a social connection between users
type Connection struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserRef         uint      `gorm:"column:user_ref;not null;uniqueIndex:idx_connection_pair" json:"userRef"`
	ConnectedRef    uint      `gorm:"column:connected_ref;not null;uniqueIndex:idx_connection_pair" json:"connectedRef"`
	Type            string    `gorm:"size:20;not null" json:"type"`   // friend, partner
	Status          string    `gorm:"size:20;not null" json:"status"` // pending, accepted, blocked
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Event represents a curated event
type Event struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	EventID               string    `gorm:"uniqueIndex;not null;size:50" json:"eventId"`
	Category              string    `gorm:"size:50" json:"category"`
	Tier                  string    `gorm:"size:50" json:"tier"`
	Title                 string    `gorm:"size:255" json:"title"`
	Vendor                string    `gorm:"size:255" json:"vendor"`
	ImageURL              string    `gorm:"size:500" json:"imageUrl"`
	DateText              string    `gorm:"size:100" json:"dateText"`
	Day                   string    `gorm:"size:100" json:"day"`
	Time                  string    `gorm:"size:100" json:"time"`
	Duration              string    `gorm:"size:100" json:"duration"`
	Neighbourhood         string    `gorm:"size:100" json:"neighbourhood"`
	Address               string    `gorm:"size:500" json:"address"`
	ShortDesc             string    `gorm:"size:500" json:"shortDesc"`
	LongDesc              string    `gorm:"type:text" json:"longDesc"`
	Highlights            string    `gorm:"type:text" json:"highlights"`
	AdditionalInformation string    `gorm:"size:500" json:"additionalInformation"`
	Status                string    `gorm:"size:50;default:'Draft'" json:"status"`
	StatusQualifier       string    `gorm:"size:50" json:"statusQualifier"`
	Cancellation          string    `gorm:"size:255" json:"cancellation"`
	Language              string    `gorm:"size:50" json:"language"`
	CurationOrder         int       `gorm:"default:0" json:"curationOrder"`
	RetailPrice           *float64  `json:"retailPrice"`
	MysteryEligible       bool      `gorm:"default:false" json:"mysteryEligible"`
	SourceURL             string    `gorm:"size:500" json:"sourceUrl"`
	SourceName            string    `gorm:"size:50" json:"sourceName"`
	ExternalID            string    `gorm:"size:255;index" json:"externalId"`
	VibeTags              string    `gorm:"size:500" json:"vibeTags"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

// Booking represents an event booking
type Booking struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	BookingID          string     `gorm:"uniqueIndex;size:100" json:"bookingId"`
	UserID             uint       `gorm:"not null;index" json:"userId"`
	EventID            uint       `gorm:"not null;index" json:"eventId"`
	Status             string     `gorm:"size:50;default:'Requested'" json:"status"`
	RequestNotes       string     `gorm:"type:text" json:"requestNotes"`
	BookedAt           *time.Time `json:"bookedAt"`
	BookingPrice       *float64   `json:"bookingPrice"`
	EventDateConfirmed string     `gorm:"size:100" json:"eventDateConfirmed"`
	EventTimeConfirmed string     `gorm:"size:100" json:"eventTimeConfirmed"`
	GuestName          string     `gorm:"size:255" json:"guestName"`
	GuestEmail         string     `gorm:"size:255" json:"guestEmail"`
	IsMystery          bool       `gorm:"default:false" json:"isMystery"`
	MysteryRevealed    bool       `gorm:"default:false" json:"mysteryRevealed"`
	FeedbackRating     *int       `json:"feedbackRating"`
	FeedbackFreeText   string     `gorm:"type:text" json:"feedbackFreeText"`
	Notes              string     `gorm:"type:text" json:"notes"`
	User               User       `gorm:"-" json:"user"`
	Event              Event      `gorm:"-" json:"event"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// ActivityLog represents user activity tracking
type ActivityLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"size:255;index" json:"sessionId"`
	UserID    uint      `gorm:"index" json:"userId"`
	EventID   *uint     `gorm:"index" json:"eventId"`
	PageURL   string    `gorm:"size:500" json:"pageUrl"`
	Action    string    `gorm:"size:100" json:"action"`
	Metadata  string    `gorm:"type:jsonb" json:"metadata"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	User      User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

// EventUser is the junction table for assigning events to users (curation)
type EventUser struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	EventID uint `gorm:"not null;uniqueIndex:idx_event_user" json:"eventId"`
	UserID  uint `gorm:"not null;uniqueIndex:idx_event_user" json:"userId"`
	Event   Event `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	User    User  `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

// MagicLinkToken represents a one-time-use magic link for passwordless auth
type MagicLinkToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	TokenHash string     `gorm:"uniqueIndex;not null;size:64" json:"-"`
	UserID    uint       `gorm:"not null;index" json:"userId"`
	ExpiresAt time.Time  `gorm:"not null" json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt"`
	User      User       `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time  `json:"createdAt"`
}

// Reaction represents a thumbs up/down reaction on an event
type Reaction struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;uniqueIndex:idx_user_event_reaction" json:"userId"`
	EventID      uint      `gorm:"not null;uniqueIndex:idx_user_event_reaction" json:"eventId"`
	ReactionType string    `gorm:"not null;size:10" json:"reactionType"` // "up" or "down"
	User         User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Event        Event     `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Invite represents a guest invite link for an event booking
type Invite struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Token       string    `gorm:"uniqueIndex;not null;size:64" json:"token"`
	BookingRef  uint      `gorm:"not null;index" json:"bookingRef"`
	EventRef    uint      `gorm:"not null;index" json:"eventRef"`
	HostRef     uint      `gorm:"not null;index" json:"hostRef"`
	GuestName   string    `gorm:"size:255" json:"guestName"`
	GuestEmail  string    `gorm:"size:255" json:"guestEmail"`
	Status      string    `gorm:"size:50;default:'Pending'" json:"status"` // Pending, Accepted
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// UserVibe represents a weekly mood/vibe poll response
type UserVibe struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	Vibe      string    `gorm:"not null;size:50" json:"vibe"`
	WeekStart string    `gorm:"not null;size:10;index" json:"weekStart"` // YYYY-MM-DD of Monday
	User      User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserBadge represents a badge earned by a user
type UserBadge struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_badge" json:"userId"`
	BadgeKey  string    `gorm:"not null;size:50;uniqueIndex:idx_user_badge" json:"badgeKey"`
	EarnedAt  time.Time `gorm:"not null" json:"earnedAt"`
	User      User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

// NotificationLog tracks sent notifications to avoid duplicates
type NotificationLog struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	BookingID        uint      `gorm:"not null;index" json:"bookingId"`
	NotificationType string    `gorm:"not null;size:50" json:"notificationType"` // reminder_7d, reminder_1d, feedback_request
	SentAt           time.Time `gorm:"not null" json:"sentAt"`
	CreatedAt        time.Time `json:"createdAt"`
}

// OnboardingSubmission represents a quiz submission from the onboarding flow
type OnboardingSubmission struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	SessionID             string    `gorm:"uniqueIndex;not null;size:255" json:"sessionId"`
	Timestamp             time.Time `gorm:"index" json:"timestamp"`
	Device                string    `gorm:"size:50" json:"device"`
	SourceURL             string    `gorm:"size:500" json:"sourceUrl"`
	Name                  string    `gorm:"size:255" json:"name"`
	Email                 string    `gorm:"size:255;index" json:"email"`
	NameParam             string    `gorm:"size:255" json:"nameParam"`
	ContactPreference     string    `gorm:"size:50" json:"contactPreference"`
	WeekdayVibe           string    `gorm:"size:50" json:"weekdayVibe"`
	WeekendFlavor         string    `gorm:"size:50" json:"weekendFlavor"`
	SocialStyle           string    `gorm:"size:50" json:"socialStyle"`
	EnergyLevel           string    `gorm:"size:50" json:"energyLevel"`
	Companionship         string    `gorm:"size:50" json:"companionship"`
	CulturalPersonality   string    `gorm:"size:50" json:"culturalPersonality"`
	CreativeEngagement    string    `gorm:"size:50" json:"creativeEngagement"`
	Motto                 string    `gorm:"size:50" json:"motto"`
	Language              string    `gorm:"size:50" json:"language"`
	PilotValueExpectation string    `gorm:"size:50" json:"pilotValueExpectation"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}
