#!/bin/sh
set -e

# Railway: metrics HTTP must listen on $PORT if METRICS_PORT is not set (see applyRailwayBotMetricsPort in setup).
# Pin METRICS_PORT=5050 on the bot service to match default prometheus BOT_SCRAPE_PORT, or set BOT_SCRAPE_PORT on prometheus.
if [ -z "${METRICS_PORT}" ] && [ -n "${PORT}" ]; then
  export METRICS_PORT="${PORT}"
fi

echo "bot: METRICS_PORT=${METRICS_PORT:-5050} (Prometheus: bot.railway.internal:<BOT_SCRAPE_PORT>/metrics)"

exec ./bot -c config.yaml
