#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source $SCRIPT_DIR/vars.sh
source $SCRIPT_DIR/account_vars.sh

OSMO_CHAIN="osmo-test-4"
OSMO_CHAIN="osmosis-1"

# start Hermes
# $HERMES_CMD keys restore --mnemonic "$HERMES_OSMO_MNEMONIC" $OSMO_CHAIN
# $HERMES_CMD start

# set up relayers 
# rly tx link test_path -d -t 3s
# rly start test_path -p events

# $STRIDE_CMD tx ibc-transfer transfer transfer channel-0 osmo1lajwg95utv75fny0w39806xuk92ky57csvj6f5 100ustrd --from val1 --keyring-backend=test --chain-id=STRIDE
# $OSMO_CMD tx ibc-transfer transfer transfer channel-417 stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 100uosmo --from oval1 --keyring-backend=test --chain-id=osmo-test-4 --node https://rpc-test.osmosis.zone:443

# rly tx relay-pkts test_path channel-3 -d
# rly tx relay-acks test_path channel-0 -d

# rly tx relay-pkts osmo_path channel-403 -d
# rly tx relay-acks osmo_path channel-403 -d

# IBC transfers
# $STRIDE_CMD tx bank send $STRIDE_VAL_ADDR $STRIDE_ADMIN_ADDRESS 10000000ustrd --from $STRIDE_VAL_ACCT -y
# sleep 15
# $STRIDE_CMD tx ibc-transfer transfer transfer channel-0 osmo1lajwg95utv75fny0w39806xuk92ky57csvj6f5 100ustrd --from val1 --keyring-backend=test --chain-id=STRIDE
# $OSMO_CMD tx ibc-transfer transfer transfer channel-417 stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 100uosmo --from oval1 --keyring-backend=test --chain-id=osmo-test-4 --node https://rpc-test.osmosis.zone:443 -y

# check hash at this point
# https://testnet.ping.pub/osmosis/account/osmo1lajwg95utv75fny0w39806xuk92ky57csvj6f5

IBC_OSMO_DENOM="ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
IBC_OSMO_DENOM="ibc/13B2C536BB057AC79D5616B8EA1B9540EC1F2170718CAFF6F0083C966FFFED0B"
# echo $STRIDE_ADMIN_MNEMONIC | $STRIDE_CMD keys add stride --recover 

# # register host zone
# $STRIDE_CMD tx stakeibc register-host-zone \
# connection-0 $OSMO_DENOM osmo $IBC_OSMO_DENOM channel-0 1 \
# --chain-id $STRIDE_CHAIN --home $STATE/stride --from stride \
# -y --keyring-backend test --gas 1000000
# sleep 15
# $STRIDE_CMD tx stakeibc add-validator osmo-test-4 osmo osmovaloper1c584m4lq25h83yp6ag8hh4htjr92d954kphp96 10 10 --from stride --keyring-backend=test --chain-id STRIDE

# liquid stake
# $STRIDE_CMD tx stakeibc 
# $STRIDE_CMD tx stakeibc liquid-stake 10 uosmo --from val1 --keyring-backend=test --chain-id STRIDE



# MAINNET 
# rly tx link test_path -d -t 10s

# $HERMES_CMD create channel --port-a transfer --port-b transfer STRIDE connection-0

# $STRIDE_CMD tx ibc-transfer transfer transfer channel-2 osmo1am99pcvynqqhyrwqfvfmnvxjk96rn46lj4pkkx 2000000ustrd --from val1 --keyring-backend=test --chain-id=STRIDE

# $HERMES_CMD keys restore --mnemonic "$KEPLR_SEED" $OSMO_CHAIN
# $OSMO_CMD keys add vish-keplr --recover
# $OSMO_CMD tx ibc-transfer transfer transfer channel-292 stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 100uosmo --from vish-keplr --chain-id=osmosis-1 --node https://rpc.osmosis.zone:443
# $STRIDE_CMD keys list --keyring-backend=test
# $STRIDE_CMD tx bank send val1 stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq 1000000000ustrd --from val1
# $STRIDE_CMD tx stakeibc register-host-zone \
# connection-0 $OSMO_DENOM osmo $IBC_OSMO_DENOM channel-2 1 \
# --chain-id $STRIDE_CHAIN --home $STATE/stride --from stride \
# -y --keyring-backend test --gas 1000000

# $STRIDE_CMD tx stakeibc add-validator osmosis-1 osmo osmovaloper15urq2dtp9qce4fyc85m6upwm9xul3049wh9czc 10 10 --from stride --keyring-backend=test --chain-id STRIDE


# $STRIDE_CMD tx stakeibc liquid-stake 10 uosmo --from val1 --keyring-backend=test --chain-id STRIDE

# $HERMES_CMD tx raw chan-close-confirm --dst-chan-id channel-296 --src-chan-id channel-3 osmosis-1 STRIDE connection-1622 icahost icacontroller-osmosis-1.DELEGATION

# $HERMES_CMD tx raw chan-open-init --dst-chan-id channel-296 --src-chan-id channel-3 osmosis-1 STRIDE connection-1617 icahost icacontroller-osmosis-1.DELEGATION
$HERMES_CMD tx raw chan-open-try --src-chan-id channel-7 osmosis-1 STRIDE connection-1622 icahost icacontroller-osmosis-1.DELEGATION

# $STRIDE_CMD tx stakeibc restore-interchain-account osmosis-1 DELEGATION --from stride --keyring-backend test