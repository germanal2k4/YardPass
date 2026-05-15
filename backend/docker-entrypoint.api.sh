#!/bin/sh
set -e

# Railway injects PORT; app reads SERVER_PORT.
if [ -n "${PORT}" ] && [ -z "${SERVER_PORT}" ]; then
  export SERVER_PORT="${PORT}"
fi

: "${SERVER_HOST:=0.0.0.0}"
export SERVER_HOST

if [ -z "${JWT_SECRET}" ]; then
  echo "ERROR: JWT_SECRET is required" >&2
  exit 1
fi

if [ -z "${DATABASE_URL}" ]; then
  echo "ERROR: DATABASE_URL is required" >&2
  exit 1
fi

echo "api: listen ${SERVER_HOST}:${SERVER_PORT:-8080}"

./migrate up
exec ./api -c config.yaml
