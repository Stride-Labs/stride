#!/bin/bash

set -eu
DOCKERNET_HOME=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$DOCKERNET_HOME/state
LOGS=$DOCKERNET_HOME/logs
UPGRADES=$DOCKERNET_HOME/upgrades
SRC=$DOCKERNET_HOME/src
PEER_PORT=26656
DOCKER_COMPOSE="docker-compose -f $DOCKERNET_HOME/docker-compose.yml"

# Logs
STRIDE_LOGS=$LOGS/stride.log
TX_LOGS=$DOCKERNET_HOME/logs/tx.log
KEYS_LOGS=$DOCKERNET_HOME/logs/keys.log

# List of hosts enabled
HOST_CHAINS=() 

# If no host zones are specified above:
#  `start-docker` defaults to just GAIA if HOST_CHAINS is empty
#  `start-docker-all` always runs all hosts
# Available host zones:
#  - GAIA
#  - JUNO
#  - OSMO
#  - STARS
#  - EVMOS
#  - HOST (Stride chain enabled as a host zone)
if [[ "${ALL_HOST_CHAINS:-false}" == "true" ]]; then 
  HOST_CHAINS=(GAIA EVMOS HOST)
elif [[ "${#HOST_CHAINS[@]}" == "0" ]]; then 
  HOST_CHAINS=(GAIA)
fi

# Sets up upgrade if {UPGRADE_NAME} is non-empty
UPGRADE_NAME=""
UPGRADE_OLD_COMMIT_HASH=""

# DENOMS
STRD_DENOM="ustrd"
ATOM_DENOM="uatom"
JUNO_DENOM="ujuno"
OSMO_DENOM="uosmo"
STARS_DENOM="ustars"
WALK_DENOM="uwalk"
EVMOS_DENOM="aevmos"
STATOM_DENOM="stuatom"
STJUNO_DENOM="stujuno"
STOSMO_DENOM="stuosmo"
STSTARS_DENOM="stustars"
STWALK_DENOM="stuwalk"
STEVMOS_DENOM="staevmos"

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

IBC_EVMOS_CHANNEL_0_DENOM='ibc/8EAC8061F4499F03D2D1419A3E73D346289AE9DB89CAB1486B72539572B1915E'
IBC_EVMOS_CHANNEL_1_DENOM='ibc/6993F2B27985C9363D3B94D702111940055833A2BA86DA93F33A67D03E4D1B7D'
IBC_EVMOS_CHANNEL_2_DENOM='ibc/0E8BF52B5A990E16C4AF2E5ED426503F3F0B12067FB2B4B660015A64CCE38EA0'
IBC_EVMOS_CHANNEL_3_DENOM='ibc/5590FF5DA750B007818BB275A9CDC8B6704414F8411E2EF8CC6C43A913B6CE88'

IBC_HOST_CHANNEL_0_DENOM='ibc/82DBA832457B89E1A344DA51761D92305F7581B7EA6C18D85037910988953C58'
IBC_HOST_CHANNEL_1_DENOM='ibc/FB7E2520A1ED6890E1632904A4ACA1B3D2883388F8E2B88F2D6A54AA15E4B49E'
IBC_HOST_CHANNEL_2_DENOM='ibc/D664DC1D38648FC4C697D9E9CF2D26369318DFE668B31F81809383A8A88CFCF4'
IBC_HOST_CHANNEL_3_DENOM='ibc/FD7AA7EB2C1D5D97A8693CCD71FFE3F5AFF12DB6756066E11E69873DE91A33EA'

# COIN TYPES
# Coin types can be found at https://github.com/satoshilabs/slips/blob/master/slip-0044.md
COSMOS_COIN_TYPE=118
ETH_COIN_TYPE=60
TERRA_COIN_TYPE=330

# INTEGRATION TEST IBC DENOM
IBC_ATOM_DENOM=$IBC_GAIA_CHANNEL_0_DENOM
IBC_JUNO_DENOM=$IBC_JUNO_CHANNEL_1_DENOM
IBC_OSMO_DENOM=$IBC_OSMO_CHANNEL_2_DENOM
IBC_STARS_DENOM=$IBC_STARS_CHANNEL_3_DENOM

# CHAIN PARAMS
BLOCK_TIME='1s'
STRIDE_HOUR_EPOCH_DURATION="90s"
STRIDE_DAY_EPOCH_DURATION="100s"
STRIDE_EPOCH_EPOCH_DURATION="40s"
STRIDE_MINT_EPOCH_DURATION="20s"
HOST_DAY_EPOCH_DURATION="60s"
HOST_HOUR_EPOCH_DURATION="60s"
HOST_WEEK_EPOCH_DURATION="60s"
HOST_MINT_EPOCH_DURATION="60s"
UNBONDING_TIME="120s"
MAX_DEPOSIT_PERIOD="30s"
VOTING_PERIOD="30s"
INITIAL_ANNUAL_PROVISIONS="10000000000000.000000000000000000"

# Tokens are denominated in the macro-unit 
# (e.g. 5000000STRD implies 5000000000000ustrd)
VAL_TOKENS=5000000
STAKE_TOKENS=5000
ADMIN_TOKENS=1000

# CHAIN MNEMONICS
VAL_MNEMONIC_1="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
VAL_MNEMONIC_2="turkey miss hurry unable embark hospital kangaroo nuclear outside term toy fall buffalo book opinion such moral meadow wing olive camp sad metal banner"
VAL_MNEMONIC_3="tenant neck ask season exist hill churn rice convince shock modify evidence armor track army street stay light program harvest now settle feed wheat"
VAL_MNEMONIC_4="tail forward era width glory magnet knock shiver cup broken turkey upgrade cigar story agent lake transfer misery sustain fragile parrot also air document"
VAL_MNEMONIC_5="crime lumber parrot enforce chimney turtle wing iron scissors jealous indicate peace empty game host protect juice submit motor cause second picture nuclear area"
VAL_MNEMONICS=(
    "$VAL_MNEMONIC_1"
    "$VAL_MNEMONIC_2"
    "$VAL_MNEMONIC_3"
    "$VAL_MNEMONIC_4"
    "$VAL_MNEMONIC_5"
)
REV_MNEMONIC="tonight bonus finish chaos orchard plastic view nurse salad regret pause awake link bacon process core talent whale million hope luggage sauce card weasel"

# STRIDE 
STRIDE_CHAIN_ID=STRIDE
STRIDE_NODE_PREFIX=stride
STRIDE_NUM_NODES=3
STRIDE_VAL_PREFIX=val
STRIDE_ADDRESS_PREFIX=stride
STRIDE_DENOM=$STRD_DENOM
STRIDE_RPC_PORT=26657
STRIDE_ADMIN_ACCT=admin
STRIDE_ADMIN_ADDRESS=stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
STRIDE_ADMIN_MNEMONIC="tone cause tribe this switch near host damage idle fragile antique tail soda alien depth write wool they rapid unfold body scan pledge soft"
STRIDE_FEE_ADDRESS=stride1czvrk3jkvtj8m27kqsqu2yrkhw3h3ykwj3rxh6

# Binaries are contigent on whether we're doing an upgrade or not
if [[ "$UPGRADE_NAME" == "" ]]; then 
  STRIDE_BINARY="$DOCKERNET_HOME/../build/strided"
else
  if [[ "${NEW_BINARY:-false}" == "false" ]]; then
    STRIDE_BINARY="$UPGRADES/binaries/strided1"
  else
    STRIDE_BINARY="$UPGRADES/binaries/strided2"
  fi
fi
STRIDE_MAIN_CMD="$STRIDE_BINARY --home $DOCKERNET_HOME/state/${STRIDE_NODE_PREFIX}1"

# GAIA 
GAIA_CHAIN_ID=GAIA
GAIA_NODE_PREFIX=gaia
GAIA_NUM_NODES=1
GAIA_BINARY="$DOCKERNET_HOME/../build/gaiad"
GAIA_VAL_PREFIX=gval
GAIA_REV_ACCT=grev1
GAIA_ADDRESS_PREFIX=cosmos
GAIA_DENOM=$ATOM_DENOM
GAIA_RPC_PORT=26557
GAIA_MAIN_CMD="$GAIA_BINARY --home $DOCKERNET_HOME/state/${GAIA_NODE_PREFIX}1"
GAIA_RECEIVER_ADDRESS='cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf'

# JUNO 
JUNO_CHAIN_ID=JUNO
JUNO_NODE_PREFIX=juno
JUNO_NUM_NODES=1
JUNO_BINARY="$DOCKERNET_HOME/../build/junod"
JUNO_VAL_PREFIX=jval
JUNO_REV_ACCT=jrev1
JUNO_ADDRESS_PREFIX=juno
JUNO_DENOM=$JUNO_DENOM
JUNO_RPC_PORT=26457
JUNO_MAIN_CMD="$JUNO_BINARY --home $DOCKERNET_HOME/state/${JUNO_NODE_PREFIX}1"
JUNO_RECEIVER_ADDRESS='juno1sy0q0jpaw4t3hnf6k5wdd4384g0syzlp7rrtsg'

# OSMO 
OSMO_CHAIN_ID=OSMO
OSMO_NODE_PREFIX=osmo
OSMO_NUM_NODES=1
OSMO_BINARY="$DOCKERNET_HOME/../build/osmosisd"
OSMO_VAL_PREFIX=oval
OSMO_REV_ACCT=orev1
OSMO_ADDRESS_PREFIX=osmo
OSMO_DENOM=$OSMO_DENOM
OSMO_RPC_PORT=26357
OSMO_MAIN_CMD="$OSMO_BINARY --home $DOCKERNET_HOME/state/${OSMO_NODE_PREFIX}1"
OSMO_RECEIVER_ADDRESS='osmo1w6wdc2684g9h3xl8nhgwr282tcxx4kl06n4sjl'

# STARS
STARS_CHAIN_ID=STARS
STARS_NODE_PREFIX=stars
STARS_NUM_NODES=1
STARS_BINARY="$DOCKERNET_HOME/../build/starsd"
STARS_VAL_PREFIX=sgval
STARS_REV_ACCT=sgrev1
STARS_ADDRESS_PREFIX=stars
STARS_DENOM=$STARS_DENOM
STARS_RPC_PORT=26257
STARS_MAIN_CMD="$STARS_BINARY --home $DOCKERNET_HOME/state/${STARS_NODE_PREFIX}1"
STARS_RECEIVER_ADDRESS='stars15dywcmy6gzsc8wfefkrx0c9czlwvwrjenqthyq'

# HOST (Stride running as a host zone)
HOST_CHAIN_ID=HOST
HOST_NODE_PREFIX=host
HOST_NUM_NODES=1
HOST_BINARY="$DOCKERNET_HOME/../build/strided"
HOST_VAL_PREFIX=hval
HOST_ADDRESS_PREFIX=stride
HOST_REV_ACCT=hrev1
HOST_DENOM=$WALK_DENOM
HOST_RPC_PORT=26157
HOST_MAIN_CMD="$HOST_BINARY --home $DOCKERNET_HOME/state/${HOST_NODE_PREFIX}1"
HOST_RECEIVER_ADDRESS='stride1trm75t8g83f26u4y8jfds7pms9l587a7q227k9'

# EVMOS
EVMOS_CHAIN_ID=evmos_9001-2
EVMOS_NODE_PREFIX=evmos
EVMOS_NUM_NODES=1
EVMOS_BINARY="$DOCKERNET_HOME/../build/evmosd"
EVMOS_VAL_PREFIX=eval
EVMOS_ADDRESS_PREFIX=evmos
EVMOS_REV_ACCT=erev1
EVMOS_DENOM=$EVMOS_DENOM
EVMOS_RPC_PORT=26057
EVMOS_MAIN_CMD="$EVMOS_BINARY --home $DOCKERNET_HOME/state/${EVMOS_NODE_PREFIX}1"
EVMOS_RECEIVER_ADDRESS='evmos123z469cfejeusvk87ufrs5520wmdxmmlc7qzuw'
EVMOS_MICRO_DENOM_UNITS="000000000000000000000000"

# RELAYER
RELAYER_CMD="$DOCKERNET_HOME/../build/relayer --home $STATE/relayer"
RELAYER_GAIA_EXEC="$DOCKER_COMPOSE run --rm relayer-gaia"
RELAYER_JUNO_EXEC="$DOCKER_COMPOSE run --rm relayer-juno"
RELAYER_OSMO_EXEC="$DOCKER_COMPOSE run --rm relayer-osmo"
RELAYER_STARS_EXEC="$DOCKER_COMPOSE run --rm relayer-stars"
RELAYER_EVMOS_EXEC="$DOCKER_COMPOSE run --rm relayer-evmos"
RELAYER_HOST_EXEC="$DOCKER_COMPOSE run --rm relayer-host"

RELAYER_STRIDE_ACCT=rly1
RELAYER_GAIA_ACCT=rly2
RELAYER_JUNO_ACCT=rly3
RELAYER_OSMO_ACCT=rly4
RELAYER_STARS_ACCT=rly5
RELAYER_HOST_ACCT=rly6
RELAYER_EVMOS_ACCT=rly7
RELAYER_ACCTS=(
  $RELAYER_GAIA_ACCT 
  $RELAYER_JUNO_ACCT 
  $RELAYER_OSMO_ACCT 
  $RELAYER_STARS_ACCT 
  $RELAYER_HOST_ACCT 
  $RELAYER_EVMOS_ACCT
)

RELAYER_GAIA_MNEMONIC="fiction perfect rapid steel bundle giant blade grain eagle wing cannon fever must humble dance kitchen lazy episode museum faith off notable rate flavor"
RELAYER_JUNO_MNEMONIC="kiwi betray topple van vapor flag decorate cement crystal fee family clown cry story gain frost strong year blanket remain grass pig hen empower"
RELAYER_OSMO_MNEMONIC="unaware wine ramp february bring trust leaf beyond fever inside option dilemma save know captain endless salute radio humble chicken property culture foil taxi"
RELAYER_STARS_MNEMONIC="deposit dawn erosion talent old broom flip recipe pill hammer animal hill nice ten target metal gas shoe visual nephew soda harbor child simple"
RELAYER_HOST_MNEMONIC="renew umbrella teach spoon have razor knee sock divert inner nut between immense library inhale dog truly return run remain dune virus diamond clinic"
RELAYER_EVMOS_MNEMONIC="science depart where tell bus ski laptop follow child bronze rebel recall brief plug razor ship degree labor human series today embody fury harvest"
RELAYER_MNEMONICS=(
  "$RELAYER_GAIA_MNEMONIC"
  "$RELAYER_JUNO_MNEMONIC"
  "$RELAYER_OSMO_MNEMONIC"
  "$RELAYER_STARS_MNEMONIC"
  "$RELAYER_HOST_MNEMONIC"
  "$RELAYER_EVMOS_MNEMONIC"
)

STRIDE_ADDRESS() { 
  # After an upgrade, the keys query can sometimes print migration info, 
  # so we need to filter by valid addresses using the prefix
  $STRIDE_MAIN_CMD keys show ${STRIDE_VAL_PREFIX}1 --keyring-backend test -a | grep $STRIDE_ADDRESS_PREFIX
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
HOST_ADDRESS() { 
  $HOST_MAIN_CMD keys show ${HOST_VAL_PREFIX}1 --keyring-backend test -a 
}
EVMOS_ADDRESS() { 
  $EVMOS_MAIN_CMD keys show ${EVMOS_VAL_PREFIX}1 --keyring-backend test -a 
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
    ( tail -f -n0 $1 & ) | grep -q "executed block.*height="
  done
}

WAIT_FOR_STRING() {
  ( tail -f -n0 $1 & ) | grep -q "$2"
}

WAIT_FOR_BALANCE_CHANGE() {
  chain=$1
  address=$2
  denom=$3

  max_blocks=30

  main_cmd=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
  initial_balance=$($main_cmd q bank balances $address --denom $denom | grep amount)
  for i in $(seq $max_blocks); do
    new_balance=$($main_cmd q bank balances $address --denom $denom | grep amount)

    if [[ "$new_balance" != "$initial_balance" ]]; then
      break
    fi

    WAIT_FOR_BLOCK $STRIDE_LOGS 1
  done
}

GET_VAL_ADDR() {
  chain=$1
  val_index=$2

  MAIN_CMD=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
  $MAIN_CMD q staking validators | grep ${chain}_${val_index} -A 5 | grep operator | awk '{print $2}'
}

GET_ICA_ADDR() {
  chain_id="$1"
  ica_type="$2" #delegation, fee, redemption, or withdrawal

  $STRIDE_MAIN_CMD q stakeibc show-host-zone $chain_id | grep ${ica_type}_account -A 1 | grep address | awk '{print $2}'
}

TRIM_TX() {
  grep -E "code:|txhash:" | sed 's/^/  /'
}