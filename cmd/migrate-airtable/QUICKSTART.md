# Airtable Migration Quick Start Guide

Get your data from Airtable to PostgreSQL in 5 minutes.

## Step 1: Get Your Credentials

### Airtable API Key
1. Go to https://airtable.com/account
2. Scroll to "API" section
3. Click "Generate API key"
4. Copy the key (starts with `key...`)

### Airtable Base ID
1. Go to https://airtable.com/api
2. Click on your base
3. Copy the Base ID from the URL or intro text (starts with `app...`)

## Step 2: Set Up Environment

```bash
cd /Users/jonathanstaude/Documents/current_work/ZP-Backend

# Copy example env file
cp cmd/migrate-airtable/.env.example .env

# Edit with your credentials
nano .env
```

Add your credentials:
```bash
AIRTABLE_API_KEY=keyXXXXXXXXXXXXXX
AIRTABLE_BASE_ID=appXXXXXXXXXXXXXX
DATABASE_URL=postgresql://user:pass@localhost:5432/zeitpass?sslmode=disable
```

## Step 3: Run Migration

### Option A: Using the helper script (Recommended)
```bash
cd cmd/migrate-airtable
./run.sh
```

### Option B: Direct execution
```bash
go run cmd/migrate-airtable/main.go
```

## Step 4: Verify Results

```bash
cd cmd/migrate-airtable
./verify.sh
```

## What Gets Migrated

1. **Users** → PostgreSQL users table
2. **Events** → PostgreSQL events table
3. **User-Event relationships** → event_users join table
4. **Bookings** → PostgreSQL bookings table (with foreign keys)
5. **ActivityLog** → PostgreSQL activity_logs table (with foreign keys)

## Expected Output

```
========================================
Starting Airtable to PostgreSQL Migration
========================================
Fetching records from table: Users
Fetched 100 records (total: 100)
========================================
Starting User migration...
========================================
Progress: 100/100 users migrated
User migration completed: 100 succeeded, 0 failed
...
Migration completed successfully in 2m15s
```

## Troubleshooting

### Common Issues

**"Missing required environment variables"**
- Check that your .env file is in the project root
- Verify all three variables are set

**"airtable API returned status 401"**
- Your API key is wrong or expired
- Generate a new key from Airtable

**"airtable API returned status 404"**
- Base ID is incorrect
- Table names don't match (they're case-sensitive!)

**"user/event not found for booking"**
- Airtable has broken linked records
- Check data integrity in Airtable first

## File Structure

```
cmd/migrate-airtable/
├── main.go              # Main migration script
├── README.md            # Full documentation
├── QUICKSTART.md        # This file
├── FIELD_MAPPINGS.md    # Detailed field mappings
├── .env.example         # Example environment file
├── run.sh              # Helper script to run migration
└── verify.sh           # Post-migration verification
```

## Next Steps

After successful migration:

1. **Verify data**: Run `./verify.sh`
2. **Test your app**: Make sure everything works
3. **Backup database**: `pg_dump zeitpass > backup.sql`
4. **Update your app**: Point to PostgreSQL instead of Airtable

## Need Help?

1. Read the full README.md for detailed documentation
2. Check FIELD_MAPPINGS.md if fields don't match
3. Review the error logs - they're very descriptive
4. Make sure your Airtable schema matches the expected structure

## Safety Notes

- Always backup your database first
- Test on a copy before production
- The script doesn't delete data, but creates new records
- Running multiple times will create duplicates (no upsert logic)

## Performance

- Small datasets (<1,000): ~1-2 minutes
- Medium datasets (1,000-10,000): ~5-15 minutes
- Large datasets (>10,000): ~15-60 minutes

The script respects Airtable's rate limit (5 req/sec) so larger migrations take longer.
