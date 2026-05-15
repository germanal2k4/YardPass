# Railway: Settings → Source → Root Directory = (repo root, empty)
#          Dockerfile Path = deployment/railway/prometheus/Dockerfile
#          Healthcheck Path = /-/healthy
#          Variables: PORT=9090  (required — Grafana uses prometheus.railway.internal:9090)
#          Optional: BACKEND_SCRAPE_PORT=8080, BOT_SCRAPE_PORT=${{Bot.PORT}} (if bot metrics port != 5050)
# Service name in Railway must be: prometheus

FROM prom/prometheus:latest

USER root

COPY backend/deployment/prometheus/prometheus.railway.yml /etc/prometheus/prometheus.yml
COPY --chmod=755 deployment/railway/prometheus/docker-entrypoint.sh /docker-entrypoint-railway.sh

# prom/prometheus image runs as nobody (uid 65534)
USER nobody

EXPOSE 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
  CMD sh -c 'wget --no-verbose --tries=1 --spider "http://127.0.0.1:$${PORT:-9090}/-/healthy" || exit 1'

ENTRYPOINT ["/docker-entrypoint-railway.sh"]
