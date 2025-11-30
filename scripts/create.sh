#!/usr/bin/env bash
set -e

CLUSTER_NAME="justalternate-eks-cluster"

echo "================================"
echo "Starting environment creation..."
echo "Using cluster: ${CLUSTER_NAME}, region: ${AWS_DEFAULT_REGION}"
echo "================================"

echo "🧱 Applying infrastructure (OpenTofu)..."
tofu -chdir=./iac init -upgrade -input=false
tofu -chdir=./iac apply -auto-approve
echo "✅ Infrastructure applied."

echo "🔗 Updating kubeconfig..."
aws eks update-kubeconfig --region "${AWS_DEFAULT_REGION}" --name "${CLUSTER_NAME}"
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

echo "⏳ Waiting for EKS nodes to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s

echo "🔐 Installing External Secrets Operator..."
helm repo add external-secrets https://charts.external-secrets.io
helm repo update

helm upgrade --install external-secrets \
  external-secrets/external-secrets \
  --namespace external-secrets \
  --create-namespace \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::${AWS_ACCOUNT_ID}:role/${CLUSTER_NAME}-eso-role" \
  --wait \
  --timeout 5m

echo "📊 Installing observability stack..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null 2>&1 || true
helm repo add grafana https://grafana.github.io/helm-charts >/dev/null 2>&1 || true
helm repo update >/dev/null 2>&1

kubectl get ns monitoring >/dev/null 2>&1 || kubectl create ns monitoring

echo "🧩 Creating Grafana dashboard ConfigMaps..."
kubectl create configmap my-rds-dashboard-cm \
  --from-file=my-dashboard.json=./observability-stack-config/dashboards/RDS.json \
  -n monitoring --dry-run=client -o yaml | kubectl apply -f -
kubectl label configmap my-rds-dashboard-cm grafana_dashboard=1 -n monitoring --overwrite

kubectl create configmap my-microservices-dashboard-cm \
  --from-file=my-dashboard.json=./observability-stack-config/dashboards/microservices.json \
  -n monitoring --dry-run=client -o yaml | kubectl apply -f -
kubectl label configmap my-microservices-dashboard-cm grafana_dashboard=1 -n monitoring --overwrite

echo "📦 Installing kube-prometheus-stack..."
helm upgrade --install kube-prometheus-stack \
  prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --values ./observability-stack-config/prometheus-stack-values.yaml \
  --set grafana.serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::${AWS_ACCOUNT_ID}:role/grafana-cloudwatch-role" \
  --wait --timeout 10m

# echo "📦 Installing Loki..."
# helm upgrade --install loki \
#   grafana/loki \
#   --namespace monitoring \
#   --values ./observability-stack-config/loki-values.yaml \
#   --wait --timeout 10m
#
# echo "📦 Installing Promtail..."
# helm upgrade --install promtail \
#   grafana/promtail \
#   --namespace monitoring \
#   --values ./observability-stack-config/promtail-values.yaml \
#   --wait --timeout 5m

echo "✅ Observability stack installation finished."

echo "🎨 Access Grafana UI"
cat <<EOF
==============================================================
To access Grafana locally:
  kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80 &
  firefox http://localhost:3000

Credentials:
  User: admin
  Pass: admin
==============================================================
EOF

echo "✅ Environment creation completed successfully!"
