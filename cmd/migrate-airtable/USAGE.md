# Complete Usage Instructions

## Overview

This migration script transfers all data from your Airtable base to PostgreSQL, preserving relationships and data integrity.

## What You Need

1. **Airtable Credentials**
   - API Key (from https://airtable.com/account)
   - Base ID (from https://airtable.com/api)

2. **PostgreSQL Database**
   - Connection URL with credentials
   - Database should exist (script will create tables)

3. **Go 1.22+** installed

## Quick Start (3 Commands)

```bash
# 1. Set up environment
cp cmd/migrate-airtable/.env.example .env
# Edit .env with your credentials

# 2. Run migration
cd cmd/migrate-airtable && ./run.sh

# 3. Verify results
./verify.sh
```

## Detailed Steps

### 1. Configure Environment Variables

Create a `.env` file in the project root:

```bash
cd /Users/jonathanstaude/Documents/current_work/ZP-Backend
cp cmd/migrate-airtable/.env.example .env
```

Edit `.env`:

```bash
# Required
AIRTABLE_API_KEY=keyXXXXXXXXXXXXXX    # From https://airtable.com/account
AIRTABLE_BASE_ID=appXXXXXXXXXXXXXX    # From https://airtable.com/api
DATABASE_URL=postgresql://username:password@localhost:5432/zeitpass?sslmode=disable

# Optional
ENVIRONMENT=development  # Enables verbose logging
```

### 2. Verify Airtable Schema

Your Airtable base must have these tables:

- **Users** (with User ID field)
- **Events** (with Event ID field)
- **Bookings** (with User and Event linked records)
- **ActivityLog** (with User linked record)

See `FIELD_MAPPINGS.md` for complete field requirements.

### 3. Run the Migration

**Method A: Using the helper script (Recommended)**

```bash
cd cmd/migrate-airtable
./run.sh
```

This script:
- Checks for required environment variables
- Shows configuration (with masked credentials)
- Asks for confirmation
- Runs the migration
- Shows success/failure summary

**Method B: Direct Go execution**

```bash
cd /Users/jonathanstaude/Documents/current_work/ZP-Backend
go run cmd/migrate-airtable/main.go
```

**Method C: Build and run binary**

```bash
cd /Users/jonathanstaude/Documents/current_work/ZP-Backend
go build -o migrate cmd/migrate-airtable/main.go
./migrate
```

### 4. Monitor Progress

The script provides detailed logging:

```
========================================
Starting Airtable to PostgreSQL Migration
========================================
Fetching records from table: Users
Fetched 50 records (total: 50)
========================================
Starting User migration...
========================================
Progress: 10/50 users migrated
Progress: 20/50 users migrated
Progress: 30/50 users migrated
Progress: 40/50 users migrated
Progress: 50/50 users migrated
User migration completed: 50 succeeded, 0 failed
========================================
Starting Event migration...
========================================
...
```

### 5. Verify Migration

After completion, run the verification script:

```bash
cd cmd/migrate-airtable
./verify.sh
```

This checks:
- Record counts in each table
- No orphaned records (broken foreign keys)
- Sample data from each table

Expected output:

```
========================================
Migration Verification
========================================

Record Counts:
Users:        150
Events:       200
Bookings:     450
ActivityLogs: 1250
User-Event:   300

Data Integrity Checks:
✓ No orphaned bookings (missing users)
✓ No orphaned bookings (missing events)
✓ No orphaned activity logs
✓ No orphaned user-event relationships (users)
✓ No orphaned user-event relationships (events)

Sample Records:
...
```

## Migration Process

The script executes in this order:

1. **Connect to PostgreSQL**
   - Runs auto-migration to create/update tables

2. **Migrate Users** (independent table)
   - Fetches from Airtable "Users" table
   - Creates PostgreSQL user records
   - Maps Airtable IDs → PostgreSQL IDs

3. **Migrate Events** (independent table)
   - Fetches from Airtable "Events" table
   - Creates PostgreSQL event records
   - Maps Airtable IDs → PostgreSQL IDs

4. **Migrate User-Event Relationships**
   - Re-fetches Users to get "Assigned Events" links
   - Creates many-to-many relationships in `event_users` table

5. **Migrate Bookings** (depends on Users and Events)
   - Fetches from Airtable "Bookings" table
   - Maps linked User and Event records to foreign keys
   - Creates PostgreSQL booking records

6. **Migrate Activity Logs** (depends on Users, optionally Events)
   - Fetches from Airtable "ActivityLog" table
   - Maps linked User and Event records to foreign keys
   - Creates PostgreSQL activity log records

## Error Handling

### Individual Record Failures

If a single record fails, it's logged but doesn't stop the migration:

```
Error mapping user 42 (recXXXXXXXXXXXXXX): User ID is required
```

At the end of each section, you'll see:

```
User migration completed: 149 succeeded, 1 failed
```

### Complete Section Failures

If an entire section fails, the migration stops:

```
Error: user migration failed: failed to fetch users: airtable API returned status 404
```

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| Missing required environment variables | .env not configured | Add required vars to .env |
| failed to connect to database | Database URL wrong | Check DATABASE_URL |
| airtable API returned status 401 | Invalid API key | Check AIRTABLE_API_KEY |
| airtable API returned status 404 | Wrong base ID or table name | Check AIRTABLE_BASE_ID and table names |
| User ID is required | Missing field in Airtable | Fill in required fields |
| user not found for booking | Broken linked record | Fix links in Airtable |
| duplicate key value violates unique constraint | Running migration twice | Clear database or use fresh DB |

## Advanced Usage

### Custom Field Mappings

If your Airtable schema uses different field names, edit the mapper functions in `main.go`:

```go
// Example: Your Airtable uses "UID" instead of "User ID"
if val, ok := record.Fields["UID"].(string); ok {
    user.UserID = val
}
```

### Filtering Records

To migrate only certain records, add filters to the Airtable API request:

```go
// In FetchRecords function
requestURL := fmt.Sprintf("%s?filterByFormula=Status='Active'", baseURL)
```

### Batch Size

The script fetches 100 records per request (Airtable default). To change:

```go
requestURL := fmt.Sprintf("%s?pageSize=50", baseURL)
```

### Handling Large Datasets

For very large datasets (>50,000 records):

1. **Increase database timeout**:
   ```go
   db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
       PrepareStmt: true,
       NowFunc: func() time.Time { return time.Now().UTC() },
   })
   ```

2. **Use batch inserts**:
   ```go
   db.CreateInBatches(users, 100)
   ```

3. **Monitor memory**: The script loads all records into memory

## Safety and Best Practices

### Before Migration

- [ ] **Backup your database**
  ```bash
  pg_dump zeitpass > backup_$(date +%Y%m%d_%H%M%S).sql
  ```

- [ ] **Test on a copy**
  ```bash
  createdb zeitpass_test
  DATABASE_URL=postgresql://user:pass@localhost/zeitpass_test go run cmd/migrate-airtable/main.go
  ```

- [ ] **Verify Airtable data**
  - All required fields have values
  - No broken linked records
  - Dates are in correct format

### During Migration

- Don't interrupt the script mid-migration
- Monitor the logs for errors
- Check progress indicators

### After Migration

- Run verification script
- Compare record counts
- Test your application
- Keep Airtable as backup until verified

## Performance Optimization

### Slow Migration?

1. **Check network latency**
   - Airtable API calls depend on internet speed

2. **Database performance**
   - Ensure PostgreSQL has enough resources
   - Disable unnecessary indexes during migration

3. **Rate limiting**
   - Script sleeps 200ms between requests
   - Airtable allows 5 req/sec

4. **Parallel processing** (advanced)
   - Could fetch multiple tables in parallel
   - Not implemented to keep code simple

## Troubleshooting Guide

### Script starts but no data is migrated

**Check:**
- Are table names correct? (case-sensitive: "Users" not "users")
- Does your Airtable API key have access to this base?
- Are there records in Airtable?

### Some records are skipped

**Check:**
- Error logs for specific record IDs
- Missing required fields in those records
- Data type mismatches

### Foreign key constraint violations

**Check:**
- Users and Events migrated successfully first?
- Linked records in Airtable are valid?
- No circular dependencies?

### Duplicate key violations

**Cause:** Running migration multiple times

**Solution:**
```sql
-- Clear existing data
TRUNCATE users, events, bookings, activity_logs CASCADE;
-- Then re-run migration
```

## Getting Help

1. **Check the logs** - Error messages are descriptive
2. **Review FIELD_MAPPINGS.md** - Verify field names match
3. **Run verify.sh** - Identify data integrity issues
4. **Check Airtable data** - Ensure it's valid before migration

## Files Reference

| File | Purpose |
|------|---------|
| `main.go` | Complete migration script (831 lines) |
| `README.md` | Full documentation |
| `QUICKSTART.md` | 5-minute getting started guide |
| `USAGE.md` | This file - complete usage instructions |
| `FIELD_MAPPINGS.md` | Detailed field mapping reference |
| `.env.example` | Example environment configuration |
| `run.sh` | Helper script to run migration |
| `verify.sh` | Post-migration verification script |

## Support

For issues:
1. Review error messages carefully
2. Check environment variables
3. Verify Airtable schema matches expected format
4. Ensure database connectivity
5. Review the code - it's well-commented and straightforward
