#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$SCRIPT_DIR/state
PEER_PORT=26656

# DENOMS
IBC_STRD_DENOM='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'
IBC_STRD_DENOM_JUNO='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'
IBC_STRD_DENOM_OSMO='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'
IBC_ATOM_DENOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
IBC_JUNO_DENOM='ibc/EFF323CC632EC4F747C61BCE238A758EFDB7699C3226565F7C20DA06509D59A5' 
IBC_OSMO_DENOM='ibc/13B2C536BB057AC79D5616B8EA1B9540EC1F2170718CAFF6F0083C966FFFED0B'  
ATOM_DENOM='uatom'
JUNO_DENOM='ujuno'
OSMO_DENOM='uosmo'
STRD_DENOM='ustrd'
STATOM_DENOM="stuatom"
STJUNO_DENOM="stujuno"
STOSMO_DENOM="stuosmo"

# CHAIN PARAMS
BLOCK_TIME='1s'
STRIDE_DAY_EPOCH_DURATION="100s"
STRIDE_EPOCH_EPOCH_DURATION="40s"
HOST_DAY_EPOCH_DURATION="60s"
HOST_HOUR_EPOCH_DURATION="60s"
HOST_WEEK_EPOCH_DURATION="60s"
UNBONDING_TIME="120s"
MAX_DEPOSIT_PERIOD="30s"
VOTING_PERIOD="30s"

VAL_TOKENS=5000000000000
STAKE_TOKENS=3000000000000
ADMIN_TOKENS=1000000000

# STRIDE 
STRIDE_CHAIN_ID=STRIDE
STRIDE_NODE_PREFIX=stride
STRIDE_NUM_NODES=3
STRIDE_CMD="$SCRIPT_DIR/../build/strided"
STRIDE_VAL_PREFIX=val
STRIDE_DENOM=$STRD_DENOM
STRIDE_RPC_PORT=26657
STRIDE_ADMIN_ACCT=admin
STRIDE_MAIN_CMD="$STRIDE_CMD --home $SCRIPT_DIR/state/${STRIDE_NODE_PREFIX}1"

STRIDE_MNEMONIC_1="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
STRIDE_MNEMONIC_2="timber vacant teach wedding disease fashion place merge poet produce promote renew sunny industry enforce heavy inch three call sustain deal flee athlete intact"
STRIDE_MNEMONIC_3="enjoy dignity rule multiply kitchen arrange flight rocket kingdom domain motion fire wage viable enough comic cry motor memory fancy dish sing border among"
STRIDE_MNEMONIC_4="vacant margin wave rice brush drastic false rifle tape critic volcano worry tumble assist pulp swamp sheriff stairs decorate chaos empower already obvious caught"
STRIDE_MNEMONIC_5="river spin follow make trash wreck clever increase dial divert meadow abuse victory able foot kid sell bench embody river income utility dismiss timber"
STRIDE_VAL_MNEMONICS=("$STRIDE_MNEMONIC_1","$STRIDE_MNEMONIC_2","$STRIDE_MNEMONIC_3","$STRIDE_MNEMONIC_4","$STRIDE_MNEMONIC_5")

# GAIA 
GAIA_CHAIN_ID=GAIA
GAIA_NODE_PREFIX=gaia
GAIA_NUM_NODES=2
GAIA_CMD="$SCRIPT_DIR/../build/gaiad"
GAIA_VAL_PREFIX=gval
GAIA_REV_ACCT=grev1
GAIA_ADDRESS_PREFIX=cosmos
GAIA_DENOM=$ATOM_DENOM
GAIA_IBC_DENOM=$IBC_ATOM_DENOM
GAIA_RPC_PORT=26557
GAIA_MAIN_CMD="$GAIA_CMD --home $SCRIPT_DIR/state/${GAIA_NODE_PREFIX}1"

GAIA_REV_MNEMONIC="tonight bonus finish chaos orchard plastic view nurse salad regret pause awake link bacon process core talent whale million hope luggage sauce card weasel"
GAIA_VAL_MNEMONIC_1="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
GAIA_VAL_MNEMONIC_2="guilt leader matrix lecture infant axis method grain diesel sting reflect brave estate surge october candy busy crash parade club practice sure gentle celery"
GAIA_VAL_MNEMONIC_3="fire tape spread wing click winter awful ozone visa spray swear color table settle review rival meadow gauge speed tide timber disease float live"
GAIA_VAL_MNEMONIC_4="curtain mom patrol rifle list lamp interest hard lock stairs display world disagree ten fantasy engine van explain chunk social smile detail initial typical"
GAIA_VAL_MNEMONIC_5="invite close edit quick effort mosquito ocean north term spread dial throw human review west bike mandate learn cabin bubble remove unlock lab unique"
GAIA_VAL_MNEMONICS=("$GAIA_VAL_MNEMONIC_1","$GAIA_VAL_MNEMONIC_2","$GAIA_VAL_MNEMONIC_3","$GAIA_VAL_MNEMONIC_4","$GAIA_VAL_MNEMONIC_5")

# JUNO 
JUNO_CHAIN_ID=JUNO
JUNO_NODE_PREFIX=juno
JUNO_NUM_NODES=2
JUNO_CMD="$SCRIPT_DIR/../build/junod"
JUNO_VAL_PREFIX=jval
JUNO_REV_ACCT=jrev1
JUNO_ADDRESS_PREFIX=juno
JUNO_DENOM=$JUNO_DENOM
JUNO_IBC_DENOM=$IBC_JUNO_DENOM
JUNO_RPC_PORT=26457
JUNO_MAIN_CMD="$JUNO_CMD --home $SCRIPT_DIR/state/${JUNO_NODE_PREFIX}1"

JUNO_REV_MNEMONIC="tonight bonus finish chaos orchard plastic view nurse salad regret pause awake link bacon process core talent whale million hope luggage sauce card weasel"
JUNO_VAL_MNEMONIC_1="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
JUNO_VAL_MNEMONIC_2="acoustic prize donkey space pitch visa labor enable sting sort safe conduct key name electric toddler disagree abandon impose chest marine three try sense"
JUNO_VAL_MNEMONIC_3="almost east skate high judge that marriage below slush olympic exercise medal utility recall meadow control siren deliver umbrella bid biology input common item"
JUNO_VAL_MNEMONIC_4="language planet neck gold garment day foam bomb roof crystal marble office they hospital party bargain horror disease enforce icon fruit describe sorry universe"
JUNO_VAL_MNEMONIC_5="embrace possible empower remove arrest escape stadium behave bulb bright time drum casual seminar remind science feel absurd isolate beef hidden peace usage sort"
JUNO_VAL_MNEMONICS=("$JUNO_VAL_MNEMONIC_1","$JUNO_VAL_MNEMONIC_2","$JUNO_VAL_MNEMONIC_3","$JUNO_VAL_MNEMONIC_4","$JUNO_VAL_MNEMONIC_5")

# OSMO 
OSMO_CHAIN_ID=OSMO
OSMO_NODE_PREFIX=osmo
OSMO_NUM_NODES=2
OSMO_CMD="$SCRIPT_DIR/../build/osmosisd"
OSMO_VAL_PREFIX=oval
OSMO_REV_ACCT=orev1
OSMO_ADDRESS_PREFIX=osmo
OSMO_DENOM=$OSMO_DENOM
OSMO_IBC_DENOM=$IBC_OSMO_DENOM
OSMO_RPC_PORT=26357
OSMO_MAIN_CMD="$OSMO_CMD --home $SCRIPT_DIR/state/${OSMO_NODE_PREFIX}1"

OSMO_REV_MNEMONIC="furnace spell ring dinosaur paper thank sketch social mystery tissue upgrade voice advice peasant quote surge meat december level broom clock hurdle portion predict"
OSMO_VAL_MNEMONIC_1="hand cheese heavy recall nose toss west finger concert crop rich disorder miss torch photo sport door sausage creek dentist movie course wasp brand"
OSMO_VAL_MNEMONIC_2="tattoo fade gloom boring review actual pluck wrestle desk update mandate grow spawn people blush gym inner voice reform glue shiver screen train august"
OSMO_VAL_MNEMONIC_3="immune acid hurry impose mechanic forward bitter square curtain busy couple hollow calm pole flush deer bird one normal fish loyal upgrade town rail"
OSMO_VAL_MNEMONIC_4="ridge round key spawn address anchor file local athlete pioneer eyebrow flush chase visa awake claim test device chimney roast tent excess profit gaze"
OSMO_VAL_MNEMONIC_5="federal garden bundle rebel museum donor hello oak daring argue talk sing chief burst rigid corn zone gather tell opera nominee desk select shine"
OSMO_VAL_MNEMONICS=("$OSMO_VAL_MNEMONIC_1","$OSMO_VAL_MNEMONIC_2","$OSMO_VAL_MNEMONIC_3","$OSMO_VAL_MNEMONIC_4","$OSMO_VAL_MNEMONIC_5")

# HERMES
HERMES_CMD="$SCRIPT_DIR/../build/hermes/release/hermes --config $STATE/hermes/config.toml"
HERMES_EXEC="docker-compose run --rm hermes hermes"

HERMES_STRIDE_ACCT=rly1
HERMES_GAIA_ACCT=rly2
HERMES_JUNO_ACCT=rly3
HERMES_OSMO_ACCT=rly4

HERMES_STRIDE_MNEMONIC="alter old invest friend relief slot swear pioneer syrup economy vendor tray focus hedgehog artist legend antenna hair almost donkey spice protect sustain increase"
HERMES_GAIA_MNEMONIC="resemble accident lake amateur physical jewel taxi nut demand magnet person blanket trip entire awkward fiber usual current index limb lady lady depart train"
HERMES_JUNO_MNEMONIC="uphold decorate moon memory taste century work pride force genius width ripple myself year steel ivory type sweet tree ignore danger pudding owner discover"
HERMES_OSMO_MNEMONIC="lawn inside color february double myth depart invite miracle nest silver spider spray recall theme loan exotic puzzle uncover dial young earn disagree fee"

# RELAYER
RELAYER_CMD="$SCRIPT_DIR/../build/relayer --home $STATE/relayer"
RELAYER_EXEC="docker-compose run --rm relayer rly"

RELAYER_STRIDE_ACCT=rly1
RELAYER_GAIA_ACCT=rly2
RELAYER_JUNO_ACCT=rly3
RELAYER_OSMO_ACCT=rly4

RELAYER_STRIDE_MNEMONIC="pride narrow breeze fitness sign bounce dose smart squirrel spell length federal replace coral lunar thunder vital push nuclear crouch fun accident hood need"
RELAYER_GAIA_MNEMONIC="fiction perfect rapid steel bundle giant blade grain eagle wing cannon fever must humble dance kitchen lazy episode museum faith off notable rate flavor"
RELAYER_JUNO_MNEMONIC="kiwi betray topple van vapor flag decorate cement crystal fee family clown cry story gain frost strong year blanket remain grass pig hen empower"
RELAYER_OSMO_MNEMONIC="unaware wine ramp february bring trust leaf beyond fever inside option dilemma save know captain endless salute radio humble chicken property culture foil taxi"

# ICQ
ICQ_CMD="$SCRIPT_DIR/../build/interchain-queries --home $STATE/icq"
ICQ_EXEC="docker-compose run --rm icq interchain-queries"

ICQ_STRIDE_ACCT=icq1
ICQ_GAIA_ACCT=icq2
ICQ_JUNO_ACCT=icq3
ICQ_OSMO_ACCT=icq4

ICQ_STRIDE_MNEMONIC="helmet say goat special plug umbrella finger night flip axis resource tuna trigger angry shove essay point laundry horror eager forget depend siren alarm"
ICQ_GAIA_MNEMONIC="capable later bamboo snow drive afraid cheese practice latin brush hand true visa drama mystery bird client nature jealous guess tank marriage volume fantasy"
ICQ_JUNO_MNEMONIC="divorce loop depth announce strategy goddess short cash private raise spatial parent deal acid casual love inner bind ozone picnic fee earn scene galaxy"
ICQ_OSMO_MNEMONIC="mix deal extend cargo office intact illegal cage fabric must upset yellow put any shaft area use piece patrol tobacco village guilt iron program"

DELEGATION_ICA_ADDR='cosmos1sy63lffevueudvvlvh2lf6s387xh9xq72n3fsy6n2gr5hm6u2szs2v0ujm'
REDEMPTION_ICA_ADDR='cosmos1xmcwu75s8v7s54k79390wc5gwtgkeqhvzegpj0nm2tdwacv47tmqg9ut30'
WITHDRAWAL_ICA_ADDR='cosmos1x5p8er7e2ne8l54tx33l560l8djuyapny55pksctuguzdc00dj7saqcw2l'
REVENUE_EOA_ADDR='cosmos1wdplq6qjh2xruc7qqagma9ya665q6qhcwju3ng'
FEE_ICA_ADDR='cosmos1lkgt5sfshn9shm7hd7chtytkq4mvwvswgmyl0hkacd4rmusu9wwq60cezx'
GAIA_DELEGATE_VAL_1='cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne'
GAIA_DELEGATE_VAL_2='cosmosvaloper133lfs9gcpxqj6er3kx605e3v9lqp2pg5syhvsz'
GAIA_RECEIVER_ACCT='cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf'

JUNO_DELEGATE_VAL='junovaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuued3knlr0'
JUNO_DELEGATION_ICA_ADDR='juno1xan7vt4nurz6c7x0lnqnvpmuc0lljz7rycqmuz2kk6wxv4k69d0sfats35'
JUNO_REDEMPTION_ICA_ADDR='juno1y6haxdt03cgkc7aedxrlaleeteel7fgc0nvtu2kggee3hnrlvnvs4kw2v9'
JUNO_WITHDRAWAL_ICA_ADDR='juno104n6h822n6n7psqjgjl7emd2uz67lptggp5cargh6mw0gxpch2gsk53qk5'
JUNO_FEE_ICA_ADDR='juno1rp8qgfq64wmjg7exyhjqrehnvww0t9ev3f3p2ls82umz2fxgylqsz3vl9h'

OSMO_DELEGATE_VAL='osmovaloper12ffkl30v0ghtyaezvedazquhtsf4q5ng8khuv4'
OSMO_DELEGATION_ICA_ADDR='osmo1cx04p5974f8hzh2lqev48kjrjugdxsxy7mzrd0eyweycpr90vk8q8d6f3h'
OSMO_REDEMPTION_ICA_ADDR='osmo1uy9p9g609676rflkjnnelaxatv8e4sd245snze7qsxzlk7dk7s8qrcjaez'
OSMO_WITHDRAWAL_ICA_ADDR='osmo10arcf5r89cdmppntzkvulc7gfmw5lr66y2m25c937t6ccfzk0cqqz2l6xv'
OSMO_FEE_ICA_ADDR='osmo1n4r77qsmu9chvchtmuqy9cv3s539q87r398l6ugf7dd2q5wgyg9su3wd4g'

CSLEEP() {
  for i in $(seq $1); do
    sleep 1
    printf "\r\t$(($1 - $i))s left..."
  done
}

GET_VAR_VALUE() {
  var_name="$1"
  echo "${!var_name}"
}

WAIT_FOR_BLOCK () {
  num_blocks="${2:-1}"
  for i in $(seq $num_blocks); do
    ( tail -f -n0 $1 & ) | grep -q "INF executed block height="
  done
}

WAIT_FOR_NONEMPTY_BLOCK () {
  ( tail -f -n0 $1 & ) | grep -q -E "num_valid_txs=[1-9]"
}

WAIT_FOR_STRING () {
  ( tail -f -n0 $1 & ) | grep -q "$2"
}

WAIT_FOR_IBC_TRANSFER () {
  success_string="packet_cmd{src_chain=(.*)port=transfer(.*): success"
  ( tail -f -n0 $HERMES_LOGS & ) | grep -q -E "$success_string"
  ( tail -f -n0 $HERMES_LOGS & ) | grep -q -E "$success_string"
}

STRIDE_STATE=$SCRIPT_DIR/state/stride
STRIDE_LOGS=$SCRIPT_DIR/logs/stride.log
STRIDE_LOGS_2=$SCRIPT_DIR/logs/stride2.log
STRIDE_LOGS_3=$SCRIPT_DIR/logs/stride3.log
STRIDE_LOGS_4=$SCRIPT_DIR/logs/stride4.log
STRIDE_LOGS_5=$SCRIPT_DIR/logs/stride5.log
GAIA_STATE=$SCRIPT_DIR/state/gaia
GAIA_LOGS=$SCRIPT_DIR/logs/gaia.log
GAIA_LOGS_2=$SCRIPT_DIR/logs/gaia2.log
GAIA_LOGS_3=$SCRIPT_DIR/logs/gaia3.log
HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log
RLY_GAIA_LOGS=$SCRIPT_DIR/logs/rly_gaia.log
RLY_OSMO_LOGS=$SCRIPT_DIR/logs/rly_osmo.log
RLY_JUNO_LOGS=$SCRIPT_DIR/logs/rly_juno.log
ICQ_LOGS=$SCRIPT_DIR/logs/icq.log
JUNO_LOGS=$SCRIPT_DIR/logs/juno.log
OSMO_LOGS=$SCRIPT_DIR/logs/osmo.log
TX_LOGS=$SCRIPT_DIR/logs/tx.log
KEYS_LOGS=$SCRIPT_DIR/logs/keys.log