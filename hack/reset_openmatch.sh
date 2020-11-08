#!/usr/bin/env bash

kubectl -n open-match delete pod "$(kubectl -n open-match get pods -l role=master -o jsonpath="{.items[0].metadata.name}")" &
kubectl -n open-match delete pod "$(kubectl -n open-match get pods -l role=slave -o jsonpath="{.items[0].metadata.name}")" &
kubectl -n open-match delete pod "$(kubectl -n open-match get pods -l role=slave -o jsonpath="{.items[0].metadata.name}")" &
kubectl -n open-match delete pod "$(kubectl -n open-match get pods -l component=backend -o jsonpath="{.items[0].metadata.name}")" &
kubectl -n open-match delete pod "$(kubectl -n open-match get pods -l component=frontend -o jsonpath="{.items[0].metadata.name}")" &
kubectl -n open-match delete pod "$(kubectl -n open-match get pods -l component=query -o jsonpath="{.items[0].metadata.name}")" &
kubectl -n open-match delete pod "$(kubectl -n open-match get pods -l component=evaluator -o jsonpath="{.items[0].metadata.name}")" &
kubectl -n open-match delete pod "$(kubectl -n open-match get pods -l component=synchronizer -o jsonpath="{.items[0].metadata.name}")" &
