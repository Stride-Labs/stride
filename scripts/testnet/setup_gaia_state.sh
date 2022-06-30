#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

NETWORK_NAME=gaia
CHAIN_NAME=GAIA 
NODE_NAME=gaia
VAL_TOKENS=500000000uatom
STAKE_TOKENS=300000000uatom
VAL_ACCT=gval1
ENDPOINT=$GAIA_MAIN_ENDPOINT

echo "Initializing gaia..."
$GAIA_CMD init test --chain-id $CHAIN_NAME --overwrite 2> /dev/null

sed -i -E 's|"stake"|"uatom"|g' "${STATE}/${NODE_NAME}/config/genesis.json"
sed -i -E 's|"full"|"validator"|g' "${STATE}/${NODE_NAME}/config/config.toml"

$GAIA_CMD keys add $VAL_ACCT --keyring-backend=test >> $STATE/keys.txt 2>&1

# get validator address
VAL_ADDR=$($GAIA_CMD keys show $VAL_ACCT --keyring-backend test -a) > /dev/null

# add money for this validator account
$GAIA_CMD add-genesis-account ${VAL_ADDR} $VAL_TOKENS
$GAIA_CMD gentx $VAL_ACCT $STAKE_TOKENS --chain-id $CHAIN_NAME --keyring-backend test 2> /dev/null

# now we grab the relevant node id
GAIA_NODE_ID=$($GAIA_CMD tendermint show-node-id)@$ENDPOINT:$PORT_ID
echo "Node ID: $GAIA_NODE_ID"

# add Hermes and ICQ relayer accounts on Stride
$GAIA_CMD keys add $HERMES_GAIA_ACCT --keyring-backend=test >> $STATE/keys.txt 2>&1
$GAIA_CMD keys add $ICQ_GAIA_ACCT --keyring-backend=test >> $STATE/keys.txt 2>&1
HERMES_GAIA_ADDRESS=$($GAIA_CMD keys show $HERMES_GAIA_ACCT --keyring-backend test -a)
ICQ_GAIA_ADDRESS=$($GAIA_CMD keys show $ICQ_GAIA_ACCT --keyring-backend test -a)

# Give relayer account token balance
$GAIA_CMD add-genesis-account ${HERMES_GAIA_ADDRESS} $VAL_TOKENS
$GAIA_CMD add-genesis-account ${ICQ_GAIA_ADDRESS} $VAL_TOKENS

# process gentx txs
$GAIA_CMD collect-gentxs 2> /dev/null

# add small changes to config.toml
# use blind address (not loopback) to allow incoming connections from outside networks for local debugging
sed -i -E "s|127.0.0.1|0.0.0.0|g" "${STATE}/${NODE_NAME}/config/config.toml"
sed -i -E "s|minimum-gas-prices = \"\"|minimum-gas-prices = \"0uatom\"|g" "${STATE}/${NODE_NAME}/config/app.toml"
# allow CORS and API endpoints for block explorer
sed -i -E 's|enable = false|enable = true|g' "${STATE}/${NODE_NAME}/config/app.toml"
sed -i -E 's|unsafe-cors = false|unsafe-cors = true|g' "${STATE}/${NODE_NAME}/config/app.toml"

## add the message types ICA should allow to the host chain
ALLOW_MESSAGES='\"/cosmos.bank.v1beta1.MsgSend\", \"/cosmos.bank.v1beta1.MsgMultiSend\", \"/cosmos.staking.v1beta1.MsgDelegate\", \"/cosmos.staking.v1beta1.MsgUndelegate\", \"/cosmos.staking.v1beta1.MsgRedeemTokensforShares\", \"/cosmos.staking.v1beta1.MsgTokenizeShares\", \"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"/ibc.applications.transfer.v1.MsgTransfer\"'
sed -i -E "s|\"allow_messages\": \[\]|\"allow_messages\": \[${ALLOW_MESSAGES}\]|g" "${STATE}/${NODE_NAME}/config/genesis.json"
