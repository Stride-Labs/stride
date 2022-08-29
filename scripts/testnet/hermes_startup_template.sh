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
hermes create connection --a-chain STRIDE_CHAIN_ID --b-chain GAIA

echo "Creating transfer channel"
hermes create channel --a-chain GAIA --a-connection connection-0 --a-port transfer --b-port transfer

hermes start 