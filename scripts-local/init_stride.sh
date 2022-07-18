#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# first, we need to create some saved state, so that we can copy to docker files
mkdir -p $STATE/$STRIDE_NODE_NAME

# then, we initialize our chains 
echo 'Initializing stride state...'

# initialize the chain
$STRIDE_CMD init test --chain-id $STRIDE_CHAIN --overwrite 2> /dev/null
# change the denom
sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${STRIDE_NODE_NAME}/config/genesis.json"
sed -i -E "s|timeout_commit = \"5s\"|timeout_commit = \"${BLOCK_TIME}\"|g" "${STATE}/${STRIDE_NODE_NAME}/config/config.toml"
sed -i -E "s|cors_allowed_origins = \[\]|cors_allowed_origins = [\"\*\"]|g" "${STATE}/${STRIDE_NODE_NAME}/config/config.toml"
# modify Stride epoch to be 3s
main_config=$STATE/$STRIDE_NODE_NAME/config/genesis.json
# NOTE: If you add new epochs, these indexes will need to be updated
jq '.app_state.epochs.epochs[$epochIndex].duration = $epochLen' --arg epochLen $DAY_EPOCH_LEN --argjson epochIndex $DAY_EPOCH_INDEX  $main_config > json.tmp && mv json.tmp $main_config
jq '.app_state.epochs.epochs[$epochIndex].duration = $epochLen' --arg epochLen $STRIDE_EPOCH_LEN --argjson epochIndex $STRIDE_EPOCH_INDEX $main_config > json.tmp && mv json.tmp $main_config
jq '.app_state.stakeibc.params.rewards_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
jq '.app_state.stakeibc.params.delegate_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
jq '.app_state.stakeibc.params.deposit_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
jq '.app_state.stakeibc.params.redemption_rate_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
jq '.app_state.stakeibc.params.reinvest_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
jq '.app_state.stakeibc.params.withdrawal_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config

# add validator account
echo $STRIDE_VAL_MNEMONIC | $STRIDE_CMD keys add $STRIDE_VAL_ACCT --recover --keyring-backend=test 
# get validator address
val_addr=$($STRIDE_CMD keys show $STRIDE_VAL_ACCT --keyring-backend test -a) > /dev/null
# add money for this validator account
$STRIDE_CMD add-genesis-account ${val_addr} 500000000000ustrd
# actually set this account as a validator
$STRIDE_CMD gentx $STRIDE_VAL_ACCT 1000000000ustrd --chain-id $STRIDE_CHAIN --keyring-backend test 2> /dev/null

# source $SCRIPT_DIR/genesis.sh

# Add hermes relayer account
echo $HERMES_STRIDE_MNEMONIC | $STRIDE_CMD keys add $HERMES_STRIDE_ACCT --recover --keyring-backend=test 
HERMES_STRIDE_ADDRESS=$($STRIDE_CMD keys show $HERMES_STRIDE_ACCT --keyring-backend test -a)
# Give relayer account token balance
$STRIDE_CMD add-genesis-account ${HERMES_STRIDE_ADDRESS} 500000000000ustrd

# Add ICQ relayer account
echo $ICQ_STRIDE_MNEMONIC | $STRIDE_CMD keys add $ICQ_STRIDE_ACCT --recover --keyring-backend=test 
ICQ_STRIDE_ADDRESS=$($STRIDE_CMD keys show $ICQ_STRIDE_ACCT --keyring-backend test -a)
# Give relayer account token balance
$STRIDE_CMD add-genesis-account ${ICQ_STRIDE_ADDRESS} 500000000000ustrd

sed -i -E "s|snapshot-interval = 0|snapshot-interval = 300|g" "${STATE}/${STRIDE_NODE_NAME}/config/app.toml"

# Collect genesis transactions
$STRIDE_CMD collect-gentxs 2> /dev/null
