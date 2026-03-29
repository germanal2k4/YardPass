#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

MINIKUBE_PROFILE="${MINIKUBE_PROFILE:-minikube}"
BACKEND_IMAGE="localhost/yardpass/backend:latest"
FRONTEND_IMAGE="localhost/yardpass/frontend:latest"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}==> $1${NC}"; }
warn() { echo -e "${YELLOW}==> $1${NC}"; }
err()  { echo -e "${RED}==> ERROR: $1${NC}" >&2; }

usage() {
    cat <<EOF
Usage: $(basename "$0") [COMMAND]

Commands:
  up        Full deploy: start minikube, build images, helmfile sync (default)
  down      Tear down: delete namespaces and all helm releases
  rebuild   Rebuild images and redeploy app charts only
  clean     Full cleanup: delete minikube cluster entirely
  images    Build and load images only
  status    Show cluster and pod status

Environment variables:
  MINIKUBE_PROFILE  Minikube profile name (default: minikube)
EOF
}

check_deps() {
    local missing=()
    for cmd in podman minikube kubectl helm helmfile; do
        command -v "$cmd" &>/dev/null || missing+=("$cmd")
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        err "Missing required tools: ${missing[*]}"
        exit 1
    fi
}

ensure_minikube() {
    if ! minikube status -p "$MINIKUBE_PROFILE" &>/dev/null; then
        log "Starting minikube (driver=podman, runtime=cri-o)..."
        minikube start \
            -p "$MINIKUBE_PROFILE" \
            --driver=podman \
            --container-runtime=cri-o \
            --cpus=4 \
            --memory=8g \
            --disk-size=30g
    else
        log "Minikube already running"
    fi
}

build_images() {
    log "Building backend image (target=api)..."
    podman build --target api -t "$BACKEND_IMAGE" ./backend

    log "Building frontend image..."
    podman build --build-arg VITE_API_BASE_URL="" -t "$FRONTEND_IMAGE" ./frontend
}

load_images() {
    log "Saving images as tarballs..."
    local tmpdir
    tmpdir=$(mktemp -d)

    podman save -o "$tmpdir/backend.tar" "$BACKEND_IMAGE"
    podman save -o "$tmpdir/frontend.tar" "$FRONTEND_IMAGE"

    log "Loading images into minikube..."
    minikube -p "$MINIKUBE_PROFILE" image load "$tmpdir/backend.tar"
    minikube -p "$MINIKUBE_PROFILE" image load "$tmpdir/frontend.tar"

    rm -rf "$tmpdir"
    log "Images loaded"
}

helmfile_sync() {
    log "Running helmfile sync (environment=local)..."
    helmfile -f helmfile.yaml.gotmpl -e local sync
}

deploy_app_only() {
    log "Redeploying application charts..."
    helmfile -f helmfile.yaml.gotmpl -e local -l name=yardpass-backend -l name=yardpass-frontend sync
}

wait_for_pods() {
    local ns="$1"
    local timeout="${2:-120}"
    log "Waiting for pods in namespace $ns (timeout=${timeout}s)..."
    kubectl wait --for=condition=Ready pods --all -n "$ns" --timeout="${timeout}s" 2>/dev/null || {
        warn "Some pods not ready in $ns after ${timeout}s"
        kubectl get pods -n "$ns"
    }
}

show_status() {
    echo ""
    log "Cluster status:"
    minikube status -p "$MINIKUBE_PROFILE" 2>/dev/null || true
    echo ""

    log "Pods - istio-system:"
    kubectl get pods -n istio-system --no-headers 2>/dev/null | \
        awk '{printf "  %-50s %s\n", $1, $3}'
    echo ""

    log "Pods - yardpass:"
    kubectl get pods -n yardpass --no-headers 2>/dev/null | \
        awk '{printf "  %-50s %s\n", $1, $3}'
    echo ""

    log "Services - yardpass:"
    kubectl get svc -n yardpass --no-headers 2>/dev/null | \
        awk '{printf "  %-30s %s\n", $1, $5}'
}

show_access_info() {
    local minikube_ip
    minikube_ip=$(minikube ip -p "$MINIKUBE_PROFILE" 2>/dev/null || echo "<minikube-ip>")

    cat <<EOF

${GREEN}==> Deployment complete!${NC}

Access (run port-forwards in separate terminals):

  ${YELLOW}# Application (via Istio Ingress Gateway)${NC}
  kubectl port-forward -n istio-system svc/istio-ingress 3000:80
  Add to /etc/hosts: 127.0.0.1 yardpass.example.com
  Open: http://yardpass.example.com:3000

  ${YELLOW}# Grafana (dashboards, logs, metrics)${NC}
  kubectl port-forward -n istio-system svc/prometheus-grafana 3001:80
  Open: http://localhost:3001  (admin / admin)

  ${YELLOW}# Kiali (service mesh observability)${NC}
  kubectl port-forward -n istio-system svc/kiali 20001:20001
  Open: http://localhost:20001

  ${YELLOW}# Jaeger (distributed tracing)${NC}
  kubectl port-forward -n istio-system svc/jaeger-query 16686:16686
  Open: http://localhost:16686

  ${YELLOW}# Prometheus${NC}
  kubectl port-forward -n istio-system svc/prometheus-kube-prometheus-prometheus 9090:9090
  Open: http://localhost:9090

EOF
}

cmd_up() {
    check_deps
    ensure_minikube
    build_images
    load_images
    helmfile_sync
    wait_for_pods "istio-system" 180
    wait_for_pods "yardpass" 180
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

    log "Teardown complete"
}

cmd_rebuild() {
    check_deps
    build_images
    load_images
    deploy_app_only

    log "Restarting deployments..."
    kubectl rollout restart deployment -n yardpass 2>/dev/null || true
    wait_for_pods "yardpass" 120
    show_status
}

cmd_clean() {
    log "Deleting minikube cluster ($MINIKUBE_PROFILE)..."
    minikube delete -p "$MINIKUBE_PROFILE" 2>/dev/null || true

    log "Pruning podman images..."
    podman image prune -f 2>/dev/null || true

    log "Cleaning helm cache..."
    rm -rf ~/.cache/helm/repository/*

    log "Full cleanup complete. Run '$(basename "$0") up' to start fresh."
}

cmd_images() {
    check_deps
    ensure_minikube
    build_images
    load_images
    log "Images built and loaded"
}

case "${1:-up}" in
    up)      cmd_up ;;
    down)    cmd_down ;;
    rebuild) cmd_rebuild ;;
    clean)   cmd_clean ;;
    images)  cmd_images ;;
    status)  show_status ;;
    -h|--help|help) usage ;;
    *)
        err "Unknown command: $1"
        usage
        exit 1
        ;;
esac
