#!/bin/bash

echo "Starting demo workload client installation"

echo -ne "Enter server ip: " 

read server_ip

sed -i "s/192.168.0.158/$server_ip/g" examples/client.yaml

kubectl apply -f examples/client.yaml

echo "Waiting for demo worklaod client to start."

kubectl wait --for=jsonpath='{.status.phase}'=Running -n client po -l app=feeder-client

echo "demo workload client up and running"

