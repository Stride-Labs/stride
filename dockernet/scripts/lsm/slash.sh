#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh


validator_address_4=$(GET_VAL_ADDR GAIA 4)
$GAIA_MAIN_CMD q staking validator $validator_address_4
echo ">>> Sleeping validator 4 for 60 seconds to incur a slash"
docker pause dockernet-gaia4-1
sleep 60
docker unpause dockernet-gaia4-1
$GAIA_MAIN_CMD q staking validator $validator_address_4
