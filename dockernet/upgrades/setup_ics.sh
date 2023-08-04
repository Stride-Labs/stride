#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

#======== 1. ADD CONSUMER ========
PROVIDER_HOME="$DOCKERNET_HOME/state/${GAIA_NODE_PREFIX}1"
CONSUMER_HOME_PREFIX="$DOCKERNET_HOME/state/${STRIDE_NODE_PREFIX}"
CONSUMER_HOME="${CONSUMER_HOME_PREFIX}1"
SOVEREIGN_HOME="$DOCKERNET_HOME/state/sovereign"
PROVIDER_BINARY=$GAIA_BINARY
PROVIDER_CHAIN_ID=$GAIA_CHAIN_ID
PROVIDER_RPC_ADDR="localhost:$GAIA_RPC_PORT"
VALIDATOR1="${GAIA_VAL_PREFIX}1"
VALIDATOR2="${GAIA_VAL_PREFIX}2"
VALIDATOR3="${GAIA_VAL_PREFIX}3"
DENOM=$ATOM_DENOM
PROVIDER_MAIN_CMD="$PROVIDER_BINARY --home $PROVIDER_HOME"
SOVEREIGN_CHAIN_ID=$STRIDE_CHAIN_ID
UPGRADE_HEIGHT="${UPGRADE_HEIGHT:-150}"
REVISION_HEIGHT=$((UPGRADE_HEIGHT + 3))
NUM_NODES=$(GET_VAR_VALUE   STRIDE_NUM_NODES)

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
    "unbonding_period": 240000000000
}
EOF

PROPOSAL_ID=1

printf "PROPOSAL\n"
$PROVIDER_MAIN_CMD tx gov submit-proposal consumer-addition $PROVIDER_HOME/consumer-proposal.json \
	--gas=100000000 --chain-id $PROVIDER_CHAIN_ID --node tcp://$PROVIDER_RPC_ADDR \
  --from $VALIDATOR1 --home $PROVIDER_HOME --keyring-backend test -b block -y | TRIM_TX

sleep 5
printf "\nVOTING\n"
# Vote yes to proposal
$PROVIDER_MAIN_CMD query gov proposals --node tcp://$PROVIDER_RPC_ADDR
$PROVIDER_MAIN_CMD tx gov vote 1 yes --from $VALIDATOR1 --chain-id $PROVIDER_CHAIN_ID --node tcp://$PROVIDER_RPC_ADDR --home $PROVIDER_HOME -b block -y --keyring-backend test
$PROVIDER_MAIN_CMD tx gov vote 1 yes --from $VALIDATOR2 --chain-id $PROVIDER_CHAIN_ID --node tcp://$PROVIDER_RPC_ADDR --home $PROVIDER_HOME -b block -y --keyring-backend test
$PROVIDER_MAIN_CMD tx gov vote 1 yes --from $VALIDATOR3 --chain-id $PROVIDER_CHAIN_ID --node tcp://$PROVIDER_RPC_ADDR --home $PROVIDER_HOME -b block -y --keyring-backend test

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
        sleep 3
        break
    elif [[ "$status" == "PROPOSAL_STATUS_REJECTED" ]]; then
        echo "Proposal Failed!"
        exit 1
    else 
        echo "Unknown proposal status: $status"
        exit 1
    fi
done

# Add ccv section to SOVEREIGN_HOME genesis to be used on upgrade handler
mkdir -p "$SOVEREIGN_HOME"/config
if ! $PROVIDER_MAIN_CMD q provider consumer-genesis "$SOVEREIGN_CHAIN_ID" --output json > "$SOVEREIGN_HOME"/consumer_section.json; 
then
       echo "Failed to get consumer genesis for the chain-id '$SOVEREIGN_CHAIN_ID'! Finalize genesis failed. For more details please check the log file in output directory."
       exit 1
fi

# This portion needs to be enabled for only above gaia v9.1.0(321d15a574def0f338ceacc5c060159ebba95edc)
# Path to the JSON file
json_file="$SOVEREIGN_HOME"/consumer_section.json

# Use jq to remove the "field2" key from the JSON file
jq 'del(.params.reward_denoms, .params.provider_reward_denoms)' "$json_file" > "$json_file.tmp"

# Replace the original file with the modified version
mv "$json_file.tmp" "$json_file"

cp $CONSUMER_HOME/config/genesis.json "$SOVEREIGN_HOME"/config/genesis.json
jq -s '.[0].app_state.ccvconsumer = .[1] | .[0]' "$SOVEREIGN_HOME"/config/genesis.json "$SOVEREIGN_HOME"/consumer_section.json > "$SOVEREIGN_HOME"/genesis_consumer.json && \
	mv "$SOVEREIGN_HOME"/genesis_consumer.json "$CONSUMER_HOME"/config/ccv.json

# Modify genesis params
jq ".app_state.ccvconsumer.params.blocks_per_distribution_transmission = \"70\" | .app_state.tokenfactory.paused = { \"paused\": false }" \
  $CONSUMER_HOME/config/ccv.json > \
   $SOVEREIGN_HOME/edited_genesis.json && mv $SOVEREIGN_HOME/edited_genesis.json $CONSUMER_HOME/config/ccv.json

for (( i=2; i <= $NUM_NODES; i++ )); do
    cp $CONSUMER_HOME/config/ccv.json "${CONSUMER_HOME_PREFIX}${i}"/config/ccv.json
done



#======== 2. CREATE ICS CONNECTION AFTER CHANGEOVER ========
echo "Waiting for the upgrade..."
BLOCK_NUM=$((REVISION_HEIGHT + 2))
WAIT_FOR_STRING $STRIDE_LOGS "height=$BLOCK_NUM module=txindex"

# Create new connections and channels for sharing voting power between two chains
relayer_logs=${LOGS}/relayer-gaia-ics.log
relayer_exec=$(GET_VAR_VALUE RELAYER_GAIA_ICS_EXEC)
relayer_config=$STATE/relayer-gaia-ics/config
mnemonic=$(GET_VAR_VALUE     RELAYER_GAIA_ICS_MNEMONIC)
chain_name=gaia
account_name=$(GET_VAR_VALUE RELAYER_GAIA_ICS_ACCT)
coin_type=$(GET_VAR_VALUE    COSMOS_COIN_TYPE)

mkdir -p $relayer_config
chmod -R 777 $STATE/relayer-gaia-ics
cp ${DOCKERNET_HOME}/config/relayer_config_ics.yaml $relayer_config/config.yaml

printf "STRIDE <> GAIA(ICS) - Adding relayer keys..."
$relayer_exec rly keys restore stride $RELAYER_STRIDE_ICS_ACCT "$mnemonic" >> $relayer_logs 2>&1
$relayer_exec rly keys restore $chain_name $account_name "$mnemonic" --coin-type $coin_type >> $relayer_logs 2>&1
echo "Done restoring relayer keys"

printf "STRIDE <> GAIA - Creating ICS channel..." | tee -a $relayer_logs
$relayer_exec rly transact link stride-gaia-ics --src-port consumer --dst-port provider --order ordered --version 1 >> $relayer_logs 2>&1

$DOCKER_COMPOSE up -d relayer-gaia-ics
$DOCKER_COMPOSE logs -f relayer-gaia-ics | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $relayer_logs 2>&1 &

printf "STRIDE <> GAIA - Registering reward denom to provider..."
val_addr=$($STRIDE_MAIN_CMD keys show ${STRIDE_VAL_PREFIX}1 --keyring-backend test -a | tr -cd '[:alnum:]._-')
$STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $val_addr 10000ustrd --from ${STRIDE_VAL_PREFIX}1 -y
WAIT_FOR_BLOCK $STRIDE_LOGS 5
$GAIA_MAIN_CMD tx provider register-consumer-reward-denom ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA --from ${GAIA_VAL_PREFIX}1 -y
