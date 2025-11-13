#!/usr/bin/env bash
set -e

aws eks update-kubeconfig --region "${AWS_DEFAULT_REGION}" --name justalternate-eks-cluster 

# Uninstall Helm releases
echo "Removing all Helm releases..."
helm uninstall promtail -n monitoring || true
helm uninstall loki -n monitoring || true
helm uninstall kube-prometheus-stack -n monitoring || true

sleep 10

kubectl delete pvc --all -n monitoring || true

echo "Tofu destroy..."
tofu -chdir=./iac destroy -auto-approve
