#!/bin/bash

echo "Starting demo workload server installation"

kubectl apply -f examples/server.yaml

echo "Waiting for demo worklaod server to start."

kubectl wait --for=jsonpath='{.status.phase}'=Running -n spire -l app=knoxgrpc

echo "demo workload server up and running on IP $(kubectl -n server get svc --no-headers | awk '{print $4}')"

echo "Getting logs of demo workload server"

kubectl -n server logs -f -l app=knoxgrpc