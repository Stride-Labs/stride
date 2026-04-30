#!/bin/bash

# This script starts a single stride node locally
# It can be used in debugging scenarios where dockernet is unncessary 
# (ex1: debugging an issue that does not require host chains or relayers)
# (ex2: debugging via adding logs to the SDK)
# (ex3: testing CLI commands)

set -eu 

STRIDE_HOME=~/.stride-local
STRIDED="build/strided --home ${STRIDE_HOME}"
# SDK 0.53's `keys` subcommands no longer reliably pick up keyring-backend from client.toml
# (help text shows default "test" but the runtime silently falls back to "os"), so pass it
# explicitly anywhere we touch the keyring
KEYRING="--keyring-backend test"
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
validator_json="${STRIDE_HOME}/validator.json"

rm -rf ${STRIDE_HOME}

$STRIDED init stride-local --chain-id $CHAIN_ID --overwrite

$STRIDED config set client chain-id $CHAIN_ID
$STRIDED config set client keyring-backend test
$STRIDED config set client node http://127.0.0.1:26657

sed -i -E "s|minimum-gas-prices = \".*\"|minimum-gas-prices = \"0${DENOM}\"|g" $app_toml
sed -i -E '/\[api\]/,/^enable = .*$/ s/^enable = .*$/enable = true/' $app_toml

sed -i -E "s|\"stake\"|\"${DENOM}\"|g" $genesis_json 

jq '(.app_state.epochs.epochs[] | select(.identifier=="day") ).duration = $epochLen' --arg epochLen $STRIDE_DAY_EPOCH_DURATION $genesis_json > json.tmp && mv json.tmp $genesis_json
jq '(.app_state.epochs.epochs[] | select(.identifier=="stride_epoch") ).duration = $epochLen' --arg epochLen $STRIDE_EPOCH_EPOCH_DURATION $genesis_json > json.tmp && mv json.tmp $genesis_json
jq '.app_state.gov.params.max_deposit_period = $newVal' --arg newVal "$MAX_DEPOSIT_PERIOD" $genesis_json > json.tmp && mv json.tmp $genesis_json
jq '.app_state.gov.params.voting_period = $newVal' --arg newVal "$VOTING_PERIOD" $genesis_json > json.tmp && mv json.tmp $genesis_json
jq '.app_state.staking.params.unbonding_time = $newVal' --arg newVal "$UNBONDING_TIME" $genesis_json > json.tmp && mv json.tmp $genesis_json

jq "del(.app_state.interchain_accounts)" $genesis_json > json.tmp && mv json.tmp $genesis_json
interchain_accts=$(cat dockernet/config/ica_controller.json)
jq ".app_state += $interchain_accts" $genesis_json > json.tmp && mv json.tmp $genesis_json

echo "$STRIDE_VAL_MNEMONIC" | $STRIDED keys add val --recover $KEYRING
$STRIDED genesis add-genesis-account $($STRIDED keys show val -a $KEYRING) 100000000000${DENOM}

# Seed POA genesis with the local validator so it produces blocks (post-v33,
# ccvconsumer is no longer in the module manager and POA is the sole source
# of the InitChain validator set). The bech32 below is
# authtypes.NewModuleAddress("gov").String() against Stride's address prefix
# — re-derive via app/test_setup.go if the prefix or module name ever changes.
POA_ADMIN="stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl"
val_op_addr=$($STRIDED keys show val -a $KEYRING)
val_pubkey_json=$($STRIDED tendermint show-validator 2>/dev/null)

jq --arg admin "$POA_ADMIN" \
   --arg op "$val_op_addr" \
   --argjson pk "$val_pubkey_json" \
   '.app_state.poa.params.admin = $admin
    | .app_state.poa.validators = [{
        pub_key: $pk,
        power: "1",
        metadata: {operator_address: $op, moniker: "stride-local"}
      }]' \
   $genesis_json > json.tmp && mv json.tmp $genesis_json

echo "$STRIDE_ADMIN_MNEMONIC" | $STRIDED keys add admin --recover $KEYRING
$STRIDED genesis add-genesis-account $($STRIDED keys show admin -a $KEYRING) 100000000000${DENOM}

# Start the daemon in the background
$STRIDED start & 
pid=$!
sleep 10

# Add a governator
echo "Adding governator..."
cat > $validator_json << EOF
{
  "pubkey": $($STRIDED tendermint show-validator),
  "amount": "1000000000${DENOM}",
  "moniker": "val1",
  "commission-rate": "0.10",
  "commission-max-rate": "0.20",
  "commission-max-change-rate": "0.01",
  "min-self-delegation": "1"
}
EOF
$STRIDED tx staking create-validator $validator_json --from val -y --chain-id $CHAIN_ID --node http://127.0.0.1:26657 $KEYRING

# Bring the daemon back to the foreground
wait $pid