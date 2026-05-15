#!/bin/sh
set -e

# Railway injects $PORT; Grafana listens on GF_SERVER_HTTP_PORT.
: "${PORT:=3000}"
export GF_SERVER_HTTP_PORT="$PORT"

exec /run.sh
