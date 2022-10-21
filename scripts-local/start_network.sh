#!/bin/bash

# kill previous networks
# loop three times in bash
echo "Killing previous networks..."
for i in {1..5}; do
    make stop-local &> /dev/null
    sleep 1
done

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/vars.sh

mkdir -p $SCRIPT_DIR/logs
mkdir -p $SCRIPT_DIR/logs/rly

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
for log in $STRIDE_LOGS $STRIDE_LOGS_2 $STRIDE_LOGS_3 $STRIDE_LOGS_4 $STRIDE_LOGS_5 $GAIA_LOGS $GAIA_LOGS_2 $HERMES_LOGS $ICQ_LOGS $JUNO_LOGS $TX_LOGS $KEYS_LOGS $OSMO_LOGS $RLY_GAIA_LOGS $RLY_OSMO_LOGS $RLY_JUNO_LOGS; do
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
#  and halts the script until Stride/Gaia have each finalized a block
printf '\n%s' "Starting Stride, Gaia, Osmo, and Juno...   "
nohup $STRIDE_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS 2>&1 &
nohup $STRIDE_CMD_2 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS_2 2>&1 &
nohup $STRIDE_CMD_3 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS_3 2>&1 &
nohup $STRIDE_CMD_4 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS_4 2>&1 &
nohup $STRIDE_CMD_5 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS_5 2>&1 &
nohup $GAIA_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS 2>&1 &
nohup $GAIA_CMD_2 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS_2 2>&1 &
nohup $JUNO_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $JUNO_LOGS 2>&1 &
nohup $OSMO_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $OSMO_LOGS 2>&1 &

( tail -f -n0 $STRIDE_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $STRIDE_LOGS_2 & ) | grep -q "finalizing commit of block"
( tail -f -n0 $STRIDE_LOGS_3 & ) | grep -q "finalizing commit of block"
( tail -f -n0 $STRIDE_LOGS_4 & ) | grep -q "finalizing commit of block"
( tail -f -n0 $STRIDE_LOGS_5 & ) | grep -q "finalizing commit of block"
( tail -f -n0 $GAIA_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $GAIA_LOGS_2 & ) | grep -q "finalizing commit of block"
( tail -f -n0 $JUNO_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $OSMO_LOGS & ) | grep -q "finalizing commit of block"
sleep 2
echo "Done"

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
        ( tail -f -n0 $HERMES_LOGS & ) | grep -q "SUCCESS Channel"
        echo "Done"
    done
fi

# Start hermes in the background and pause until the log message shows that it is up and running
printf '%s' "Starting Hermes...            "
nohup $HERMES_CMD start >> $HERMES_LOGS 2>&1 &
( tail -f -n0 $HERMES_LOGS & ) | grep -q -E "Hermes has started"
echo "Done"

# Start ICQ in the background
printf '%s' "Starting ICQ...     "
source $SCRIPT_DIR/icq_startup.sh &
sleep 5
echo "Done"

# Create a copy of the state that can be used for the "cache" option
echo "Network is ready for transactions!\n"
rm -rf $SCRIPT_DIR/.state.backup
sleep 1
# cp -r $SCRIPT_DIR/state $SCRIPT_DIR/.state.backup

if [ "$CACHE" != "true" ]; then
    bash $SCRIPT_DIR/register_host.sh
fi

# Add more detailed log files
$SCRIPT_DIR/create_logs_gaia.sh &
$SCRIPT_DIR/create_logs_osmo.sh &
$SCRIPT_DIR/create_logs_juno.sh &

echo "Done! Go get em.\n\n"

# bash $SCRIPT_DIR/tests/run_all_tests.sh