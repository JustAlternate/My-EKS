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

kubectl get svc --all-namespaces -o json | jq -r '.items[] | select(.spec.type=="LoadBalancer") | .metadata.name + " -n " + .metadata.namespace' | xargs -L 1 kubectl delete svc

kubectl delete ingress --all --all-namespaces

echo "Tofu destroy..."
tofu -chdir=./iac destroy -auto-approve
