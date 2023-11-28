#!/bin/bash

# This script starts a single stride node locally
# It can be used in debugging scenarios where dockernet is unncessary 
# (ex1: debugging an issue that does not require host chains or relayers)
# (ex2: debugging via adding logs to the SDK)
# (ex3: testing CLI commands)

set -eu 

STRIDE_HOME=~/.stride-local
STRIDED="build/strided --home ${STRIDE_HOME}"
CHAIN_ID=stride-local-1
DENOM=ustrd

STRIDE_ADMIN_MNEMONIC="tone cause tribe this switch near host damage idle fragile antique tail soda alien depth write wool they rapid unfold body scan pledge soft"
STRIDE_VAL_MNEMONIC="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"

STRIDE_DAY_EPOCH_DURATION="140s"
STRIDE_EPOCH_EPOCH_DURATION="35s"
MAX_DEPOSIT_PERIOD="30s"
VOTING_PERIOD="30s"
UNBONDING_TIME="240s"

config_toml="${STRIDE_HOME}/config/config.toml"
client_toml="${STRIDE_HOME}/config/client.toml"
app_toml="${STRIDE_HOME}/config/app.toml"
genesis_json="${STRIDE_HOME}/config/genesis.json"

rm -rf ${STRIDE_HOME}

$STRIDED init stride-local --chain-id $CHAIN_ID --overwrite

sed -i -E "s|minimum-gas-prices = \".*\"|minimum-gas-prices = \"0${DENOM}\"|g" $app_toml
sed -i -E '/\[api\]/,/^enable = .*$/ s/^enable = .*$/enable = true/' $app_toml

sed -i -E "s|chain-id = \"\"|chain-id = \"${CHAIN_ID}\"|g" $client_toml
sed -i -E "s|keyring-backend = \"os\"|keyring-backend = \"test\"|g" $client_toml
sed -i -E "s|node = \".*\"|node = \"tcp://localhost:26657\"|g" $client_toml

jq '(.app_state.epochs.epochs[] | select(.identifier=="day") ).duration = $epochLen' --arg epochLen $STRIDE_DAY_EPOCH_DURATION $genesis_json > json.tmp && mv json.tmp $genesis_json
jq '(.app_state.epochs.epochs[] | select(.identifier=="stride_epoch") ).duration = $epochLen' --arg epochLen $STRIDE_EPOCH_EPOCH_DURATION $genesis_json > json.tmp && mv json.tmp $genesis_json
jq '.app_state.gov.params.max_deposit_period = $newVal' --arg newVal "$MAX_DEPOSIT_PERIOD" $genesis_json > json.tmp && mv json.tmp $genesis_json
jq '.app_state.gov.params.voting_period = $newVal' --arg newVal "$VOTING_PERIOD" $genesis_json > json.tmp && mv json.tmp $genesis_json

jq "del(.app_state.interchain_accounts)" $genesis_json > json.tmp && mv json.tmp $genesis_json
interchain_accts=$(cat dockernet/config/ica_controller.json)
jq ".app_state += $interchain_accts" $genesis_json > json.tmp && mv json.tmp $genesis_json

# hack since add-comsumer-section is built for dockernet
rm -rf ~/.stride-loca1
cp -r ${STRIDE_HOME} ~/.stride-loca1

$STRIDED add-consumer-section 1
jq '.app_state.ccvconsumer.params.unbonding_period = $newVal' --arg newVal "$UNBONDING_TIME" $genesis_json > json.tmp && mv json.tmp $genesis_json

rm -rf ~/.stride-loca1

echo "$STRIDE_VAL_MNEMONIC" | $STRIDED keys add val --recover --keyring-backend=test 
$STRIDED add-genesis-account $($STRIDED keys show val -a) 100000000000${DENOM}

echo "$STRIDE_ADMIN_MNEMONIC" | $STRIDED keys add admin --recover --keyring-backend=test 
$STRIDED add-genesis-account $($STRIDED keys show admin -a) 100000000000${DENOM}

$STRIDED start


