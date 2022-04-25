#!/usr/bin/env bash
set -e
echo 'Deploying scm server with spi operator and oauth service'

kubectl apply -k  server/config/k8s

kubectl rollout status deployment/spi-system-file-retriever-server  -n spi-system
kubectl rollout status deployment/spi-controller-manager  -n spi-system
kubectl rollout status deployment/spi-oauth-service  -n spi-system
