#!/usr/bin/env bash
set -e

# deploy
kubectl apply -f ./services/api
kubectl apply -f ./services/web-server
kubectl apply -f ./services/service-monitor.yaml
