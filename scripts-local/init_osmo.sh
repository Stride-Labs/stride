

#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# first, we need to create some saved state, so that we can copy to docker files
mkdir -p $STATE/osmo

# then, we initialize our chains 
echo 'Initializing Osmo state...'

# initialize the chain
$OSMO_CMD init test --chain-id $OSMO_CHAIN --overwrite 2> /dev/null

GENFILE=$OSMO_HOME/config/genesis.json
ICAFILE=$SCRIPT_DIR/juno/ica.json
ica_obj=`cat $ICAFILE`
jq ".app_state += $ica_obj" $GENFILE > json.tmp && mv json.tmp $GENFILE

for NODE_NAME in osmo; do
    sed -i -E 's|"stake"|"uosmo"|g' "${STATE}/${NODE_NAME}/config/genesis.json"
    sed -i -E 's|"full"|"validator"|g' "${STATE}/${NODE_NAME}/config/config.toml"
    sed -i -E "s|timeout_commit = \"5s\"|timeout_commit = \"${BLOCK_TIME}\"|g" "${STATE}/${NODE_NAME}/config/config.toml"
done

MAIN_NODE_ID=$($OSMO_CMD tendermint show-node-id)@localhost:$OSMO_PEER_PORT,

# sed -i -E 's|enable = false|enable = true|g'  "${STATE}/${JUNO_NODE_NAME}/config/app.toml"
MAIN_CONFIG="scripts-local/state/osmo/config/app.toml"
sed -i -E 's|unsafe-cors = false|unsafe-cors = true|g' $MAIN_CONFIG
# enable RPC endpoint for OSMO
sed -i -e '1,/enable = false/ s/enable = false/enable = true/' "${STATE}/${OSMO_NODE_NAME}/config/app.toml"

# set the unbonding time
OSMO_CFG_TMP="${STATE}/${OSMO_NODE_NAME}/config/genesis.json"
jq '.app_state.staking.params.unbonding_time = $newVal' --arg newVal "$UNBONDING_TIME" $OSMO_CFG_TMP > json.tmp && mv json.tmp $OSMO_CFG_TMP
# day epoch is index 0
jq '(.app_state.epochs.epochs[] | select(.identifier=="day") ).duration = $epochLen' --arg epochLen $DAY_EPOCH_LEN $OSMO_CFG_TMP > json.tmp && mv json.tmp $OSMO_CFG_TMP
jq '(.app_state.epochs.epochs[] | select(.identifier=="hour") ).duration = $epochLen' --arg epochLen $DAY_EPOCH_LEN $OSMO_CFG_TMP > json.tmp && mv json.tmp $OSMO_CFG_TMP
jq '(.app_state.epochs.epochs[] | select(.identifier=="week") ).duration = $epochLen' --arg epochLen $DAY_EPOCH_LEN $OSMO_CFG_TMP > json.tmp && mv json.tmp $OSMO_CFG_TMP

# add validator account
echo $OSMO_VAL_MNEMONIC | $OSMO_CMD keys add $OSMO_VAL_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 
# get validator address
val_addr=$($OSMO_CMD keys show $OSMO_VAL_ACCT --keyring-backend test -a) > /dev/null
# add money for this validator account
$OSMO_CMD add-genesis-account ${val_addr} 500000000000000uosmo
# actually set this account as a validator
$OSMO_CMD gentx $OSMO_VAL_ACCT 10000000000uosmo --chain-id $OSMO_CHAIN --keyring-backend test 2> /dev/null

# Add hermes relayer account
echo $HERMES_OSMO_MNEMONIC | $OSMO_CMD keys add $HERMES_OSMO_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 
HERMES_OSMO_ADDRESS=$($OSMO_CMD keys show $HERMES_OSMO_ACCT --keyring-backend test -a)
# Give relayer account token balance
$OSMO_CMD add-genesis-account ${HERMES_OSMO_ADDRESS} 5000000000000uosmo

# Add ICQ relayer account
echo $ICQ_OSMO_MNEMONIC | $OSMO_CMD keys add $ICQ_OSMO_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 
ICQ_OSMO_ADDRESS=$($OSMO_CMD keys show $ICQ_OSMO_ACCT --keyring-backend test -a)
# Give relayer account token balance
$OSMO_CMD add-genesis-account ${ICQ_OSMO_ADDRESS} 5000000000000uosmo

# Add ibc-go relayer account
echo $RLY_OSMO_MNEMONIC | $OSMO_CMD keys add $RLY_OSMO_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 
# Give relayer account token balance
$OSMO_CMD add-genesis-account ${RLY_OSMO_ADDR} 5000000000000uosmo

# add revenue account
echo $OSMO_REV_MNEMONIC | $OSMO_CMD keys add $OSMO_REV_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 
# get revenue address
rev_addr=$($OSMO_CMD keys show $OSMO_REV_ACCT --keyring-backend test -a) > /dev/null

# Collect genesis transactions
$OSMO_CMD collect-gentxs 2> /dev/null

## add the message types ICA should allow to the host chain
ALLOW_MESSAGES='\"/cosmos.bank.v1beta1.MsgSend\", \"/cosmos.bank.v1beta1.MsgMultiSend\", \"/cosmos.staking.v1beta1.MsgDelegate\", \"/cosmos.staking.v1beta1.MsgUndelegate\", \"/cosmos.staking.v1beta1.MsgBeginRedelegate\", \"/cosmos.staking.v1beta1.MsgRedeemTokensforShares\", \"/cosmos.staking.v1beta1.MsgTokenizeShares\", \"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"/ibc.applications.transfer.v1.MsgTransfer\"'
sed -i -E "s|\"allow_messages\": \[\]|\"allow_messages\": \[${ALLOW_MESSAGES}\]|g" "${STATE}/${OSMO_NODE_NAME}/config/genesis.json"

# Update ports so they don't conflict with the stride chain
sed -i -E "s|1317|1117|g" "${STATE}/${OSMO_NODE_NAME}/config/app.toml"
sed -i -E "s|9090|9040|g" "${STATE}/${OSMO_NODE_NAME}/config/app.toml"
sed -i -E "s|9091|9041|g" "${STATE}/${OSMO_NODE_NAME}/config/app.toml"

sed -i -E "s|26657|23657|g" "${STATE}/${OSMO_NODE_NAME}/config/client.toml"
sed -i -E "s|26657|23657|g" "${STATE}/${OSMO_NODE_NAME}/config/config.toml"

sed -i -E "s|26656|23656|g" "${STATE}/${OSMO_NODE_NAME}/config/config.toml"
sed -i -E "s|26658|23658|g" "${STATE}/${OSMO_NODE_NAME}/config/config.toml"
sed -i -E "s|26660|23660|g" "${STATE}/${OSMO_NODE_NAME}/config/config.toml"
sed -i -E "s|6060|6022|g" "${STATE}/${OSMO_NODE_NAME}/config/config.toml"