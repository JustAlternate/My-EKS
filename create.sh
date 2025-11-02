#!/usr/bin/env bash
set -e

echo "Tofu apply..."
tofu -chdir ./iac apply -auto-approve

echo "Installing observability stack..."
aws eks update-kubeconfig --region "${AWS_DEFAULT_REGION}" --name justalternate-eks-cluster 

kubectl wait --for=condition=Ready nodes --all --timeout=300s

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update


AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)


helm upgrade --install kube-prometheus-stack \
  prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --version 65.0.0 \
  --values iac/observability-stack-config/prometheus-stack-values.yaml \
  --set grafana.serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::${AWS_ACCOUNT_ID}:role/grafana-cloudwatch-role" \
  --wait \
  --timeout 10m

helm upgrade --install loki \
  grafana/loki \
  --namespace monitoring \
  --version 6.6.0 \
  --values iac/observability-stack-config/loki-values.yaml \
  --wait \
  --timeout 10m

helm upgrade --install promtail \
  grafana/promtail \
  --namespace monitoring \
  --version 6.16.0 \
  --values iac/observability-stack-config/promtail-values.yaml \
  --wait \
  --timeout 5m

kubectl get pods -n monitoring

echo "Access Grafana:"
echo "===="
echo "kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80 &"
echo "firefox http://localhost:3000"
echo "===="
echo "  User: admin"
echo "  Pass: admin"
echo "===="
