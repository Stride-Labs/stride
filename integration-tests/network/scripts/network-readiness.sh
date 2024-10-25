#!/bin/bash

set -eu

namespace="$1"

echo "---" && sleep 1
printf "Waiting for network to startup..."

# TODO: fix this
trap 'printf "\nInterrupted by user.\n"; exit' INT

elapsed=0
while true; do 
    not_ready_pods=$(kubectl get pods --no-headers -n $namespace | grep -v '1/1 *Running' | wc -l)

    if [[ $not_ready_pods -eq 0 ]]; then 
        printf "Ready! 🚀\n"
        break
    fi
    if [[ $elapsed -eq 60 ]]; then 
        printf "\nThe network's taking longer than expected to startup. Please investigate via `kubectl get pods`\n"
        exit 1
    fi

    sleep 2 && printf "."
    elapsed=$((elapsed + 1))
done