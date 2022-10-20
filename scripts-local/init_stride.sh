#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# Optionally pass an argument to override the stride binary
STRIDE_BINARY="${1:-$SCRIPT_DIR/../build/strided}"
STRIDE_CMD="$STRIDE_BINARY --home $SCRIPT_DIR/state/stride"

# first, we need to create some saved state, so that we can copy to docker files
mkdir -p $STATE/$STRIDE_NODE_NAME

# then, we initialize our chains 
echo 'Initializing Stride state...'

# initialize the chain
$STRIDE_CMD init stride1 --chain-id $STRIDE_CHAIN --overwrite 2> /dev/null
$STRIDE_CMD_2 init stride2 --chain-id $STRIDE_CHAIN --overwrite 2> /dev/null
$STRIDE_CMD_3 init stride3 --chain-id $STRIDE_CHAIN --overwrite 2> /dev/null
$STRIDE_CMD_4 init stride4 --chain-id $STRIDE_CHAIN --overwrite 2> /dev/null
$STRIDE_CMD_5 init stride5 --chain-id $STRIDE_CHAIN --overwrite 2> /dev/null

for NODE_NAME in stride stride2 stride3 stride4 stride5; do
    # change the denom
    sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${NODE_NAME}/config/genesis.json"
    # sed -i -E 's|timeout_propose = "3s"|timeout_propose = "5s"|g' "${STATE}/${NODE_NAME}/config/config.toml"
    sed -i -E "s|timeout_commit = \"5s\"|timeout_commit = \"${BLOCK_TIME}\"|g" "${STATE}/${NODE_NAME}/config/config.toml"
    sed -i -E "s|cors_allowed_origins = \[\]|cors_allowed_origins = [\"\*\"]|g" "${STATE}/${NODE_NAME}/config/config.toml"
    sed -i -E "s|allow_duplicate_ip = false|allow_duplicate_ip = true|g" "${STATE}/${NODE_NAME}/config/config.toml"
    sed -i -E "s|addr_book_strict = true|addr_book_strict = false|g" "${STATE}/${NODE_NAME}/config/config.toml"
    # sed -i -E "s|skip_timeout_commit = false|skip_timeout_commit = true|g" "${STATE}/${NODE_NAME}/config/config.toml"

    # update the client config
    clienttoml="${STATE}/${NODE_NAME}/config/client.toml"
    sed -i -E "s|chain-id = \"\"|chain-id = \"${STRIDE_CHAIN}\"|g" $clienttoml
    sed -i -E "s|keyring-backend = \"os\"|keyring-backend = \"test\"|g" $clienttoml
    # modify Stride epoch to be 3s
    main_config=$STATE/$NODE_NAME/config/genesis.json
    # NOTE: If you add new epochs, these indexes will need to be updated
    jq '.app_state.epochs.epochs[$epochIndex].duration = $epochLen' --arg epochLen $DAY_EPOCH_LEN --argjson epochIndex $DAY_EPOCH_INDEX  $main_config > json.tmp && mv json.tmp $main_config
    jq '.app_state.epochs.epochs[$epochIndex].duration = $epochLen' --arg epochLen $STRIDE_EPOCH_LEN --argjson epochIndex $STRIDE_EPOCH_INDEX $main_config > json.tmp && mv json.tmp $main_config
    jq '.app_state.epochs.epochs[$epochIndex].duration = $epochLen' --arg epochLen $MINT_EPOCH_LEN --argjson epochIndex $MINT_EPOCH_INDEX $main_config > json.tmp && mv json.tmp $main_config
    jq '.app_state.stakeibc.params.rewards_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
    jq '.app_state.stakeibc.params.delegate_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
    jq '.app_state.stakeibc.params.deposit_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
    jq '.app_state.stakeibc.params.redemption_rate_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
    jq '.app_state.stakeibc.params.reinvest_interval = $interval' --arg interval $INTERVAL_LEN $main_config > json.tmp && mv json.tmp $main_config
done


MAIN_NODE_ID=$($STRIDE_CMD tendermint show-node-id)@localhost:26656,

# ================= MAP PORTS FOR NODE 2 SO IT DOESN'T CONFLICT WITH NODE 1 =================
sed -i -E 's|6060|6020|g' "${STATE}/stride2/config/config.toml"
sed -i -E "s|26657|$STRIDE_PORT_ID_2|g" "${STATE}/stride2/config/client.toml"
sed -i -E "s|26657|$STRIDE_PORT_ID_2|g" "${STATE}/stride2/config/config.toml"
sed -i -E "s|26656|$STRIDE_PORT_ID_2|g" "${STATE}/stride2/config/config.toml"
sed -i -E "s|26658|26258|g" "${STATE}/stride2/config/config.toml"
sed -i -E "s|external_address = \"\"|external_address = \"localhost:${STRIDE_EXT_ADR_2}\"|g" "${STATE}/stride2/config/config.toml"

sed -i -E "s|9090|9020|g" "${STATE}/stride2/config/app.toml"
sed -i -E "s|9091|9021|g" "${STATE}/stride2/config/app.toml"
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$MAIN_NODE_ID\"|g" "${STATE}/stride2/config/config.toml"
sed -i -E 's|enable = true|enable = false|g' "${STATE}/stride2/config/app.toml"

mkdir $STRIDE_HOME/config/gentx/
# ============================== SETUP CHAIN 2 ======================================
echo $STRIDE_VAL_MNEMONIC_2 | $STRIDE_CMD_2 keys add $STRIDE_VAL_ACCT_2 --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 
$STRIDE_CMD_2 add-genesis-account $STRIDE_VAL_2_ADDR 500000000000000ustrd
$STRIDE_CMD add-genesis-account $STRIDE_VAL_2_ADDR 500000000000000ustrd
$STRIDE_CMD_2 gentx $STRIDE_VAL_ACCT_2 1000000000ustrd --chain-id $STRIDE_CHAIN --keyring-backend test --output-document=$STRIDE_HOME/config/gentx/val2.json >> $TX_LOGS 2>&1 


# ================= MAP PORTS FOR NODE 3 SO IT DOESN'T CONFLICT WITH NODE 1 =================
sed -i -E 's|6060|6010|g' "${STATE}/stride3/config/config.toml"
sed -i -E "s|26657|$STRIDE_PORT_ID_3|g" "${STATE}/stride3/config/client.toml"
sed -i -E "s|26657|$STRIDE_PORT_ID_3|g" "${STATE}/stride3/config/config.toml"
sed -i -E "s|26656|$STRIDE_PORT_ID_3|g" "${STATE}/stride3/config/config.toml"
sed -i -E "s|26658|26158|g" "${STATE}/stride3/config/config.toml"
sed -i -E "s|external_address = \"\"|external_address = \"localhost:${STRIDE_EXT_ADR_3}\"|g" "${STATE}/stride3/config/config.toml"

sed -i -E "s|9090|9010|g" "${STATE}/stride3/config/app.toml"
sed -i -E "s|9091|9011|g" "${STATE}/stride3/config/app.toml"
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$MAIN_NODE_ID\"|g" "${STATE}/stride3/config/config.toml"
sed -i -E 's|enable = true|enable = false|g' "${STATE}/stride3/config/app.toml"

# ============================== SETUP CHAIN 3 ======================================
echo $STRIDE_VAL_MNEMONIC_3 | $STRIDE_CMD_3 keys add $STRIDE_VAL_ACCT_3 --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 
$STRIDE_CMD_3 add-genesis-account $STRIDE_VAL_3_ADDR 500000000000000ustrd
$STRIDE_CMD add-genesis-account $STRIDE_VAL_3_ADDR 500000000000000ustrd
$STRIDE_CMD_3 gentx $STRIDE_VAL_ACCT_3 1000000000ustrd --chain-id $STRIDE_CHAIN --keyring-backend test --output-document=$STRIDE_HOME/config/gentx/val3.json >> $TX_LOGS 2>&1 


# ================= MAP PORTS FOR NODE 4 SO IT DOESN'T CONFLICT WITH NODE 1 =================
sed -i -E 's|6060|6000|g' "${STATE}/stride4/config/config.toml"
sed -i -E "s|26657|$STRIDE_PORT_ID_4|g" "${STATE}/stride4/config/client.toml"
sed -i -E "s|26657|$STRIDE_PORT_ID_4|g" "${STATE}/stride4/config/config.toml"
sed -i -E "s|26656|$STRIDE_PORT_ID_4|g" "${STATE}/stride4/config/config.toml"
sed -i -E "s|26658|26058|g" "${STATE}/stride4/config/config.toml"
sed -i -E "s|external_address = \"\"|external_address = \"localhost:${STRIDE_EXT_ADR_4}\"|g" "${STATE}/stride4/config/config.toml"

sed -i -E "s|9090|9000|g" "${STATE}/stride4/config/app.toml"
sed -i -E "s|9091|9001|g" "${STATE}/stride4/config/app.toml"
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$MAIN_NODE_ID\"|g" "${STATE}/stride4/config/config.toml"
sed -i -E 's|enable = true|enable = false|g' "${STATE}/stride4/config/app.toml"

# ============================== SETUP CHAIN 4 ======================================
echo $STRIDE_VAL_MNEMONIC_4 | $STRIDE_CMD_4 keys add $STRIDE_VAL_ACCT_4 --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 
$STRIDE_CMD_4 add-genesis-account $STRIDE_VAL_4_ADDR 500000000000000ustrd
$STRIDE_CMD add-genesis-account $STRIDE_VAL_4_ADDR 500000000000000ustrd
$STRIDE_CMD_4 gentx $STRIDE_VAL_ACCT_4 1000000000ustrd --chain-id $STRIDE_CHAIN --keyring-backend test --output-document=$STRIDE_HOME/config/gentx/val4.json >> $TX_LOGS 2>&1 


# ================= MAP PORTS FOR NODE 5 SO IT DOESN'T CONFLICT WITH NODE 1 =================
sed -i -E 's|6060|5090|g' "${STATE}/stride5/config/config.toml"
sed -i -E "s|26657|$STRIDE_PORT_ID_5|g" "${STATE}/stride5/config/client.toml"
sed -i -E "s|26657|$STRIDE_PORT_ID_5|g" "${STATE}/stride5/config/config.toml"
sed -i -E "s|26656|$STRIDE_PORT_ID_5|g" "${STATE}/stride5/config/config.toml"
sed -i -E "s|26658|25958|g" "${STATE}/stride5/config/config.toml"
sed -i -E "s|external_address = \"\"|external_address = \"localhost:${STRIDE_EXT_ADR_5}\"|g" "${STATE}/stride5/config/config.toml"

sed -i -E "s|9090|8090|g" "${STATE}/stride5/config/app.toml"
sed -i -E "s|9091|8091|g" "${STATE}/stride5/config/app.toml"
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$MAIN_NODE_ID\"|g" "${STATE}/stride5/config/config.toml"
sed -i -E 's|enable = true|enable = false|g' "${STATE}/stride5/config/app.toml"

# ============================== SETUP CHAIN 5 ======================================
echo $STRIDE_VAL_MNEMONIC_5 | $STRIDE_CMD_5 keys add $STRIDE_VAL_ACCT_5 --recover --keyring-backend=test >> $KEYS_LOGS 2>&1 
$STRIDE_CMD_5 add-genesis-account $STRIDE_VAL_5_ADDR 500000000000000ustrd
$STRIDE_CMD add-genesis-account $STRIDE_VAL_5_ADDR 500000000000000ustrd
$STRIDE_CMD_5 gentx $STRIDE_VAL_ACCT_5 1000000000ustrd --chain-id $STRIDE_CHAIN --keyring-backend test --output-document=$STRIDE_HOME/config/gentx/val5.json >> $TX_LOGS 2>&1 


# add validator account
echo $STRIDE_VAL_MNEMONIC | $STRIDE_CMD keys add $STRIDE_VAL_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
# get validator address
val_addr=$($STRIDE_CMD keys show $STRIDE_VAL_ACCT -a) > /dev/null
# add money for this validator account
$STRIDE_CMD add-genesis-account ${val_addr} 500000000000ustrd 
# actually set this account as a validator
$STRIDE_CMD gentx $STRIDE_VAL_ACCT 100000000000ustrd --chain-id $STRIDE_CHAIN 2> /dev/null

# Add hermes relayer account
echo $HERMES_STRIDE_MNEMONIC | $STRIDE_CMD keys add $HERMES_STRIDE_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
HERMES_STRIDE_ADDRESS=$($STRIDE_CMD keys show $HERMES_STRIDE_ACCT --keyring-backend test -a)
# Give relayer account token balance
$STRIDE_CMD add-genesis-account ${HERMES_STRIDE_ADDRESS} 500000000000ustrd >> $KEYS_LOGS 2>&1 &

# Add ICQ relayer account
echo $ICQ_STRIDE_MNEMONIC | $STRIDE_CMD keys add $ICQ_STRIDE_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
ICQ_STRIDE_ADDRESS=$($STRIDE_CMD keys show $ICQ_STRIDE_ACCT --keyring-backend test -a)
# Give relayer account token balance
$STRIDE_CMD add-genesis-account ${ICQ_STRIDE_ADDRESS} 500000000000ustrd  # >> $KEYS_LOGS 2>&1 &

# Add rly relayer account
echo $RLY_STRIDE_MNEMONIC | $STRIDE_CMD keys add $RLY_STRIDE_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
# Give relayer account token balance
$STRIDE_CMD add-genesis-account ${RLY_STRIDE_ADDR} 500000000000ustrd # >> $KEYS_LOGS 2>&1 &

sed -i -E "s|snapshot-interval = 0|snapshot-interval = 300|g" "${STATE}/${STRIDE_NODE_NAME}/config/app.toml"

# Collect genesis transactions
$STRIDE_CMD collect-gentxs 2> /dev/null

# Shorten voting period
sed -i -E "s|max_deposit_period\": \"172800s\"|max_deposit_period\": \"${MAX_DEPOSIT_PERIOD}\"|g" "${STATE}/${STRIDE_NODE_NAME}/config/genesis.json"
sed -i -E "s|voting_period\": \"172800s\"|voting_period\": \"${VOTING_PERIOD}\"|g" "${STATE}/${STRIDE_NODE_NAME}/config/genesis.json"

cp $STRIDE_HOME/config/genesis.json $STRIDE_HOME_2/config/genesis.json
cp $STRIDE_HOME/config/genesis.json $STRIDE_HOME_3/config/genesis.json
cp $STRIDE_HOME/config/genesis.json $STRIDE_HOME_4/config/genesis.json
cp $STRIDE_HOME/config/genesis.json $STRIDE_HOME_5/config/genesis.json