# Railway: Settings → Source → Root Directory = (repo root, empty)
#          Dockerfile Path = deployment/railway/grafana/Dockerfile
# Service name in Railway must be: grafana
# Generate a public domain for this service (Grafana UI).
#
# Required Railway variables (set in UI):
#   GF_SECURITY_ADMIN_USER=admin
#   GF_SECURITY_ADMIN_PASSWORD=<your password>
#   GF_USERS_ALLOW_SIGN_UP=false
#
# Also required in the same Railway project (internal DNS):
#   elasticsearch, jaeger, prometheus, backend (metrics on :5050)

FROM grafana/grafana:11.4.0

COPY backend/deployment/grafana/provisioning/dashboards/dashboards.yml \
  /etc/grafana/provisioning/dashboards/dashboards.yml
# Prod dashboard: Explore links only (no localhost — browser is not on Railway network).
COPY backend/deployment/grafana/provisioning/dashboards/yardpass-backend.railway.json \
  /etc/grafana/provisioning/dashboards/yardpass-backend.json
COPY backend/deployment/grafana/provisioning/datasources/datasources.railway.yml \
  /etc/grafana/provisioning/datasources/datasources.yml

COPY --chmod=755 deployment/railway/grafana/docker-entrypoint.sh /docker-entrypoint-railway.sh

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
  CMD sh -c 'wget --no-verbose --tries=1 --spider "http://127.0.0.1:$${PORT:-3000}/api/health" || exit 1'

ENTRYPOINT ["/docker-entrypoint-railway.sh"]
