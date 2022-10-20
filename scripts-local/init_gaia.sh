#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# first, we need to create some saved state, so that we can copy to docker files
mkdir -p $STATE/$GAIA_NODE_NAME

# then, we initialize our chains
echo 'Initializing Gaia state...'

# initialize the chain
$GAIA_CMD init test --chain-id $GAIA_CHAIN --overwrite 2> /dev/null
$GAIA_CMD_2 init test --chain-id $GAIA_CHAIN --overwrite 2> /dev/null
# $GAIA_CMD_3 init test --chain-id $GAIA_CHAIN --overwrite 2> /dev/null

for NODE_NAME in gaia gaia2; do
    sed -i -E 's|"stake"|"uatom"|g' "${STATE}/${NODE_NAME}/config/genesis.json"
    sed -i -E 's|"full"|"validator"|g' "${STATE}/${NODE_NAME}/config/config.toml"
    sed -i -E "s|timeout_commit = \"5s\"|timeout_commit = \"${BLOCK_TIME}\"|g" "${STATE}/${NODE_NAME}/config/config.toml"
    sed -i -E "s|chain-id = \"\"|chain-id = \"${GAIA_CHAIN}\"|g" "${STATE}/${NODE_NAME}/config/client.toml"
    sed -i -E "s|keyring-backend = \"os\"|keyring-backend = \"test\"|g" "${STATE}/${NODE_NAME}/config/client.toml"
done

MAIN_NODE_ID=$($GAIA_CMD tendermint show-node-id)@localhost:$GAIA_PEER_PORT,

# sed -i -E 's|enable = false|enable = true|g'  "${STATE}/${GAIA_NODE_NAME}/config/app.toml"
MAIN_CONFIG="scripts-local/state/gaia/config/app.toml"
sed -i -E 's|unsafe-cors = false|unsafe-cors = true|g' $MAIN_CONFIG
# enable RPC endpoint for GAIA
sed -i -e '1,/enable = false/ s/enable = false/enable = true/' "${STATE}/${GAIA_NODE_NAME}/config/app.toml"

# ================= MAP PORTS FOR NODES 2 & 3 SO THEY DON'T CONFLICT WITH NODE 1 =================
# change all port on additional nodes
sed -i -E 's|6060|6061|g' "${STATE}/gaia2/config/config.toml"
sed -i -E "s|26657|$GAIA_PORT_ID_2|g" "${STATE}/gaia2/config/client.toml"
sed -i -E "s|26657|$GAIA_PORT_ID_2|g" "${STATE}/gaia2/config/config.toml"
sed -i -E "s|26656|$GAIA_PEER_PORT_2|g" "${STATE}/gaia2/config/config.toml"
sed -i -E "s|26658|26458|g" "${STATE}/gaia2/config/config.toml"
sed -i -E "s|external_address = \"\"|external_address = \"localhost:${GAIA_EXT_ADR_2}\"|g" "${STATE}/gaia2/config/config.toml"

sed -i -E "s|9090|9070|g" "${STATE}/gaia2/config/app.toml"
sed -i -E "s|9091|9071|g" "${STATE}/gaia2/config/app.toml"

# sed -i -E 's|6060|6062|g' "${STATE}/gaia3/config/config.toml"
# sed -i -E "s|26657|$GAIA_PORT_ID_3|g" "${STATE}/gaia3/config/client.toml"
# sed -i -E "s|26657|$GAIA_PORT_ID_3|g" "${STATE}/gaia3/config/config.toml"
# sed -i -E "s|26656|$GAIA_PEER_PORT_3|g" "${STATE}/gaia3/config/config.toml"
# sed -i -E "s|26658|26358|g" "${STATE}/gaia3/config/config.toml"
# sed -i -E "s|external_address = \"\"|external_address = \"localhost:26355\"|g" "${STATE}/gaia3/config/config.toml"

# sed -i -E "s|9090|9060|g" "${STATE}/gaia3/config/app.toml"
# sed -i -E "s|9091|9061|g" "${STATE}/gaia3/config/app.toml"

# tell nodes 2 and 3 to look for node 1
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$MAIN_NODE_ID\"|g" "${STATE}/gaia2/config/config.toml"
# sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$MAIN_NODE_ID\"|g" "${STATE}/gaia3/config/config.toml"

mkdir $GAIA_HOME/config/gentx/

# ============================== SETUP CHAIN 2 ======================================
echo $GAIA_VAL_MNEMONIC_2 | $GAIA_CMD_2 keys add $GAIA_VAL_ACCT_2 --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 &
$GAIA_CMD_2 add-genesis-account $GAIA_VAL_2_ADDR 500000000000000uatom
$GAIA_CMD add-genesis-account $GAIA_VAL_2_ADDR 500000000000000uatom
$GAIA_CMD_2 gentx $GAIA_VAL_ACCT_2 5000000000uatom --chain-id $GAIA_CHAIN --output-document=$GAIA_HOME/config/gentx/gval2.json 2> /dev/null

# ============================== SETUP CHAIN 3 ======================================
# echo $GAIA_VAL_MNEMONIC_3 | $GAIA_CMD_3 keys add $GAIA_VAL_ACCT_3 --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 &
# $GAIA_CMD_3 add-genesis-account $GAIA_VAL_3_ADDR 500000000000000uatom
# $GAIA_CMD add-genesis-account $GAIA_VAL_3_ADDR 500000000000000uatom
# $GAIA_CMD_3 gentx $GAIA_VAL_ACCT_3 1000000000uatom --output-document=$GAIA_HOME/config/gentx/gval3.json

# set the unbonding time
GAIA_CFG_TMP="${STATE}/${GAIA_NODE_NAME}/config/genesis.json"
jq '.app_state.staking.params.unbonding_time = $newVal' --arg newVal "$UNBONDING_TIME" $GAIA_CFG_TMP > json.tmp && mv json.tmp $GAIA_CFG_TMP

# add validator account
echo $GAIA_VAL_MNEMONIC | $GAIA_CMD keys add $GAIA_VAL_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
# get validator address
val_addr=$($GAIA_CMD keys show $GAIA_VAL_ACCT -a) > /dev/null
# add money for this validator account
$GAIA_CMD add-genesis-account ${val_addr} 500000000000000uatom
# actually set this account as a validator
$GAIA_CMD gentx $GAIA_VAL_ACCT 5000000000uatom --chain-id $GAIA_CHAIN 2> /dev/null

# Add hermes relayer account
echo $HERMES_GAIA_MNEMONIC | $GAIA_CMD keys add $HERMES_GAIA_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
HERMES_GAIA_ADDRESS=$($GAIA_CMD keys show $HERMES_GAIA_ACCT --keyring-backend test -a)
# Give relayer account token balance
$GAIA_CMD add-genesis-account ${HERMES_GAIA_ADDRESS} 5000000000000uatom >> $KEYS_LOGS 2>&1 &

# Add ICQ relayer account
echo $ICQ_GAIA_MNEMONIC | $GAIA_CMD keys add $ICQ_GAIA_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
ICQ_GAIA_ADDRESS=$($GAIA_CMD keys show $ICQ_GAIA_ACCT --keyring-backend test -a)
# Give relayer account token balance
$GAIA_CMD add-genesis-account ${ICQ_GAIA_ADDRESS} 5000000000000uatom >> $KEYS_LOGS 2>&1 &

# Add rly relayer account
echo $RLY_GAIA_MNEMONIC | $STRIDE_CMD keys add $RLY_GAIA_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
# Give relayer account token balance
$GAIA_CMD add-genesis-account ${RLY_GAIA_ADDR} 500000000000uatom >> $KEYS_LOGS 2>&1 &

# add revenue account
echo $GAIA_REV_MNEMONIC | $GAIA_CMD keys add $GAIA_REV_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
# get revenue address
rev_addr=$($GAIA_CMD keys show $GAIA_REV_ACCT -a) > /dev/null

# Collect genesis transactions
$GAIA_CMD collect-gentxs 2> /dev/null

## add the message types ICA should allow to the host chain
ALLOW_MESSAGES='\"/cosmos.bank.v1beta1.MsgSend\", \"/cosmos.bank.v1beta1.MsgMultiSend\", \"/cosmos.staking.v1beta1.MsgDelegate\", \"/cosmos.staking.v1beta1.MsgUndelegate\", \"/cosmos.staking.v1beta1.MsgBeginRedelegate\", \"/cosmos.staking.v1beta1.MsgRedeemTokensforShares\", \"/cosmos.staking.v1beta1.MsgTokenizeShares\", \"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"/ibc.applications.transfer.v1.MsgTransfer\"'
sed -i -E "s|\"allow_messages\": \[\]|\"allow_messages\": \[${ALLOW_MESSAGES}\]|g" "${STATE}/${GAIA_NODE_NAME}/config/genesis.json"

cp $GAIA_HOME/config/genesis.json $GAIA_HOME_2/config/genesis.json
# cp $GAIA_HOME/config/genesis.json $GAIA_HOME_3/config/genesis.json

# Update ports so they don't conflict with the stride chain
sed -i -E "s|1317|1307|g" "${STATE}/${GAIA_NODE_NAME}/config/app.toml"
sed -i -E "s|9090|9080|g" "${STATE}/${GAIA_NODE_NAME}/config/app.toml"
sed -i -E "s|9091|9081|g" "${STATE}/${GAIA_NODE_NAME}/config/app.toml"

sed -i -E "s|26657|26557|g" "${STATE}/${GAIA_NODE_NAME}/config/client.toml"
sed -i -E "s|26657|26557|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"

sed -i -E "s|26656|26556|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
sed -i -E "s|26658|26558|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
sed -i -E "s|26660|26560|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
sed -i -E "s|6060|6062|g" "${STATE}/${GAIA_NODE_NAME}/config/config.toml"
