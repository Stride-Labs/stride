#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

echo "Creating connection $STRIDE_CHAIN <> $GAIA_CHAIN"
$HERMES_CMD create connection --a-chain $STRIDE_CHAIN --b-chain $GAIA_CHAIN

echo "Creating transfer channel for Gaia"
$HERMES_CMD create channel --a-chain $GAIA_CHAIN --a-connection connection-0 --a-port transfer --b-port transfer

sleep 3
nohup rly start gaia_path -p events --home $SCRIPT_DIR/go-rly >> $RLY_GAIA_LOGS 2>&1 &

# echo "Creating connection $STRIDE_CHAIN <> $OSMO_CHAIN"
# $HERMES_CMD create connection --a-chain $STRIDE_CHAIN --b-chain $OSMO_CHAIN

# echo "Creating transfer channel for Osmo"
# $HERMES_CMD create channel --a-chain $OSMO_CHAIN --a-connection connection-0 --a-port transfer --b-port transfer

# sleep 3
# $STRIDE_CMD tx ibc-transfer transfer transfer channel-0 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2 1000000ustrd --from val1 --keyring-backend=test --chain-id=STRIDE -y >> $TX_LOGS 2>&1 &
# nohup rly start osmo_path -p events --home $SCRIPT_DIR/go-rly --debug-addr localhost:7598 >> $RLY_OSMO_LOGS 2>&1 &

# echo "Creating connection $STRIDE_CHAIN <> $JUNO_CHAIN"
# $HERMES_CMD create connection --a-chain $STRIDE_CHAIN --b-chain $JUNO_CHAIN

# echo "Creating transfer channel for Juno"
# $HERMES_CMD create channel --a-chain $JUNO_CHAIN --a-connection connection-0 --a-port transfer --b-port transfer

# sleep 3
# $STRIDE_CMD tx ibc-transfer transfer transfer channel-0 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2 1000000ustrd --from val1 --keyring-backend=test --chain-id=STRIDE -y >> $TX_LOGS 2>&1 &
# WAIT_FOR_BLOCK $STRIDE_LOGS 2
# $STRIDE_CMD tx ibc-transfer transfer transfer channel-1 osmo1zwj4yr264fr9au20gur3qapt3suwkgp0w039jd 1000000ustrd --from val1 --keyring-backend=test --chain-id=STRIDE -y >> $TX_LOGS 2>&1 &
# nohup rly start juno_path -p events --home $SCRIPT_DIR/go-rly --debug-addr localhost:7599 >> $RLY_JUNO_LOGS 2>&1 &