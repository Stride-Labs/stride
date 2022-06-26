#!/bin/bash

set -eu 

while true; do
    ping -c 1 STRIDE_ENDPOINT
    if [ "$?" == "0" ]; then 
        echo "Stride endpoint found."
        break
    fi
    echo "Stride endpoint not available yet. Trying again in 30 seconds..."
    sleep 30
done

echo "Restoring Hermes Accounts"
hermes keys restore -m "HERMES_STRIDE_MNEMONIC" STRIDE
hermes keys restore -m "HERMES_GAIA_MNEMONIC" GAIA

hermes start &
sleep 30

echo "Creating hermes identifiers"
hermes tx raw create-client STRIDE GAIA
sleep 15 

hermes tx raw conn-init STRIDE GAIA 07-tendermint-0 07-tendermint-0
sleep 15

echo "Creating connection STRIDE <> GAIA"
hermes create connection STRIDE GAIA
sleep 15

echo "Creating transfer channel"
hermes create channel --port-a transfer --port-b transfer GAIA connection-0 

wait