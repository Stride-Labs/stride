#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# NOTE: Requires GAIA and OSMO as the host zones (configured in config.sh)
memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'", "stakeibc": { "action": "LiquidStake", "ibc_receiver": "'$(OSMO_ADDRESS)'", "transfer_channel": "channel-1" } } }'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 10000uatom --from ${GAIA_VAL_PREFIX}1 -y