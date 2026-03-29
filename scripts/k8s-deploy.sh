#!/bin/bash
set -e

# Deploy YardPass to Kubernetes with Istio
# Prerequisites: kubectl, helm, helmfile, minikube or existing k8s cluster

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

echo "==> Building images..."
docker build -t yardpass/backend:latest ./backend
# For same-origin API in k8s (empty = use gateway host):
docker build -t yardpass/frontend:latest --build-arg VITE_API_BASE_URL="" ./frontend

echo "==> Loading images into minikube (if running)..."
if command -v minikube &>/dev/null && minikube status &>/dev/null; then
  minikube image load yardpass/backend:latest
  minikube image load yardpass/frontend:latest
fi

echo "==> Running helmfile sync..."
helmfile sync

echo ""
echo "==> Deployment complete. Check status:"
echo "  kubectl get pods -n yardpass"
echo "  kubectl get pods -n istio-system"
echo ""
echo "==> Access:"
echo "  Add to /etc/hosts: $(minikube ip 2>/dev/null || echo '<ingress-ip>') yardpass.example.com"
echo "  Grafana: kubectl port-forward -n istio-system svc/prometheus-grafana 3030:80"
echo "  Kiali:   kubectl port-forward -n istio-system svc/kiali 20001:20001"
echo "  App:     http://yardpass.example.com (after hosts + ingress)"
