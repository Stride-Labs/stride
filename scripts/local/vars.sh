#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

DEPENDENCIES="jq"

# check and install dependencies
echo "\nChecking dependencies... "
deps=0
for name in $DEPENDENCIES
do
    [[ $(type $name 2>/dev/null) ]] || { echo "\n    * $name is required to run this script;";deps=1; }
done
[[ $deps -ne 1 ]] && echo "OK\n" || { echo "\nInstall the missing dependencies and rerun this script...\n"; exit 1; }

STATE=$SCRIPT_DIR/state

# define STRIDE vars
STRIDE_PORT_ID=26657  # 36564 
STRIDE_CHAIN=STRIDE
STRIDE_NODE_NAME=stride
STRIDE_VAL_ACCT=val1
STRIDE_VAL_KEY="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
STRIDE_CMD="build/strided --home $STATE/stride"

# define GAIA vars
GAIA_CHAIN=GAIA
GAIA_PORT_ID=26658
GAIA_NODE_NAME=gaia
GAIA_VAL_ACCT=gval1
GAIA_VAL_KEY="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
GAIA_CMD="build/gaiad --home $STATE/gaia"

# define relayer vars
RLY_NAME_1=rly1
RLY_NAME_2=rly2
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
