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

sleep 60

echo "Creating connection STRIDE <> GAIA"
hermes create connection STRIDE GAIA

echo "Creating transfer channel"
hermes create channel --port-a transfer --port-b transfer GAIA connection-0 

hermes start 