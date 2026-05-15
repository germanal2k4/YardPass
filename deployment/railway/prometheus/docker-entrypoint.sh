#!/bin/sh
set -e

# Railway routes healthchecks and traffic to $PORT; Prometheus defaults to 9090.
: "${PORT:=9090}"

exec /bin/prometheus \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/prometheus \
  --web.listen-address="0.0.0.0:${PORT}"
