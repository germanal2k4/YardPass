# Railway: Settings → Source → Root Directory = (repo root, empty)
#          Dockerfile Path = deployment/railway/prometheus/Dockerfile
#          Healthcheck Path = /-/healthy
# Service name in Railway must be: prometheus

FROM prom/prometheus:latest

COPY backend/deployment/prometheus/prometheus.railway.yml /etc/prometheus/prometheus.yml
COPY deployment/railway/prometheus/docker-entrypoint.sh /docker-entrypoint-railway.sh
RUN chmod +x /docker-entrypoint-railway.sh

EXPOSE 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
  CMD sh -c 'wget --no-verbose --tries=1 --spider "http://127.0.0.1:$${PORT:-9090}/-/healthy" || exit 1'

ENTRYPOINT ["/docker-entrypoint-railway.sh"]
