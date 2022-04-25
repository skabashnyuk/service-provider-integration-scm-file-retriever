#!/usr/bin/env bash
set -e
echo 'Testing SPIAccessTokenBinding'


cat <<EOF | kubectl apply -n spi-system -f -
apiVersion: appstudio.redhat.com/v1beta1
kind: SPIAccessTokenBinding
metadata:
  name: acctoken-binding
spec:
  permissions:
    required:
      - type: r
        area: repository
      - type: w
        area: repository
  repoUrl: https://github.com/redhat-appstudio/service-provider-integration-operator
  secret:
    name: token-secret
    type: kubernetes.io/basic-auth
EOF
