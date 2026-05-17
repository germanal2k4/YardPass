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

if [ -z "${NGINX_RESOLVER_DIRECTIVE}" ]; then
  if [ -n "${NGINX_RESOLVER}" ]; then
    NGINX_RESOLVER_DIRECTIVE="resolver ${NGINX_RESOLVER} valid=10s;"
  else
    # Parse /etc/resolv.conf for nameserver entries (covers k8s CoreDNS, Railway
    # private DNS, plain compose). Wrap IPv6 in [] for nginx resolver syntax.
    _ns=$(awk '/^nameserver[[:space:]]/ {
      ip=$2
      if (index(ip, ":") > 0) printf "[%s] ", ip
      else printf "%s ", ip
    }' /etc/resolv.conf)
    if [ -n "${_ns}" ]; then
      _ipv6_flag=""
      case "${_ns}" in *"["*) _ipv6_flag=" ipv6=on" ;; esac
      NGINX_RESOLVER_DIRECTIVE="resolver ${_ns}${_ipv6_flag} valid=10s;"
    else
      NGINX_RESOLVER_DIRECTIVE=""
    fi
  fi
fi
export NGINX_RESOLVER_DIRECTIVE

echo "nginx: PORT=${PORT} BACKEND_URL=${BACKEND_URL}"
echo "nginx: ${NGINX_RESOLVER_DIRECTIVE:-<no resolver>}"

envsubst '${BACKEND_URL} ${PORT} ${NGINX_RESOLVER_DIRECTIVE}' < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf

exec nginx -g 'daemon off;'
