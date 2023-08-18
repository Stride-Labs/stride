#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

LSM_TOKEN_DENOM=ibc/19825915130745DAC8CBC51A6DBE4FC7644463CF4254CD46D49B15AABEE73FB8
$STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $LSM_TOKEN_DENOM --from staker --gas auto -y