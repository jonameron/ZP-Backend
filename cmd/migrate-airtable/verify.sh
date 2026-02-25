#!/bin/bash

# Migration Verification Script
# Run this after migration to verify data integrity

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Migration Verification${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Load environment
if [ ! -f "../../.env" ]; then
    echo "Error: .env file not found"
    exit 1
fi

source ../../.env

# Extract database connection details
DB_URL=$DATABASE_URL

echo -e "${GREEN}Running verification queries...${NC}"
echo ""

# Function to run SQL query
run_query() {
    psql "$DB_URL" -t -c "$1"
}

# Count records in each table
echo -e "${YELLOW}Record Counts:${NC}"
echo -n "Users:        "
run_query "SELECT COUNT(*) FROM users;"
echo -n "Events:       "
run_query "SELECT COUNT(*) FROM events;"
echo -n "Bookings:     "
run_query "SELECT COUNT(*) FROM bookings;"
echo -n "ActivityLogs: "
run_query "SELECT COUNT(*) FROM activity_logs;"
echo -n "User-Event:   "
run_query "SELECT COUNT(*) FROM event_users;"
echo ""

# Check for orphaned records
echo -e "${YELLOW}Data Integrity Checks:${NC}"

orphaned_bookings=$(run_query "SELECT COUNT(*) FROM bookings b LEFT JOIN users u ON b.user_id = u.id WHERE u.id IS NULL;" | xargs)
if [ "$orphaned_bookings" -eq 0 ]; then
    echo -e "${GREEN}✓ No orphaned bookings (missing users)${NC}"
else
    echo -e "${YELLOW}⚠ Found $orphaned_bookings orphaned bookings${NC}"
fi

orphaned_bookings_events=$(run_query "SELECT COUNT(*) FROM bookings b LEFT JOIN events e ON b.event_id = e.id WHERE e.id IS NULL;" | xargs)
if [ "$orphaned_bookings_events" -eq 0 ]; then
    echo -e "${GREEN}✓ No orphaned bookings (missing events)${NC}"
else
    echo -e "${YELLOW}⚠ Found $orphaned_bookings_events orphaned bookings${NC}"
fi

orphaned_activities=$(run_query "SELECT COUNT(*) FROM activity_logs a LEFT JOIN users u ON a.user_id = u.id WHERE u.id IS NULL;" | xargs)
if [ "$orphaned_activities" -eq 0 ]; then
    echo -e "${GREEN}✓ No orphaned activity logs${NC}"
else
    echo -e "${YELLOW}⚠ Found $orphaned_activities orphaned activity logs${NC}"
fi

orphaned_event_users=$(run_query "SELECT COUNT(*) FROM event_users eu LEFT JOIN users u ON eu.user_id = u.id WHERE u.id IS NULL;" | xargs)
if [ "$orphaned_event_users" -eq 0 ]; then
    echo -e "${GREEN}✓ No orphaned user-event relationships (users)${NC}"
else
    echo -e "${YELLOW}⚠ Found $orphaned_event_users orphaned user-event relationships${NC}"
fi

orphaned_event_users_events=$(run_query "SELECT COUNT(*) FROM event_users eu LEFT JOIN events e ON eu.event_id = e.id WHERE e.id IS NULL;" | xargs)
if [ "$orphaned_event_users_events" -eq 0 ]; then
    echo -e "${GREEN}✓ No orphaned user-event relationships (events)${NC}"
else
    echo -e "${YELLOW}⚠ Found $orphaned_event_users_events orphaned user-event relationships${NC}"
fi

echo ""

# Sample data
echo -e "${YELLOW}Sample Records:${NC}"
echo ""
echo "Latest 5 Users:"
run_query "SELECT user_id, first_name, last_name, email FROM users ORDER BY created_at DESC LIMIT 5;"
echo ""

echo "Latest 5 Events:"
run_query "SELECT event_id, title, category, status FROM events ORDER BY created_at DESC LIMIT 5;"
echo ""

echo "Latest 5 Bookings:"
run_query "SELECT booking_id, status, booked_at FROM bookings ORDER BY created_at DESC LIMIT 5;"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Verification Complete${NC}"
echo -e "${GREEN}========================================${NC}"
