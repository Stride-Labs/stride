#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )


source $SCRIPT_DIR/vars.sh

mkdir -p $SCRIPT_DIR/logs

CACHE="${1:-false}"

# Stop processes and clear state and logs
make stop 2>/dev/null || true
rm -rf $SCRIPT_DIR/state $SCRIPT_DIR/logs/*.log $SCRIPT_DIR/logs/temp

# Recreate each log file
for log in $STRIDE_LOGS $GAIA_LOGS $GAIA_LOGS_2 $HERMES_LOGS $ICQ_LOGS $JUNO_LOGS $TX_LOGS $KEYS_LOGS $OSMO_LOGS; do
    touch $log
done


if [ "$CACHE" != "true" ]; then
    # If not caching, initialize state for Stride, Gaia, and relayers
    sh ${SCRIPT_DIR}/init_stride.sh
    sh ${SCRIPT_DIR}/init_gaia.sh
    sh ${SCRIPT_DIR}/init_relayers.sh
    sh ${SCRIPT_DIR}/init_juno.sh
    sh ${SCRIPT_DIR}/init_osmo.sh
else
    # Otherwise, restore from the backup file
    echo "Restoring state from cache..."
    cp -r $SCRIPT_DIR/.state.backup $SCRIPT_DIR/state
fi

# Starts Stride and Gaia in the background using nohup, pipes the logs to their corresponding log files,
#   and halts the script until Stride/Gaia have each finalized a block
printf '\n%s' "Starting Stride, Gaia, Osmo, and Juno...   "
nohup $STRIDE_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS 2>&1 &
nohup $GAIA_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS 2>&1 &
nohup $GAIA_CMD_2 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS_2 2>&1 &
nohup $JUNO_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $JUNO_LOGS 2>&1 &
nohup $OSMO_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $OSMO_LOGS 2>&1 &

# nohup $GAIA_CMD_3 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS_3 2>&1 &

( tail -f -n0 $STRIDE_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $GAIA_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $JUNO_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $OSMO_LOGS & ) | grep -q "finalizing commit of block"
sleep 2
echo "Done"
exit 0
if [ "$CACHE" != "true" ]; then
    # If cache mode is disabled, create the hermes connection and channels, 
    # Logs are piped to the hermes log file and the script is halted until:
    #  1)  "Creating transfer channel" is printed (indicating the connection has been created)
    #  2)  "Message ChanOpenInit" is printed (indicating the channnel has been created)
    bash $SCRIPT_DIR/init_channel.sh >> $HERMES_LOGS 2>&1 &
    for i in {1..3}
    do
        printf '%s' "Creating Hermes Connection $i... "
        ( tail -f -n0 $HERMES_LOGS & ) | grep -q "Creating transfer channel"
        echo "Done"

        printf '%s' "Creating Hermes Channel $i...    "
        # continuation of logs from above command
        ( tail -f -n0 $HERMES_LOGS & ) | grep -q "Success: Channel"
        echo "Done"
    done
fi

# Start hermes in the background and pause until the log message shows that it is up and running
printf '%s' "Starting Hermes...            "

nohup $HERMES_CMD start >> $HERMES_LOGS 2>&1 &
( tail -f -n0 $HERMES_LOGS & ) | grep -q -E "Hermes has started"
echo "Done"

# Start ICQ in the background
printf '%s' "Starting ICQ...               "
nohup $ICQ_CMD run --local >> $ICQ_LOGS 2>&1 &
sleep 5
echo "Done"

# Create a copy of the state that can be used for the "cache" option
echo "Network is ready for transactions.\n"
rm -rf $SCRIPT_DIR/.state.backup
sleep 1
cp -r $SCRIPT_DIR/state $SCRIPT_DIR/.state.backup

if [ "$CACHE" != "true" ]; then
    # Submit a transaction on stride to register the gaia host zone
    echo "Creating Gaia host zone..." | tee -a $TX_LOGS
    $STRIDE_CMD tx stakeibc register-host-zone \
        connection-0 $ATOM_DENOM cosmos $IBC_ATOM_DENOM channel-0 1 \
        --chain-id $STRIDE_CHAIN --home $STATE/stride \
        --keyring-backend test --from $STRIDE_VAL_ACCT --gas 1000000 -y >> $TX_LOGS 2>&1
    WAIT_FOR_BLOCK $STRIDE_LOGS
    echo "Creating Juno host zone..." | tee -a $TX_LOGS
    $STRIDE_CMD tx stakeibc register-host-zone \
    connection-1 $JUNO_DENOM juno $IBC_JUNO_DENOM channel-1 1 \
    --chain-id $STRIDE_CHAIN --home $STATE/stride \
    --keyring-backend test --from $STRIDE_VAL_ACCT --gas 1000000 -y >> $TX_LOGS 2>&1
    WAIT_FOR_BLOCK $STRIDE_LOGS
    echo "Creating Osmo host zone..." | tee -a $TX_LOGS
    $STRIDE_CMD tx stakeibc register-host-zone \
    connection-2 $OSMO_DENOM osmo $IBC_OSMO_DENOM channel-2 1 \
    --chain-id $STRIDE_CHAIN --home $STATE/stride \
    --keyring-backend test --from $STRIDE_VAL_ACCT --gas 1000000 -y >> $TX_LOGS 2>&1
fi
# sleep a while longer to wait for ICA accounts to set up
GAIA_CONFIRM="GAIA.WITHDRAWAL:STRIDE->GAIA}: channel handshake step completed with events: OpenConfirmChannel"
JUNO_CONFIRM="JUNO.WITHDRAWAL:STRIDE->JUNO}: channel handshake step completed with events: OpenConfirmChannel"
OSMO_CONFIRM="OSMO.WITHDRAWAL:STRIDE->OSMO}: channel handshake step completed with events: OpenConfirmChannel"
( tail -f -n0 $HERMES_LOGS & ) | grep -q "$GAIA_CONFIRM"
( tail -f -n1000 $HERMES_LOGS & ) | grep -q "$JUNO_CONFIRM"
( tail -f -n2000 $HERMES_LOGS & ) | grep -q "$OSMO_CONFIRM"

echo "Registering validators on host zones..." | tee -a $TX_LOGS

# send gaia validator 2 money
$GAIA_CMD tx bank send $GAIA_VAL_ACCT $GAIA_VAL_2_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y >> $TX_LOGS 2>&1
WAIT_FOR_NONEMPTY_BLOCK $GAIA_LOGS
# add juno validator
$STRIDE_CMD tx stakeibc add-validator JUNO $JUNO_VAL_ACCT $JUNO_DELEGATE_VAL 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y >> $TX_LOGS 2>&1
WAIT_FOR_NONEMPTY_BLOCK $STRIDE_LOGS
# add osmo validator
$STRIDE_CMD tx stakeibc add-validator OSMO $OSMO_VAL_ACCT $OSMO_DELEGATE_VAL 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y >> $TX_LOGS 2>&1
WAIT_FOR_NONEMPTY_BLOCK $STRIDE_LOGS
# send gaia validator 3 money
# $GAIA_CMD tx bank send gval1 $GAIA_VAL_3_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y >> $TX_LOGS 2>&1
# WAIT_FOR_NONEMPTY_BLOCK $GAIA_LOGS
# add validator 2 for gaia
$STRIDE_CMD tx stakeibc add-validator GAIA gval1 $GAIA_DELEGATE_VAL 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y >> $TX_LOGS 2>&1
WAIT_FOR_NONEMPTY_BLOCK $STRIDE_LOGS
# add validator 2 for gaia
$STRIDE_CMD tx stakeibc add-validator GAIA gval2 $GAIA_DELEGATE_VAL_2 10 10 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y >> $TX_LOGS 2>&1
WAIT_FOR_NONEMPTY_BLOCK $STRIDE_LOGS

# Add more detailed log files
$SCRIPT_DIR/create_logs.sh &

echo "Done! Go get em."
