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
STRIDE_CHAINS=(STRIDE STRIDE STRIDE)
STRIDE_DOCKER_NAMES=(stride1 stride2 stride3)
MAIN_ID=0
main_chain=${STRIDE_CHAINS[$MAIN_ID]}
main_node=${STRIDE_DOCKER_NAMES[$MAIN_ID]}
PORT_ID=26656  # 36564 

VAL_ACCTS=(val1 val2 val3)

V1="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
V2="timber vacant teach wedding disease fashion place merge poet produce promote renew sunny industry enforce heavy inch three call sustain deal flee athlete intact"
V3="enjoy dignity rule multiply kitchen arrange flight rocket kingdom domain motion fire wage viable enough comic cry motor memory fancy dish sing border among"
VKEYS=("$V1" "$V2" "$V3")

BASE_RUN=strided

ST_CMDS=()
for chain_name in "${STRIDE_DOCKER_NAMES[@]}"; do
  ST_CMDS+=( "$BASE_RUN --home $STATE/$chain_name" )
done
main_cmd=${ST_CMDS[$MAIN_ID]}


# define GAIA vars
GAIA_CHAINS=(GAIA GAIA GAIA)
GAIA_DOCKER_NAMES=(gaia1 gaia2 gaia3)
GVAL_ACCTS=(gval1 gval2 gval3)
main_gaia_chain=${GAIA_CHAINS[$MAIN_ID]}
main_gaia_node=${GAIA_DOCKER_NAMES[$MAIN_ID]}

GAIA_RUN="docker-compose --ansi never run -T"

GV1="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
GV2="social smooth replace total room drip donor science wheel source scare hammer affair fade opinion injury mandate then orbit work worry exhaust diagram hotel"
GV3="spike expire grant chef cheese cave someone blue price juice crash field sell camera true wet card saddle oblige where inject process dismiss soft"
GVKEYS=("$GV1" "$GV2" "$GV3")

RLY_MNEMONIC_1="alter old invest friend relief slot swear pioneer syrup economy vendor tray focus hedgehog artist legend antenna hair almost donkey spice protect sustain increase"
RLY_MNEMONIC_2="resemble accident lake amateur physical jewel taxi nut demand magnet person blanket trip entire awkward fiber usual current index limb lady lady depart train"


GAIA_CMDS=()
for docker_name in "${GAIA_DOCKER_NAMES[@]}"; do
  GAIA_CMDS+=( "$GAIA_RUN $docker_name gaiad --home=/gaia/.gaiad" )
done
main_gaia_cmd=${GAIA_CMDS[$MAIN_ID]}

ICQ_STRIDE_KEY="helmet say goat special plug umbrella finger night flip axis resource tuna trigger angry shove essay point laundry horror eager forget depend siren alarm"
ICQ_GAIA_KEY="capable later bamboo snow drive afraid cheese practice latin brush hand true visa drama mystery bird client nature jealous guess tank marriage volume fantasy"

CSLEEP() {
  for i in $(seq $1); do
    sleep 1
    printf "\r\t$(($1 - $i))s left..."
  done
}

# ICQ
ICQ_RUN="docker-compose --ansi never run -T icq interchain-queries"
