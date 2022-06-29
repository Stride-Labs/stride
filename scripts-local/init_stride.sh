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
jq '.app_state.epochs.epochs[2].duration = $newVal' --arg newVal "3s" $main_config > json.tmp && mv json.tmp $main_config
jq '.app_state.epochs.epochs[1].duration = $newVal' --arg newVal "60s" $main_config > json.tmp && mv json.tmp $main_config

# add validator account
echo $STRIDE_VAL_MNEMONIC | $STRIDE_CMD keys add $STRIDE_VAL_ACCT --recover --keyring-backend=test 
# get validator address
val_addr=$($STRIDE_CMD keys show $STRIDE_VAL_ACCT --keyring-backend test -a) > /dev/null
# add money for this validator account
$STRIDE_CMD add-genesis-account ${val_addr} 500000000000ustrd
# actually set this account as a validator
$STRIDE_CMD gentx $STRIDE_VAL_ACCT 1000000000ustrd --chain-id $STRIDE_CHAIN --keyring-backend test 2> /dev/null

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

# Collect genesis transactions
$STRIDE_CMD collect-gentxs 2> /dev/null
