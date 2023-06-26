#!/bin/bash

CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../config.sh

# Note: For first time use, we need to create the wallet keys. Set the `DRY_RUN` flag `false` to actually create the keys.
DRY_RUN=true

WALLET_MNEMONIC_1="banner spread envelope side kite person disagree path silver will brother under couch edit food venture squirrel civil budget number acquire point work mass"
WALLET_MNEMONIC_2="veteran try aware erosion drink dance decade comic dawn museum release episode original list ability owner size tuition surface ceiling depth seminar capable only"

VAL_KEY_1=$($STRIDE_MAIN_CMD keys show val1 --keyring-backend=test --output json | jq -r '.address')
echo "-> Validator key 1: $VAL_KEY_1"

VAL_KEY_2=$($OSMO_MAIN_CMD keys show oval1 --keyring-backend=test --output json | jq -r '.address')
echo "-> Validator key 2: $VAL_KEY_2"

RUN_TYPE=""
if $DRY_RUN; then
    RUN_TYPE="--dry-run "
fi

WALLET_KEY_1=$(echo $WALLET_MNEMONIC_1 | $STRIDE_MAIN_CMD keys add --recover wallet1 --keyring-backend=test $RUN_TYPE --output json | jq -r '.address')
echo "-> Wallet key 1: $WALLET_KEY_1"

WALLET_KEY_2=$(echo $WALLET_MNEMONIC_1 | $OSMO_MAIN_CMD keys add --recover wallet2 --keyring-backend=test $RUN_TYPE | grep "address: " | awk '{print $2}')
echo "-> Wallet key 2: $WALLET_KEY_2"

WALLET_KEY_3=$(echo $WALLET_MNEMONIC_2 | $OSMO_MAIN_CMD keys add --recover wallet3 --keyring-backend=test $RUN_TYPE | grep "address: " | awk '{print $2}')
echo "-> Wallet key 3: $WALLET_KEY_3"

echo "-> Check balance"
$STRIDE_MAIN_CMD query bank balances $VAL_KEY_1
$OSMO_MAIN_CMD query bank balances $VAL_KEY_2
$STRIDE_MAIN_CMD query bank balances $WALLET_KEY_1
$OSMO_MAIN_CMD query bank balances $WALLET_KEY_2
$OSMO_MAIN_CMD query bank balances $WALLET_KEY_3

echo "-> Send tokens"
$STRIDE_MAIN_CMD tx bank send $VAL_KEY_1 $WALLET_KEY_1 100000ustrd --from $VAL_KEY_1 -y | TRIM_TX
$OSMO_MAIN_CMD tx bank send $VAL_KEY_2 $WALLET_KEY_2 100000uosmo --from $VAL_KEY_2 -y | TRIM_TX
$OSMO_MAIN_CMD tx bank send $VAL_KEY_2 $WALLET_KEY_3 100000uosmo --from $VAL_KEY_2 -y | TRIM_TX

echo "-> Check balance again"
$STRIDE_MAIN_CMD query bank balances $VAL_KEY_1
$OSMO_MAIN_CMD query bank balances $VAL_KEY_2
$STRIDE_MAIN_CMD query bank balances $WALLET_KEY_1
$OSMO_MAIN_CMD query bank balances $WALLET_KEY_2
$OSMO_MAIN_CMD query bank balances $WALLET_KEY_3

echo "-> Register an interchain account on behalf of $WALLET_KEY_1"
$STRIDE_MAIN_CMD tx ica controller register connection-0 --from $WALLET_KEY_1 --keyring-backend test -y | TRIM_TX

echo "-> Wait until the relayer has relayed the packet"
sleep 30

echo "-> Store the interchain account address by parsing the query result"
ICA_ADDR=$($STRIDE_MAIN_CMD q ica controller interchain-account $WALLET_KEY_1 connection-0 -o json | jq -r '.address')
echo $ICA_ADDR

echo "-> Query the interchain account balance on the host chain."
$OSMO_MAIN_CMD q bank balances $ICA_ADDR

echo "-> Send funds to the interchain account."
$OSMO_MAIN_CMD tx bank send $WALLET_KEY_2 $ICA_ADDR 10000uosmo --from $WALLET_KEY_2 -y | TRIM_TX

echo "-> Wait until the relayer has relayed the packet"
sleep 30

echo "-> Query the balance once again and observe the changes"
$OSMO_MAIN_CMD q bank balances $ICA_ADDR

echo "-> Submit a staking delegation tx using the interchain account via ibc"
OSMO_VAL=$(cat $CURRENT_DIR/../state/osmo1/config/genesis.json | jq -r '.app_state.genutil.gen_txs[0].body.messages[0].validator_address')
$STRIDE_MAIN_CMD tx ica host generate-packet-data '{
    "@type":"/cosmos.staking.v1beta1.MsgDelegate",
    "delegator_address":"'$ICA_ADDR'",
    "validator_address":"'$OSMO_VAL'",
    "amount": {
        "denom": "uosmo",
        "amount": "1000"
    }
}' > $CURRENT_DIR/ica_test_delegate.json

$STRIDE_MAIN_CMD tx ica controller send-tx \
    connection-0 $CURRENT_DIR/ica_test_delegate.json \
    --from $WALLET_KEY_1 --keyring-backend test -y | TRIM_TX

echo "-> Wait until the relayer has relayed the packet"
sleep 30

echo "-> Inspect the staking delegations on the host chain"
$OSMO_MAIN_CMD q staking delegations-to $OSMO_VAL

echo "-> Submit a bank send tx using the interchain account via ibc"
$STRIDE_MAIN_CMD tx ica host generate-packet-data '{
    "@type":"/cosmos.bank.v1beta1.MsgSend",
    "from_address":"'$ICA_ADDR'",
    "to_address":"'$WALLET_KEY_3'",
    "amount": [
        {
            "denom": "uosmo",
            "amount": "1000"
        }
    ]
}' > $CURRENT_DIR/ica_test_send.json

$STRIDE_MAIN_CMD tx ica controller send-tx \
    connection-0 $CURRENT_DIR/ica_test_send.json \
    --from $WALLET_KEY_1 --keyring-backend test -y | TRIM_TX

echo "-> Wait until the relayer has relayed the packet"
sleep 30

echo "-> Query the interchain account balance on the host chain"
$OSMO_MAIN_CMD q bank balances $ICA_ADDR
$OSMO_MAIN_CMD q bank balances $WALLET_KEY_3
