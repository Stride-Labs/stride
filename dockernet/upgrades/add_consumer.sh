#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

PROVIDER_HOME="$DOCKERNET_HOME/state/${GAIA_NODE_PREFIX}1"
PROVIDER_BINARY=$GAIA_BINARY
PROVIDER_CHAIN_ID=$GAIA_CHAIN_ID
PROVIDER_RPC_ADDR="localhost:$GAIA_RPC_PORT"
VALIDATOR="${GAIA_VAL_PREFIX}1"
DENOM=$ATOM_DENOM
PROVIDER_MAIN_CMD="$PROVIDER_BINARY --home $PROVIDER_HOME"
SOVEREIGN_CHAIN_ID=$STRIDE_CHAIN_ID
REVISION_HEIGHT=10000

# Build consumer chain proposal file - unbonding period 21 days
tee $PROVIDER_HOME/consumer-proposal.json<<EOF
{
    "title": "Create a chain",
    "description": "Gonna be a great chain",
    "chain_id": "$SOVEREIGN_CHAIN_ID",
    "initial_height": {
        "revision_number": 0,
        "revision_height": $REVISION_HEIGHT
    },
    "genesis_hash": "519df96a862c30f53e67b1277e6834ab4bd59dfdd08c781d1b7cf3813080fb28",
    "binary_hash": "09184916f3e85aa6fa24d3c12f1e5465af2214f13db265a52fa9f4617146dea5",
    "spawn_time": "2022-06-01T09:10:00.000000000-00:00", 
    "deposit": "10000001$DENOM",
    "consumer_redistribution_fraction": "0.75",
    "blocks_per_distribution_transmission": 1000,
    "ccv_timeout_period": 2419200000000000,
    "transfer_timeout_period": 3600000000000,
    "historical_entries": 10000,
    "unbonding_period": 1814400000000000
}
EOF

PROPOSAL_ID=1

printf "PROPOSAL\n"
$PROVIDER_MAIN_CMD tx gov submit-proposal consumer-addition $PROVIDER_HOME/consumer-proposal.json \
	--gas=100000000 --chain-id $PROVIDER_CHAIN_ID --node tcp://$PROVIDER_RPC_ADDR \
  --from $VALIDATOR --home $PROVIDER_HOME --keyring-backend test -b block -y | TRIM_TX

sleep 5
printf "\nVOTING\n"
# Vote yes to proposal
$PROVIDER_MAIN_CMD query gov proposals --node tcp://$PROVIDER_RPC_ADDR
$PROVIDER_MAIN_CMD tx gov vote 1 yes --from $VALIDATOR --chain-id $PROVIDER_CHAIN_ID --node tcp://$PROVIDER_RPC_ADDR --home $PROVIDER_HOME -b block -y --keyring-backend test

sleep 5
printf "\nVOTE CONFIRMATION\n"
echo "$PROVIDER_MAIN_CMD query gov tally $PROPOSAL_ID"
$PROVIDER_MAIN_CMD query gov tally $PROPOSAL_ID

printf "\nPROPOSAL STATUS\n"
while true; do
    status=$($PROVIDER_MAIN_CMD query gov proposal $PROPOSAL_ID | grep "status" | awk '{printf $2}')
    if [[ "$status" == "PROPOSAL_STATUS_VOTING_PERIOD" ]]; then
        echo "Proposal still in progress..."
        sleep 5
    elif [[ "$status" == "PROPOSAL_STATUS_PASSED" ]]; then
        echo "Proposal passed!"
        exit 0
    elif [[ "$status" == "PROPOSAL_STATUS_REJECTED" ]]; then
        echo "Proposal Failed!"
        exit 1
    else 
        echo "Unknown proposal status: $status"
        exit 1
    fi
done
