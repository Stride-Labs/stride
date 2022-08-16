#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$SCRIPT_DIR/state
PEER_PORT=26656

# DENOMS
IBC_STRD_DENOM='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'
IBC_ATOM_DENOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
ATOM_DENOM='uatom'
STRD_DENOM='ustrd'
STATOM_DENOM="stuatom"

BLOCK_TIME='5s'
DAY_EPOCH_INDEX=1
DAY_EPOCH_DURATION="100s"
STRIDE_EPOCH_INDEX=2
STRIDE_EPOCH_DURATION="40s"
UNBONDING_TIME="200s"
MAX_DEPOSIT_PERIOD="30s"
VOTING_PERIOD="30s"

VAL_TOKENS=5000000000000
STAKE_TOKENS=3000000000000
ADMIN_TOKENS=1000000000

# STRIDE vars
STRIDE_CHAIN_ID=STRIDE
STRIDE_NODE_PREFIX=stride
STRIDE_NUM_NODES=3
STRIDE_CMD="$SCRIPT_DIR/../build/strided"
STRIDE_VAL_PREFIX=val
STRIDE_DENOM=$STRD_DENOM
STRIDE_RPC_PORT=26657
STRIDE_ADMIN_ACCT=admin
MAIN_STRIDE_CMD="$STRIDE_CMD --home $SCRIPT_DIR/state/stride1"

STRIDE_MNEMONIC_1="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
STRIDE_MNEMONIC_2="timber vacant teach wedding disease fashion place merge poet produce promote renew sunny industry enforce heavy inch three call sustain deal flee athlete intact"
STRIDE_MNEMONIC_3="enjoy dignity rule multiply kitchen arrange flight rocket kingdom domain motion fire wage viable enough comic cry motor memory fancy dish sing border among"
STRIDE_MNEMONIC_4="vacant margin wave rice brush drastic false rifle tape critic volcano worry tumble assist pulp swamp sheriff stairs decorate chaos empower already obvious caught"
STRIDE_MNEMONIC_5="river spin follow make trash wreck clever increase dial divert meadow abuse victory able foot kid sell bench embody river income utility dismiss timber"
STRIDE_VAL_MNEMONICS=("$STRIDE_MNEMONIC_1" "$STRIDE_MNEMONIC_2" "$STRIDE_MNEMONIC_3" "$STRIDE_MNEMONIC_4" "$STRIDE_MNEMONIC_5")

# GAIA vars
GAIA_CHAIN_ID=GAIA
GAIA_NODE_PREFIX=gaia
GAIA_NUM_NODES=3
GAIA_CMD="$SCRIPT_DIR/../build/gaiad"
GAIA_VAL_PREFIX=gval
GAIA_DENOM=$ATOM_DENOM
GAIA_RPC_PORT=26557
MAIN_GAIA_CMD="$GAIA_CMD --home $SCRIPT_DIR/state/gaia1"

GAIA_MNEMONIC_1="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
GAIA_MNEMONIC_2="guilt leader matrix lecture infant axis method grain diesel sting reflect brave estate surge october candy busy crash parade club practice sure gentle celery"
GAIA_MNEMONIC_3="fire tape spread wing click winter awful ozone visa spray swear color table settle review rival meadow gauge speed tide timber disease float live"
GAIA_MNEMONIC_4="curtain mom patrol rifle list lamp interest hard lock stairs display world disagree ten fantasy engine van explain chunk social smile detail initial typical"
GAIA_MNEMONIC_5="invite close edit quick effort mosquito ocean north term spread dial throw human review west bike mandate learn cabin bubble remove unlock lab unique"
GAIA_VAL_MNEMONICS=("$GAIA_MNEMONIC_1" "$GAIA_MNEMONIC_2" "$GAIA_MNEMONIC_3" "$GAIA_MNEMONIC_4" "$GAIA_MNEMONIC_5")

# define relayer vars
HERMES_CMD="$SCRIPT_DIR/../build/hermes/release/hermes -c $STATE/hermes/config.toml"
HERMES_EXEC="docker-compose run --rm hermes hermes"

HERMES_STRIDE_ACCT=rly1
HERMES_GAIA_ACCT=rly2
HERMES_OSMOSIS_ACCT=rly3

HERMES_STRIDE_MNEMONIC="alter old invest friend relief slot swear pioneer syrup economy vendor tray focus hedgehog artist legend antenna hair almost donkey spice protect sustain increase"
HERMES_GAIA_MNEMONIC="resemble accident lake amateur physical jewel taxi nut demand magnet person blanket trip entire awkward fiber usual current index limb lady lady depart train"
HERMES_OSMOSIS_MNEMONIC="artwork ranch dinosaur maple unhappy office bone vote rebel slot outside benefit innocent wrist certain cradle almost fat trial build chicken enroll strike milk"

RELAYER_CMD="$SCRIPT_DIR/../build/relayer --home $STATE/relayer"
RELAYER_EXEC="docker-compose run --rm relayer rly"

RELAYER_STRIDE_ACCT=rly1
RELAYER_GAIA_ACCT=rly2
RELAYER_OSMOSIS_ACCT=rly3

RELAYER_STRIDE_MNEMONIC="alter old invest friend relief slot swear pioneer syrup economy vendor tray focus hedgehog artist legend antenna hair almost donkey spice protect sustain increase"
RELAYER_GAIA_MNEMONIC="resemble accident lake amateur physical jewel taxi nut demand magnet person blanket trip entire awkward fiber usual current index limb lady lady depart train"
RELAYER_OSMOSIS_MNEMONIC="artwork ranch dinosaur maple unhappy office bone vote rebel slot outside benefit innocent wrist certain cradle almost fat trial build chicken enroll strike milk"

ICQ_CMD="$SCRIPT_DIR/../build/interchain-queries --home $STATE/icq"
ICQ_EXEC="docker-compose run --rm icq interchain-queries"

ICQ_STRIDE_ACCT=icq1
ICQ_GAIA_ACCT=icq2
ICQ_OSMOSIS_ACCT=icq3

ICQ_STRIDE_MNEMONIC="helmet say goat special plug umbrella finger night flip axis resource tuna trigger angry shove essay point laundry horror eager forget depend siren alarm"
ICQ_GAIA_MNEMONIC="capable later bamboo snow drive afraid cheese practice latin brush hand true visa drama mystery bird client nature jealous guess tank marriage volume fantasy"
ICQ_OSMOSIS_MNEMONIC="rival inch buzz slow high dynamic antique idle switch evolve math virus direct health simple capital place mutual air orphan champion prefer garage over"


CSLEEP() {
  for i in $(seq $1); do
    sleep 1
    printf "\r\t$(($1 - $i))s left..."
  done
}
