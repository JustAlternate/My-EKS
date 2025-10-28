#!/usr/bin/env bash

# Install all at once
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Prometheus + Grafana
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace \
  --values ./services/observability-stack/prometheus-stack-values.yaml

# Loki
helm install loki grafana/loki \
  --namespace monitoring \
  --values ./services/observability-stack/loki-values.yaml

# Promtail
helm install promtail grafana/promtail \
  --namespace monitoring \
  --values ./services/observability-stack/promtail-values.yaml

# Access Grafana
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
