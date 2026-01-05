#!/bin/bash

# YardPass Database Rollback Script
# This script rolls back database migrations
# Dynamically discovers and handles any set of migrations

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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
DOWN_MIGRATIONS_DIR="$MIGRATIONS_DIR/down"

echo -e "${RED}╔════════════════════════════════════════╗${NC}"
echo -e "${RED}║   YardPass Database Rollback Script   ║${NC}"
echo -e "${RED}╚════════════════════════════════════════╝${NC}"
echo ""

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${RED}Error: Migrations directory not found at $MIGRATIONS_DIR${NC}"
    exit 1
fi

# Check if down migrations directory exists
if [ ! -d "$DOWN_MIGRATIONS_DIR" ]; then
    echo -e "${RED}Error: Down migrations directory not found at $DOWN_MIGRATIONS_DIR${NC}"
    exit 1
fi

# Function to get all down migrations sorted by number (descending)
get_down_migrations() {
    find "$DOWN_MIGRATIONS_DIR" -name "*_down.sql" -type f | sort -r
}

# Function to extract migration number from filename
get_migration_number() {
    local filename=$(basename "$1")
    echo "$filename" | grep -o '^[0-9]\+' || echo "000"
}

# Function to get migration name without extension and number
get_migration_name() {
    local filename=$(basename "$1")
    # Remove _down.sql extension and leading number with underscore
    echo "$filename" | sed 's/_down\.sql$//' | sed 's/^[0-9]\{3\}_//'
}

# Get list of available down migrations
AVAILABLE_MIGRATIONS=$(get_down_migrations)

# Check if any migrations exist
if [ -z "$AVAILABLE_MIGRATIONS" ]; then
    echo -e "${YELLOW}No down migrations found in $DOWN_MIGRATIONS_DIR${NC}"
    exit 0
fi

# Count migrations
TOTAL_MIGRATIONS=$(echo "$AVAILABLE_MIGRATIONS" | wc -l | xargs)
echo -e "${GREEN}Found $TOTAL_MIGRATIONS down migration(s)${NC}"
echo ""

# Get latest (highest number) migration
LATEST_MIGRATION=$(echo "$AVAILABLE_MIGRATIONS" | head -1)
LATEST_NUMBER=$(get_migration_number "$LATEST_MIGRATION")
LATEST_NAME=$(get_migration_name "$LATEST_MIGRATION")

# Menu
echo -e "${YELLOW}Выберите действие:${NC}"
echo "1) Откатить последнюю миграцию (${LATEST_NUMBER}_${LATEST_NAME})"
echo "2) Откатить все миграции (полная очистка БД)"
echo "3) Откатить конкретную миграцию"
echo "4) Показать список доступных миграций"
echo "5) Отмена"
echo ""
read -p "Ваш выбор (1-5): " choice

case $choice in
    1)
        # Rollback latest migration
        echo -e "${YELLOW}Откат последней миграции: ${LATEST_NUMBER}_${LATEST_NAME}${NC}"
        echo ""
        echo "Файл: $(basename "$LATEST_MIGRATION")"
        echo ""
        
        # Show first few lines of migration for preview
        echo -e "${BLUE}Предпросмотр (первые 5 строк комментариев):${NC}"
        grep -E "^--" "$LATEST_MIGRATION" | head -5 || echo "Комментарии не найдены"
        echo ""
        
        read -p "Продолжить откат? (yes/no): " confirm
        
        if [ "$confirm" = "yes" ]; then
            echo -e "${YELLOW}Выполнение отката...${NC}"
            if $PSQL_CMD -f "$LATEST_MIGRATION"; then
                echo -e "${GREEN}✓ Миграция ${LATEST_NUMBER} успешно откачена${NC}"
            else
                echo -e "${RED}✗ Ошибка при откате миграции${NC}"
                exit 1
            fi
        else
            echo -e "${YELLOW}Отменено${NC}"
        fi
        ;;
        
    2)
        # Rollback all migrations
        echo -e "${RED}⚠️  ВНИМАНИЕ: Это удалит ВСЕ ДАННЫЕ из базы!${NC}"
        echo ""
        echo "Будут откачены следующие миграции (в обратном порядке):"
        echo "$AVAILABLE_MIGRATIONS" | while read -r migration; do
            migration_num=$(get_migration_number "$migration")
            migration_name=$(get_migration_name "$migration")
            echo "  - ${migration_num}_${migration_name}"
        done
        echo ""
        read -p "Вы УВЕРЕНЫ? Введите 'DELETE ALL' для подтверждения: " confirm
        
        if [ "$confirm" = "DELETE ALL" ]; then
            echo -e "${YELLOW}Откат всех миграций в обратном порядке...${NC}"
            echo ""
            
            CURRENT=0
            # Rollback in reverse order (already sorted descending)
            echo "$AVAILABLE_MIGRATIONS" | while read -r migration; do
                CURRENT=$((CURRENT + 1))
                migration_num=$(get_migration_number "$migration")
                migration_name=$(get_migration_name "$migration")
                migration_file=$(basename "$migration")
                
                echo -e "${YELLOW}[$CURRENT/$TOTAL_MIGRATIONS] Откат: ${migration_num}_${migration_name}${NC}"
                
                if $PSQL_CMD -f "$migration" 2>&1 | grep -v "NOTICE"; then
                    echo -e "${GREEN}✓ ${migration_file} откачена${NC}"
                else
                    echo -e "${RED}✗ Ошибка при откате ${migration_file}${NC}"
                    echo -e "${YELLOW}Продолжить откат остальных миграций? (yes/no): ${NC}"
                    read -p "" continue_rollback
                    if [ "$continue_rollback" != "yes" ]; then
                        exit 1
                    fi
                fi
                echo ""
            done
            
            echo -e "${GREEN}✓ Все миграции откачены${NC}"
            echo -e "${YELLOW}База данных очищена${NC}"
        else
            echo -e "${YELLOW}Отменено${NC}"
        fi
        ;;
        
    3)
        # Rollback specific migration
        echo -e "${YELLOW}Доступные миграции для отката:${NC}"
        echo ""
        
        # Display available migrations with numbers
        counter=1
        echo "$AVAILABLE_MIGRATIONS" | sort | while read -r migration; do
            migration_num=$(get_migration_number "$migration")
            migration_name=$(get_migration_name "$migration")
            echo "$counter) ${migration_num}_${migration_name}"
            counter=$((counter + 1))
        done
        echo ""
        
        # Also allow direct number input
        echo -e "${BLUE}Вы можете ввести:${NC}"
        echo "  - Номер из списка выше (1-${TOTAL_MIGRATIONS})"
        echo "  - Номер миграции (например: 003)"
        echo ""
        read -p "Ваш выбор: " migration_choice
        
        # Try to find migration by number or by list position
        SELECTED_MIGRATION=""
        
        # Check if it's a migration number (e.g., 003)
        if echo "$migration_choice" | grep -qE '^[0-9]{3}$'; then
            SELECTED_MIGRATION=$(find "$DOWN_MIGRATIONS_DIR" -name "${migration_choice}_*_down.sql" -type f | head -1)
        # Check if it's a list position
        elif echo "$migration_choice" | grep -qE '^[0-9]+$' && [ "$migration_choice" -ge 1 ] && [ "$migration_choice" -le "$TOTAL_MIGRATIONS" ]; then
            # Get migration by position
            SELECTED_MIGRATION=$(echo "$AVAILABLE_MIGRATIONS" | sort | sed -n "${migration_choice}p")
        fi
        
        if [ -n "$SELECTED_MIGRATION" ] && [ -f "$SELECTED_MIGRATION" ]; then
            migration_num=$(get_migration_number "$SELECTED_MIGRATION")
            migration_name=$(get_migration_name "$SELECTED_MIGRATION")
            
            echo ""
            echo -e "${YELLOW}Откат: ${migration_num}_${migration_name}${NC}"
            echo "Файл: $(basename "$SELECTED_MIGRATION")"
            echo ""
            
            # Show preview
            echo -e "${BLUE}Предпросмотр:${NC}"
            grep -E "^--" "$SELECTED_MIGRATION" | head -5 || echo "Комментарии не найдены"
            echo ""
            
            read -p "Продолжить? (yes/no): " confirm
            
            if [ "$confirm" = "yes" ]; then
                if $PSQL_CMD -f "$SELECTED_MIGRATION"; then
                    echo -e "${GREEN}✓ Миграция ${migration_num} успешно откачена${NC}"
                else
                    echo -e "${RED}✗ Ошибка при откате миграции${NC}"
                    exit 1
                fi
            else
                echo -e "${YELLOW}Отменено${NC}"
            fi
        else
            echo -e "${RED}Миграция не найдена${NC}"
            exit 1
        fi
        ;;
        
    4)
        # List available migrations
        echo -e "${BLUE}Доступные down-миграции:${NC}"
        echo ""
        printf "%-6s %-40s %s\n" "Номер" "Название" "Файл"
        echo "─────────────────────────────────────────────────────────────────────"
        
        echo "$AVAILABLE_MIGRATIONS" | sort | while read -r migration; do
            migration_num=$(get_migration_number "$migration")
            migration_name=$(get_migration_name "$migration")
            migration_file=$(basename "$migration")
            printf "%-6s %-40s %s\n" "$migration_num" "$migration_name" "$migration_file"
        done
        echo ""
        echo -e "${GREEN}Всего: $TOTAL_MIGRATIONS миграций${NC}"
        ;;
        
    5)
        # Cancel
        echo -e "${YELLOW}Отменено${NC}"
        exit 0
        ;;
        
    *)
        echo -e "${RED}Неверный выбор${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${BLUE}Готово! 🔄${NC}"

