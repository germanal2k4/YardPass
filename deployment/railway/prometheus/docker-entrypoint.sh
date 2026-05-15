#!/bin/sh
set -e

# Railway routes healthchecks and traffic to $PORT; Prometheus defaults to 9090.
: "${PORT:=9090}"

# Host = Railway service slug (must match Networking / service name, not display title).
: "${BACKEND_SCRAPE_HOST:=backend}"
: "${BACKEND_SCRAPE_PORT:=8080}"
: "${BOT_SCRAPE_HOST:=bot}"
: "${BOT_SCRAPE_PORT:=5050}"

BACKEND_TARGET="${BACKEND_SCRAPE_HOST}.railway.internal:${BACKEND_SCRAPE_PORT}"
BOT_TARGET="${BOT_SCRAPE_HOST}.railway.internal:${BOT_SCRAPE_PORT}"

CFG_OUT=/tmp/prometheus.railway.generated.yml
sed -e "s|__BACKEND_TARGET__|${BACKEND_TARGET}|g" \
    -e "s|__BOT_TARGET__|${BOT_TARGET}|g" \
    /etc/prometheus/prometheus.yml >"${CFG_OUT}"

echo "prometheus: scrape targets (generated)"
grep -E "^\s+-\s+targets:" -A1 "${CFG_OUT}" || true
echo "prometheus: if yardpass-bot is DOWN, set BOT_SCRAPE_HOST to your bot service slug and BOT_SCRAPE_PORT to its METRICS_PORT."

exec /bin/prometheus \
  --config.file="${CFG_OUT}" \
  --storage.tsdb.path=/prometheus \
  --web.listen-address="0.0.0.0:${PORT}"
