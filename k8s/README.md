# YardPass Kubernetes + Istio

Развёртывание YardPass в Kubernetes с Istio, Prometheus, Grafana, Kiali, Jaeger, Elasticsearch, Fluent Bit.

## Структура

```
.
├── helmfile.yaml.gotmpl       # Главный файл деплоя (environments: default, local)
├── values-local.yaml          # Overrides для локального деплоя (minikube + podman)
├── scripts/
│   ├── k8s-deploy.sh          # Быстрый деплой (docker)
│   └── k8s-local-deploy.sh    # Полный деплой через podman + minikube
├── istio/                     # Values для Istio и стека мониторинга
├── istio-config/              # Istio CR: Gateway, VirtualServices, PeerAuth
├── jaeger/                    # Helm chart — Jaeger all-in-one
├── elasticsearch/             # Helm chart — Elasticsearch single-node
├── fluent-bit/                # Helm chart — Fluent Bit DaemonSet
├── grafana-dashboards/        # Helm chart — Grafana dashboards via ConfigMaps
├── yardpass-backend/          # Helm chart — API
├── yardpass-frontend/         # Helm chart — SPA (Nginx)
└── yardpass-bot/              # Helm chart — Telegram бот (отключён по умолчанию)
```

## Требования

- `podman`, `minikube`, `kubectl`, `helm`, `helmfile`
- macOS / Linux

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

## Облачный деплой

```bash
# Использует default environment (без local overrides)
helmfile -f helmfile.yaml.gotmpl sync
```

## Environments (helmfile)

| Environment | Команда | Образы | Описание |
|-------------|---------|--------|----------|
| `default` | `helmfile -f helmfile.yaml.gotmpl sync` | `yardpass/backend:latest` | Продакшн / облако |
| `local` | `helmfile -f helmfile.yaml.gotmpl -e local sync` | `localhost/yardpass/backend:latest` | Локальный (minikube + podman) |

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
