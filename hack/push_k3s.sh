#!/usr/bin/env bash

set -eu
set -o pipefail

TAG=$(git rev-parse --short HEAD)
make docker
docker save -o ${DOCKER_IMAGE_TAR_PATH} ${DOCKER_IMAGE_NAME}:${TAG}
scp ${DOCKER_IMAGE_TAR_PATH} ${REMOTE_SSH}:./
ssh ${REMOTE_SSH} docker load -i ./${DOCKER_IMAGE_TAR_NAME}
#kubectl set image deployment/agones-inmemory-store agones-store=${DOCKER_IMAGE_NAME}:${TAG}
envsubst < deploy/install.yaml | kubectl -n agones-store apply -f -