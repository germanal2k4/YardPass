#!/bin/sh
set -e

# Railway routes healthchecks and traffic to $PORT; Prometheus defaults to 9090.
: "${PORT:=9090}"

# Host = Railway service slug (must match Networking / service name, not display title).
# Optional full overrides (recommended for polling bot: metrics bind to bot's $PORT, not 5050):
#   BACKEND_SCRAPE_TARGET=${{Backend.RAILWAY_PRIVATE_DOMAIN}}:8080
#   BOT_SCRAPE_TARGET=${{Bot.RAILWAY_PRIVATE_DOMAIN}}:${{Bot.PORT}}
: "${BACKEND_SCRAPE_HOST:=backend}"
: "${BACKEND_SCRAPE_PORT:=8080}"
: "${BOT_SCRAPE_HOST:=bot}"
: "${BOT_SCRAPE_PORT:=5050}"

if [ -n "${BACKEND_SCRAPE_TARGET:-}" ]; then
  BACKEND_TARGET="${BACKEND_SCRAPE_TARGET}"
else
  BACKEND_TARGET="${BACKEND_SCRAPE_HOST}.railway.internal:${BACKEND_SCRAPE_PORT}"
fi

if [ -n "${BOT_SCRAPE_TARGET:-}" ]; then
  BOT_TARGET="${BOT_SCRAPE_TARGET}"
else
  BOT_TARGET="${BOT_SCRAPE_HOST}.railway.internal:${BOT_SCRAPE_PORT}"
fi

CFG_OUT=/tmp/prometheus.railway.generated.yml
sed -e "s|__BACKEND_TARGET__|${BACKEND_TARGET}|g" \
    -e "s|__BOT_TARGET__|${BOT_TARGET}|g" \
    /etc/prometheus/prometheus.yml >"${CFG_OUT}"

echo "prometheus: scrape targets (generated)"
grep -E "^\s+-\s+targets:" -A1 "${CFG_OUT}" || true
echo "prometheus: yardpass-bot polling mode exposes metrics on the bot service PORT; set BOT_SCRAPE_TARGET or BOT_SCRAPE_PORT to match (see deployment/railway/env.example)."

exec /bin/prometheus \
  --config.file="${CFG_OUT}" \
  --storage.tsdb.path=/prometheus \
  --web.listen-address="0.0.0.0:${PORT}"
