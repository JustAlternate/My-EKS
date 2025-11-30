#!/usr/bin/env bash
set -e

# deploy
kubectl apply -f ./services/service-monitor.yaml
kubectl apply -f ./services/cluster-secret-store.yaml
kubectl apply -f ./services/api
kubectl apply -f ./services/web-server
