#!/bin/sh
set -e

: "${PORT:=80}"
export PORT

if [ -z "${BACKEND_URL}" ]; then
  echo "ERROR: BACKEND_URL is not set (e.g. http://backend.railway.internal:8080)" >&2
  exit 1
fi

case "${BACKEND_URL}" in
  http://*|https://*) ;;
  *)
    echo "ERROR: BACKEND_URL must start with http:// or https:// (got: ${BACKEND_URL})" >&2
    exit 1
    ;;
esac

echo "nginx: PORT=${PORT} BACKEND_URL=${BACKEND_URL}"

envsubst '${BACKEND_URL} ${PORT}' < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf

exec nginx -g 'daemon off;'
