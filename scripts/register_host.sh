#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/vars.sh

CHAIN_ID="$1"
HOST_ZONE_NUM="$2"

CONNECTION="connection-$HOST_ZONE_NUM"
CHANNEL="channel-$HOST_ZONE_NUM"

MAIN_CMD=$(GET_VAR_VALUE       ${CHAIN_ID}_MAIN_CMD)
VAL_PREFIX=$(GET_VAR_VALUE     ${CHAIN_ID}_VAL_PREFIX)
IBC_DENOM=$(GET_VAR_VALUE      IBC_${CHAIN_ID}_CHANNEL_${HOST_ZONE_NUM}_DENOM)
HOST_DENOM=$(GET_VAR_VALUE     ${CHAIN_ID}_DENOM)
ADDRESS_PREFIX=$(GET_VAR_VALUE ${CHAIN_ID}_ADDRESS_PREFIX)

# Get validator addresses
DELEGATE_VAL_1="$($MAIN_CMD q staking validators | grep ${CHAIN_ID}_1 -A 5 | grep operator | awk '{print $2}')"
DELEGATE_VAL_2="$($MAIN_CMD q staking validators | grep ${CHAIN_ID}_2 -A 5 | grep operator | awk '{print $2}')"

echo "$CHAIN_ID - Registering host zone..."
$STRIDE_MAIN_CMD tx stakeibc register-host-zone \
    $CONNECTION $HOST_DENOM $ADDRESS_PREFIX $IBC_DENOM $CHANNEL 1 \
    --gas 1000000 --from $STRIDE_ADMIN_ACCT --home $SCRIPT_DIR/state/stride1 -y | grep -E "code:|txhash:" | sed 's/^/  /'
sleep 4

echo "$CHAIN_ID - Registering validators..."
$STRIDE_MAIN_CMD tx stakeibc add-validator $CHAIN_ID ${VAL_PREFIX}1 $DELEGATE_VAL_1 10 5 \
    --from $STRIDE_ADMIN_ACCT -y | grep -E "code:|txhash:" | sed 's/^/  /'
sleep 4

$STRIDE_MAIN_CMD tx stakeibc add-validator $CHAIN_ID ${VAL_PREFIX}2 $DELEGATE_VAL_2 10 10 \
    --from $STRIDE_ADMIN_ACCT -y | grep -E "code:|txhash:" | sed 's/^/  /'
sleep 4

while true; do
    if ! $STRIDE_MAIN_CMD q stakeibc list-host-zone | grep Account | grep -q null; then
        sleep 1
        break
    fi
done