#!/usr/bin/env bash
# Поднять Postgres и Redis (docker-compose.data.yml) перед helmfile sync / minikube.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

docker compose -f docker-compose.data.yml up -d

echo "Ожидание Postgres..."
until docker compose -f docker-compose.data.yml exec -T postgres pg_isready -U "${POSTGRES_USER:-yardpass}" >/dev/null 2>&1; do
  sleep 1
done
echo "Готово. K8s: host.minikube.internal → этот Postgres/Redis на хосте."
