#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/vars.sh

# Submit a transaction on stride to register the gaia host zone
printf "\nCreating host zone...\n"
$MAIN_STRIDE_CMD tx stakeibc register-host-zone \
    connection-0 $ATOM_DENOM cosmos $IBC_ATOM_DENOM channel-0 1 \
    --from $STRIDE_ADMIN_ACCT --gas 1000000 --home $SCRIPT_DIR/state/stride1 -y

# sleep a while longer to wait for ICA accounts to set up
sleep 60

printf "\nRegistering validators on host zone...\n"
GAIA_VAL_2_ADDR="$($MAIN_GAIA_CMD keys show gval2 -a)"
GAIA_VAL_3_ADDR="$($MAIN_GAIA_CMD keys show gval3 -a)"
GAIA_DELEGATE_VAL_1="$($MAIN_GAIA_CMD q staking validators | grep GAIA_1 -A 5 | grep operator | awk '{print $2}')"
GAIA_DELEGATE_VAL_2="$($MAIN_GAIA_CMD q staking validators | grep GAIA_2 -A 5 | grep operator | awk '{print $2}')"

$MAIN_STRIDE_CMD tx stakeibc add-validator $GAIA_CHAIN_ID gval1 $GAIA_DELEGATE_VAL_1 10 5 --from $STRIDE_ADMIN_ACCT -y
CSLEEP 10
$MAIN_STRIDE_CMD tx stakeibc add-validator $GAIA_CHAIN_ID gval2 $GAIA_DELEGATE_VAL_2 10 10 --from $STRIDE_ADMIN_ACCT -y
CSLEEP 30

echo "Done"