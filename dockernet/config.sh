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
# HOST_CHAINS have liquid staking support, ACCESSORY_CHAINS do not
HOST_CHAINS=()
ACCESSORY_CHAINS=() 

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
#  - DYDX
#  - NOBLE (only runs as an accessory chain - does not have liquid staking functionality)
if [[ "${ALL_HOST_CHAINS:-false}" == "true" ]]; then 
  HOST_CHAINS=(GAIA EVMOS HOST)
elif [[ "${#HOST_CHAINS[@]}" == "0" ]]; then 
  HOST_CHAINS=(GAIA)
fi
REWARD_CONVERTER_HOST_ZONE=${HOST_CHAINS[0]}

# DENOMS
STRD_DENOM="ustrd"
ATOM_DENOM="uatom"
JUNO_DENOM="ujuno"
OSMO_DENOM="uosmo"
STARS_DENOM="ustars"
WALK_DENOM="uwalk"
EVMOS_DENOM="aevmos"
DYDX_DENOM="udydx"
NOBLE_DENOM="utoken"
USDC_DENOM="uusdc"
STATOM_DENOM="stuatom"
STJUNO_DENOM="stujuno"
STOSMO_DENOM="stuosmo"
STSTARS_DENOM="stustars"
STWALK_DENOM="stuwalk"
STEVMOS_DENOM="staevmos"
STDYDX_DENOM="studydx"

IBC_GAIA_CHANNEL_0_STATOM_DENOM='ibc/054A44EC8D9B68B9A6F0D5708375E00A5569A28F21E0064FF12CADC3FEF1D04F'
IBC_GAIA_CHANNEL_1_STATOM_DENOM='ibc/8B21DA0E34A49AE151FEEBCCF3AFE1188E24BA8E19439FB93434DF6008E7E228'
IBC_GAIA_CHANNEL_2_STATOM_DENOM='ibc/60CB7A5465C318C8F68F603D78721A2ECC1DA2D0E905C6AD9ACD1CAC3F0DB22D'
IBC_GAIA_CHANNEL_3_STATOM_DENOM='ibc/0C0FD07C29EB075C18EA77B73CF9FCE68A268E0738C9F5B11D13E418AD889437'

IBC_GAIA_STATOM_DENOM=$IBC_GAIA_CHANNEL_0_STATOM_DENOM

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

IBC_DYDX_CHANNEL_0_DENOM='ibc/815D14313C85CADBDFCEB13C8028DB853BE16CF6600D6B3A90ECFB7DCF1FAAF9'
IBC_DYDX_CHANNEL_1_DENOM='ibc/78B7A771A2ECBF5D10DC6AB35568A7AC4161DB21B3A848DA470655358A6DD854'
IBC_DYDX_CHANNEL_2_DENOM='ibc/748465E0D883217048DB25F4C3825D03F682A06FE292E21072BF678E249DAC18'
IBC_DYDX_CHANNEL_3_DENOM='ibc/6301148031C0AC9A392C2DDB1B2D1F11B3B9D0A3ECF20C6B5122685D9E4CC631'

IBC_GAIA_STDENOM='ibc/054A44EC8D9B68B9A6F0D5708375E00A5569A28F21E0064FF12CADC3FEF1D04F'
IBC_HOST_STDENOM='ibc/E3AF56419340E719710C088D3855F65C4717E1A0C3B405F0C1D16F2A54E89421'
IBC_EVMOS_STDENOM='ibc/04CDA5EBB8A7E94BB60879B7F43EF0EDD2604990D8AB5BA18ADCB173F66FF874'

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
STRIDE_HOUR_EPOCH_DURATION="60s"
STRIDE_DAY_EPOCH_DURATION="140s"
STRIDE_EPOCH_EPOCH_DURATION="35s"
STRIDE_MINT_EPOCH_DURATION="20s"
HOST_DAY_EPOCH_DURATION="60s"
HOST_HOUR_EPOCH_DURATION="60s"
HOST_WEEK_EPOCH_DURATION="60s"
HOST_MINT_EPOCH_DURATION="60s"
UNBONDING_TIME="240s"
MAX_DEPOSIT_PERIOD="30s"
VOTING_PERIOD="30s"
INITIAL_ANNUAL_PROVISIONS="10000000000000.000000000000000000"

# LSM Params
LSM_VALIDATOR_BOND_FACTOR="250"
LSM_GLOBAL_LIQUID_STAKING_CAP="0.25"
LSM_VALIDATOR_LIQUID_STAKING_CAP="0.50"

# Tokens are denominated in the macro-unit 
# (e.g. 5000000STRD implies 5000000000000ustrd)
VAL_TOKENS=5000000
STAKE_TOKENS=5000
ADMIN_TOKENS=1000
USER_TOKENS=100

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
USER_MNEMONIC="brief play describe burden half aim soccer carbon hope wait output play vacuum joke energy crucial output mimic cruise brother document rail anger leaf"
USER_ACCT=user

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
if [[ "${UPGRADE_NAME:-}" == "" ]]; then 
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

# DYDX
DYDX_CHAIN_ID=DYDX
DYDX_NODE_PREFIX=dydx
DYDX_NUM_NODES=1
DYDX_BINARY="$DOCKERNET_HOME/../build/dydxprotocold"
DYDX_VAL_PREFIX=val
DYDX_ADDRESS_PREFIX=dydx
DYDX_REV_ACCT=rev
DYDX_DENOM=$DYDX_DENOM
DYDX_RPC_PORT=25957
DYDX_MAIN_CMD="$DYDX_BINARY --home $DOCKERNET_HOME/state/${DYDX_NODE_PREFIX}1"
DYDX_RECEIVER_ADDRESS='dydx1q9caajs6wrfu2yhytvkqd2csxycx6revdcme9y'
# The micro denom is actually the same as default cosmos chains but there's a 
# minimum stake amount so this effectively gets the validator over the minimum
DYDX_MICRO_DENOM_UNITS="000000000000000" 

# NOBLE
NOBLE_CHAIN_ID=NOBLE
NOBLE_NODE_PREFIX=noble
NOBLE_NUM_NODES=1
NOBLE_BINARY="$DOCKERNET_HOME/../build/nobled"
NOBLE_VAL_PREFIX=val
NOBLE_ADDRESS_PREFIX=noble
NOBLE_REV_ACCT=rev
NOBLE_DENOM=$NOBLE_DENOM
NOBLE_RPC_PORT=25857
NOBLE_MAIN_CMD="$NOBLE_BINARY --home $DOCKERNET_HOME/state/${NOBLE_NODE_PREFIX}1"
NOBLE_RECEIVER_ADDRESS='noble1dd9sxkz3wr723lsf65h549ykdh4npxzh5qawmg'
NOBLE_AUTHORITHY_MNEMONIC="giant screen unit high agree swing impact switch lend universe sand myself conduct sustain august barely misery lawsuit honey social version window demise palace"

# RELAYER
RELAYER_GAIA_EXEC="$DOCKER_COMPOSE run --rm relayer-gaia"
RELAYER_GAIA_ICS_EXEC="$DOCKER_COMPOSE run --rm relayer-gaia-ics"
RELAYER_JUNO_EXEC="$DOCKER_COMPOSE run --rm relayer-juno"
RELAYER_OSMO_EXEC="$DOCKER_COMPOSE run --rm relayer-osmo"
RELAYER_STARS_EXEC="$DOCKER_COMPOSE run --rm relayer-stars"
RELAYER_HOST_EXEC="$DOCKER_COMPOSE run --rm relayer-host"
RELAYER_EVMOS_EXEC="$DOCKER_COMPOSE run --rm relayer-evmos"
RELAYER_DYDX_EXEC="$DOCKER_COMPOSE run --rm relayer-dydx"
RELAYER_NOBLE_EXEC="$DOCKER_COMPOSE run --rm relayer-noble"

# Accounts for relay paths with stride
RELAYER_STRIDE_ACCT=rly1
RELAYER_GAIA_ACCT=rly2
RELAYER_JUNO_ACCT=rly3
RELAYER_OSMO_ACCT=rly4
RELAYER_STARS_ACCT=rly5
RELAYER_HOST_ACCT=rly6
RELAYER_EVMOS_ACCT=rly7
RELAYER_STRIDE_ICS_ACCT=rly8
RELAYER_GAIA_ICS_ACCT=rly9
RELAYER_DYDX_ACCT=rly10
STRIDE_RELAYER_ACCTS=(
  $RELAYER_GAIA_ACCT 
  $RELAYER_JUNO_ACCT 
  $RELAYER_OSMO_ACCT 
  $RELAYER_STARS_ACCT 
  $RELAYER_HOST_ACCT 
  $RELAYER_EVMOS_ACCT
  $RELAYER_GAIA_ICS_ACCT
  $RELAYER_DYDX_ACCT
)

# Mnemonics for connections with stride
RELAYER_GAIA_MNEMONIC="fiction perfect rapid steel bundle giant blade grain eagle wing cannon fever must humble dance kitchen lazy episode museum faith off notable rate flavor"
RELAYER_JUNO_MNEMONIC="kiwi betray topple van vapor flag decorate cement crystal fee family clown cry story gain frost strong year blanket remain grass pig hen empower"
RELAYER_OSMO_MNEMONIC="unaware wine ramp february bring trust leaf beyond fever inside option dilemma save know captain endless salute radio humble chicken property culture foil taxi"
RELAYER_STARS_MNEMONIC="deposit dawn erosion talent old broom flip recipe pill hammer animal hill nice ten target metal gas shoe visual nephew soda harbor child simple"
RELAYER_HOST_MNEMONIC="renew umbrella teach spoon have razor knee sock divert inner nut between immense library inhale dog truly return run remain dune virus diamond clinic"
RELAYER_GAIA_ICS_MNEMONIC="size chimney clog job robot thunder gaze vapor economy smooth kit denial alter merit produce front force eager outside mansion believe fan tonight detect"
RELAYER_EVMOS_MNEMONIC="science depart where tell bus ski laptop follow child bronze rebel recall brief plug razor ship degree labor human series today embody fury harvest"
RELAYER_DYDX_MNEMONIC="mother depth nature rapid draw west afraid depend allow fee siren useful catalog sun biology cabbage busy science front smile nurse balcony medal burst"
STRIDE_RELAYER_MNEMONICS=(
  "$RELAYER_GAIA_MNEMONIC"
  "$RELAYER_JUNO_MNEMONIC"
  "$RELAYER_OSMO_MNEMONIC"
  "$RELAYER_STARS_MNEMONIC"
  "$RELAYER_HOST_MNEMONIC"
  "$RELAYER_EVMOS_MNEMONIC"
  "$RELAYER_GAIA_ICS_MNEMONIC"
  "$RELAYER_DYDX_MNEMONIC"
)
# Mnemonics for connections between accessory chains
RELAYER_STRIDE_OSMO_MNEMONIC="father october lonely ticket leave regret pudding buffalo return asthma plastic piano beef orient ill clip right phone ready pottery helmet hip solid galaxy"
RELAYER_OSMO_STRIDE_MNEMONIC="narrow assist come feel canyon anxiety three reason satoshi inspire region little attend impulse what student dog armor economy faculty dutch distance upon calm"
RELAYER_STRIDE_NOBLE_MNEMONIC="absent confirm lumber hobby glide alter remain yard mixed fiscal series kitchen effort protect pistol hire bless police year struggle near hour wisdom jewel"
RELAYER_NOBLE_STRIDE_MNEMONIC="jar point equal question fatigue frog disorder wasp labor obtain head print orbit entire frown high sadness dash retire idea coffee rubber rough until"
RELAYER_NOBLE_OSMO_MNEMONIC="actual field visual wage orbit add human unit happy rich evil chair entire person february cactus deputy impact gasp elbow sunset brand possible fly"
RELAYER_OSMO_NOBLE_MNEMONIC="obey clinic miss grunt inflict laugh sell moral kitchen tumble gold song flavor rather horn exhaust state amazing poverty differ approve spike village device"
# Mnemonics between host zone and accessory chains when running with GAIA as the host
RELAYER_GAIA_NOBLE_MNEMONIC="aerobic breeze claw climb bounce morning tank victory eight funny employ bracket hire reduce fine flee lava domain warfare loop theme fly tattoo must"
RELAYER_NOBLE_GAIA_MNEMONIC="sentence fruit crumble sail bar knife exact flame apart prosper hint myth clean among tiny burden depart purity select envelope identify cross physical emerge"
RELAYER_GAIA_OSMO_MNEMONIC="small fire step promote fox reward book seek arctic session illegal loyal because brass spoil minute wonder jazz shoe price muffin churn evil monitor"
RELAYER_OSMO_GAIA_MNEMONIC="risk wool reason sweet current strategy female miracle squeeze that wire develop ocean rapid domain lift blame monkey sick round museum item maze trumpet"
# Mnemonics between host zone and accessory chains when running with DYDX as the host
RELAYER_NOBLE_DYDX_MNEMONIC="sentence fruit crumble sail bar knife exact flame apart prosper hint myth clean among tiny burden depart purity select envelope identify cross physical emerge"
RELAYER_DYDX_NOBLE_MNEMONIC="aerobic breeze claw climb bounce morning tank victory eight funny employ bracket hire reduce fine flee lava domain warfare loop theme fly tattoo must"
RELAYER_DYDX_OSMO_MNEMONIC="small fire step promote fox reward book seek arctic session illegal loyal because brass spoil minute wonder jazz shoe price muffin churn evil monitor"
RELAYER_OSMO_DYDX_MNEMONIC="risk wool reason sweet current strategy female miracle squeeze that wire develop ocean rapid domain lift blame monkey sick round museum item maze trumpet"

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
DYDX_ADDRESS() { 
  $DYDX_MAIN_CMD keys show ${DYDX_VAL_PREFIX}1 --keyring-backend test -a 
}
NOBLE_ADDRESS() { 
  $NOBLE_MAIN_CMD keys show ${NOBLE_VAL_PREFIX}1 --keyring-backend test -a 
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

SAVE_DOCKER_LOGS() {
  service_name=$1
  log_path=$2
  $DOCKER_COMPOSE logs -f $service_name | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $log_path 2>&1 &
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

# Helper function to ensure there's enough time left in the epoch for operations to complete
# This will check how much time is remaining in the epoch, and if there's enough time,
# it will do nothing, otherwise it will sleep until the next epoch begins
# Ex: if you need at least 30 seconds in the day epoch to complete the test,
#     you can run `AVOID_EPOCH_BOUNDARY day 30``
AVOID_EPOCH_BOUNDARY() {
  epoch_type="$1"
  buffer_required="$2"

  seconds_remaining_in_epoch=$($STRIDE_MAIN_CMD q epochs seconds-remaining $epoch_type)

  # If there's enough time left, no need to sleep
  if [[ $seconds_remaining_in_epoch -gt $buffer_required ]]; then
    return
  fi

  # Otherwise, wait for the next epoch
  sleep $((seconds_remaining_in_epoch+5))
}

# Sleep until the balance has changed
# Optionally provide a minimum amount it must change by (to ignore interest)
WAIT_FOR_BALANCE_CHANGE() {
  chain=$1
  address=$2
  denom=$3
  minimum_change=${4:-1} # defaults to 1

  max_blocks=30

  main_cmd=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
  initial_balance=$($main_cmd q bank balances $address --denom $denom | grep amount | NUMBERS_ONLY)
  for i in $(seq $max_blocks); do
    new_balance=$($main_cmd q bank balances $address --denom $denom | grep amount | NUMBERS_ONLY)
    balance_change=$(echo "$new_balance - $initial_balance" | bc)

    if [[ $(echo "$balance_change >= $minimum_change" | bc -l) == "1" ]]; then
      break
    fi

    WAIT_FOR_BLOCK $STRIDE_LOGS 1
  done
}

# Sleep until the total delegation amount has changed
# Optionally provide a minimum amount it must change by  (to ignore interest)
WAIT_FOR_DELEGATION_CHANGE() {
  chain_id=$1
  minimum_change=${2:-1} # defaults to 1

  max_blocks=30

  initial_delegation=$($STRIDE_MAIN_CMD q stakeibc show-host-zone $chain_id | grep "total_delegations" | NUMBERS_ONLY)
  for i in $(seq $max_blocks); do
    new_delegation=$($STRIDE_MAIN_CMD q stakeibc show-host-zone $chain_id | grep "total_delegations" | NUMBERS_ONLY)
    delegation_change=$(echo "$new_delegation - $initial_delegation" | bc)

    if [[ $(echo "$delegation_change >= $minimum_change" | bc -l) == "1" ]]; then
      break
    fi

    WAIT_FOR_BLOCK $STRIDE_LOGS 1
  done
}

GET_VAL_ADDR() {
  chain=$1
  val_index=$2

  MAIN_CMD=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
  $MAIN_CMD q staking validators | grep ${chain}_${val_index} -A 6 | grep operator | awk '{print $2}'
}

GET_ICA_ADDR() {
  chain_id="$1"
  ica_type="$2" #delegation, fee, redemption, or withdrawal

  $STRIDE_MAIN_CMD q stakeibc show-host-zone $chain_id | grep ${ica_type}_ica_address | awk '{print $2}'
}

GET_HOST_ZONE_FIELD() {
  chain_id="$1"
  field="$2"

  $STRIDE_MAIN_CMD q stakeibc show-host-zone $chain_id | grep $field | awk '{print $2}'
}

GET_IBC_DENOM() {
  chain="$1"
  transfer_channel_id="$2"
  base_denom="$3"

  main_cmd=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
  echo "ibc/$($main_cmd q ibc-transfer denom-hash transfer/${transfer_channel_id}/${base_denom} | awk '{print $2}')"
}

GET_CLIENT_ID_FROM_CHAIN_ID() {
  src_chain="$1"
  counterparty_chain_id="$2"

  main_cmd=$(GET_VAR_VALUE ${src_chain}_MAIN_CMD)
  $main_cmd q ibc client states | grep $counterparty_chain_id -B 6 | grep client_id | awk '{print $3}'
}

GET_CONNECTION_ID_FROM_CLIENT_ID() {
  src_chain="$1"
  client_id="$2"

  main_cmd=$(GET_VAR_VALUE ${src_chain}_MAIN_CMD)
  $main_cmd q ibc connection path $client_id | grep connection- | awk '{print $2}'
}

GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID() {
  src_chain="$1"
  connection_id="$2"

  main_cmd=$(GET_VAR_VALUE ${src_chain}_MAIN_CMD)
  $main_cmd q ibc channel connections $connection_id | grep -m 1 "channel_id" | awk '{print $3}'
}

GET_COUNTERPARTY_TRANSFER_CHANNEL_ID() {
  src_chain="$1"
  channel_id="$2"

  main_cmd=$(GET_VAR_VALUE ${src_chain}_MAIN_CMD)
  $main_cmd q ibc channel end transfer $channel_id | grep -A 2 counterparty | grep channel_id | awk '{print $2}'
}

GET_LATEST_PROPOSAL_ID() {
  chain="$1"

  main_cmd=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
  $main_cmd q gov proposals | grep '  id:' | tail -1 | awk '{printf $2}' | tr -d '"'
}

WATCH_PROPOSAL_STATUS() {
  chain="$1"
  proposal_id="$2"

  main_cmd=$(GET_VAR_VALUE ${chain}_MAIN_CMD)

  # Continually polls the proposal status until it passes or fails
  while true; do
    status=$($main_cmd query gov proposal $proposal_id | grep "status" | awk '{printf $2}')
    if [[ "$status" == "PROPOSAL_STATUS_VOTING_PERIOD" ]]; then
        echo "  Proposal still in progress..."
        sleep 5
    elif [[ "$status" == "PROPOSAL_STATUS_PASSED" ]]; then
        echo "  Proposal passed!"
        exit 0
    elif [[ "$status" == "PROPOSAL_STATUS_REJECTED" ]]; then
        echo "  Proposal rejected!"
        exit 1
    elif [[ "$status" == "PROPOSAL_STATUS_FAILED" ]]; then
        echo "  Proposal failed!"
        exit 1
    else 
        echo "ERROR: Unknown proposal status: $status"
        exit 1
    fi
  done
}

TRIM_TX() {
  grep -E "code:|txhash:" | sed 's/^/  /'
}

NUMBERS_ONLY() {
  tr -cd '[:digit:]'
}

GETBAL() {
  head -n 1 | grep -o -E '[0-9]+' || echo "0"
}

GETSTAKE() {
  tail -n 2 | head -n 1 | grep -o -E '[0-9]+' | head -n 1
}
