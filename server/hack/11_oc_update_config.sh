#!/usr/bin/env bash
set -e
echo 'Deploying SPI OAuth2 config'

OAUTH_URL=$(oc get route/spi-oauth-route --namespace=spi-system -o=jsonpath={'.spec.host'})
tmpfile=/tmp/config.yaml

spiConfig=$(cat <<EOF

sharedSecret: $(openssl rand -hex 20)
serviceProviders:
  - type: GitHub
    clientId: $SPI_GITHUB_CLIENT_ID
    clientSecret: $SPI_GITHUB_CLIENT_SECRET
  - type: Quay
    clientId: $SPI_QUAY_CLIENT_ID
    clientSecret: $SPI_QUAY_CLIENT_SECRET
baseUrl: https://$OAUTH_URL

EOF
)

echo "Please go to https://github.com/settings/developers."
echo "And register new Github OAuth application for callback https://"$OAUTH_URL"/github/callback"

CONFIG_SECRET='oauth-config' #$(kubectl get secrets  -l app.kubernetes.io/part-of=service-provider-integration-operator  -n spi-system -o json | jq '.items[0].metadata.name' -r)
echo $CONFIG_SECRET
#kubectl delete secret/$CONFIG_SECRET -n spi-system
echo "$spiConfig" > "$tmpfile"
cat $tmpfile


kubectl create secret generic  $CONFIG_SECRET \
--save-config --dry-run=client \
--from-file="$tmpfile"  \
-o yaml |
kubectl apply -n spi-system  -f -


rm "$tmpfile"


kubectl rollout restart  deployment/spi-controller-manager  -n spi-system
kubectl rollout restart  deployment/spi-oauth-service  -n spi-system
