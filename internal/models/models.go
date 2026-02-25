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
	Notes           string     `gorm:"type:text" json:"notes"`
	AssignedEvents  []Event    `gorm:"many2many:event_users;" json:"-"`
	Bookings        []Booking  `json:"-"`
	ActivityLogs    []ActivityLog `json:"-"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
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
	AssignedUsers         []User    `gorm:"many2many:event_users;" json:"-"`
	Bookings              []Booking `json:"-"`
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
	FeedbackRating     *int       `json:"feedbackRating"`
	FeedbackFreeText   string     `gorm:"type:text" json:"feedbackFreeText"`
	Notes              string     `gorm:"type:text" json:"notes"`
	User               User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	Event              Event      `gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE" json:"event"`
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
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}
