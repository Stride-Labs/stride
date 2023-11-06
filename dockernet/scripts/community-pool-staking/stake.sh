#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

$GAIA_MAIN_CMD tx bank send $(GAIA_ADDRESS) $(GET_ICA_ADDR GAIA community_pool_deposit) 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y