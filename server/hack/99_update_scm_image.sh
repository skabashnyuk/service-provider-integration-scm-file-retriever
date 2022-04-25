#!/usr/bin/env bash
set -e
echo 'Updating SPI SCM image'


SPIS_TAG_NAME=$(git branch --show-current)'_'$(date '+%Y_%m_%d_%H_%M_%S')
make docker-build SPIS_TAG_NAME=$SPIS_TAG_NAME
minikube image load 'quay.io/redhat-appstudio/service-provider-integration-scm-file-retriever-server:'$SPIS_TAG_NAME
kubectl set image deployment/spi-system-file-retriever-server server=quay.io/redhat-appstudio/service-provider-integration-scm-file-retriever-server':'$SPIS_TAG_NAME -n spi-system


