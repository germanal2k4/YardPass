#!/bin/bash

# YardPass Database Migration Script
# This script applies all database migrations in order

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default database connection parameters
DB_NAME="${DB_NAME:-yardpass}"
DB_USER="${DB_USER:-yardpass}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

# Parse DATABASE_URL if provided
if [ -n "$DATABASE_URL" ]; then
    echo -e "${YELLOW}Using DATABASE_URL for connection${NC}"
    PSQL_CMD="psql $DATABASE_URL"
else
    echo -e "${YELLOW}Using connection parameters: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME${NC}"
    PSQL_CMD="psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
fi

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/../migrations"

echo -e "${GREEN}=== YardPass Database Migration ===${NC}"
echo ""

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${RED}Error: Migrations directory not found at $MIGRATIONS_DIR${NC}"
    exit 1
fi

# Get list of migration files
MIGRATIONS=$(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort)

if [ -z "$MIGRATIONS" ]; then
    echo -e "${RED}Error: No migration files found in $MIGRATIONS_DIR${NC}"
    exit 1
fi

# Count migrations
TOTAL_MIGRATIONS=$(echo "$MIGRATIONS" | wc -l | xargs)
echo -e "${GREEN}Found $TOTAL_MIGRATIONS migration(s) to apply${NC}"
echo ""

# Apply each migration
CURRENT=0
for migration in $MIGRATIONS; do
    CURRENT=$((CURRENT + 1))
    MIGRATION_NAME=$(basename "$migration")
    
    echo -e "${YELLOW}[$CURRENT/$TOTAL_MIGRATIONS] Applying migration: $MIGRATION_NAME${NC}"
    
    if $PSQL_CMD -f "$migration" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Successfully applied $MIGRATION_NAME${NC}"
    else
        echo -e "${RED}✗ Failed to apply $MIGRATION_NAME${NC}"
        echo -e "${RED}Error details:${NC}"
        $PSQL_CMD -f "$migration"
        exit 1
    fi
    echo ""
done

echo -e "${GREEN}=== All migrations applied successfully! ===${NC}"
