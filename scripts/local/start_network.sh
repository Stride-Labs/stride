#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source $SCRIPT_DIR/vars.sh

mkdir -p $SCRIPT_DIR/logs

cache="${1:-false}"

STRIDE_STATE=$SCRIPT_DIR/state/stride
STRIDE_LOGS=$SCRIPT_DIR/logs/stride.log
GAIA_STATE=$SCRIPT_DIR/state/gaia
GAIA_LOGS=$SCRIPT_DIR/logs/gaia.log
HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log
ICQ_LOGS=$SCRIPT_DIR/logs/icq.log

if [ "$cache" == "true" ]; then
    echo "Restoring from cache..."
    rm -rf $SCRIPT_DIR/state 
    mv $SCRIPT_DIR/.state.backup $SCRIPT_DIR/state
fi

printf '%s' "Starting Stride and Gaia...   "
nohup $STRIDE_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS 2>&1 &
nohup $GAIA_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS 2>&1 &
( tail -f -n0 $STRIDE_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $GAIA_LOGS & ) | grep -q "finalizing commit of block"
echo "Done"

if [ "$cache" != "true" ]; then
    printf '%s' "Creating Hermes Connection... "
    bash $SCRIPT_DIR/init_channel.sh > tr -d '\000' > $HERMES_LOGS 2>&1 &
    ( tail -f -n0 $HERMES_LOGS & ) | grep -q "Creating transfer channel"
    echo "Done"

    printf '%s' "Creating Hermes Channel...    "
    # contiuation of logs from above command
    ( tail -f -n0 $HERMES_LOGS & ) | grep -q "Message ChanOpenInit"
    echo "Done"
else 
    echo "" > $HERMES_LOGS
fi

printf '%s' "Starting Hermes...            "
nohup $HERMES_CMD start >> $HERMES_LOGS 2>&1 &
( tail -f -n0 $HERMES_LOGS & ) | grep -q "Hermes has started"
echo "Done"

printf '%s' "Starting ICQ...               "
nohup $ICQ_CMD run --local > $ICQ_LOGS 2>&1 &
echo "Done"

echo "Network is ready for transactions."
cp -r $SCRIPT_DIR/state $SCRIPT_DIR/.state.backup

echo "Creating host zone..."
ATOM='uatom'
IBCATOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
$STRIDE_CMD tx stakeibc register-host-zone \
    connection-0 $ATOM $IBCATOM channel-0 \
    --chain-id $STRIDE_CHAIN --home $STATE/stride \
    --keyring-backend test --from $STRIDE_VAL_ACCT --gas 500000 -y

