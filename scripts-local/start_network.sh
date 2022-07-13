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
GAIA_LOGS_2=$SCRIPT_DIR/logs/gaia2.log
GAIA_LOGS_3=$SCRIPT_DIR/logs/gaia3.log
HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log
ICQ_LOGS=$SCRIPT_DIR/logs/icq.log

# Stop processes and clear state and logs
make stop 2>/dev/null || true
rm -rf $SCRIPT_DIR/state $SCRIPT_DIR/logs/*.log $SCRIPT_DIR/logs/temp

# Recreate each log file
for log in $STRIDE_LOGS $GAIA_LOGS $GAIA_LOGS_2 $HERMES_LOGS $ICQ_LOGS; do
    touch $log
done


if [ "$CACHE" != "true" ]; then
    # If not caching, initialize state for Stride, Gaia, and relayers
    sh ${SCRIPT_DIR}/init_stride.sh
    sh ${SCRIPT_DIR}/init_gaia.sh
    sh ${SCRIPT_DIR}/init_relayers.sh
else
    # Otherwise, restore from the backup file
    echo "Restoring state from cache..."
    cp -r $SCRIPT_DIR/.state.backup $SCRIPT_DIR/state
fi

# Starts Stride and Gaia in the background using nohup, pipes the logs to their corresponding log files,
#   and halts the script until Stride/Gaia have each finalized a block
printf '\n%s' "Starting Stride and Gaia...   "
nohup $STRIDE_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS 2>&1 &
nohup $GAIA_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS 2>&1 &
nohup $GAIA_CMD_2 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS_2 2>&1 &
# nohup $GAIA_CMD_3 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS_3 2>&1 &

( tail -f -n0 $STRIDE_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $GAIA_LOGS & ) | grep -q "finalizing commit of block"
sleep 5
echo "Done"

if [ "$CACHE" != "true" ]; then
    # If cache mode is disabled, create the hermes connection and channels, 
    # Logs are piped to the hermes log file and the script is halted until:
    #  1)  "Creating transfer channel" is printed (indicating the connection has been created)
    #  2)  "Message ChanOpenInit" is printed (indicating the channnel has been created)
    printf '%s' "Creating Hermes Connection... "
    bash $SCRIPT_DIR/init_channel.sh >> $HERMES_LOGS 2>&1 &
    ( tail -f -n0 $HERMES_LOGS & ) | grep -q "Creating transfer channel"
    echo "Done"

    printf '%s' "Creating Hermes Channel...    "
    # continuation of logs from above command
    ( tail -f -n0 $HERMES_LOGS & ) | grep -q "Message ChanOpenInit"
    echo "Done"

    # If we're starting for the first time, confirm with the connection handshake
    hermes_start_msg_indicator="STRIDE->GAIA}: connection handshake step completed with events: OpenConfirmConnection"
else
    # If we're starting from cache, no need to wait for a handshake
    hermes_start_msg_indicator="Hermes has started"
fi

# Start hermes in the background and pause until the log message shows that it is up and running
printf '%s' "Starting Hermes...            "

nohup $HERMES_CMD start >> $HERMES_LOGS 2>&1 &
( tail -f -n0 $HERMES_LOGS & ) | grep -q -E "$hermes_start_msg_indicator"
echo "Done"

# Start ICQ in the background
printf '%s' "Starting ICQ...               "
nohup $ICQ_CMD run --local >> $ICQ_LOGS 2>&1 &
sleep 5
echo "Done"

# Create a copy of the state that can be used for the "cache" option
echo "Network is ready for transactions.\n"
rm -rf $SCRIPT_DIR/.state.backup
cp -r $SCRIPT_DIR/state $SCRIPT_DIR/.state.backup

if [ "$CACHE" != "true" ]; then
    # Submit a transaction on stride to register the gaia host zone
    echo "Creating host zone..."
    $STRIDE_CMD tx stakeibc register-host-zone \
        connection-0 $ATOM_DENOM $IBC_ATOM_DENOM channel-0 3 \
        --chain-id $STRIDE_CHAIN --home $STATE/stride \
        --keyring-backend test --from $STRIDE_VAL_ACCT --gas 1000000 -y
fi
# sleep a while longer to wait for ICA accounts to set up
sleep 60

echo "Registering validators on host zone..."

CSLEEP 10
$GAIA_CMD tx bank send gval1 $GAIA_VAL_2_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y
CSLEEP 10
$GAIA_CMD tx bank send gval1 $GAIA_VAL_3_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y

CSLEEP 10
$STRIDE_CMD tx stakeibc add-validator GAIA gval1 $GAIA_DELEGATE_VAL 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
CSLEEP 30
$STRIDE_CMD tx stakeibc add-validator GAIA gval2 $GAIA_DELEGATE_VAL_2 10 10 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
CSLEEP 30


# Add more detailed log files
$SCRIPT_DIR/create_logs.sh &