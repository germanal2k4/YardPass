#!/bin/bash

# YardPass Database Migration Script
# This script applies all database migrations in order

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATIONS_DIR="$PROJECT_ROOT/migrations"

# Load .env file if exists
ENV_FILE="$PROJECT_ROOT/../.env"
if [ ! -f "$ENV_FILE" ]; then
    ENV_FILE="$PROJECT_ROOT/.env"
fi

if [ -f "$ENV_FILE" ]; then
    echo -e "${YELLOW}Loading environment from: $ENV_FILE${NC}"
    set -a
    source "$ENV_FILE" 2>/dev/null || true
    set +a
fi

# Parse DATABASE_URL
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}ОШИБКА: DATABASE_URL не установлен!${NC}"
    echo "Установите DATABASE_URL в .env файле или как переменную окружения"
    exit 1
fi

echo -e "${GREEN}=== YardPass Database Migration ===${NC}"
echo ""

# Parse DATABASE_URL to extract components
if [[ $DATABASE_URL =~ postgres://([^:]+):([^@]+)@([^:]+):([^/]+)/([^?]+) ]]; then
    DB_USER="${BASH_REMATCH[1]}"
    DB_PASS="${BASH_REMATCH[2]}"
    DB_HOST="${BASH_REMATCH[3]}"
    DB_PORT="${BASH_REMATCH[4]}"
    DB_NAME="${BASH_REMATCH[5]}"
    
    echo -e "${YELLOW}Подключение к БД:${NC}"
    echo "  Host: $DB_HOST"
    echo "  Port: $DB_PORT"
    echo "  Database: $DB_NAME"
    echo "  User: $DB_USER"
    echo ""
else
    echo -e "${RED}ОШИБКА: Неверный формат DATABASE_URL${NC}"
    echo "Ожидается формат: postgres://user:password@host:port/database"
    exit 1
fi

# Check if psql is available, if not try Docker
USE_DOCKER=false
if ! command -v psql &> /dev/null; then
    echo -e "${YELLOW}psql не найден локально, пробую использовать Docker...${NC}"
    
    # Check if Docker is available and container exists
    if command -v docker &> /dev/null; then
        if docker ps --format "{{.Names}}" | grep -q "yardpass-postgres"; then
            USE_DOCKER=true
            echo -e "${GREEN}✓ Найден контейнер yardpass-postgres, использую Docker${NC}"
        else
            echo -e "${RED}ОШИБКА: Контейнер yardpass-postgres не запущен!${NC}"
            echo "Запустите: docker compose up -d"
            exit 1
        fi
    else
        echo -e "${RED}ОШИБКА: psql и Docker не найдены!${NC}"
        echo ""
        echo "Установите PostgreSQL client:"
        echo "  Ubuntu/Debian: sudo apt-get install postgresql-client"
        echo "  macOS: brew install postgresql"
        exit 1
    fi
fi

# Setup PSQL command based on method
if [ "$USE_DOCKER" = true ]; then
    # Use Docker to execute psql
    PSQL_CMD="docker exec yardpass-postgres psql -U $DB_USER -d $DB_NAME"
    # No need for PGPASSWORD with Docker exec
else
    export PGPASSWORD="$DB_PASS"
    PSQL_CMD="psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
fi

# Check database connection
echo -e "${YELLOW}Проверка подключения к БД...${NC}"
if ! $PSQL_CMD -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}ОШИБКА: Не удалось подключиться к базе данных!${NC}"
    echo ""
    echo "Проверьте:"
    echo "  1. PostgreSQL запущен"
    echo "  2. DATABASE_URL правильный"
    echo "  3. База данных создана"
    echo "  4. Пользователь имеет права доступа"
    unset PGPASSWORD
    exit 1
fi
echo -e "${GREEN}✓ Подключение успешно${NC}"
echo ""

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${RED}ОШИБКА: Директория миграций не найдена: $MIGRATIONS_DIR${NC}"
    unset PGPASSWORD
    exit 1
fi

# Get list of migration files (excluding down directory)
MIGRATIONS=$(find "$MIGRATIONS_DIR" -maxdepth 1 -name "*.sql" -type f | sort)

if [ -z "$MIGRATIONS" ]; then
    echo -e "${RED}ОШИБКА: Файлы миграций не найдены в $MIGRATIONS_DIR${NC}"
    unset PGPASSWORD
    exit 1
fi

# Count migrations
TOTAL_MIGRATIONS=$(echo "$MIGRATIONS" | wc -l | xargs)
echo -e "${GREEN}Найдено миграций: $TOTAL_MIGRATIONS${NC}"
echo ""

# Check existing tables
echo -e "${YELLOW}Проверка существующих таблиц...${NC}"
EXISTING_TABLES=$($PSQL_CMD -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs)

if [ "$EXISTING_TABLES" -gt 0 ]; then
    echo -e "${YELLOW}⚠ Найдено таблиц в БД: $EXISTING_TABLES${NC}"
    read -p "  Продолжить выполнение миграций? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Отменено"
        unset PGPASSWORD
        exit 0
    fi
else
    echo -e "${GREEN}✓ База данных пуста${NC}"
fi
echo ""

# Apply each migration
CURRENT=0
for migration in $MIGRATIONS; do
    CURRENT=$((CURRENT + 1))
    MIGRATION_NAME=$(basename "$migration")
    
    echo -e "${YELLOW}[$CURRENT/$TOTAL_MIGRATIONS] Применение миграции: $MIGRATION_NAME${NC}"
    
    # Use different approach for Docker vs local psql
    if [ "$USE_DOCKER" = true ]; then
        # For Docker, pipe the file content
        if docker exec -i yardpass-postgres psql -U $DB_USER -d $DB_NAME < "$migration" 2>&1 | grep -v "ERROR:" | grep -v "NOTICE:" | head -5; then
            # Check if there were actual errors (not just warnings)
            ERRORS=$(docker exec -i yardpass-postgres psql -U $DB_USER -d $DB_NAME < "$migration" 2>&1 | grep -c "ERROR:" || true)
            if [ "$ERRORS" -gt 0 ]; then
                # Show full output if there are errors
                docker exec -i yardpass-postgres psql -U $DB_USER -d $DB_NAME < "$migration" 2>&1 | grep "ERROR:"
                echo -e "${YELLOW}⚠ Некоторые объекты уже существуют (это нормально)${NC}"
            else
                echo -e "${GREEN}✓ Успешно применена: $MIGRATION_NAME${NC}"
            fi
        else
            echo -e "${GREEN}✓ Применена: $MIGRATION_NAME${NC}"
        fi
    else
        # For local psql, use -f flag
        OUTPUT=$($PSQL_CMD -f "$migration" 2>&1)
        ERRORS=$(echo "$OUTPUT" | grep -c "ERROR:" || true)
        if [ "$ERRORS" -gt 0 ]; then
            # Filter out "already exists" errors as they're usually OK
            CRITICAL=$(echo "$OUTPUT" | grep "ERROR:" | grep -v "already exists" || true)
            if [ -n "$CRITICAL" ]; then
                echo "$OUTPUT"
                echo -e "${RED}✗ Ошибка при применении: $MIGRATION_NAME${NC}"
                unset PGPASSWORD
                exit 1
            else
                echo -e "${YELLOW}⚠ Некоторые объекты уже существуют (это нормально)${NC}"
                echo -e "${GREEN}✓ Миграция применена: $MIGRATION_NAME${NC}"
            fi
        else
            echo -e "${GREEN}✓ Успешно применена: $MIGRATION_NAME${NC}"
        fi
    fi
    echo ""
done

# Verify scan_events table exists
echo -e "${YELLOW}Проверка создания таблиц...${NC}"
if $PSQL_CMD -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'scan_events');" 2>/dev/null | grep -q "t"; then
    echo -e "${GREEN}✓ Таблица scan_events создана${NC}"
else
    echo -e "${RED}✗ Таблица scan_events НЕ найдена!${NC}"
    echo -e "${YELLOW}Возможно, миграция выполнилась с ошибками${NC}"
fi

# Show all tables
echo ""
echo -e "${GREEN}Созданные таблицы:${NC}"
$PSQL_CMD -c "\dt" 2>/dev/null || true

unset PGPASSWORD

echo ""
echo -e "${GREEN}=== Все миграции применены успешно! ===${NC}"
