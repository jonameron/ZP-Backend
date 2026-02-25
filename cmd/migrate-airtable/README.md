# Airtable to PostgreSQL Migration Script

This script migrates data from Airtable to PostgreSQL, handling all relationships and data transformations.

## Features

- Fetches data from all 4 Airtable tables (Users, Events, Bookings, ActivityLog)
- Maps Airtable records to GORM models
- Handles many-to-many relationships between Users and Events
- Includes progress logging for each table
- Graceful error handling with detailed error messages
- Rate limiting to respect Airtable API limits (5 requests/second)
- Automatic database schema migration

## Prerequisites

1. Go 1.22 or higher
2. PostgreSQL database
3. Airtable account with API access

## Environment Variables

Create a `.env` file in the project root or set these environment variables:

```bash
# Airtable Configuration
AIRTABLE_API_KEY=your_airtable_api_key_here
AIRTABLE_BASE_ID=your_airtable_base_id_here

# PostgreSQL Configuration
DATABASE_URL=postgresql://username:password@localhost:5432/zeitpass?sslmode=disable
```

### Getting Airtable Credentials

1. **API Key**:
   - Go to https://airtable.com/account
   - In the "API" section, click "Generate API key"
   - Copy your API key

2. **Base ID**:
   - Go to https://airtable.com/api
   - Select your base
   - The Base ID is shown in the URL and documentation (starts with `app`)
   - Example: `appXXXXXXXXXXXXXX`

## Airtable Schema Requirements

The script expects the following table names and field structures:

### Users Table
- User ID (text, required)
- First Name (text)
- Last Name (text)
- Email (text)
- Phone (text)
- Plan (text)
- Plan Duration (text)
- Plan Renewal Date (date)
- Total Events (number)
- Guest Access (checkbox)
- Events Upcoming (number)
- Events Attended (number)
- Notes (long text)
- Assigned Events (linked records to Events table)

### Events Table
- Event ID (text, required)
- Category (text)
- Tier (text)
- Title (text)
- Vendor (text)
- Image URL (text)
- Date Text (text)
- Day (text)
- Time (text)
- Duration (text)
- Neighbourhood (text)
- Address (text)
- Short Desc (text)
- Long Desc (long text)
- Highlights (long text)
- Additional Information (text)
- Status (text)
- Status Qualifier (text)
- Cancellation (text)
- Language (text)
- Curation Order (number)
- Retail Price (number)

### Bookings Table
- Booking ID (text)
- User (linked record to Users table, required)
- Event (linked record to Events table, required)
- Status (text)
- Request Notes (long text)
- Booked At (date/time)
- Booking Price (number)
- Event Date Confirmed (text)
- Event Time Confirmed (text)
- Feedback Rating (number)
- Feedback Free Text (long text)
- Notes (long text)

### ActivityLog Table
- Session ID (text)
- User (linked record to Users table, required)
- Event (linked record to Events table, optional)
- Page URL (text)
- Action (text)
- Metadata (long text or JSON)
- Timestamp (date/time)

## Usage

### 1. Install Dependencies

```bash
cd /Users/jonathanstaude/Documents/current_work/ZP-Backend
go mod tidy
```

### 2. Set Up Environment Variables

```bash
# Copy the example env file
cp .env.example .env

# Edit .env and add your Airtable credentials
nano .env
```

### 3. Run the Migration

```bash
# From the project root
go run cmd/migrate-airtable/main.go
```

The script will:
1. Wait 5 seconds before starting (press Ctrl+C to cancel)
2. Connect to PostgreSQL and run migrations
3. Fetch and migrate Users
4. Fetch and migrate Events
5. Create User-Event relationships (many-to-many)
6. Fetch and migrate Bookings (with foreign keys)
7. Fetch and migrate Activity Logs (with foreign keys)

### 4. Monitor Progress

The script provides detailed logging:

```
========================================
Starting Airtable to PostgreSQL Migration
========================================
Fetching records from table: Users
Fetched 50 records (total: 50)
Total records fetched from Users: 50
========================================
Starting User migration...
========================================
Progress: 10/50 users migrated
Progress: 20/50 users migrated
...
User migration completed: 50 succeeded, 0 failed
```

## Migration Order

The script follows this order to maintain referential integrity:

1. **Users** - Base table with no dependencies
2. **Events** - Base table with no dependencies
3. **User-Event Relationships** - Many-to-many join table
4. **Bookings** - Depends on Users and Events
5. **Activity Logs** - Depends on Users and optionally Events

## Error Handling

- Individual record failures are logged but don't stop the migration
- Each section reports success/failure counts
- If a table migration fails completely, the script stops
- All errors include detailed messages for debugging

## Data Transformation Notes

1. **Airtable IDs to PostgreSQL IDs**:
   - The script maintains internal maps to convert Airtable record IDs to PostgreSQL auto-increment IDs

2. **Linked Records**:
   - Airtable linked records are converted to foreign key relationships
   - Many-to-many relationships use GORM's association management

3. **Date/Time Fields**:
   - Airtable date strings are parsed to Go time.Time
   - Timestamps use RFC3339 format

4. **Optional Fields**:
   - Missing fields are handled gracefully with zero values
   - Null pointers are used for optional numeric fields

5. **Metadata JSON**:
   - JSON metadata fields are preserved as JSONB in PostgreSQL

## Troubleshooting

### "Missing required environment variables"
- Ensure your `.env` file is in the project root
- Verify all three environment variables are set

### "failed to connect to database"
- Check your DATABASE_URL is correct
- Ensure PostgreSQL is running
- Verify network connectivity and credentials

### "airtable API returned status 401"
- Your API key is invalid or expired
- Generate a new API key from Airtable

### "airtable API returned status 404"
- Base ID is incorrect
- Table names don't match (case-sensitive)
- Check permissions on the Airtable base

### "user/event not found for booking"
- Linked records in Airtable reference non-existent records
- Check data integrity in Airtable

### Rate Limiting
- The script respects Airtable's rate limit (5 requests/second)
- If you see rate limit errors, the built-in delays should handle this
- For very large datasets, the migration may take several minutes

## Development

To modify the mapping logic:

1. Edit the `mapAirtable*` functions in `main.go`
2. Update field mappings to match your Airtable schema
3. Test with a small dataset first

## Building a Binary

```bash
# Build for current platform
go build -o migrate-airtable cmd/migrate-airtable/main.go

# Run the binary
./migrate-airtable
```

## Safety Notes

- **Backup your database before running**: This script creates records and should not delete data, but always backup first
- **Test with a copy**: Run the migration on a test database first
- **Idempotency**: Running the script multiple times will create duplicate records (no upsert logic)
- **Dry Run**: Consider adding a dry-run flag for testing (not currently implemented)

## Performance

- **Small datasets (< 1,000 records)**: ~1-2 minutes
- **Medium datasets (1,000-10,000 records)**: ~5-15 minutes
- **Large datasets (> 10,000 records)**: ~15-60 minutes

Performance depends on:
- Network latency to Airtable
- Database write speed
- Number of relationships
- Airtable API rate limits

## Next Steps

After successful migration:

1. Verify record counts match
2. Check relationships are correct
3. Test your application with the migrated data
4. Set up regular backups
5. Consider implementing incremental sync if needed

## Support

For issues or questions:
1. Check the error messages in the logs
2. Verify your Airtable schema matches the expected format
3. Ensure all required fields are present in Airtable
4. Review the mapping functions for field name mismatches
