#!/bin/bash
set -euo pipefail

# Deploy YardPass to Kubernetes (docker) with Istio.
# Postgres + Redis запускаются на хосте через docker-compose.data.yml,
# K8s обращается к ним через host.minikube.internal.
#
# Usage:
#   k8s-deploy.sh            — full deploy (default = up)
#   k8s-deploy.sh up         — start data services + minikube + build + helmfile
#   k8s-deploy.sh down       — tear down helm releases + stop data services
#   k8s-deploy.sh rebuild    — rebuild images + redeploy app charts
#   k8s-deploy.sh clean      — delete minikube cluster + data services
#   k8s-deploy.sh data       — start Postgres + Redis only
#   k8s-deploy.sh status     — show cluster and data service status

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# shellcheck disable=SC1091
[[ -f .env ]] && set -a && source .env && set +a

BACKEND_IMAGE="yardpass/backend:latest"
BOT_IMAGE="yardpass/bot:latest"
FRONTEND_IMAGE="yardpass/frontend:latest"

log()  { echo "==> $1"; }
err()  { echo "==> ERROR: $1" >&2; }

check_deps() {
    local missing=()
    for cmd in docker minikube kubectl helm helmfile; do
        command -v "$cmd" &>/dev/null || missing+=("$cmd")
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        err "Missing required tools: ${missing[*]}"
        exit 1
    fi
}

ensure_minikube() {
    if ! minikube status &>/dev/null; then
        log "Starting minikube..."
        minikube start --cpus=4 --memory=8g
    else
        log "Minikube already running"
    fi
}

build_images() {
    log "Building backend image (target=api)..."
    docker build --target api -t "$BACKEND_IMAGE" ./backend

    log "Building bot image (target=bot)..."
    docker build --target bot -t "$BOT_IMAGE" ./backend

    log "Building frontend image..."
    docker build --build-arg VITE_API_BASE_URL="" -t "$FRONTEND_IMAGE" ./frontend
}

load_images() {
    log "Loading images into minikube..."
    if command -v minikube &>/dev/null && minikube status &>/dev/null; then
        minikube image load "$BACKEND_IMAGE"
        minikube image load "$BOT_IMAGE"
        minikube image load "$FRONTEND_IMAGE"
    fi
}

start_data_services() {
    log "Starting Postgres + Redis (docker-compose.data.yml)..."
    "$SCRIPT_DIR/compose-db.sh" up
}

stop_data_services() {
    log "Stopping Postgres + Redis..."
    "$SCRIPT_DIR/compose-db.sh" down 2>/dev/null || true
}

helmfile_sync() {
    log "Running helmfile sync..."
    helmfile -f helmfile.yaml.gotmpl sync
}

deploy_app_only() {
    log "Redeploying application charts..."
    helmfile -f helmfile.yaml.gotmpl -l name=yardpass-backend -l name=yardpass-frontend sync
}

show_status() {
    echo ""
    log "Data services (Postgres + Redis):"
    "$SCRIPT_DIR/compose-db.sh" status 2>/dev/null || echo "  compose status unavailable"
    echo ""

    log "Pods - istio-system:"
    kubectl get pods -n istio-system 2>/dev/null || true
    echo ""

    log "Pods - yardpass:"
    kubectl get pods -n yardpass 2>/dev/null || true
    echo ""

    log "Services - yardpass:"
    kubectl get svc -n yardpass 2>/dev/null || true
}

show_access_info() {
    local minikube_ip
    minikube_ip=$(minikube ip 2>/dev/null || echo '<ingress-ip>')

    cat <<EOF

==> Deployment complete!

Access:
  Add to /etc/hosts: $minikube_ip yardpass.example.com

  # Application (via Istio Ingress Gateway)
  kubectl port-forward -n istio-system svc/istio-ingress 3000:80
  Open: http://yardpass.example.com:3000

  # Grafana
  kubectl port-forward -n istio-system svc/prometheus-grafana 3030:80
  Open: http://localhost:3030

  # Kiali
  kubectl port-forward -n istio-system svc/kiali 20001:20001
  Open: http://localhost:20001

EOF
}

cmd_up() {
    check_deps
    start_data_services
    ensure_minikube
    build_images
    load_images
    helmfile_sync
    show_status
    show_access_info
}

cmd_down() {
    log "Tearing down all Helm releases..."
    helmfile -f helmfile.yaml.gotmpl destroy 2>/dev/null || true

    log "Deleting namespaces..."
    kubectl delete namespace yardpass --ignore-not-found --timeout=60s 2>/dev/null || true

    log "Cleaning up CRDs and webhooks..."
    kubectl delete validatingwebhookconfiguration istiod-default-validator --ignore-not-found 2>/dev/null || true

    stop_data_services

    log "Teardown complete"
}

cmd_rebuild() {
    check_deps
    build_images
    load_images
    deploy_app_only

    log "Restarting deployments..."
    kubectl rollout restart deployment -n yardpass 2>/dev/null || true
    show_status
}

cmd_clean() {
    log "Deleting minikube cluster..."
    minikube delete 2>/dev/null || true

    stop_data_services

    log "Pruning docker images..."
    docker image prune -f 2>/dev/null || true

    log "Cleaning helm cache..."
    rm -rf ~/.cache/helm/repository/*

    log "Full cleanup complete. Run '$(basename "$0") up' to start fresh."
}

cmd_data() {
    start_data_services
}

case "${1:-up}" in
    up)      cmd_up ;;
    down)    cmd_down ;;
    rebuild) cmd_rebuild ;;
    clean)   cmd_clean ;;
    data)    cmd_data ;;
    images)  build_images && load_images ;;
    status)  show_status ;;
    -h|--help|help)
        cat <<EOF
Usage: $(basename "$0") [COMMAND]

Commands:
  up        Full deploy: data services + minikube + build + helmfile sync (default)
  down      Tear down: delete helm releases + stop data services
  rebuild   Rebuild images and redeploy app charts only
  clean     Full cleanup: delete minikube + data services
  data      Start Postgres + Redis only
  images    Build and load images only
  status    Show cluster and data service status
EOF
        ;;
    *)
        err "Unknown command: $1"
        exit 1
        ;;
esac
