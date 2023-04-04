#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/../config.sh

CHAIN="$1"
HOST_ZONE_NUM="$2"

CONNECTION="connection-$HOST_ZONE_NUM"
CHANNEL="channel-$HOST_ZONE_NUM"

CHAIN_ID=$(GET_VAR_VALUE       ${CHAIN}_CHAIN_ID)
VAL_PREFIX=$(GET_VAR_VALUE     ${CHAIN}_VAL_PREFIX)
NODE_PREFIX=$(GET_VAR_VALUE    ${CHAIN}_NODE_PREFIX)
IBC_DENOM=$(GET_VAR_VALUE      IBC_${CHAIN}_CHANNEL_${HOST_ZONE_NUM}_DENOM)
HOST_DENOM=$(GET_VAR_VALUE     ${CHAIN}_DENOM)
ADDRESS_PREFIX=$(GET_VAR_VALUE ${CHAIN}_ADDRESS_PREFIX)
NUM_VALS=$(GET_VAR_VALUE       ${CHAIN}_NUM_NODES)

echo "$CHAIN - Registering host zone..."
$STRIDE_MAIN_CMD tx stakeibc register-host-zone \
    $CONNECTION $HOST_DENOM $ADDRESS_PREFIX $IBC_DENOM $CHANNEL 1 \
    --gas 1000000 --from $STRIDE_ADMIN_ACCT --home $DOCKERNET_HOME/state/stride1 -y | TRIM_TX
sleep 10

echo "$CHAIN - Registering validators..."
# Build array of validators of the form:
# {"name": "...", "address": "...", "weight": "..."}
validators=()
weights=(5 10 5 10 5) # alternate weights across vals
for (( i=1; i <= $NUM_VALS; i++ )); do
    delegate_val=$(GET_VAL_ADDR $CHAIN $i)
    weight=${weights[$((i-1))]}

    validator="{\"name\":\"${VAL_PREFIX}${i}\",\"address\":\"$delegate_val\",\"weight\":$weight}"
    if [[ "$i" != $NUM_VALS ]]; then
        validator="${validator},"
    fi
    validators+=("$validator")
done

# Write validators list to json file  of the form:
# {"validators": [{"name": "...", "address": "...", "weight": "..."}, {"name": ... }] }
validator_json=$DOCKERNET_HOME/state/${NODE_PREFIX}1/validators.json
echo "{\"validators\": [${validators[*]}]}" > $validator_json

# Add host zone validators to Stride's host zone struct
$STRIDE_MAIN_CMD tx stakeibc add-validators $CHAIN_ID $validator_json \
    --from $STRIDE_ADMIN_ACCT -y | TRIM_TX
sleep 5

# Confirm the ICA accounts have been registered before continuing
timeout=100
while true; do
    if ! $STRIDE_MAIN_CMD q stakeibc show-host-zone $CHAIN_ID | grep Account | grep -q null; then
        break
    else
        if [[ "$timeout" == "0" ]]; then 
            echo "ERROR - Unable to register host zones."
            exit 1
        fi
        timeout=$((timeout-1))
        sleep 1
    fi
done

echo "Done"