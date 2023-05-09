#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

LSM_TOKEN_DENOM=ibc/457E522AB5A620091319195479CADC0638D42A24B49B420BB728F424E9CA60A1
$STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $LSM_TOKEN_DENOM --from staker --gas auto