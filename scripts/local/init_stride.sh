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
# modify Stride epoch to be 3s
main_config=$STATE/$STRIDE_NODE_NAME/config/genesis.json
jq '.app_state.epochs.epochs[2].duration = $newVal' --arg newVal "3s" $main_config > json.tmp && mv json.tmp $main_config

# add validator account
echo $STRIDE_VAL_KEY | $STRIDE_CMD keys add $STRIDE_VAL_ACCT --recover --keyring-backend=test 
# get validator address
val_addr=$($STRIDE_CMD keys show $STRIDE_VAL_ACCT --keyring-backend test -a) > /dev/null
# add money for this validator account
$STRIDE_CMD add-genesis-account ${val_addr} 500000000000ustrd
# actually set this account as a validator
$STRIDE_CMD gentx $STRIDE_VAL_ACCT 1000000000ustrd --chain-id $STRIDE_CHAIN --keyring-backend test 2> /dev/null

# Add relayer account
echo $RLY_MNEMONIC_1 | $STRIDE_CMD keys add $RLY_NAME_1 --recover --keyring-backend=test 
RLY_ADDRESS_1=$($STRIDE_CMD keys show $RLY_NAME_1 --keyring-backend test -a)
# Give relayer account token balance
$STRIDE_CMD add-genesis-account ${RLY_ADDRESS_1} 500000000000ustrd

# Collect genesis transactions
$STRIDE_CMD collect-gentxs 2> /dev/null
