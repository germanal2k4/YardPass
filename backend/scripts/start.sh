#!/bin/bash

# Скрипт для запуска YardPass сервисов

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "Использование: $0 [api|bot|both]"
    echo ""
    echo "  api  - запустить только API сервер"
    echo "  bot  - запустить только Telegram бота"
    echo "  both - запустить оба сервиса"
    exit 1
fi

# Определяем правильный Go
if [ -f "$HOME/go/go/bin/go" ]; then
    GO_CMD="$HOME/go/go/bin/go"
elif command -v go &> /dev/null; then
    GO_CMD="go"
else
    echo -e "${RED}ОШИБКА: Go не найден!${NC}"
    echo "Установите Go или добавьте в PATH: export PATH=\"\$HOME/go/go/bin:\$PATH\""
    exit 1
fi

# Переходим в директорию backend
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"

if [ ! -d "$BACKEND_DIR" ]; then
    echo -e "${RED}ОШИБКА: Директория backend не найдена!${NC}"
    exit 1
fi

cd "$BACKEND_DIR"

# Проверка .env (может быть в корне проекта или в backend)
ENV_FILE="$PROJECT_ROOT/.env"
if [ ! -f "$ENV_FILE" ]; then
    ENV_FILE="$BACKEND_DIR/.env"
fi

if [ ! -f "$ENV_FILE" ]; then
    echo -e "${YELLOW}ПРЕДУПРЕЖДЕНИЕ: .env файл не найден!${NC}"
    echo "Создайте .env файл в корне проекта или в backend/"
else
    # Загружаем переменные окружения
    set -a
    source "$ENV_FILE" 2>/dev/null || true
    set +a
fi

# Проверка обязательных переменных для API
if [ "$SERVICE" == "api" ] || [ "$SERVICE" == "both" ]; then
    if [ -z "$JWT_SECRET" ] || [ "$JWT_SECRET" == "your-secret-key-change-this-in-production" ]; then
        echo -e "${RED}ОШИБКА: JWT_SECRET не установлен или использует значение по умолчанию!${NC}"
        exit 1
    fi
fi

# Проверка обязательных переменных для бота
if [ "$SERVICE" == "bot" ] || [ "$SERVICE" == "both" ]; then
    if [ -z "$TELEGRAM_BOT_TOKEN" ] || [ "$TELEGRAM_BOT_TOKEN" == "your-telegram-bot-token" ]; then
        echo -e "${RED}ОШИБКА: TELEGRAM_BOT_TOKEN не установлен или использует значение по умолчанию!${NC}"
        exit 1
    fi
fi

CONFIG_FILE="$BACKEND_DIR/config/config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${YELLOW}ПРЕДУПРЕЖДЕНИЕ: config.yaml файл не найден!${NC}"
    exit 1
fi

# Функция для запуска API
start_api() {
    echo -e "${GREEN}Запуск API сервера...${NC}"
    echo "Директория: $BACKEND_DIR"
    echo "Go: $GO_CMD"
    $GO_CMD run ./cmd/api -c "$CONFIG_FILE"
}

# Функция для запуска бота
start_bot() {
    echo -e "${GREEN}Запуск Telegram бота...${NC}"
    echo "Директория: $BACKEND_DIR"
    echo "Go: $GO_CMD"
    $GO_CMD run ./cmd/bot -c "$CONFIG_FILE"
}

case $SERVICE in
    api)
        start_api
        ;;
    bot)
        start_bot
        ;;
    both)
        echo -e "${GREEN}Запуск обоих сервисов...${NC}"
        echo ""
        
        # Запуск API в фоне
        start_api &
        API_PID=$!
        
        # Небольшая задержка перед запуском бота
        sleep 2
        
        # Запуск бота в фоне
        start_bot &
        BOT_PID=$!
        
        # Обработка сигналов
        trap "kill $API_PID $BOT_PID 2>/dev/null; exit" INT TERM
        
        echo ""
        echo -e "${GREEN}Оба сервиса запущены${NC}"
        echo "  API PID: $API_PID"
        echo "  Bot PID: $BOT_PID"
        echo ""
        echo "Нажмите Ctrl+C для остановки"
        
        # Ожидание завершения
        wait
        ;;
    *)
        echo -e "${RED}Неизвестный сервис: $SERVICE${NC}"
        exit 1
        ;;
esac
