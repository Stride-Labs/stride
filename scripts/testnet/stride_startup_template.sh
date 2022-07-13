#!/bin/bash

while true; do
    ping -c 1 STRIDE_ENDPOINT
    if [ "$?" == "0" ]; then 
        echo "Stride endpoint found."
        break
    fi
    echo "Stride endpoint not available yet. Trying again in 30 seconds..."
    sleep 30
done

cosmovisor run start