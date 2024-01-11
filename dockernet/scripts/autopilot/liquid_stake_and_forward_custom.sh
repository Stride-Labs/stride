#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh


# Check balances before
GAIA_ADDRESS_BALANCE_BEFORE=$($GAIA_MAIN_CMD q bank balances $(GAIA_ADDRESS))
OSMO_ADDRESS_BALANCE_BEFORE=$($OSMO_MAIN_CMD q bank balances $(OSMO_ADDRESS))
echo "GAIA_ADDRESS_BALANCE_BEFORE: $GAIA_ADDRESS_BALANCE_BEFORE"
echo "OSMO_ADDRESS_BALANCE_BEFORE: $OSMO_ADDRESS_BALANCE_BEFORE"

# NOTE: Requires GAIA and OSMO as the host zones (configured in config.sh)
memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'", "stakeibc": { "action": "LiquidStake", "ibc_receiver": "'$(OSMO_ADDRESS)'", "transfer_channel": "channel-1" } } }'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 10000uatom --from ${GAIA_VAL_PREFIX}1 -y
sleep 10

# Check balances before
GAIA_ADDRESS_BALANCE_AFTER=$($GAIA_MAIN_CMD q bank balances $(GAIA_ADDRESS))
OSMO_ADDRESS_BALANCE_AFTER=$($OSMO_MAIN_CMD q bank balances $(OSMO_ADDRESS))
echo "GAIA_ADDRESS_BALANCE_AFTER: $GAIA_ADDRESS_BALANCE_AFTER"
echo "OSMO_ADDRESS_BALANCE_AFTER: $OSMO_ADDRESS_BALANCE_AFTER"
