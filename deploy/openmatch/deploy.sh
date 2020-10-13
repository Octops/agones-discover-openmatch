#!/usr/bin/env bash

set -eu
set -o pipefail

BASEDIR=$(dirname "$0")

kubectl create ns open-match
kubectl -n open-match create -f ${BASEDIR}/

