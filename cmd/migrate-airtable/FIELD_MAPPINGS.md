# Field Mappings: Airtable to PostgreSQL

This document shows the exact field mappings from Airtable to PostgreSQL models.

## Users Table

| Airtable Field Name | PostgreSQL Column | Type | Required | Notes |
|---------------------|-------------------|------|----------|-------|
| User ID | user_id | string | Yes | Unique identifier |
| First Name | first_name | string | No | |
| Last Name | last_name | string | No | |
| Email | email | string | No | Unique index |
| Phone | phone | string | No | |
| Plan | plan | string | No | e.g., "Pilot", "Premium" |
| Plan Duration | plan_duration | string | No | e.g., "Monthly", "Annual" |
| Plan Renewal Date | plan_renewal_date | *time.Time | No | Parsed from ISO date |
| Total Events | total_events | int | No | Default: 0 |
| Guest Access | guest_access | bool | No | Default: false |
| Events Upcoming | events_upcoming | int | No | Default: 0 |
| Events Attended | events_attended | int | No | Default: 0 |
| Notes | notes | string | No | Long text |
| Assigned Events | assigned_events | []Event | No | Many-to-many relationship |

## Events Table

| Airtable Field Name | PostgreSQL Column | Type | Required | Notes |
|---------------------|-------------------|------|----------|-------|
| Event ID | event_id | string | Yes | Unique identifier |
| Category | category | string | No | e.g., "Food & Drink", "Arts & Culture" |
| Tier | tier | string | No | e.g., "Essentials", "Premium" |
| Title | title | string | No | Event name |
| Vendor | vendor | string | No | Business/venue name |
| Image URL | image_url | string | No | Full URL to event image |
| Date Text | date_text | string | No | Human-readable date |
| Day | day | string | No | Day of week or specific date |
| Time | time | string | No | Start time |
| Duration | duration | string | No | e.g., "2 hours" |
| Neighbourhood | neighbourhood | string | No | Area/district |
| Address | address | string | No | Full address |
| Short Desc | short_desc | string | No | Brief description |
| Long Desc | long_desc | string | No | Full description |
| Highlights | highlights | string | No | Key features (can be bullet points) |
| Additional Information | additional_information | string | No | Extra details |
| Status | status | string | No | e.g., "Published", "Draft" |
| Status Qualifier | status_qualifier | string | No | Additional status info |
| Cancellation | cancellation | string | No | Cancellation policy |
| Language | language | string | No | e.g., "English", "German" |
| Curation Order | curation_order | int | No | Display order (lower = higher priority) |
| Retail Price | retail_price | *float64 | No | Original price in EUR |

## Bookings Table

| Airtable Field Name | PostgreSQL Column | Type | Required | Notes |
|---------------------|-------------------|------|----------|-------|
| Booking ID | booking_id | string | No | Auto-generated if missing |
| User | user_id | uint | Yes | Foreign key to Users |
| Event | event_id | uint | Yes | Foreign key to Events |
| Status | status | string | No | e.g., "Requested", "Confirmed", "Attended" |
| Request Notes | request_notes | string | No | User's booking notes |
| Booked At | booked_at | *time.Time | No | Booking timestamp |
| Booking Price | booking_price | *float64 | No | Actual price paid |
| Event Date Confirmed | event_date_confirmed | string | No | Final confirmed date |
| Event Time Confirmed | event_time_confirmed | string | No | Final confirmed time |
| Feedback Rating | feedback_rating | *int | No | 1-5 star rating |
| Feedback Free Text | feedback_free_text | string | No | Written feedback |
| Notes | notes | string | No | Admin notes |

## ActivityLog Table

| Airtable Field Name | PostgreSQL Column | Type | Required | Notes |
|---------------------|-------------------|------|----------|-------|
| Session ID | session_id | string | No | Browser session identifier |
| User | user_id | uint | Yes | Foreign key to Users |
| Event | event_id | *uint | No | Optional foreign key to Events |
| Page URL | page_url | string | No | Full URL visited |
| Action | action | string | No | e.g., "page_view", "click", "booking" |
| Metadata | metadata | string (jsonb) | No | Additional data as JSON |
| Timestamp | timestamp | time.Time | No | When action occurred |

## Data Type Conversions

### Airtable to Go/PostgreSQL

| Airtable Type | Go Type | PostgreSQL Type | Notes |
|---------------|---------|-----------------|-------|
| Single line text | string | VARCHAR | Max 255 chars in DB |
| Long text | string | TEXT | Unlimited length |
| Number | float64 → int | INTEGER | Cast to int for counts |
| Number (decimal) | float64 | DOUBLE PRECISION | For prices |
| Checkbox | bool | BOOLEAN | |
| Date | string → time.Time | TIMESTAMP | Parsed with time.Parse |
| Linked record (single) | uint | INTEGER | Foreign key |
| Linked record (multiple) | []uint | many2many table | Join table created |

### Date/Time Parsing

The script handles these date formats:

```go
// ISO 8601 / RFC3339 (Airtable default)
"2024-01-15T14:30:00.000Z"

// Simple date
"2024-01-15"
```

### JSON Metadata

If the Metadata field in ActivityLog contains JSON:

```json
{
  "browser": "Chrome",
  "device": "mobile",
  "custom_field": "value"
}
```

It's stored as JSONB in PostgreSQL and can be queried:

```sql
SELECT * FROM activity_logs WHERE metadata->>'browser' = 'Chrome';
```

## Relationship Mapping

### Many-to-Many: Users ↔ Events

**Airtable**:
- Users table has "Assigned Events" (linked records)
- Events table has "Assigned Users" (linked records)

**PostgreSQL**:
- Join table: `event_users`
- Columns: `user_id`, `event_id`
- Managed automatically by GORM

### One-to-Many: User → Bookings

**Airtable**:
- Bookings table has "User" (single linked record)

**PostgreSQL**:
- Bookings table has `user_id` foreign key
- Cascade delete: if User deleted, their Bookings are deleted

### One-to-Many: Event → Bookings

**Airtable**:
- Bookings table has "Event" (single linked record)

**PostgreSQL**:
- Bookings table has `event_id` foreign key
- Cascade delete: if Event deleted, related Bookings are deleted

### One-to-Many: User → ActivityLogs

**Airtable**:
- ActivityLog table has "User" (single linked record)

**PostgreSQL**:
- ActivityLogs table has `user_id` foreign key
- Cascade delete: if User deleted, their ActivityLogs are deleted

## Custom Field Mappings

If your Airtable schema uses different field names, update these functions in `main.go`:

### For Users
```go
func mapAirtableUser(record AirtableRecord) (*models.User, error) {
    // Change "User ID" to your field name
    if val, ok := record.Fields["User ID"].(string); ok {
        user.UserID = val
    }
    // ... repeat for other fields
}
```

### For Events
```go
func mapAirtableEvent(record AirtableRecord) (*models.Event, error) {
    // Change "Event ID" to your field name
    if val, ok := record.Fields["Event ID"].(string); ok {
        event.EventID = val
    }
    // ... repeat for other fields
}
```

## Special Cases

### Handling Missing Required Fields

If Airtable records are missing required fields:

```
Error mapping user 25 (recXXXXXXXXXXXXXX): User ID is required
```

**Solution**:
1. Fix data in Airtable before migration
2. Or modify the mapper to generate IDs:
   ```go
   if val, ok := record.Fields["User ID"].(string); ok {
       user.UserID = val
   } else {
       user.UserID = fmt.Sprintf("user_%s", record.ID) // Generate from record ID
   }
   ```

### Handling Invalid Dates

If date parsing fails, the field is left as zero value (nil for pointers).

**To fix**: Update dates in Airtable to ISO 8601 format.

### Handling Price Strings

Some Airtable bases store prices as text (e.g., "€25.00").

The script handles both:
- Number fields: Direct conversion
- Text fields: Attempts to parse as float64

For currency symbols, you may need to strip them:

```go
// In mapAirtableEvent
if val, ok := record.Fields["Retail Price"].(string); ok && val != "" {
    // Remove currency symbols and spaces
    cleaned := strings.TrimSpace(strings.TrimPrefix(val, "€"))
    price, err := strconv.ParseFloat(cleaned, 64)
    if err == nil {
        event.RetailPrice = &price
    }
}
```

## Validation Checklist

Before running the migration:

- [ ] All table names match exactly (case-sensitive)
- [ ] All required fields have data in Airtable
- [ ] User IDs are unique
- [ ] Event IDs are unique
- [ ] Email addresses are unique (if used)
- [ ] Linked records point to existing records
- [ ] Date fields use ISO 8601 format
- [ ] Numeric fields don't contain text
- [ ] Price fields are numbers, not currency-formatted strings

After running the migration:

- [ ] Record counts match between Airtable and PostgreSQL
- [ ] Sample records look correct
- [ ] Relationships are preserved
- [ ] No orphaned records (bookings without users/events)
- [ ] Dates and times are in correct timezone
- [ ] JSON metadata is valid
