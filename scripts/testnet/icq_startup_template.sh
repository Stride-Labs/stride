#!/bin/sh

while true; do
    ping -c 1 STRIDE_ENDPOINT
    if [ "$?" == "0" ]; then 
        echo "Stride endpoint found!"
        break
    fi
    echo "Stride endpoint not available yet. Trying again in 30 seconds..."
    sleep 30
done

echo "Restoring ICQ Accounts"
echo "ICQ_STRIDE_MNEMONIC" | icq keys restore icq1 --chain stride
echo "ICQ_GAIA_MNEMONIC" | icq keys restore icq2 --chain gaia

echo "Starting ICQ..."
icq run