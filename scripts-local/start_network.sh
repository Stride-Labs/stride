#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source $SCRIPT_DIR/vars.sh

mkdir -p $SCRIPT_DIR/logs

CACHE="${1:-false}"

STRIDE_STATE=$SCRIPT_DIR/state/stride
STRIDE_LOGS=$SCRIPT_DIR/logs/stride.log
GAIA_STATE=$SCRIPT_DIR/state/gaia
GAIA_LOGS=$SCRIPT_DIR/logs/gaia.log
HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log
ICQ_LOGS=$SCRIPT_DIR/logs/icq.log

# Clear logs
rm -rf $SCRIPT_DIR/logs/*.log

# TODO (TEST-122) Fix "caching" option in local mode
if [ "$CACHE" == "true" ]; then
    echo "Restoring from cache..."
    rm -rf $SCRIPT_DIR/state 
    cp -r $SCRIPT_DIR/.state.backup $SCRIPT_DIR/state
fi

# Starts Stride and Gaia in the background using nohup, pipes the logs to their corresponding log files,
#   and halts the script until Stride/Gaia have each finalized a block
printf '%s' "Starting Stride and Gaia...   "
nohup $STRIDE_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS 2>&1 &
nohup $GAIA_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS 2>&1 &
( tail -f -n0 $STRIDE_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $GAIA_LOGS & ) | grep -q "finalizing commit of block"
echo "Done"

if [ "$CACHE" != "true" ]; then
    # If cache mode is disabled, create the hermes connection and channels, 
    # Logs are piped to the hermes log file and the script is halted until:
    #  1)  "Creating transfer channel" is printed (indicating the connection has been created)
    #  2)  "Message ChanOpenInit" is printed (indicating the channnel has been created)
    printf '%s' "Creating Hermes Connection... "
    bash $SCRIPT_DIR/init_channel.sh > $HERMES_LOGS 2>&1 &
    ( tail -f -n0 $HERMES_LOGS & ) | grep -q "Creating transfer channel"
    echo "Done"

    printf '%s' "Creating Hermes Channel...    "
    # continuation of logs from above command
    ( tail -f -n0 $HERMES_LOGS & ) | grep -q "Message ChanOpenInit"
    echo "Done"
fi

# Start hermes in the background and pause until the log message shows that it is up and running
if [ "$CACHE" != "true" ]; then
    # If we're starting for the first time, confirm with the connection handshake
    hermes_start_msg_indicator="STRIDE->GAIA}: connection handshake step completed with events: OpenConfirmConnection"
else
    # If we're starting from cache, no need to wait for a handshake
    hermes_start_msg_indicator="Hermes has started"
fi
printf '%s' "Starting Hermes...            "
nohup $HERMES_CMD start >> $HERMES_LOGS 2>&1 &
( tail -f -n0 $HERMES_LOGS & ) | grep -q -E "$hermes_start_msg_indicator"
echo "Done"

# Start ICQ in the background
printf '%s' "Starting ICQ...               "
nohup $ICQ_CMD run --local > $ICQ_LOGS 2>&1 &
sleep 5
echo "Done"

# Create a copy of the state that can be used for the "cache" option
echo "Network is ready for transactions.\n"
cp -r $SCRIPT_DIR/state $SCRIPT_DIR/.state.backup

if [ "$CACHE" != "true" ]; then
    # Submit a transaction on stride to register the gaia host zone
    echo "Creating host zone..."
    $STRIDE_CMD tx stakeibc register-host-zone \
        connection-0 $ATOM $IBCATOM channel-0 \
        --chain-id $STRIDE_CHAIN --home $STATE/stride \
        --keyring-backend test --from $STRIDE_VAL_ACCT --gas 500000 -y
fi
