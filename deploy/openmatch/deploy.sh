#!/usr/bin/env bash

set -eu
set -o pipefail

BASEDIR=$(dirname "$0")

kubectl create ns open-match
kubectl -n open-match apply -f ${BASEDIR}/open-match-services-nodeport.yaml

kubectl -n open-match create -f ${BASEDIR}/00-open-match-override-configmap.yaml \
  -f ${BASEDIR}/01-open-match-core.yml \
  -f ${BASEDIR}/07-open-match-default-evaluator.yaml \
  -f ${BASEDIR}/02-prometheus-chart.yaml  \
  -f ${BASEDIR}/03-grafana-chart.yaml


