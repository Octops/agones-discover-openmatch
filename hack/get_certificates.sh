#!/usr/bin/env bash

NAMESPACE=default # replace with any namespace
#EXTERNAL_IP=$(kubectl get services agones-allocator -n agones-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
EXTERNAL_IP=$(kubectl get services agones-allocator -n agones-system -o jsonpath='{.spec.loadBalancerIP}')
KEY_FILE=${CERTS_PATH}/client.key
CERT_FILE=${CERTS_PATH}/client.crt
TLS_CA_FILE=${CERTS_PATH}/ca.crt
TLS_FILE=${CERTS_PATH}/tls.crt

# allocator-client.default secret is created only when using helm installation. Otherwise generate the client certificate and replace the following.
# In case of MacOS replace "base64 -d" with "base64 -D"
kubectl get secret allocator-client.default -n "${NAMESPACE}" -ojsonpath="{.data.tls\.crt}" | base64 -d > "${CERT_FILE}"
kubectl get secret allocator-client.default -n "${NAMESPACE}" -ojsonpath="{.data.tls\.key}" | base64 -d > "${KEY_FILE}"
kubectl get secret allocator-tls-ca -n agones-system -ojsonpath="{.data.tls-ca\.crt}" | base64 -d > "${TLS_CA_FILE}"
kubectl get secret allocator-tls -n agones-system -ojsonpath="{.data.tls\.crt}" | base64 -d > "${TLS_FILE}"

#go run examples/allocator-client/main.go --ip ${EXTERNAL_IP} \
#    --port 443 \
#    --namespace ${NAMESPACE} \
#    --key ${KEY_FILE} \
#    --cert ${CERT_FILE} \
#    --cacert ${TLS_CA_FILE}
