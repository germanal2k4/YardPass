#!/bin/sh
set -e

# Railway routes healthchecks and traffic to $PORT; Prometheus defaults to 9090.
: "${PORT:=9090}"

# Scrape targets (override in Railway Variables on the prometheus service).
# Backend exposes /metrics on API port (default 8080).
: "${BACKEND_SCRAPE_PORT:=8080}"
# Bot metrics listen on METRICS_PORT, or Railway $PORT when METRICS_PORT is unset (see applyRailwayBotMetricsPort).
# If your bot uses a dynamic Railway PORT, set BOT_SCRAPE_PORT=${{Bot.PORT}} on this service.
: "${BOT_SCRAPE_PORT:=5050}"

CFG_OUT=/tmp/prometheus.railway.generated.yml
sed -e "s/__BACKEND_PORT__/${BACKEND_SCRAPE_PORT}/g" \
    -e "s/__BOT_PORT__/${BOT_SCRAPE_PORT}/g" \
    /etc/prometheus/prometheus.yml >"${CFG_OUT}"

exec /bin/prometheus \
  --config.file="${CFG_OUT}" \
  --storage.tsdb.path=/prometheus \
  --web.listen-address="0.0.0.0:${PORT}"
