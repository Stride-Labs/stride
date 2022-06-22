#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$SCRIPT_DIR/state

# define STRIDE vars
STRIDE_PORT_ID=26657  # 36564 
STRIDE_CHAIN=STRIDE
STRIDE_NODE_NAME=stride
STRIDE_VAL_ACCT=val1
STRIDE_VAL_MNEMONIC="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
STRIDE_CMD="build/strided --home $STATE/stride"

# define GAIA vars
GAIA_CHAIN=GAIA
GAIA_PORT_ID=26658
GAIA_NODE_NAME=gaia
GAIA_VAL_ACCT=gval1
GAIA_VAL_MNEMONIC="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
GAIA_CMD="build/gaiad --home $STATE/gaia"

HERMES_CMD="build/hermes/release/hermes -c $SCRIPT_DIR/hermes/config.toml"

# define relayer vars
HERMES_STRIDE_ACCT=rly1
HERMES_GAIA_ACCT=rly2
HERMES_STRIDE_MNEMONIC="alter old invest friend relief slot swear pioneer syrup economy vendor tray focus hedgehog artist legend antenna hair almost donkey spice protect sustain increase"
HERMES_GAIA_MNEMONIC="resemble accident lake amateur physical jewel taxi nut demand magnet person blanket trip entire awkward fiber usual current index limb lady lady depart train"

ICQ_CMD="build/interchain-queries --home $STATE/icq"

ICQ_STRIDE_ACCT=icq1
ICQ_GAIA_ACCT=icq2
ICQ_STRIDE_MNEMONIC="helmet say goat special plug umbrella finger night flip axis resource tuna trigger angry shove essay point laundry horror eager forget depend siren alarm"
ICQ_GAIA_MNEMONIC="capable later bamboo snow drive afraid cheese practice latin brush hand true visa drama mystery bird client nature jealous guess tank marriage volume fantasy"


CSLEEP() {
  for i in $(seq $1); do
    sleep 1
    printf "\r\t$(($1 - $i))s left..."
  done
}
