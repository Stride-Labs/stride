#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/vars.sh

# Override stride command to use old binary
STRIDE_CMD="$SCRIPT_DIR/upgrades/binaries/strided 1 --home $SCRIPT_DIR/state/stride"

export DAEMON_NAME=strided
export DAEMON_HOME=$STATE/stride
export DAEMON_RESTART_AFTER_UPGRADE=true

MAX_DEPOSIT_PERIOD="60s"
VOTING_PERIOD="60s"
PROPOSAL_NAME=v2
UPGRADE_HEIGHT=65

mkdir -p $SCRIPT_DIR/logs

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

# Initialize state for Stride, Gaia, and relayers
sh ${SCRIPT_DIR}/init_stride.sh
sh ${SCRIPT_DIR}/init_gaia.sh
sh ${SCRIPT_DIR}/init_relayers.sh

# Shorten stride voting period
sed -i -E "s|max_deposit_period\": \"172800s\"|max_deposit_period\": \"${MAX_DEPOSIT_PERIOD}\"|g" "${STRIDE_STATE}/config/genesis.json"
sed -i -E "s|voting_period\": \"172800s\"|voting_period\": \"${VOTING_PERIOD}\"|g" "${STRIDE_STATE}/config/genesis.json"

# Setup upgrade and cosmovisor directories
mkdir -p $SCRIPT_DIR/upgrades/cosmovisor/genesis/bin/
mkdir -p $SCRIPT_DIR/upgrades/cosmovisor/upgrades/v2/bin/
mkdir -p $STATE/stride/cosmovisor

cp $SCRIPT_DIR/../build/strided $SCRIPT_DIR/upgrades/binaries/strided2
cp $SCRIPT_DIR/upgrades/binaries/strided1 $SCRIPT_DIR/upgrades/cosmovisor/genesis/bin/strided
cp $SCRIPT_DIR/upgrades/binaries/strided2 $SCRIPT_DIR/upgrades/cosmovisor/upgrades/v2/bin/strided

printf '\n%s' "Starting Stride and Gaia...   "
cp -r $SCRIPT_DIR/../../upgrades-stride/cosmovisor/* $STATE/stride/cosmovisor/
nohup cosmovisor run start --home $STATE/stride | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS 2>&1 &

nohup $GAIA_CMD start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS 2>&1 &
nohup $GAIA_CMD_2 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS_2 2>&1 &
# nohup $GAIA_CMD_3 start | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS_3 2>&1 &

( tail -f -n0 $STRIDE_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $GAIA_LOGS & ) | grep -q "finalizing commit of block"
sleep 5
echo "Done"

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
( tail -f -n0 $HERMES_LOGS & ) | grep -q "Success: Channel"
echo "Done"

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

# Add more detailed log files
$SCRIPT_DIR/create_logs.sh &

# Propose upgrades
bash $SCRIPT_DIR/upgrades/submit_upgrade.sh

tail -f $SCRIPT_DIR/state/stride/scripts-local/logs/stride.log