#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$SCRIPT_DIR/state

ATOM='uatom'
IBCATOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
BLOCK_TIME='1s'

# define STRIDE vars
STRIDE_PORT_ID=26657  # 36564 
STRIDE_CHAIN=STRIDE
STRIDE_NODE_NAME=stride
STRIDE_VAL_ACCT=val1
STRIDE_VAL_MNEMONIC="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
STRIDE_CMD="$SCRIPT_DIR/../build/strided --home $STATE/stride"

# define GAIA vars
GAIA_CHAIN=GAIA
GAIA_PORT_ID=26658
GAIA_NODE_NAME=gaia
GAIA_VAL_ACCT=gval1
GAIA_REV_ACCT=grev1
GAIA_VAL_MNEMONIC="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
GAIA_REV_MNEMONIC="tonight bonus finish chaos orchard plastic view nurse salad regret pause awake link bacon process core talent whale million hope luggage sauce card weasel"
GAIA_VAL_ADDR="cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2"
GAIA_CMD="$SCRIPT_DIR/../build/gaiad --home $STATE/gaia"

GAIA_VAL_ACCT_2=gval2
GAIA_VAL_MNEMONIC_2="guilt leader matrix lecture infant axis method grain diesel sting reflect brave estate surge october candy busy crash parade club practice sure gentle celery"
GAIA_VAL_2_ADDR="cosmos133lfs9gcpxqj6er3kx605e3v9lqp2pg54sreu3"
GAIA_VAL_ACCT_3=gval3
GAIA_VAL_MNEMONIC_3="fire tape spread wing click winter awful ozone visa spray swear color table settle review rival meadow gauge speed tide timber disease float live"
GAIA_VAL_3_ADDR="cosmos1fumal3j4lxzjp22fzffge8mw56qm33h9ez0xy2"

HERMES_CMD="$SCRIPT_DIR/../build/hermes/release/hermes -c $SCRIPT_DIR/hermes/config.toml"

# define relayer vars
HERMES_STRIDE_ACCT=rly1
HERMES_GAIA_ACCT=rly2
HERMES_STRIDE_MNEMONIC="alter old invest friend relief slot swear pioneer syrup economy vendor tray focus hedgehog artist legend antenna hair almost donkey spice protect sustain increase"
HERMES_GAIA_MNEMONIC="resemble accident lake amateur physical jewel taxi nut demand magnet person blanket trip entire awkward fiber usual current index limb lady lady depart train"

ICQ_CMD="$SCRIPT_DIR/../build/interchain-queries --home $STATE/icq"

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