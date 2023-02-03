#!/bin/bash

echo "Starting SaaS spire-agent installation"

echo -ne "Enter the Join Token: " 

read join_token

sed -i "s/fe7caa1f-52f1-4278-a37d-03d7b084d606/$join_token/g" deployment/agent/agent-configmap.yaml

kubectl apply -f deployment/agent/.

echo "Waiting for agent to start."

kubectl wait --for=jsonpath='{.status.phase}'=Running -n spire -l app=spire-agent

echo "Agent up and running"

echo "Bringing up logs"

kubectl -n spire logs logs -f -lapp=spire-agent