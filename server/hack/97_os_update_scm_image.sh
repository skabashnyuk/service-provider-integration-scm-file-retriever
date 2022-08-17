#!/usr/bin/env bash
set -e
echo 'Updating SPI SCM image'


#REGISTRY=$(oc registry info)
#INTERNAL_REGISTRY=$(oc registry info --internal=true)
#TOKEN=$(oc whoami -t)
#CUR_USER=$(oc whoami)
#docker login -u $CUR_USER -p $TOKEN $REGISTRY

SPIS_TAG_NAME=$(git branch --show-current)'_'$(date '+%Y_%m_%d_%H_%M_%S')
SPIS_IMAGE_TAG_BASE="quay.io/skabashn/service-provider-integration-scm-file-retriever-server"
make docker-push SPIS_TAG_NAME=$SPIS_TAG_NAME SPIS_IMAGE_TAG_BASE=$SPIS_IMAGE_TAG_BASE
#docker tag "quay.io/redhat-appstudio/service-provider-integration-scm-file-retriever-server:"$SPIS_TAG_NAME $REGISTRY/"redhat-appstudio/service-provider-integration-scm-file-retriever-server:"$SPIS_TAG_NAME
#docker push $REGISTRY/$IMAGE
oc set image deployment/spi-system-file-retriever-server server=$SPIS_IMAGE_TAG_BASE':'$SPIS_TAG_NAME -n spi-system
