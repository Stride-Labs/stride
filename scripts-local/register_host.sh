#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/vars.sh

if [ -z ${STRIDE_ADMIN_MNEMONIC+x} ]; then
    echo "ERROR: STRIDE_ADMIN_MNEMONIC must be set as an environment variable in order to register a host zone."
    exit 1
fi

# Add stride admin to keychain
STRIDE_ADMIN_ACCT=admin
echo $STRIDE_ADMIN_MNEMONIC | $STRIDE_CMD keys add $STRIDE_ADMIN_ACCT --recover &> /dev/null

echo "Funding stride admin account..."
STRIDE_ADMIN_ADDRESS=$($STRIDE_CMD keys show $STRIDE_ADMIN_ACCT -a) 
$STRIDE_CMD tx bank send $STRIDE_VAL_ADDR $STRIDE_ADMIN_ADDRESS 10000000ustrd --from $STRIDE_VAL_ACCT -y
sleep 10

# Submit a transaction on stride to register the gaia host zone
printf "\nCreating host zone...\n"
$STRIDE_CMD tx stakeibc register-host-zone \
    connection-0 $ATOM_DENOM cosmos $IBC_ATOM_DENOM channel-0 1 \
    --home $STATE/stride --from $STRIDE_ADMIN_ACCT --gas 1000000 -y

# sleep a while longer to wait for ICA accounts to set up
sleep 60

printf "\nRegistering validators on host zone...\n"
CSLEEP 10
$GAIA_CMD tx bank send gval1 $GAIA_VAL_2_ADDR 10000uatom -y
CSLEEP 10
$GAIA_CMD tx bank send gval1 $GAIA_VAL_3_ADDR 10000uatom -y

CSLEEP 10
# weights must be high so that we can slash them with reasonable precision
$STRIDE_CMD tx stakeibc add-validator GAIA gval1 $GAIA_DELEGATE_VAL 10 5000000 --from $STRIDE_ADMIN_ACCT -y
CSLEEP 30
$STRIDE_CMD tx stakeibc add-validator GAIA gval2 $GAIA_DELEGATE_VAL_2 10 10000000 --from $STRIDE_ADMIN_ACCT -y
CSLEEP 30

echo "Done"