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

register_proposal_json=$DOCKERNET_HOME/state/${NODE_PREFIX}1/register_proposal.json
tee $register_proposal_json<<EOF
{
	"title": "Register gaia as a host zone",
    "description": "Proposal to register gaia as host zone.",
	"connection_id": "$CONNECTION",
	"bech32prefix": "$ADDRESS_PREFIX",
	"host_denom": "$HOST_DENOM",
	"ibc_denom": "$IBC_DENOM",
	"transfer_channel_id": "$CHANNEL",
	"unbonding_frequency": 1,
	"min_redemption_rate": "0.0",
	"max_redemption_rate": "0.0",
    "deposit": "10000001ustrd"
}
EOF

$STRIDE_MAIN_CMD tx gov submit-legacy-proposal register-host-zone $register_proposal_json \
  --gas 1000000 --from val1 --home $DOCKERNET_HOME/state/stride1 -y | TRIM_TX

sleep 5

proposal_id=$($STRIDE_MAIN_CMD q gov proposals | grep 'id: "' | tail -1 | awk '{printf $2}' | tr -d '"')

echo "VOTING $proposal_id"
$STRIDE_MAIN_CMD tx gov vote $proposal_id Yes --gas 1000000 --from val1 -y | TRIM_TX
$STRIDE_MAIN_CMD tx gov vote $proposal_id Yes --gas 1000000 --from val2 -y | TRIM_TX
$STRIDE_MAIN_CMD tx gov vote $proposal_id Yes --gas 1000000 --from val3 -y | TRIM_TX
sleep 35

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

add_val_proposal_json=$DOCKERNET_HOME/state/${NODE_PREFIX}1/add_validators_proposal.json
echo "{\"description\":\"Register $CHAIN_ID\", \"hostZone\":\"$CHAIN_ID\",\"validators\": [${validators[*]}],\"deposit\": \"10000001ustrd\"}" > $add_val_proposal_json
$STRIDE_MAIN_CMD tx gov submit-legacy-proposal add-validators $add_val_proposal_json --from val1 -y | TRIM_TX

sleep 5
proposal_id=$($STRIDE_MAIN_CMD q gov proposals | grep 'id: "' | tail -1 | awk '{printf $2}' | tr -d '"')
echo "VOTING $proposal_id"
$STRIDE_MAIN_CMD tx gov vote $proposal_id Yes --gas 1000000 --from val1 -y | TRIM_TX
$STRIDE_MAIN_CMD tx gov vote $proposal_id Yes --gas 1000000 --from val2 -y | TRIM_TX
$STRIDE_MAIN_CMD tx gov vote $proposal_id Yes --gas 1000000 --from val3 -y | TRIM_TX
sleep 35

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