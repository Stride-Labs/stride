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
$STRIDE_CMD tx bank send $STRIDE_VAL_ADDR $STRIDE_ADMIN_ADDRESS 10000000ustrd --from $STRIDE_VAL_ACCT -y >> $TX_LOGS 2>&1
WAIT_FOR_BLOCK $STRIDE_LOGS 2

# Submit a transaction on stride to register the gaia host zone
echo "Creating Gaia host zone..." | tee -a $TX_LOGS
$STRIDE_CMD tx stakeibc register-host-zone \
    connection-0 $ATOM_DENOM cosmos $IBC_ATOM_DENOM channel-0 1 \
    --chain-id $STRIDE_CHAIN --home $STATE/stride \
    --keyring-backend test --from $STRIDE_ADMIN_ACCT --gas 1000000 -y >> $TX_LOGS 2>&1
WAIT_FOR_BLOCK $STRIDE_LOGS 2
echo "Creating Osmo host zone..." | tee -a $TX_LOGS
$STRIDE_CMD tx stakeibc register-host-zone \
connection-1 $OSMO_DENOM osmo $IBC_OSMO_DENOM channel-1 1 \
--chain-id $STRIDE_CHAIN --home $STATE/stride \
--keyring-backend test --from $STRIDE_ADMIN_ACCT --gas 1000000 -y >> $TX_LOGS 2>&1
# echo "Creating Juno host zone..." | tee -a $TX_LOGS
# $STRIDE_CMD tx stakeibc register-host-zone \
# connection-1 $JUNO_DENOM juno $IBC_JUNO_DENOM channel-1 1 \
# --chain-id $STRIDE_CHAIN --home $STATE/stride \
# --keyring-backend test --from $STRIDE_ADMIN_ACCT --gas 1000000 -y >> $TX_LOGS 2>&1
# WAIT_FOR_BLOCK $STRIDE_LOGS 2

# sleep a while longer to wait for ICA accounts to set up
GAIA_CONFIRM="GAIA.WITHDRAWAL:STRIDE->GAIA}: channel handshake step completed with events: OpenConfirmChannel"
JUNO_CONFIRM="JUNO.WITHDRAWAL:STRIDE->JUNO}: channel handshake step completed with events: OpenConfirmChannel"
OSMO_CONFIRM="OSMO.WITHDRAWAL:STRIDE->OSMO}: channel handshake step completed with events: OpenConfirmChannel"
( tail -f -n500 $HERMES_LOGS & ) | grep -q "$GAIA_CONFIRM"
# ( tail -f -n1000 $HERMES_LOGS & ) | grep -q "$JUNO_CONFIRM"
( tail -f -n2000 $HERMES_LOGS & ) | grep -q "$OSMO_CONFIRM"

echo "Registering validators on host zones..." | tee -a $TX_LOGS

# send gaia validator 2 money
$GAIA_CMD tx bank send $GAIA_VAL_ACCT $GAIA_VAL_2_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y >> $TX_LOGS 2>&1
WAIT_FOR_BLOCK $GAIA_LOGS
# add juno validator
# $STRIDE_CMD tx stakeibc add-validator JUNO $JUNO_VAL_ACCT $JUNO_DELEGATE_VAL 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_ADMIN_ACCT -y >> $TX_LOGS 2>&1
# WAIT_FOR_BLOCK $STRIDE_LOGS 2
# # add osmo validator
$STRIDE_CMD tx stakeibc add-validator OSMO $OSMO_VAL_ACCT $OSMO_DELEGATE_VAL 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_ADMIN_ACCT -y >> $TX_LOGS 2>&1
WAIT_FOR_BLOCK $STRIDE_LOGS 2
# add validator 2 for gaia
$STRIDE_CMD tx stakeibc add-validator GAIA gval1 $GAIA_DELEGATE_VAL 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_ADMIN_ACCT -y >> $TX_LOGS 2>&1
WAIT_FOR_BLOCK $STRIDE_LOGS 2
# add validator 2 for gaia
$STRIDE_CMD tx stakeibc add-validator GAIA gval2 $GAIA_DELEGATE_VAL_2 10 10 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_ADMIN_ACCT -y >> $TX_LOGS 2>&1
WAIT_FOR_BLOCK $STRIDE_LOGS 2

echo "Done"
