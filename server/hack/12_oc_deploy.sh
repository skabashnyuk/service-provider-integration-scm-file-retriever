#!/usr/bin/env bash
set -e
echo 'Deploying scm server with spi operator and oauth service'

MY_PATH=$(dirname "$0")





kustomize build server/config/os | oc apply -f -
echo 'Vault init'
$MY_PATH/vault-init.sh
echo 'updating provider configuration'
$MY_PATH/11_oc_update_config.sh
echo 'restarting services '
oc rollout status deployment/spi-system-file-retriever-server  -n spi-system
oc rollout status deployment/spi-controller-manager  -n spi-system
oc rollout status deployment/spi-oauth-service  -n spi-system
