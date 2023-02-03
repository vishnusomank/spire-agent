#!/bin/bash

echo "Sending hello to demo workload server"

kubectl -n client exec -it $(kubectl -n client get po -o name) -- feeder-client