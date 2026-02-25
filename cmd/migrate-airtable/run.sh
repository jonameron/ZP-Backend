#!/bin/bash

# Airtable to PostgreSQL Migration Runner
# This script helps you run the migration safely

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Airtable to PostgreSQL Migration${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if .env file exists
if [ ! -f "../../.env" ]; then
    echo -e "${RED}Error: .env file not found in project root${NC}"
    echo -e "${YELLOW}Please create a .env file with the following variables:${NC}"
    echo "  AIRTABLE_API_KEY"
    echo "  AIRTABLE_BASE_ID"
    echo "  DATABASE_URL"
    echo ""
    echo -e "${YELLOW}You can copy .env.example as a starting point:${NC}"
    echo "  cp .env.example ../../.env"
    exit 1
fi

# Load environment variables
source ../../.env

# Validate required environment variables
MISSING_VARS=()
if [ -z "$AIRTABLE_API_KEY" ]; then
    MISSING_VARS+=("AIRTABLE_API_KEY")
fi
if [ -z "$AIRTABLE_BASE_ID" ]; then
    MISSING_VARS+=("AIRTABLE_BASE_ID")
fi
if [ -z "$DATABASE_URL" ]; then
    MISSING_VARS+=("DATABASE_URL")
fi

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
    echo -e "${RED}Error: Missing required environment variables:${NC}"
    for var in "${MISSING_VARS[@]}"; do
        echo "  - $var"
    done
    echo ""
    echo -e "${YELLOW}Please add these to your .env file${NC}"
    exit 1
fi

# Show configuration (masked)
echo -e "${GREEN}Configuration:${NC}"
echo "  Airtable API Key: ${AIRTABLE_API_KEY:0:10}..."
echo "  Airtable Base ID: $AIRTABLE_BASE_ID"
echo "  Database: ${DATABASE_URL%%\?*}" # Show URL without query params
echo ""

# Confirm before proceeding
echo -e "${YELLOW}This will migrate data from Airtable to PostgreSQL.${NC}"
echo -e "${YELLOW}Make sure you have backed up your database!${NC}"
echo ""
read -p "Do you want to proceed? (yes/no): " -r
echo ""

if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
    echo -e "${RED}Migration cancelled${NC}"
    exit 0
fi

# Run the migration
echo -e "${GREEN}Starting migration...${NC}"
echo ""

cd ../..
go run cmd/migrate-airtable/main.go

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Migration completed!${NC}"
echo -e "${GREEN}========================================${NC}"
