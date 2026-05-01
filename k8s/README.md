# YardPass Kubernetes + Istio

Развёртывание YardPass в Kubernetes с Istio, Prometheus, Grafana, Kiali, Jaeger, Elasticsearch, Fluent Bit.

## Структура

```
.
├── helmfile.yaml.gotmpl       # Главный файл деплоя (environments: default, local, cd)
├── values-local.yaml          # Overrides для локального деплоя (minikube + podman)
├── values-cd.yaml             # Overrides для CI/CD деплоя (ghcr.io образы)
├── .github/workflows/
│   ├── ci.yml                 # CI: тесты, линтинг, сборка
│   └── cd.yml                 # CD: build/push образов в GHCR, деплой в K8s
├── scripts/
│   ├── k8s-deploy.sh          # Быстрый деплой (docker)
│   └── k8s-local-deploy.sh    # Полный деплой через podman + minikube
├── istio/                     # Values для Istio и стека мониторинга
├── docker-compose.data.yml    # Только Postgres + Redis (хост для minikube / общий с полным compose)
├── docker-compose.yml         # include data.yml + приложение (backend, frontend, …)
├── istio-config/              # Istio CR: Gateway, VS, ServiceEntry (Postgres/Redis на хосте), PeerAuth, DestinationRule, ServiceMonitor/PodMonitor, Thanos secret
├── values-default.yaml        # Общие значения helmfile
├── jaeger/                    # Helm chart — Jaeger all-in-one
├── elasticsearch/             # Helm chart — Elasticsearch single-node
├── fluent-bit/                # Helm chart — Fluent Bit DaemonSet
├── grafana-dashboards/        # Helm chart — Grafana dashboards via ConfigMaps
├── yardpass-backend/          # Helm chart — API
├── yardpass-frontend/         # Helm chart — SPA (Nginx)
└── yardpass-bot/              # Helm chart — Telegram бот
```

## Требования

- `podman`, `minikube`, `kubectl`, `helm`, `helmfile`
- macOS / Linux

## Postgres и Redis

Helm-чартов Bitnami Postgres/Redis **нет** — базы только в **docker-compose**.

Если раньше ставились релизы `postgres` / `redis` в `yardpass`, сними: `helm uninstall postgres redis -n yardpass` (или `helmfile destroy` по старому манифесту).

1. Поднять данные на хосте: `./scripts/compose-db.sh` или `docker compose -f docker-compose.data.yml up -d` (либо `docker compose up -d postgres redis` из корня — сервисы приходят из `include`).
2. Поды в minikube подключаются к `host.minikube.internal:5432` и `:6379` (см. `yardpass-backend/values.yaml`, `values-local.yaml`).
3. В **istio-config** заданы **ServiceEntry** для Postgres и Redis (`MESH_EXTERNAL`), sidecar маршрутизирует egress.
4. **mTLS (STRICT)** — между подами в mesh. До БД на хосте — TCP через sidecar; при необходимости SSL на стороне Postgres (`sslmode` в DSN).

Дополнительно: PeerAuthentication на метриках 5050, DestinationRule для Prometheus, ServiceMonitor/PodMonitor Istio, Thanos secret по флагу — см. `istio-config/templates/`.

## Локальный деплой (podman + minikube)

```bash
# Полный деплой с нуля
./scripts/k8s-local-deploy.sh up

# Пересборка и передеплой только приложения
./scripts/k8s-local-deploy.sh rebuild

# Остановка (удаление Helm releases и namespace)
./scripts/k8s-local-deploy.sh down

# Полная очистка (удаление кластера minikube)
./scripts/k8s-local-deploy.sh clean

# Только сборка и загрузка образов
./scripts/k8s-local-deploy.sh images

# Статус кластера
./scripts/k8s-local-deploy.sh status
```

## CI/CD (GitHub Actions)

Пайплайн из двух workflows:

1. **CI** (`ci.yml`) — запускается на push/PR в `main`: тесты backend (Go), тесты frontend (Node), E2E (Playwright)
2. **CD** (`cd.yml`) — запускается после успешного CI на `main`:
   - Собирает Docker-образы backend и frontend
   - Пушит в GitHub Container Registry (`ghcr.io/germanal2k4/yardpass-*`)
   - Деплоит в K8s через `helmfile -e cd sync`

### Необходимые GitHub Secrets

| Secret | Описание |
|--------|----------|
| `KUBE_CONFIG` | Base64-encoded kubeconfig для доступа к кластеру (`cat ~/.kube/config \| base64`) |
| `DB_URL` | PostgreSQL connection string (опционально, для переопределения дефолта) |
| `REDIS_URL` | Redis connection string (опционально) |

`GITHUB_TOKEN` для ghcr.io предоставляется автоматически.

### Как настроить

1. Добавьте секреты в GitHub: Settings -> Secrets and variables -> Actions
2. Убедитесь, что K8s кластер доступен из GitHub Actions runner
3. При push в `main` — CI прогоняет тесты, CD собирает образы и деплоит

## Environments (helmfile)

| Environment | Команда | Образы | Описание |
|-------------|---------|--------|----------|
| `default` | `helmfile -f helmfile.yaml.gotmpl sync` | `ghcr.io/germanal2k4/yardpass-*:latest` | Продакшн / облако |
| `local` | `helmfile -f helmfile.yaml.gotmpl -e local sync` | `localhost/yardpass/*:latest` | Локальный (minikube + podman) |
| `cd` | `helmfile -f helmfile.yaml.gotmpl -e cd sync` | `ghcr.io/germanal2k4/yardpass-*:<sha>` | CI/CD (GitHub Actions) |

## Порты и доступ

| Сервис | Namespace | Port-forward |
|--------|-----------|--------------|
| App (через Gateway) | istio-system | `kubectl port-forward -n istio-system svc/istio-ingress 3000:80` |
| Grafana | istio-system | `kubectl port-forward -n istio-system svc/prometheus-grafana 3001:80` |
| Kiali | istio-system | `kubectl port-forward -n istio-system svc/kiali 20001:20001` |
| Jaeger | istio-system | `kubectl port-forward -n istio-system svc/jaeger-query 16686:16686` |
| Prometheus | istio-system | `kubectl port-forward -n istio-system svc/prometheus-kube-prometheus-prometheus 9090:9090` |

После port-forward для Gateway добавьте в `/etc/hosts`:

```
127.0.0.1 yardpass.example.com
```

Приложение: http://yardpass.example.com:3000

## Секреты

Перед деплоем настройте в соответствующих `values.yaml`:

1. **yardpass-backend**: `secrets.databaseUrl`, `secrets.redisUrl`
2. **yardpass-bot**: `secrets.telegramToken` (при включении бота)
