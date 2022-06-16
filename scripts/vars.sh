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
PORT_ID=26656  # 36564 
MAIN_ID=0

STRIDE_CHAIN=STRIDE
STRIDE_NODE_NAMES=(stride1 stride2 stride3)
STRIDE_MAIN_NODE=${STRIDE_NODE_NAMES[$MAIN_ID]}

STRIDE_VAL_ACCTS=(val1 val2 val3)
STRIDE_MNEMONIC_1="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
STRIDE_MNEMONIC_2="timber vacant teach wedding disease fashion place merge poet produce promote renew sunny industry enforce heavy inch three call sustain deal flee athlete intact"
STRIDE_MNEMONIC_3="enjoy dignity rule multiply kitchen arrange flight rocket kingdom domain motion fire wage viable enough comic cry motor memory fancy dish sing border among"
STRIDE_VAL_KEYS=("$STRIDE_MNEMONIC_1" "$STRIDE_MNEMONIC_2" "$STRIDE_MNEMONIC_3")

stride_run=strided
stride_exec="docker-compose --ansi never exec -T"

STRIDE_RUN_CMDS=()
for node_name in "${STRIDE_NODE_NAMES[@]}"; do
  STRIDE_RUN_CMDS+=( "$stride_run --home $STATE/$node_name" )
done
STRIDE_MAIN_CMD=${STRIDE_RUN_CMDS[$MAIN_ID]}

STRIDE1_EXEC="$stride_exec ${STRIDE_NODE_NAMES[0]} strided --home /stride/.strided --chain-id $STRIDE_CHAIN"
STRIDE2_EXEC="$stride_exec ${STRIDE_NODE_NAMES[1]} strided --home /stride/.strided --chain-id $STRIDE_CHAIN"
STRIDE3_EXEC="$stride_exec ${STRIDE_NODE_NAMES[2]} strided --home /stride/.strided --chain-id $STRIDE_CHAIN"


# define GAIA vars
GAIA_CHAIN=GAIA
GAIA_NODE_NAMES=(gaia1 gaia2 gaia3)
GAIA_MAIN_NODE=${GAIA_NODE_NAMES[$MAIN_ID]}

GAIA_VAL_ACCTS=(gval1 gval2 gval3)
GAIA_MNEMONIC_1="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
GAIA_MNEMONIC_2="social smooth replace total room drip donor science wheel source scare hammer affair fade opinion injury mandate then orbit work worry exhaust diagram hotel"
GAIA_MNEMONIC_3="spike expire grant chef cheese cave someone blue price juice crash field sell camera true wet card saddle oblige where inject process dismiss soft"
GAIA_VAL_KEYS=("$GAIA_MNEMONIC_1" "$GAIA_MNEMONIC_2" "$GAIA_MNEMONIC_3")

gaia_run="docker-compose --ansi never run --rm -T"
gaia_exec="docker-compose --ansi never exec -T"

GAIA_RUN_CMDS=()
for node_name in "${GAIA_NODE_NAMES[@]}"; do
  GAIA_RUN_CMDS+=( "$gaia_run $node_name gaiad --home=/gaia/.gaiad" )
done
GAIA_MAIN_CMD=${GAIA_RUN_CMDS[$MAIN_ID]}

GAIA1_EXEC="$gaia_exec ${GAIA_NODE_NAMES[0]} gaiad --home /gaia/.gaiad"
GAIA2_EXEC="$gaia_exec ${GAIA_NODE_NAMES[1]} gaiad --home /gaia/.gaiad"
GAIA3_EXEC="$gaia_exec ${GAIA_NODE_NAMES[2]} gaiad --home /gaia/.gaiad"

RLY_MNEMONIC_1="alter old invest friend relief slot swear pioneer syrup economy vendor tray focus hedgehog artist legend antenna hair almost donkey spice protect sustain increase"
RLY_MNEMONIC_2="resemble accident lake amateur physical jewel taxi nut demand magnet person blanket trip entire awkward fiber usual current index limb lady lady depart train"

ICQ_RUN="docker-compose --ansi never run --rm -T icq interchain-queries"

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
