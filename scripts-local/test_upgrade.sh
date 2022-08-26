#!/bin/bash

# kill previous networks
# loop three times in bash
echo "Killing previous networks..."
for i in {1..3}; do
    make stop &> /dev/null
    sleep 1
done

UPGRADE_NAME="v2"
export DAEMON_NAME=strided
export DAEMON_HOME=$SCRIPT_DIR/state/stride
export DAEMON_RESTART_AFTER_UPGRADE=true

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
rm -rf $SCRIPT_DIR/logs/*.log $SCRIPT_DIR/logs/temp

# Recreate each log file
for log in $STRIDE_LOGS $STRIDE_LOGS_2 $STRIDE_LOGS_3 $STRIDE_LOGS_4 $STRIDE_LOGS_5 $GAIA_LOGS $GAIA_LOGS_2 $HERMES_LOGS $ICQ_LOGS $JUNO_LOGS $TX_LOGS $KEYS_LOGS $OSMO_LOGS $RLY_GAIA_LOGS $RLY_OSMO_LOGS $RLY_JUNO_LOGS; do
    touch $log
done

# Setup upgrade and cosmovisor directories
mkdir -p $SCRIPT_DIR/upgrades/cosmovisor/genesis/bin/
mkdir -p $SCRIPT_DIR/upgrades/cosmovisor/upgrades/$UPGRADE_NAME/bin/
mkdir -p $SCRIPT_DIR/state/stride/cosmovisor

rm -f $SCRIPT_DIR/upgrades/binaries/strided2
cp $SCRIPT_DIR/upgrades/binaries/strided1 $SCRIPT_DIR/upgrades/cosmovisor/genesis/bin/strided
cp $SCRIPT_DIR/../build/strided $SCRIPT_DIR/upgrades/cosmovisor/upgrades/$UPGRADE_NAME/bin/strided
cp $SCRIPT_DIR/../build/strided $SCRIPT_DIR/upgrades/binaries/strided

printf '\n%s' "Starting Stride validators...   "
for sname in stride stride2 stride3 stride4 stride5; do
    mkdir -p $SCRIPT_DIR/state/$sname/cosmovisor
    cp -r $SCRIPT_DIR/upgrades/cosmovisor/* $STATE/$sname/cosmovisor/
done

DAEMON_HOME="$STATE/stride"; nohup cosmovisor run start --home $STATE/stride | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS 2>&1 &
DAEMON_HOME="$STATE/stride2"; nohup cosmovisor run start --home $STATE/stride2 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS_2 2>&1 &
DAEMON_HOME="$STATE/stride3"; nohup cosmovisor run start --home $STATE/stride3 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS_3 2>&1 &
DAEMON_HOME="$STATE/stride4"; nohup cosmovisor run start --home $STATE/stride4 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS_4 2>&1 &
DAEMON_HOME="$STATE/stride5"; nohup cosmovisor run start --home $STATE/stride5 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS_5 2>&1 &

( tail -f -n0 $STRIDE_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $STRIDE_LOGS_2 & ) | grep -q "finalizing commit of block"
( tail -f -n0 $STRIDE_LOGS_3 & ) | grep -q "finalizing commit of block"
( tail -f -n0 $STRIDE_LOGS_4 & ) | grep -q "finalizing commit of block"
( tail -f -n0 $STRIDE_LOGS_5 & ) | grep -q "finalizing commit of block"
sleep 5
echo "Done"

# Create a copy of the state that can be used for the "cache" option
echo "Network is ready for transactions.\n"

# Propose upgrades
bash $SCRIPT_DIR/upgrades/submit_upgrade.sh

tail -f $SCRIPT_DIR/logs/stride.log 
