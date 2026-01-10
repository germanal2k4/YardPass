#!/bin/bash

# Скрипт для добавления жильца в систему YardPass
# Использование: ./add_resident.sh <telegram_id> <phone> <name> [building_name] [apartment_number]

set -e

TELEGRAM_ID=$1
PHONE=$2
NAME=$3
BUILDING_NAME=${4:-"Тестовый дом"}
APARTMENT_NUMBER=${5:-"1"}

if [ -z "$TELEGRAM_ID" ] || [ -z "$PHONE" ]; then
    echo "Использование: $0 <telegram_id> <phone> <name> [building_name] [apartment_number]"
    echo ""
    echo "Пример:"
    echo "  $0 123456789 +79991234567 \"Иван Иванов\" \"Дом 1\" \"101\""
    echo ""
    echo "Как узнать свой Telegram ID:"
    echo "  1. Напишите боту @userinfobot в Telegram"
    echo "  2. Или используйте @getidsbot"
    exit 1
fi

echo "Добавление жильца в систему..."
echo "Telegram ID: $TELEGRAM_ID"
echo "Телефон: $PHONE"
echo "Имя: ${NAME:-не указано}"
echo "Дом: $BUILDING_NAME"
echo "Квартира: $APARTMENT_NUMBER"
echo ""

# Создаем building если не существует
BUILDING_ID=$(docker exec yardpass-postgres psql -U yardpass -d yardpass -t -A -c "SELECT id FROM buildings WHERE name = '$BUILDING_NAME' LIMIT 1;" | head -1 | xargs)

if [ -z "$BUILDING_ID" ]; then
    echo "Создание здания: $BUILDING_NAME"
    BUILDING_ID=$(docker exec yardpass-postgres psql -U yardpass -d yardpass -t -A -c "INSERT INTO buildings (name, address) VALUES ('$BUILDING_NAME', '') RETURNING id;" | head -1 | xargs)
    echo "✓ Здание создано с ID: $BUILDING_ID"
else
    echo "✓ Здание уже существует с ID: $BUILDING_ID"
fi

# Создаем apartment если не существует
APARTMENT_ID=$(docker exec yardpass-postgres psql -U yardpass -d yardpass -t -A -c "SELECT id FROM apartments WHERE building_id = $BUILDING_ID AND number = '$APARTMENT_NUMBER' LIMIT 1;" | head -1 | xargs)

if [ -z "$APARTMENT_ID" ]; then
    echo "Создание квартиры: $APARTMENT_NUMBER"
    APARTMENT_ID=$(docker exec yardpass-postgres psql -U yardpass -d yardpass -t -A -c "INSERT INTO apartments (building_id, number) VALUES ($BUILDING_ID, '$APARTMENT_NUMBER') RETURNING id;" | head -1 | xargs)
    echo "✓ Квартира создана с ID: $APARTMENT_ID"
else
    echo "✓ Квартира уже существует с ID: $APARTMENT_ID"
fi

# Проверяем, существует ли уже resident
EXISTING=$(docker exec yardpass-postgres psql -U yardpass -d yardpass -t -A -c "SELECT id FROM residents WHERE telegram_id = $TELEGRAM_ID LIMIT 1;" | head -1 | xargs)

if [ -n "$EXISTING" ]; then
    echo "Обновление существующего жильца..."
    if [ -n "$NAME" ]; then
        docker exec yardpass-postgres psql -U yardpass -d yardpass -c "UPDATE residents SET apartment_id = $APARTMENT_ID, chat_id = $TELEGRAM_ID, name = '$NAME', phone = '$PHONE' WHERE telegram_id = $TELEGRAM_ID;"
    else
        docker exec yardpass-postgres psql -U yardpass -d yardpass -c "UPDATE residents SET apartment_id = $APARTMENT_ID, chat_id = $TELEGRAM_ID, phone = '$PHONE' WHERE telegram_id = $TELEGRAM_ID;"
    fi
    echo "✓ Жилец обновлен"
else
    echo "Создание нового жильца..."
    if [ -n "$NAME" ]; then
        docker exec yardpass-postgres psql -U yardpass -d yardpass -c "INSERT INTO residents (apartment_id, telegram_id, chat_id, name, phone, status) VALUES ($APARTMENT_ID, $TELEGRAM_ID, $TELEGRAM_ID, '$NAME', '$PHONE', 'active');"
    else
        docker exec yardpass-postgres psql -U yardpass -d yardpass -c "INSERT INTO residents (apartment_id, telegram_id, chat_id, phone, status) VALUES ($APARTMENT_ID, $TELEGRAM_ID, $TELEGRAM_ID, '$PHONE', 'active');"
    fi
    echo "✓ Жилец создан"
fi

echo ""
echo "✅ Готово! Теперь вы можете взаимодействовать с ботом в Telegram."
echo "Найдите вашего бота и отправьте команду /start"

