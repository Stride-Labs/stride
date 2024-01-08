#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'", "stakeibc": { "action": "LiquidStake" } } }'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 10000uatom --from ${GAIA_VAL_PREFIX}1 -y