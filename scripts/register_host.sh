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
GAIA_VAL_2_ADDR="cosmos133lfs9gcpxqj6er3kx605e3v9lqp2pg54sreu3" 
GAIA_VAL_3_ADDR="cosmos1fumal3j4lxzjp22fzffge8mw56qm33h9ez0xy2" 
GAIA_DELEGATE_VAL_1='cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne' 
GAIA_DELEGATE_VAL_2='cosmosvaloper133lfs9gcpxqj6er3kx605e3v9lqp2pg5syhvsz' 

CSLEEP 10
$MAIN_GAIA_CMD tx bank send gval1 $GAIA_VAL_2_ADDR 10000uatom -y
CSLEEP 10
$MAIN_GAIA_CMD tx bank send gval1 $GAIA_VAL_3_ADDR 10000uatom -y

CSLEEP 10
$MAIN_STRIDE_CMD tx stakeibc add-validator $GAIA_CHAIN_ID gval1 $GAIA_DELEGATE_VAL_1 10 5 --from $STRIDE_ADMIN_ACCT -y
CSLEEP 30
$MAIN_STRIDE_CMD tx stakeibc add-validator $GAIA_CHAIN_ID gval2 $GAIA_DELEGATE_VAL_2 10 10 --from $STRIDE_ADMIN_ACCT -y
CSLEEP 30

echo "Done"