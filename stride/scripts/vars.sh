#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

DEPENDENCIES="jq"

# check and install dependencies
echo "\nChecking dependencies... "
for name in $DEPENDENCIES
do
    [[ $(type $name 2>/dev/null) ]] || { echo "\n    * $name is required to run this script;";deps=1; }
done
[[ $deps -ne 1 ]] && echo "OK\n" || { echo "\nInstall the missing dependencies and rerun this script...\n"; exit 1; }

# define vars
STATE=$SCRIPT_DIR/state
STRIDE_CHAINS=(STRIDE_1 STRIDE_2 STRIDE_3)
GAIA_NODES=(GAIA_1)
STRIDE_DOCKER_NAMES=(stride1 stride2 stride3)
MAIN_ID=0
main_chain=${STRIDE_CHAINS[$MAIN_ID]}
PORT_ID=26656  # 36564 

VAL_ACCTS=(val1 val2 val3)

V1="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
V2="timber vacant teach wedding disease fashion place merge poet produce promote renew sunny industry enforce heavy inch three call sustain deal flee athlete intact"
V3="enjoy dignity rule multiply kitchen arrange flight rocket kingdom domain motion fire wage viable enough comic cry motor memory fancy dish sing border among"
VKEYS=("$V1" "$V2" "$V3")
BASE_RUN=strided

ST_CMDS=()
for chain_name in "${STRIDE_CHAINS[@]}"; do
  ST_CMDS+=( "$BASE_RUN --home $STATE/$chain_name" )
main_cmd=${ST_CMDS[$MAIN_ID]}
done