#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# first, we need to create some saved state, so that we can copy to docker files
mkdir -p $STATE/$GAIA_NODE_NAME

# then, we initialize our chains 
echo 'Initializing gaia state...'

# initialize the chain
$GAIA_CMD init test --chain-id $GAIA_CHAIN --overwrite 2> /dev/null
sed -i -E 's|"stake"|"uatom"|g' "${STATE}/${GAIA_NODE_NAME}/config/genesis.json"
sed -i -E 's|"full"|"validator"|g' "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
sed -i -E "s|timeout_commit = \"5s\"|timeout_commit = \"${BLOCK_TIME}\"|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
sed -i -E 's|enable = false|enable = true|g'  "${STATE}/${GAIA_NODE_NAME}/config/app.toml"
sed -i -E 's|unsafe-cors = false|unsafe-cors = true|g' "${STATE}/${GAIA_NODE_NAME}/config/app.toml"

# add validator account
echo $GAIA_VAL_MNEMONIC | $GAIA_CMD keys add $GAIA_VAL_ACCT --recover --keyring-backend=test 
# get validator address
val_addr=$($GAIA_CMD keys show $GAIA_VAL_ACCT --keyring-backend test -a) > /dev/null
# add money for this validator account
$GAIA_CMD add-genesis-account ${val_addr} 500000000000000uatom
# actually set this account as a validator
$GAIA_CMD gentx $GAIA_VAL_ACCT 1000000000uatom --chain-id $GAIA_CHAIN --keyring-backend test 2> /dev/null

# Add hermes relayer account
echo $HERMES_GAIA_MNEMONIC | $GAIA_CMD keys add $HERMES_GAIA_ACCT --recover --keyring-backend=test 
HERMES_GAIA_ADDRESS=$($GAIA_CMD keys show $HERMES_GAIA_ACCT --keyring-backend test -a)
# Give relayer account token balance
$GAIA_CMD add-genesis-account ${HERMES_GAIA_ADDRESS} 5000000000000uatom

# Add ICQ relayer account
echo $ICQ_GAIA_MNEMONIC | $GAIA_CMD keys add $ICQ_GAIA_ACCT --recover --keyring-backend=test 
ICQ_GAIA_ADDRESS=$($GAIA_CMD keys show $ICQ_GAIA_ACCT --keyring-backend test -a)
# Give relayer account token balance
$GAIA_CMD add-genesis-account ${ICQ_GAIA_ADDRESS} 5000000000000uatom

# add revenue account
echo $GAIA_REV_MNEMONIC | $GAIA_CMD keys add $GAIA_REV_ACCT --recover --keyring-backend=test 
# get revenue address
rev_addr=$($GAIA_CMD keys show $GAIA_REV_ACCT --keyring-backend test -a) > /dev/null

# Collect genesis transactions
$GAIA_CMD collect-gentxs 2> /dev/null

## add the message types ICA should allow to the host chain
ALLOW_MESSAGES='\"/cosmos.bank.v1beta1.MsgSend\", \"/cosmos.bank.v1beta1.MsgMultiSend\", \"/cosmos.staking.v1beta1.MsgDelegate\", \"/cosmos.staking.v1beta1.MsgUndelegate\", \"/cosmos.staking.v1beta1.MsgRedeemTokensforShares\", \"/cosmos.staking.v1beta1.MsgTokenizeShares\", \"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"/ibc.applications.transfer.v1.MsgTransfer\"'
sed -i -E "s|\"allow_messages\": \[\]|\"allow_messages\": \[${ALLOW_MESSAGES}\]|g" "${STATE}/${GAIA_NODE_NAME}/config/genesis.json"

# Update ports so they don't conflict with the stride chain
sed -i -E "s|1317|1307|g" "${STATE}/${GAIA_NODE_NAME}/config/app.toml"
sed -i -E "s|9090|9080|g" "${STATE}/${GAIA_NODE_NAME}/config/app.toml"
sed -i -E "s|9091|9081|g" "${STATE}/${GAIA_NODE_NAME}/config/app.toml"
sed -i -E "s|26657|26557|g" "${STATE}/${GAIA_NODE_NAME}/config/client.toml"
sed -i -E "s|26656|26556|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
sed -i -E "s|26657|26557|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
sed -i -E "s|26658|26558|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
sed -i -E "s|26660|26560|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
