# Railway: Settings → Source → Root Directory = (repo root, empty)
#          Dockerfile Path = deployment/railway/prometheus/Dockerfile
# Service name in Railway must be: prometheus

FROM prom/prometheus:latest

COPY backend/deployment/prometheus/prometheus.railway.yml /etc/prometheus/prometheus.yml

EXPOSE 9090
