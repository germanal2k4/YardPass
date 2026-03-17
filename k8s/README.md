# YardPass Kubernetes + Istio

Развёртывание YardPass в Kubernetes с Istio, Prometheus, Grafana, Kiali (по аналогии с german-currency).

## Структура

```
.
├── helmfile.yaml          # Главный файл деплоя
├── istio/                 # Values для Istio и стека мониторинга
│   ├── istiod-values.yaml
│   ├── prometheus-values.yaml
│   ├── kiali-values.yaml
│   ├── cert-manager-values.yaml
│   ├── postgres-values.yaml
│   └── redis-values.yaml
├── istio-config/          # Istio CR: Gateway, VirtualServices, PeerAuth, ServiceMonitor
├── yardpass-backend/      # Helm chart для API
├── yardpass-frontend/     # Helm chart для SPA
├── yardpass-bot/          # Helm chart для Telegram бота (installed: false по умолчанию)
└── scripts/k8s-deploy.sh
```

## Требования

- Kubernetes (minikube, k3s, или облачный кластер)
- `kubectl`, `helm`, `helmfile`
- Для minikube: `minikube addons enable ingress` (опционально)

## Развёртывание

```bash
# 1. Собрать образы
docker build -t yardpass/backend:latest ./backend
docker build -t yardpass/frontend:latest --build-arg VITE_API_BASE_URL="" ./frontend

# 2. Для minikube — загрузить образы
minikube image load yardpass/backend:latest
minikube image load yardpass/frontend:latest

# 3. Деплой
helmfile sync
```

## Порты и доступ

| Сервис   | Namespace   | Доступ                          |
|----------|-------------|----------------------------------|
| Grafana  | istio-system| `kubectl port-forward -n istio-system svc/prometheus-grafana 3030:80` |
| Kiali    | istio-system| `kubectl port-forward -n istio-system svc/kiali 20001:20001` |
| Prometheus | istio-system | Внутренний, через Grafana |
| Приложение | yardpass  | http://yardpass.example.com (добавить в /etc/hosts) |

## Gateway и роутинг

- **Gateway**: `yardpass.example.com` (порт 80)
- `/`, `/login`, и статика → **yardpass-frontend**
- `/api`, `/auth`, `/service`, `/health` → **yardpass-backend**

## Секреты

Перед деплоем настройте:

1. **yardpass-backend** Secret: `database-url`, `redis-url` (по умолчанию для локального postgres/redis)
2. **yardpass-bot** Secret: `telegram-token` — замените на реальный токен при включении бота

## Включение бота

```yaml
# В helmfile.yaml для release yardpass-bot:
installed: true  # было false
```

И обновите Secret `yardpass-bot` с `TELEGRAM_BOT_TOKEN`.
