#!/usr/bin/env bash
# Поднять / остановить Postgres и Redis (docker-compose.data.yml).
# Работает и с docker compose, и с podman compose.
#
#   compose-db.sh          — up -d + wait
#   compose-db.sh up       — то же
#   compose-db.sh down     — остановить и удалить контейнеры
#   compose-db.sh status   — показать статус сервисов
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

COMPOSE_FILE="docker-compose.data.yml"

detect_compose() {
  if command -v docker &>/dev/null && docker compose version &>/dev/null 2>&1; then
    echo "docker compose"
  elif command -v podman &>/dev/null && podman compose version &>/dev/null 2>&1; then
    echo "podman compose"
  elif command -v podman-compose &>/dev/null; then
    echo "podman-compose"
  else
    echo ""
  fi
}

COMPOSE_CMD="${COMPOSE_CMD:-$(detect_compose)}"
if [[ -z "$COMPOSE_CMD" ]]; then
  echo "ERROR: ни docker compose, ни podman compose не найдены" >&2
  exit 1
fi

compose() { $COMPOSE_CMD -f "$COMPOSE_FILE" "$@"; }

wait_postgres() {
  echo "Ожидание Postgres..."
  local retries=30
  while ! compose exec -T postgres pg_isready -U "${POSTGRES_USER:-yardpass}" >/dev/null 2>&1; do
    retries=$((retries - 1))
    if [[ $retries -le 0 ]]; then
      echo "ERROR: Postgres не стартовал за 30 секунд" >&2
      exit 1
    fi
    sleep 1
  done
  echo "Postgres готов."
}

wait_redis() {
  echo "Ожидание Redis..."
  local retries=20
  while ! compose exec -T redis redis-cli ping >/dev/null 2>&1; do
    retries=$((retries - 1))
    if [[ $retries -le 0 ]]; then
      echo "ERROR: Redis не стартовал за 20 секунд" >&2
      exit 1
    fi
    sleep 1
  done
  echo "Redis готов."
}

cmd_up() {
  echo "==> Запуск Postgres + Redis ($COMPOSE_CMD)..."
  compose up -d
  wait_postgres
  wait_redis
  echo "Готово. K8s: host.minikube.internal → Postgres/Redis на хосте."
}

cmd_down() {
  echo "==> Остановка Postgres + Redis..."
  compose down
}

cmd_status() {
  compose ps
}

case "${1:-up}" in
  up)     cmd_up ;;
  down)   cmd_down ;;
  status) cmd_status ;;
  *)
    echo "Usage: $(basename "$0") [up|down|status]" >&2
    exit 1
    ;;
esac
