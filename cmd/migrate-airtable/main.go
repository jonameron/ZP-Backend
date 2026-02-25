package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"zeitpass/internal/models"
)

// AirtableClient handles interactions with Airtable API
type AirtableClient struct {
	apiKey     string
	baseID     string
	httpClient *http.Client
}

// AirtableRecord represents a generic Airtable record
type AirtableRecord struct {
	ID          string                 `json:"id"`
	CreatedTime string                 `json:"createdTime"`
	Fields      map[string]interface{} `json:"fields"`
}

// AirtableResponse represents the Airtable API response
type AirtableResponse struct {
	Records []AirtableRecord `json:"records"`
	Offset  string           `json:"offset"`
}

// NewAirtableClient creates a new Airtable client
func NewAirtableClient(apiKey, baseID string) *AirtableClient {
	return &AirtableClient{
		apiKey: apiKey,
		baseID: baseID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchRecords fetches all records from an Airtable table
func (c *AirtableClient) FetchRecords(tableName string) ([]AirtableRecord, error) {
	log.Printf("Fetching records from table: %s", tableName)

	var allRecords []AirtableRecord
	offset := ""

	for {
		// Build URL with proper encoding
		baseURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", c.baseID, url.PathEscape(tableName))
		requestURL := baseURL
		if offset != "" {
			requestURL = fmt.Sprintf("%s?offset=%s", baseURL, url.QueryEscape(offset))
		}

		// Create HTTP request
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Check status code
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("airtable API returned status %d: %s", resp.StatusCode, string(body))
		}

		// Parse response
		var response AirtableResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		allRecords = append(allRecords, response.Records...)
		log.Printf("Fetched %d records (total: %d)", len(response.Records), len(allRecords))

		if response.Offset == "" {
			break
		}
		offset = response.Offset

		// Rate limiting - Airtable allows 5 requests per second
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("Total records fetched from %s: %d", tableName, len(allRecords))
	return allRecords, nil
}

// Mapper functions to convert Airtable records to PostgreSQL models

func mapAirtableUser(record AirtableRecord) (*models.User, error) {
	user := &models.User{}

	// Map fields from Airtable to User model
	if val, ok := record.Fields["User ID"].(string); ok {
		user.UserID = val
	} else {
		return nil, fmt.Errorf("User ID is required")
	}

	if val, ok := record.Fields["First Name"].(string); ok {
		user.FirstName = val
	}

	if val, ok := record.Fields["Last Name"].(string); ok {
		user.LastName = val
	}

	if val, ok := record.Fields["Email"].(string); ok {
		user.Email = val
	}

	if val, ok := record.Fields["Phone"].(string); ok {
		user.Phone = val
	}

	if val, ok := record.Fields["Plan"].(string); ok {
		user.Plan = val
	}

	if val, ok := record.Fields["Plan Duration"].(string); ok {
		user.PlanDuration = val
	}

	if val, ok := record.Fields["Plan Renewal Date"].(string); ok && val != "" {
		t, err := time.Parse("2006-01-02", val)
		if err == nil {
			user.PlanRenewalDate = &t
		}
	}

	if val, ok := record.Fields["Total Events"].(float64); ok {
		user.TotalEvents = int(val)
	}

	if val, ok := record.Fields["Guest Access"].(bool); ok {
		user.GuestAccess = val
	}

	if val, ok := record.Fields["Events Upcoming"].(float64); ok {
		user.EventsUpcoming = int(val)
	}

	if val, ok := record.Fields["Events Attended"].(float64); ok {
		user.EventsAttended = int(val)
	}

	if val, ok := record.Fields["Notes"].(string); ok {
		user.Notes = val
	}

	// Parse created time
	if record.CreatedTime != "" {
		t, err := time.Parse(time.RFC3339, record.CreatedTime)
		if err == nil {
			user.CreatedAt = t
			user.UpdatedAt = t
		}
	}

	return user, nil
}

func mapAirtableEvent(record AirtableRecord) (*models.Event, error) {
	event := &models.Event{}

	// Map fields from Airtable to Event model
	if val, ok := record.Fields["Event ID"].(string); ok {
		event.EventID = val
	} else {
		return nil, fmt.Errorf("Event ID is required")
	}

	if val, ok := record.Fields["Category"].(string); ok {
		event.Category = val
	}

	if val, ok := record.Fields["Tier"].(string); ok {
		event.Tier = val
	}

	if val, ok := record.Fields["Title"].(string); ok {
		event.Title = val
	}

	if val, ok := record.Fields["Vendor"].(string); ok {
		event.Vendor = val
	}

	if val, ok := record.Fields["Image URL"].(string); ok {
		event.ImageURL = val
	}

	if val, ok := record.Fields["Date Text"].(string); ok {
		event.DateText = val
	}

	if val, ok := record.Fields["Day"].(string); ok {
		event.Day = val
	}

	if val, ok := record.Fields["Time"].(string); ok {
		event.Time = val
	}

	if val, ok := record.Fields["Duration"].(string); ok {
		event.Duration = val
	}

	if val, ok := record.Fields["Neighbourhood"].(string); ok {
		event.Neighbourhood = val
	}

	if val, ok := record.Fields["Address"].(string); ok {
		event.Address = val
	}

	if val, ok := record.Fields["Short Desc"].(string); ok {
		event.ShortDesc = val
	}

	if val, ok := record.Fields["Long Desc"].(string); ok {
		event.LongDesc = val
	}

	if val, ok := record.Fields["Highlights"].(string); ok {
		event.Highlights = val
	}

	if val, ok := record.Fields["Additional Information"].(string); ok {
		event.AdditionalInformation = val
	}

	if val, ok := record.Fields["Status"].(string); ok {
		event.Status = val
	}

	if val, ok := record.Fields["Status Qualifier"].(string); ok {
		event.StatusQualifier = val
	}

	if val, ok := record.Fields["Cancellation"].(string); ok {
		event.Cancellation = val
	}

	if val, ok := record.Fields["Language"].(string); ok {
		event.Language = val
	}

	if val, ok := record.Fields["Curation Order"].(float64); ok {
		event.CurationOrder = int(val)
	}

	if val, ok := record.Fields["Retail Price"].(float64); ok {
		event.RetailPrice = &val
	} else if val, ok := record.Fields["Retail Price"].(string); ok && val != "" {
		price, err := strconv.ParseFloat(val, 64)
		if err == nil {
			event.RetailPrice = &price
		}
	}

	// Parse created time
	if record.CreatedTime != "" {
		t, err := time.Parse(time.RFC3339, record.CreatedTime)
		if err == nil {
			event.CreatedAt = t
			event.UpdatedAt = t
		}
	}

	return event, nil
}

func mapAirtableBooking(record AirtableRecord, userMap map[string]uint, eventMap map[string]uint) (*models.Booking, error) {
	booking := &models.Booking{}

	// Map Booking ID
	if val, ok := record.Fields["Booking ID"].(string); ok {
		booking.BookingID = val
	} else {
		// Generate booking ID from record ID if not present
		booking.BookingID = record.ID
	}

	// Map User (from linked record)
	if userLinks, ok := record.Fields["User"].([]interface{}); ok && len(userLinks) > 0 {
		if userID, ok := userLinks[0].(string); ok {
			if dbUserID, exists := userMap[userID]; exists {
				booking.UserID = dbUserID
			} else {
				return nil, fmt.Errorf("user not found for booking: %s", userID)
			}
		}
	} else {
		return nil, fmt.Errorf("User is required for booking")
	}

	// Map Event (from linked record)
	if eventLinks, ok := record.Fields["Event"].([]interface{}); ok && len(eventLinks) > 0 {
		if eventID, ok := eventLinks[0].(string); ok {
			if dbEventID, exists := eventMap[eventID]; exists {
				booking.EventID = dbEventID
			} else {
				return nil, fmt.Errorf("event not found for booking: %s", eventID)
			}
		}
	} else {
		return nil, fmt.Errorf("Event is required for booking")
	}

	if val, ok := record.Fields["Status"].(string); ok {
		booking.Status = val
	}

	if val, ok := record.Fields["Request Notes"].(string); ok {
		booking.RequestNotes = val
	}

	if val, ok := record.Fields["Booked At"].(string); ok && val != "" {
		t, err := time.Parse(time.RFC3339, val)
		if err == nil {
			booking.BookedAt = &t
		}
	}

	if val, ok := record.Fields["Booking Price"].(float64); ok {
		booking.BookingPrice = &val
	} else if val, ok := record.Fields["Booking Price"].(string); ok && val != "" {
		price, err := strconv.ParseFloat(val, 64)
		if err == nil {
			booking.BookingPrice = &price
		}
	}

	if val, ok := record.Fields["Event Date Confirmed"].(string); ok {
		booking.EventDateConfirmed = val
	}

	if val, ok := record.Fields["Event Time Confirmed"].(string); ok {
		booking.EventTimeConfirmed = val
	}

	if val, ok := record.Fields["Feedback Rating"].(float64); ok {
		rating := int(val)
		booking.FeedbackRating = &rating
	}

	if val, ok := record.Fields["Feedback Free Text"].(string); ok {
		booking.FeedbackFreeText = val
	}

	if val, ok := record.Fields["Notes"].(string); ok {
		booking.Notes = val
	}

	// Parse created time
	if record.CreatedTime != "" {
		t, err := time.Parse(time.RFC3339, record.CreatedTime)
		if err == nil {
			booking.CreatedAt = t
			booking.UpdatedAt = t
		}
	}

	return booking, nil
}

func mapAirtableActivityLog(record AirtableRecord, userMap map[string]uint, eventMap map[string]uint) (*models.ActivityLog, error) {
	activity := &models.ActivityLog{}

	// Map User (from linked record)
	if userLinks, ok := record.Fields["User"].([]interface{}); ok && len(userLinks) > 0 {
		if userID, ok := userLinks[0].(string); ok {
			if dbUserID, exists := userMap[userID]; exists {
				activity.UserID = dbUserID
			} else {
				return nil, fmt.Errorf("user not found for activity log: %s", userID)
			}
		}
	} else {
		return nil, fmt.Errorf("User is required for activity log")
	}

	// Map Event (optional, from linked record)
	if eventLinks, ok := record.Fields["Event"].([]interface{}); ok && len(eventLinks) > 0 {
		if eventID, ok := eventLinks[0].(string); ok {
			if dbEventID, exists := eventMap[eventID]; exists {
				activity.EventID = &dbEventID
			}
		}
	}

	if val, ok := record.Fields["Session ID"].(string); ok {
		activity.SessionID = val
	}

	if val, ok := record.Fields["Page URL"].(string); ok {
		activity.PageURL = val
	}

	if val, ok := record.Fields["Action"].(string); ok {
		activity.Action = val
	}

	if val, ok := record.Fields["Metadata"].(string); ok {
		activity.Metadata = val
	} else if val, ok := record.Fields["Metadata"].(map[string]interface{}); ok {
		jsonBytes, err := json.Marshal(val)
		if err == nil {
			activity.Metadata = string(jsonBytes)
		}
	}

	// Parse timestamp
	if val, ok := record.Fields["Timestamp"].(string); ok && val != "" {
		t, err := time.Parse(time.RFC3339, val)
		if err == nil {
			activity.Timestamp = t
		}
	} else if record.CreatedTime != "" {
		t, err := time.Parse(time.RFC3339, record.CreatedTime)
		if err == nil {
			activity.Timestamp = t
		}
	}

	// Parse created time
	if record.CreatedTime != "" {
		t, err := time.Parse(time.RFC3339, record.CreatedTime)
		if err == nil {
			activity.CreatedAt = t
		}
	}

	return activity, nil
}

// Migration orchestrator
type Migrator struct {
	db             *gorm.DB
	airtableClient *AirtableClient
}

func NewMigrator(db *gorm.DB, airtableClient *AirtableClient) *Migrator {
	return &Migrator{
		db:             db,
		airtableClient: airtableClient,
	}
}

func (m *Migrator) MigrateUsers() (map[string]uint, error) {
	log.Println("========================================")
	log.Println("Starting User migration...")
	log.Println("========================================")

	records, err := m.airtableClient.FetchRecords("Users")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	userMap := make(map[string]uint) // Maps Airtable ID to PostgreSQL ID
	successCount := 0
	errorCount := 0

	for i, record := range records {
		user, err := mapAirtableUser(record)
		if err != nil {
			log.Printf("Error mapping user %d (%s): %v", i+1, record.ID, err)
			errorCount++
			continue
		}

		// Create user in database
		if err := m.db.Create(user).Error; err != nil {
			log.Printf("Error creating user %s: %v", user.UserID, err)
			errorCount++
			continue
		}

		userMap[record.ID] = user.ID
		successCount++

		if (i+1)%10 == 0 || i+1 == len(records) {
			log.Printf("Progress: %d/%d users migrated", i+1, len(records))
		}
	}

	log.Printf("User migration completed: %d succeeded, %d failed", successCount, errorCount)
	return userMap, nil
}

func (m *Migrator) MigrateEvents() (map[string]uint, error) {
	log.Println("========================================")
	log.Println("Starting Event migration...")
	log.Println("========================================")

	records, err := m.airtableClient.FetchRecords("Events")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	eventMap := make(map[string]uint) // Maps Airtable ID to PostgreSQL ID
	successCount := 0
	errorCount := 0

	for i, record := range records {
		event, err := mapAirtableEvent(record)
		if err != nil {
			log.Printf("Error mapping event %d (%s): %v", i+1, record.ID, err)
			errorCount++
			continue
		}

		// Create event in database
		if err := m.db.Create(event).Error; err != nil {
			log.Printf("Error creating event %s: %v", event.EventID, err)
			errorCount++
			continue
		}

		eventMap[record.ID] = event.ID
		successCount++

		if (i+1)%10 == 0 || i+1 == len(records) {
			log.Printf("Progress: %d/%d events migrated", i+1, len(records))
		}
	}

	log.Printf("Event migration completed: %d succeeded, %d failed", successCount, errorCount)
	return eventMap, nil
}

func (m *Migrator) MigrateUserEventRelationships(userMap map[string]uint, eventMap map[string]uint) error {
	log.Println("========================================")
	log.Println("Starting User-Event relationship migration...")
	log.Println("========================================")

	// Fetch Users table again to get event assignments
	records, err := m.airtableClient.FetchRecords("Users")
	if err != nil {
		return fmt.Errorf("failed to fetch users for relationships: %w", err)
	}

	successCount := 0
	errorCount := 0

	for i, record := range records {
		airtableUserID := record.ID
		dbUserID, exists := userMap[airtableUserID]
		if !exists {
			continue
		}

		// Get assigned events from Airtable
		assignedEvents, ok := record.Fields["Assigned Events"].([]interface{})
		if !ok || len(assignedEvents) == 0 {
			continue
		}

		// Convert to event IDs
		var eventIDs []uint
		for _, airtableEventID := range assignedEvents {
			if eventIDStr, ok := airtableEventID.(string); ok {
				if dbEventID, exists := eventMap[eventIDStr]; exists {
					eventIDs = append(eventIDs, dbEventID)
				}
			}
		}

		if len(eventIDs) == 0 {
			continue
		}

		// Load user and associate events
		var user models.User
		if err := m.db.First(&user, dbUserID).Error; err != nil {
			log.Printf("Error loading user %d: %v", dbUserID, err)
			errorCount++
			continue
		}

		// Load events
		var events []models.Event
		if err := m.db.Find(&events, eventIDs).Error; err != nil {
			log.Printf("Error loading events for user %d: %v", dbUserID, err)
			errorCount++
			continue
		}

		// Associate events with user
		if err := m.db.Model(&user).Association("AssignedEvents").Replace(&events); err != nil {
			log.Printf("Error associating events for user %d: %v", dbUserID, err)
			errorCount++
			continue
		}

		successCount++

		if (i+1)%10 == 0 || i+1 == len(records) {
			log.Printf("Progress: %d/%d user-event relationships processed", i+1, len(records))
		}
	}

	log.Printf("User-Event relationship migration completed: %d succeeded, %d failed", successCount, errorCount)
	return nil
}

func (m *Migrator) MigrateBookings(userMap map[string]uint, eventMap map[string]uint) error {
	log.Println("========================================")
	log.Println("Starting Booking migration...")
	log.Println("========================================")

	records, err := m.airtableClient.FetchRecords("Bookings")
	if err != nil {
		return fmt.Errorf("failed to fetch bookings: %w", err)
	}

	successCount := 0
	errorCount := 0

	for i, record := range records {
		booking, err := mapAirtableBooking(record, userMap, eventMap)
		if err != nil {
			log.Printf("Error mapping booking %d (%s): %v", i+1, record.ID, err)
			errorCount++
			continue
		}

		// Create booking in database
		if err := m.db.Create(booking).Error; err != nil {
			log.Printf("Error creating booking %s: %v", booking.BookingID, err)
			errorCount++
			continue
		}

		successCount++

		if (i+1)%10 == 0 || i+1 == len(records) {
			log.Printf("Progress: %d/%d bookings migrated", i+1, len(records))
		}
	}

	log.Printf("Booking migration completed: %d succeeded, %d failed", successCount, errorCount)
	return nil
}

func (m *Migrator) MigrateActivityLogs(userMap map[string]uint, eventMap map[string]uint) error {
	log.Println("========================================")
	log.Println("Starting ActivityLog migration...")
	log.Println("========================================")

	records, err := m.airtableClient.FetchRecords("ActivityLog")
	if err != nil {
		return fmt.Errorf("failed to fetch activity logs: %w", err)
	}

	successCount := 0
	errorCount := 0

	for i, record := range records {
		activity, err := mapAirtableActivityLog(record, userMap, eventMap)
		if err != nil {
			log.Printf("Error mapping activity log %d (%s): %v", i+1, record.ID, err)
			errorCount++
			continue
		}

		// Create activity log in database
		if err := m.db.Create(activity).Error; err != nil {
			log.Printf("Error creating activity log: %v", err)
			errorCount++
			continue
		}

		successCount++

		if (i+1)%50 == 0 || i+1 == len(records) {
			log.Printf("Progress: %d/%d activity logs migrated", i+1, len(records))
		}
	}

	log.Printf("ActivityLog migration completed: %d succeeded, %d failed", successCount, errorCount)
	return nil
}

func (m *Migrator) RunMigration() error {
	log.Println("========================================")
	log.Println("Starting Airtable to PostgreSQL Migration")
	log.Println("========================================")
	startTime := time.Now()

	// Step 1: Migrate Users
	userMap, err := m.MigrateUsers()
	if err != nil {
		return fmt.Errorf("user migration failed: %w", err)
	}

	// Step 2: Migrate Events
	eventMap, err := m.MigrateEvents()
	if err != nil {
		return fmt.Errorf("event migration failed: %w", err)
	}

	// Step 3: Migrate User-Event relationships
	if err := m.MigrateUserEventRelationships(userMap, eventMap); err != nil {
		return fmt.Errorf("user-event relationship migration failed: %w", err)
	}

	// Step 4: Migrate Bookings
	if err := m.MigrateBookings(userMap, eventMap); err != nil {
		return fmt.Errorf("booking migration failed: %w", err)
	}

	// Step 5: Migrate Activity Logs
	if err := m.MigrateActivityLogs(userMap, eventMap); err != nil {
		return fmt.Errorf("activity log migration failed: %w", err)
	}

	duration := time.Since(startTime)
	log.Println("========================================")
	log.Printf("Migration completed successfully in %v", duration)
	log.Println("========================================")

	return nil
}

func initDatabase(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate schemas
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Booking{},
		&models.ActivityLog{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return db, nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get environment variables
	airtableAPIKey := os.Getenv("AIRTABLE_API_KEY")
	airtableBaseID := os.Getenv("AIRTABLE_BASE_ID")
	databaseURL := os.Getenv("DATABASE_URL")

	// Validate environment variables
	var missingVars []string
	if airtableAPIKey == "" {
		missingVars = append(missingVars, "AIRTABLE_API_KEY")
	}
	if airtableBaseID == "" {
		missingVars = append(missingVars, "AIRTABLE_BASE_ID")
	}
	if databaseURL == "" {
		missingVars = append(missingVars, "DATABASE_URL")
	}

	if len(missingVars) > 0 {
		log.Fatalf("Missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	// Initialize database
	db, err := initDatabase(databaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Airtable client
	airtableClient := NewAirtableClient(airtableAPIKey, airtableBaseID)

	// Create migrator and run migration
	migrator := NewMigrator(db, airtableClient)

	// Confirm before starting
	log.Println("This will migrate data from Airtable to PostgreSQL.")
	log.Println("Press Ctrl+C to cancel or wait 5 seconds to continue...")
	time.Sleep(5 * time.Second)

	if err := migrator.RunMigration(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
