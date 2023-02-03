#!/bin/bash

echo "Starting spire-server installation"

kubectl create ns spire

kubectl -n spire apply -f deployment/server/.

echo "Waiting for server to start."

kubectl wait --for=jsonpath='{.status.phase}'=Running -n spire po -l app=spire-server

echo "Server up and running on IP $(kubectl -n spire get svc --no-headers | awk '{print $4}')"

echo "Generating Join Tokens for agents..."

kubectl -n spire exec spire-server-0 -- /opt/spire/bin/spire-server token generate -ttl 21600 -spiffeID  spiffe://accuknox.com/agent/saas


kubectl -n spire exec spire-server-0 -- /opt/spire/bin/spire-server token generate -ttl 21600 -spiffeID  spiffe://accuknox.com/agent/operator


echo "Generating entries for workloads..."

kubectl -n spire exec spire-server-0 -- /opt/spire/bin/spire-server entry create -parentID spiffe://accuknox.com/agent/saas -spiffeID spiffe://accuknox.com/knoxgrpc -selector k8s:ns:server -selector k8s:pod-label:app:knoxgrpc

kubectl -n spire exec spire-server-0 -- /opt/spire/bin/spire-server entry create -parentID spiffe://accuknox.com/agent/operator -spiffeID spiffe://accuknox.com/feeder-client -selector k8s:ns:client -selector k8s:pod-label:app:feeder-client
