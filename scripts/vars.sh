#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$SCRIPT_DIR/state
LOGS=$SCRIPT_DIR/logs
PEER_PORT=26656

# Logs
STRIDE_LOGS=$LOGS/stride.log
TX_LOGS=$SCRIPT_DIR/logs/tx.log
KEYS_LOGS=$SCRIPT_DIR/logs/keys.log

# Sets up upgrade if {UPGRADE_NAME} is non-empty
UPGRADE_NAME=""
UPGRADE_OLD_COMMIT_HASH=""

# DENOMS
ATOM_DENOM='uatom'
JUNO_DENOM='ujuno'
OSMO_DENOM='uosmo'
STRD_DENOM='ustrd'
STARS_DENOM='ustars'
STATOM_DENOM="stuatom"
STJUNO_DENOM="stujuno"
STOSMO_DENOM="stuosmo"
STSTARS_DENOM="stustars"

IBC_STRD_DENOM='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'  

IBC_GAIA_CHANNEL_0_DENOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
IBC_GAIA_CHANNEL_1_DENOM='ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9'
IBC_GAIA_CHANNEL_2_DENOM='ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E'
IBC_GAIA_CHANNEL_3_DENOM='ibc/A4DB47A9D3CF9A068D454513891B526702455D3EF08FB9EB558C561F9DC2B701'

IBC_JUNO_CHANNEL_0_DENOM='ibc/04F5F501207C3626A2C14BFEF654D51C2E0B8F7CA578AB8ED272A66FE4E48097' 
IBC_JUNO_CHANNEL_1_DENOM='ibc/EFF323CC632EC4F747C61BCE238A758EFDB7699C3226565F7C20DA06509D59A5' 
IBC_JUNO_CHANNEL_2_DENOM='ibc/4CD525F166D32B0132C095F353F4C6F033B0FF5C49141470D1EFDA1D63303D04'
IBC_JUNO_CHANNEL_3_DENOM='ibc/C814F0B662234E24248AE3B2FE2C1B54BBAF12934B757F6E7BC5AEC119963895' 

IBC_OSMO_CHANNEL_0_DENOM='ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518'
IBC_OSMO_CHANNEL_1_DENOM='ibc/0471F1C4E7AFD3F07702BEF6DC365268D64570F7C1FDC98EA6098DD6DE59817B'
IBC_OSMO_CHANNEL_2_DENOM='ibc/13B2C536BB057AC79D5616B8EA1B9540EC1F2170718CAFF6F0083C966FFFED0B'
IBC_OSMO_CHANNEL_3_DENOM='ibc/47BD209179859CDE4A2806763D7189B6E6FE13A17880FE2B42DE1E6C1E329E23'

IBC_STARS_CHANNEL_0_DENOM='ibc/49BAE4CD2172833F14000627DA87ED8024AD46A38D6ED33F6239F22B5832F958'
IBC_STARS_CHANNEL_1_DENOM='ibc/9222203B0B37D076F07B3CAC716533C80E7C4239499B6306CD9921A15D308F12'
IBC_STARS_CHANNEL_2_DENOM='ibc/C6469BA9DC791E65B3C1596CD2005941324C00659E2DF90D5E08D86B82E7E08B'
IBC_STARS_CHANNEL_3_DENOM='ibc/482A30C07803B0455B1492BAF94EC3D600E862D52A814F25A34BCCAAA132FEE9'

# INTEGRATION TEST IBC DENOM
IBC_ATOM_DENOM=$IBC_GAIA_CHANNEL_0_DENOM
IBC_JUNO_DENOM=$IBC_JUNO_CHANNEL_1_DENOM
IBC_OSMO_DENOM=$IBC_OSMO_CHANNEL_2_DENOM
IBC_STARS_DENOM=$IBC_STARS_CHANNEL_3_DENOM

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
HARSHEN_SLASHING=true

INITIAL_ANNUAL_PROVISIONS="10000000000000.000000000000000000"
VAL_TOKENS=5000000000000
STAKE_TOKENS=5000000000
ADMIN_TOKENS=1000000000

# STRIDE 
STRIDE_CHAIN_ID=STRIDE
STRIDE_NODE_PREFIX=stride
STRIDE_NUM_NODES=3
STRIDE_VAL_PREFIX=val
STRIDE_DENOM=$STRD_DENOM
STRIDE_RPC_PORT=26657
STRIDE_ADMIN_ACCT=admin
STRIDE_ADMIN_ADDRESS=stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8

# Binaries are contigent on whether we're doing an upgrade or not
if [[ "$UPGRADE_NAME" == "" ]]; then 
  STRIDE_CMD="$SCRIPT_DIR/../build/strided"
else
  STRIDE_CMD="$SCRIPT_DIR/upgrades/binaries/strided1"
fi
STRIDE_MAIN_CMD="$STRIDE_CMD --home $SCRIPT_DIR/state/${STRIDE_NODE_PREFIX}1"

STRIDE_MNEMONIC_1="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
STRIDE_MNEMONIC_2="turkey miss hurry unable embark hospital kangaroo nuclear outside term toy fall buffalo book opinion such moral meadow wing olive camp sad metal banner"
STRIDE_MNEMONIC_3="tenant neck ask season exist hill churn rice convince shock modify evidence armor track army street stay light program harvest now settle feed wheat"
STRIDE_MNEMONIC_4="tail forward era width glory magnet knock shiver cup broken turkey upgrade cigar story agent lake transfer misery sustain fragile parrot also air document"
STRIDE_MNEMONIC_5="crime lumber parrot enforce chimney turtle wing iron scissors jealous indicate peace empty game host protect juice submit motor cause second picture nuclear area"
STRIDE_VAL_MNEMONICS=("$STRIDE_MNEMONIC_1","$STRIDE_MNEMONIC_2","$STRIDE_MNEMONIC_3","$STRIDE_MNEMONIC_4","$STRIDE_MNEMONIC_5")

# GAIA 
GAIA_CHAIN_ID=GAIA
GAIA_NODE_PREFIX=gaia
GAIA_NUM_NODES=4
GAIA_CMD="$SCRIPT_DIR/../build/gaiad"
GAIA_VAL_PREFIX=gval
GAIA_REV_ACCT=grev1
GAIA_ADDRESS_PREFIX=cosmos
GAIA_DENOM=$ATOM_DENOM
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
JUNO_NUM_NODES=1
JUNO_CMD="$SCRIPT_DIR/../build/junod"
JUNO_VAL_PREFIX=jval
JUNO_REV_ACCT=jrev1
JUNO_ADDRESS_PREFIX=juno
JUNO_DENOM=$JUNO_DENOM
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
OSMO_NUM_NODES=1
OSMO_CMD="$SCRIPT_DIR/../build/osmosisd"
OSMO_VAL_PREFIX=oval
OSMO_REV_ACCT=orev1
OSMO_ADDRESS_PREFIX=osmo
OSMO_DENOM=$OSMO_DENOM
OSMO_RPC_PORT=26357
OSMO_MAIN_CMD="$OSMO_CMD --home $SCRIPT_DIR/state/${OSMO_NODE_PREFIX}1"

OSMO_REV_MNEMONIC="furnace spell ring dinosaur paper thank sketch social mystery tissue upgrade voice advice peasant quote surge meat december level broom clock hurdle portion predict"
OSMO_VAL_MNEMONIC_1="badge thumb upper scrap gift prosper milk whale journey term indicate risk acquire afford awake margin venture penalty simple fancy fluid review enrich ozone"
OSMO_VAL_MNEMONIC_2="tattoo fade gloom boring review actual pluck wrestle desk update mandate grow spawn people blush gym inner voice reform glue shiver screen train august"
OSMO_VAL_MNEMONIC_3="immune acid hurry impose mechanic forward bitter square curtain busy couple hollow calm pole flush deer bird one normal fish loyal upgrade town rail"
OSMO_VAL_MNEMONIC_4="ridge round key spawn address anchor file local athlete pioneer eyebrow flush chase visa awake claim test device chimney roast tent excess profit gaze"
OSMO_VAL_MNEMONIC_5="federal garden bundle rebel museum donor hello oak daring argue talk sing chief burst rigid corn zone gather tell opera nominee desk select shine"
OSMO_VAL_MNEMONICS=("$OSMO_VAL_MNEMONIC_1","$OSMO_VAL_MNEMONIC_2","$OSMO_VAL_MNEMONIC_3","$OSMO_VAL_MNEMONIC_4","$OSMO_VAL_MNEMONIC_5")

# STARS
STARS_CHAIN_ID=STARS
STARS_NODE_PREFIX=stars
STARS_NUM_NODES=1
STARS_CMD="$SCRIPT_DIR/../build/starsd"
STARS_VAL_PREFIX=sgval
STARS_REV_ACCT=sgrev1
STARS_ADDRESS_PREFIX=stars
STARS_DENOM=$STARS_DENOM
STARS_RPC_PORT=26257
STARS_MAIN_CMD="$STARS_CMD --home $SCRIPT_DIR/state/${STARS_NODE_PREFIX}1"

STARS_REV_MNEMONIC="furnace spell ring dinosaur paper thank sketch social mystery tissue upgrade voice advice peasant quote surge meat december level broom clock hurdle portion predict"
STARS_VAL_MNEMONIC_1="badge thumb upper scrap gift prosper milk whale journey term indicate risk acquire afford awake margin venture penalty simple fancy fluid review enrich ozone"
STARS_VAL_MNEMONIC_2="tattoo fade gloom boring review actual pluck wrestle desk update mandate grow spawn people blush gym inner voice reform glue shiver screen train august"
STARS_VAL_MNEMONIC_3="immune acid hurry impose mechanic forward bitter square curtain busy couple hollow calm pole flush deer bird one normal fish loyal upgrade town rail"
STARS_VAL_MNEMONIC_4="ridge round key spawn address anchor file local athlete pioneer eyebrow flush chase visa awake claim test device chimney roast tent excess profit gaze"
STARS_VAL_MNEMONIC_5="federal garden bundle rebel museum donor hello oak daring argue talk sing chief burst rigid corn zone gather tell opera nominee desk select shine"
STARS_VAL_MNEMONICS=("$STARS_VAL_MNEMONIC_1","$STARS_VAL_MNEMONIC_2","$STARS_VAL_MNEMONIC_3","$STARS_VAL_MNEMONIC_4","$STARS_VAL_MNEMONIC_5")

# HERMES
HERMES_CMD="$SCRIPT_DIR/../build/hermes/release/hermes --config $STATE/hermes/config.toml"
HERMES_EXEC="docker-compose run --rm hermes hermes"

HERMES_STRIDE_ACCT=hrly1
HERMES_GAIA_ACCT=hrly2
HERMES_JUNO_ACCT=hrly3
HERMES_OSMO_ACCT=hrly4
HERMES_STARS_ACCT=hrly5

HERMES_STRIDE_MNEMONIC="alter old invest friend relief slot swear pioneer syrup economy vendor tray focus hedgehog artist legend antenna hair almost donkey spice protect sustain increase"
HERMES_GAIA_MNEMONIC="resemble accident lake amateur physical jewel taxi nut demand magnet person blanket trip entire awkward fiber usual current index limb lady lady depart train"
HERMES_JUNO_MNEMONIC="uphold decorate moon memory taste century work pride force genius width ripple myself year steel ivory type sweet tree ignore danger pudding owner discover"
HERMES_OSMO_MNEMONIC="lawn inside color february double myth depart invite miracle nest silver spider spray recall theme loan exotic puzzle uncover dial young earn disagree fee"
HERMES_STARS_MNEMONIC="inherit shallow bargain explain fence vocal fury perfect jeans figure festival abstract soldier entry bubble ketchup swim useless doctor thing imitate can shock coin"

# RELAYER
RELAYER_CMD="$SCRIPT_DIR/../build/relayer --home $STATE/relayer"
RELAYER_GAIA_EXEC="docker-compose run --rm relayer-gaia"
RELAYER_JUNO_EXEC="docker-compose run --rm relayer-juno"
RELAYER_OSMO_EXEC="docker-compose run --rm relayer-osmo"
RELAYER_STARS_EXEC="docker-compose run --rm relayer-stars"

RELAYER_STRIDE_ACCT=rly1
RELAYER_GAIA_ACCT=rly2
RELAYER_JUNO_ACCT=rly3
RELAYER_OSMO_ACCT=rly4
RELAYER_STARS_ACCT=rly5

RELAYER_STRIDE_MNEMONIC="pride narrow breeze fitness sign bounce dose smart squirrel spell length federal replace coral lunar thunder vital push nuclear crouch fun accident hood need"
RELAYER_GAIA_MNEMONIC="fiction perfect rapid steel bundle giant blade grain eagle wing cannon fever must humble dance kitchen lazy episode museum faith off notable rate flavor"
RELAYER_JUNO_MNEMONIC="kiwi betray topple van vapor flag decorate cement crystal fee family clown cry story gain frost strong year blanket remain grass pig hen empower"
RELAYER_OSMO_MNEMONIC="unaware wine ramp february bring trust leaf beyond fever inside option dilemma save know captain endless salute radio humble chicken property culture foil taxi"
RELAYER_STARS_MNEMONIC="deposit dawn erosion talent old broom flip recipe pill hammer animal hill nice ten target metal gas shoe visual nephew soda harbor child simple"

GAIA_DELEGATE_VAL='cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne'
GAIA_DELEGATION_ICA_ADDR='cosmos1sy63lffevueudvvlvh2lf6s387xh9xq72n3fsy6n2gr5hm6u2szs2v0ujm'
GAIA_REDEMPTION_ICA_ADDR='cosmos1xmcwu75s8v7s54k79390wc5gwtgkeqhvzegpj0nm2tdwacv47tmqg9ut30'
GAIA_WITHDRAWAL_ICA_ADDR='cosmos1x5p8er7e2ne8l54tx33l560l8djuyapny55pksctuguzdc00dj7saqcw2l'
GAIA_FEE_ICA_ADDR='cosmos1lkgt5sfshn9shm7hd7chtytkq4mvwvswgmyl0hkacd4rmusu9wwq60cezx'
GAIA_RECEIVER_ACCT='cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf'

JUNO_DELEGATE_VAL='junovaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuued3knlr0'
JUNO_DELEGATION_ICA_ADDR='juno1xan7vt4nurz6c7x0lnqnvpmuc0lljz7rycqmuz2kk6wxv4k69d0sfats35'
JUNO_REDEMPTION_ICA_ADDR='juno1y6haxdt03cgkc7aedxrlaleeteel7fgc0nvtu2kggee3hnrlvnvs4kw2v9'
JUNO_WITHDRAWAL_ICA_ADDR='juno104n6h822n6n7psqjgjl7emd2uz67lptggp5cargh6mw0gxpch2gsk53qk5'
JUNO_FEE_ICA_ADDR='juno1rp8qgfq64wmjg7exyhjqrehnvww0t9ev3f3p2ls82umz2fxgylqsz3vl9h'
JUNO_RECEIVER_ACCT='juno1sy0q0jpaw4t3hnf6k5wdd4384g0syzlp7rrtsg'

OSMO_DELEGATE_VAL='osmovaloper12ffkl30v0ghtyaezvedazquhtsf4q5ng8khuv4'
OSMO_DELEGATION_ICA_ADDR='osmo1cx04p5974f8hzh2lqev48kjrjugdxsxy7mzrd0eyweycpr90vk8q8d6f3h'
OSMO_REDEMPTION_ICA_ADDR='osmo1uy9p9g609676rflkjnnelaxatv8e4sd245snze7qsxzlk7dk7s8qrcjaez'
OSMO_WITHDRAWAL_ICA_ADDR='osmo10arcf5r89cdmppntzkvulc7gfmw5lr66y2m25c937t6ccfzk0cqqz2l6xv'
OSMO_FEE_ICA_ADDR='osmo1n4r77qsmu9chvchtmuqy9cv3s539q87r398l6ugf7dd2q5wgyg9su3wd4g'
OSMO_RECEIVER_ACCT='osmo1w6wdc2684g9h3xl8nhgwr282tcxx4kl06n4sjl'

STARS_DELEGATE_VAL='starsvaloper12ffkl30v0ghtyaezvedazquhtsf4q5ng2c0xaf'
STARS_DELEGATION_ICA_ADDR="stars1kl6wa99e6hf97xr90m2n04rl0smv842pj9utqyvgyrksrm9aacdqyfc3en"
STARS_REDEMPTION_ICA_ADDR="stars1x07hv0hxujj6l0mfyynwyuccf8fl27vjup0y8dmyajy9ugae22hqfvmv4e"
STARS_WITHDRAWAL_ICA_ADDR="stars1x5ndl5p9tjy376a9xmqhw79gz0s678480759cdgaretcgm36akvs0a78tj"
STARS_FEE_ICA_ADDR="stars1v09y993sku5djvm0rffq0nfsk5rzke4d2vzvny5e6vmq7dz0dehqnwl4ay"
STARS_RECEIVER_ACCT='stars15dywcmy6gzsc8wfefkrx0c9czlwvwrjenqthyq'

STRIDE_ADDRESS() { 
  $STRIDE_MAIN_CMD keys show ${STRIDE_VAL_PREFIX}1 --keyring-backend test -a 
}
GAIA_ADDRESS() { 
  $GAIA_MAIN_CMD keys show ${GAIA_VAL_PREFIX}1 --keyring-backend test -a 
}
JUNO_ADDRESS() { 
  $JUNO_MAIN_CMD keys show ${JUNO_VAL_PREFIX}1 --keyring-backend test -a 
}
OSMO_ADDRESS() { 
  $OSMO_MAIN_CMD keys show ${OSMO_VAL_PREFIX}1 --keyring-backend test -a 
}
STARS_ADDRESS() { 
  $STARS_MAIN_CMD keys show ${STARS_VAL_PREFIX}1 --keyring-backend test -a 
}

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

WAIT_FOR_BLOCK() {
  num_blocks="${2:-1}"
  for i in $(seq $num_blocks); do
    ( tail -f -n0 $1 & ) | grep -q "INF executed block height="
  done
}

WAIT_FOR_STRING() {
  ( tail -f -n0 $1 & ) | grep -q "$2"
}

GET_VAL_ADDR() {
  chain_id=$1
  val_index=$2

  MAIN_CMD=$(GET_VAR_VALUE ${chain_id}_MAIN_CMD)
  $MAIN_CMD q staking validators | grep ${chain_id}_${val_index} -A 5 | grep operator | awk '{print $2}'
}

GET_ICA_ADDR() {
  chain_id="$1"
  ica_type="$2" #delegation, fee, redemption, or withdrawal

  $STRIDE_MAIN_CMD q stakeibc show-host-zone $chain_id | grep ${ica_type}Account -A 1 | grep address | awk '{print $2}'
}

TRIM_TX() {
  grep -E "code:|txhash:" | sed 's/^/  /'
}