#!/usr/bin/env bash
set -e

CLUSTER_NAME="justalternate-eks-cluster"

echo "🔗 Updating kubeconfig..."
aws eks update-kubeconfig --region "${AWS_DEFAULT_REGION}" --name "${CLUSTER_NAME}"

echo "==================================="
echo "🧹 Cleaning up Kubernetes resources..."
echo "==================================="

echo "📦 Uninstalling Helm releases (monitoring stack)..."
helm uninstall promtail -n monitoring || true
helm uninstall loki -n monitoring || true
helm uninstall kube-prometheus-stack -n monitoring || true
helm uninstall external-secrets -n external-secrets || true

echo "⏳ Waiting for cleanup..."
sleep 10

echo "💾 Removing persistent volume claims..."
kubectl delete pvc --all -n monitoring --wait=false || true

echo "🗃️ Removing custom ConfigMaps..."
kubectl delete configmap my-rds-dashboard-cm -n monitoring || true
kubectl delete configmap my-microservices-dashboard-cm -n monitoring || true

echo "🌐 Removing services..."
kubectl delete svc --all -n monitoring || true
kubectl delete svc --all -n default || true

echo "🚪 Removing ingresses..."
kubectl delete ingress --all --all-namespaces || true

echo "🗑️ Deleting namespaces monitoring and external-secrets..."
kubectl delete ns monitoring --wait=false || true
kubectl delete ns external-secrets --wait=false || true

echo "⏳ Waiting for cleanup..."
sleep 10

echo "==================================="
echo "🧨 Destroying AWS infrastructure..."
echo "==================================="
tofu -chdir=./iac destroy -auto-approve

echo "===================================="
echo "✅ Cleanup complete!"
echo "🎉 All Kubernetes and AWS resources have been destroyed."
echo "===================================="
