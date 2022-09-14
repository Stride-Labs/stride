#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$SCRIPT_DIR/state

# DENOMS
IBC_STRD_DENOM='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'
IBC_STRD_DENOM_JUNO='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'
IBC_STRD_DENOM_OSMO='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'
IBC_ATOM_DENOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
IBC_JUNO_DENOM='ibc/4CD525F166D32B0132C095F353F4C6F033B0FF5C49141470D1EFDA1D63303D04'
IBC_OSMO_DENOM='ibc/0471F1C4E7AFD3F07702BEF6DC365268D64570F7C1FDC98EA6098DD6DE59817B'
ATOM_DENOM='uatom'
JUNO_DENOM='ujuno'
OSMO_DENOM='uosmo'
STRD_DENOM='ustrd'
STATOM_DENOM="stuatom"
STJUNO_DENOM="stujuno"
STOSMO_DENOM="stuosmo"

# **************************************************************************
# WARNING: CHANGES TO THESE PARAMS COULD BREAK INTEGRATION TESTS
# **************************************************************************
# CHAIN PARAMS
BLOCK_TIME_SECONDS=1
BLOCK_TIME="${BLOCK_TIME_SECONDS}s"
# NOTE: If you add new epochs, these indexes will need to be updated
DAY_EPOCH_INDEX=1
INTERVAL_LEN=1
DAY_EPOCH_LEN="100s"
STRIDE_EPOCH_INDEX=2
STRIDE_EPOCH_LEN="40s"
MINT_EPOCH_INDEX=3
MINT_EPOCH_LEN="60s"
IBC_TX_WAIT_SECONDS=30
MAX_DEPOSIT_PERIOD="30s"
VOTING_PERIOD="30s"
UNBONDING_TIME="120s"

# define STRIDE vars
STRIDE_PORT_ID=26657  # 36564
STRIDE_CHAIN=STRIDE
STRIDE_NODE_NAME=stride
STRIDE_VAL_ACCT=val1
STRIDE_VAL_MNEMONIC="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
STRIDE_VAL_ADDR="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"
STRIDE_HOME="$STATE/stride"
STRIDE_ADMIN_ADDRESS="stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"

STRIDE_CMD="$SCRIPT_DIR/../build/strided --home $STRIDE_HOME"

# define vars for STRIDE 2
STRIDE_VAL_ACCT_2=val2
STRIDE_PORT_ID_2=26257
STRIDE_PEER_PORT_2=26256
STRIDE_EXT_ADR_2=26255
STRIDE_VAL_MNEMONIC_2="turkey miss hurry unable embark hospital kangaroo nuclear outside term toy fall buffalo book opinion such moral meadow wing olive camp sad metal banner"
STRIDE_VAL_2_ADDR="stride17kht2x2ped6qytr2kklevtvmxpw7wq9rmuc3ca"
STRIDE_VAL_2_PUBKEY='{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A8E8fiJfecuL6r67CEMk7lwXgAkFcVjI0hwimy7N1pFl"}'
STRIDE_HOME_2="$STATE/stride2"
STRIDE_CMD_2="$SCRIPT_DIR/../build/stride2/strided --home $STRIDE_HOME_2"

# define vars for STRIDE 3
STRIDE_VAL_ACCT_3=val3
STRIDE_PORT_ID_3=26157
STRIDE_PEER_PORT_3=26156
STRIDE_EXT_ADR_3=26155
STRIDE_VAL_MNEMONIC_3="tenant neck ask season exist hill churn rice convince shock modify evidence armor track army street stay light program harvest now settle feed wheat"
STRIDE_VAL_3_ADDR="stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv"
STRIDE_VAL_3_PUBKEY='{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A0GIT0Ftl4rGIjx+KPwzUIlMqm/4ZMUv6IgsmdwXbt3f"}'
STRIDE_HOME_3="$STATE/stride3"
STRIDE_CMD_3="$SCRIPT_DIR/../build/stride3/strided --home $STRIDE_HOME_3"

# define vars for STRIDE 4
STRIDE_VAL_ACCT_4=val4
STRIDE_PORT_ID_4=26057
STRIDE_PEER_PORT_4=26056
STRIDE_EXT_ADR_4=26055
STRIDE_VAL_MNEMONIC_4="tail forward era width glory magnet knock shiver cup broken turkey upgrade cigar story agent lake transfer misery sustain fragile parrot also air document"
STRIDE_VAL_4_ADDR="stride1py0fvhdtq4au3d9l88rec6vyda3e0wtt9szext"
STRIDE_VAL_4_PUBKEY='{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A0v7/E7N4O5OLD1IG8KqwDkm9i961JDUrYyoN3l2OgEg"}'
STRIDE_HOME_4="$STATE/stride4"
STRIDE_CMD_4="$SCRIPT_DIR/../build/stride4/strided --home $STRIDE_HOME_4"

# define vars for STRIDE 5
STRIDE_VAL_ACCT_5=val5
STRIDE_PORT_ID_5=25957
STRIDE_PEER_PORT_5=25956
STRIDE_EXT_ADR_5=25955
STRIDE_VAL_MNEMONIC_5="crime lumber parrot enforce chimney turtle wing iron scissors jealous indicate peace empty game host protect juice submit motor cause second picture nuclear area"
STRIDE_VAL_5_ADDR="stride1c5jnf370kaxnv009yhc3jt27f549l5u36chzem"
STRIDE_VAL_5_PUBKEY='{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A3cNnP5IXZ2llK10AyU4xa/T+YorvBXZBTT4aVxFKTCo"}'
STRIDE_HOME_5="$STATE/stride5"
STRIDE_CMD_5="$SCRIPT_DIR/../build/stride5/strided --home $STRIDE_HOME_5"

STRIDE_CMD="$SCRIPT_DIR/../build/strided --home $STATE/stride"

# define GAIA vars
GAIA_CHAIN=GAIA
GAIA_PEER_PORT=26556
GAIA_NODE_NAME=gaia
GAIA_VAL_ACCT=gval1
GAIA_REV_ACCT=grev1
GAIA_VAL_MNEMONIC="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
GAIA_REV_MNEMONIC="tonight bonus finish chaos orchard plastic view nurse salad regret pause awake link bacon process core talent whale million hope luggage sauce card weasel"
GAIA_VAL_ADDR="cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2"
GAIA_HOME="$STATE/gaia"
GAIA_CMD="$SCRIPT_DIR/../build/gaiad --home $GAIA_HOME"

GAIA_VAL_ACCT_2=gval2
GAIA_PORT_ID_2=26457
GAIA_PEER_PORT_2=26456
GAIA_EXT_ADR_2=26455
GAIA_VAL_MNEMONIC_2="guilt leader matrix lecture infant axis method grain diesel sting reflect brave estate surge october candy busy crash parade club practice sure gentle celery"
GAIA_VAL_2_ADDR="cosmos133lfs9gcpxqj6er3kx605e3v9lqp2pg54sreu3"
GAIA_VAL_2_PUBKEY='{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A2yf54F9IxosnE3MdJ+rYP96AG5mFr60UtjorA8TMq8h"}'
GAIA_HOME_2="$STATE/gaia2"
GAIA_CMD_2="$SCRIPT_DIR/../build/gaia2/gaiad --home $GAIA_HOME_2"
GAIA_VAL_ACCT_3=gval3
GAIA_PORT_ID_3=26357
GAIA_PEER_PORT_3=26356
GAIA_EXT_ADR_2=26355
GAIA_VAL_MNEMONIC_3="fire tape spread wing click winter awful ozone visa spray swear color table settle review rival meadow gauge speed tide timber disease float live"
GAIA_VAL_3_ADDR="cosmos1fumal3j4lxzjp22fzffge8mw56qm33h9ez0xy2"
GAIA_VAL_3_PUBKEY='{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A7X2X1v+pa0kIuxgfhZoPNhrVUZ5DFcYERZy4UanN1dc"}'
GAIA_HOME_3="$STATE/gaia3"
GAIA_CMD_3="$SCRIPT_DIR/../build/gaia3/gaiad --home $GAIA_HOME_3"

HERMES_CMD="$SCRIPT_DIR/../build/hermes/release/hermes --config $SCRIPT_DIR/hermes/config.toml"

# define relayer vars
HERMES_STRIDE_ACCT=rly1
HERMES_GAIA_ACCT=rly2
HERMES_JUNO_ACCT=rly3
HERMES_OSMO_ACCT=rly4
HERMES_STRIDE_MNEMONIC="alter old invest friend relief slot swear pioneer syrup economy vendor tray focus hedgehog artist legend antenna hair almost donkey spice protect sustain increase"
HERMES_STRIDE_ADDR="stride1ft20pydau82pgesyl9huhhux307s9h3078692y"
HERMES_GAIA_MNEMONIC="resemble accident lake amateur physical jewel taxi nut demand magnet person blanket trip entire awkward fiber usual current index limb lady lady depart train"
HERMES_JUNO_MNEMONIC="uphold decorate moon memory taste century work pride force genius width ripple myself year steel ivory type sweet tree ignore danger pudding owner discover"
HERMES_OSMO_MNEMONIC="lawn inside color february double myth depart invite miracle nest silver spider spray recall theme loan exotic puzzle uncover dial young earn disagree fee"
HERMES_OSMO_ADDRESS="osmo1lajwg95utv75fny0w39806xuk92ky57csvj6f5"
RLY_GAIA_MNEMONIC="fiction perfect rapid steel bundle giant blade grain eagle wing cannon fever must humble dance kitchen lazy episode museum faith off notable rate flavor"
RLY_GAIA_ADDR="cosmos16lmf7t0jhaatan6vnxlgv47h2wf0k5lnhvye5h"
RLY_GAIA_ACCT=gaiarly
RLY_STRIDE_MNEMONIC="pride narrow breeze fitness sign bounce dose smart squirrel spell length federal replace coral lunar thunder vital push nuclear crouch fun accident hood need"
RLY_STRIDE_ADDR="stride1z56v8wqvgmhm3hmnffapxujvd4w4rkw6mdrmjg"
RLY_STRIDE_ACCT=striderly
RLY_OSMO_MNEMONIC="unaware wine ramp february bring trust leaf beyond fever inside option dilemma save know captain endless salute radio humble chicken property culture foil taxi"
RLY_OSMO_ADDR="osmo1zwj4yr264fr9au20gur3qapt3suwkgp0w039jd"
RLY_OSMO_ACCT=osmorly
RLY_JUNO_MNEMONIC="kiwi betray topple van vapor flag decorate cement crystal fee family clown cry story gain frost strong year blanket remain grass pig hen empower"
RLY_JUNO_ADDR="juno1jrmtt5c6z8h5yrrwml488qnm7p3vxrrm2n40l0"
RLY_JUNO_ACCT=junorly

ICQ_CMD="$SCRIPT_DIR/../build/interchain-queries --home $STATE/icq"

ICQ_STRIDE_ACCT=icq1
ICQ_GAIA_ACCT=icq2
ICQ_JUNO_ACCT=icq3
ICQ_OSMO_ACCT=icq4
ICQ_STRIDE_MNEMONIC="helmet say goat special plug umbrella finger night flip axis resource tuna trigger angry shove essay point laundry horror eager forget depend siren alarm"
ICQ_STRIDE_ADDR="stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"
ICQ_GAIA_MNEMONIC="capable later bamboo snow drive afraid cheese practice latin brush hand true visa drama mystery bird client nature jealous guess tank marriage volume fantasy"
ICQ_JUNO_MNEMONIC="divorce loop depth announce strategy goddess short cash private raise spatial parent deal acid casual love inner bind ozone picnic fee earn scene galaxy"
ICQ_OSMO_MNEMONIC="mix deal extend cargo office intact illegal cage fabric must upset yellow put any shaft area use piece patrol tobacco village guilt iron program"

DELEGATION_ICA_ADDR='cosmos1sy63lffevueudvvlvh2lf6s387xh9xq72n3fsy6n2gr5hm6u2szs2v0ujm'
REDEMPTION_ICA_ADDR='cosmos1xmcwu75s8v7s54k79390wc5gwtgkeqhvzegpj0nm2tdwacv47tmqg9ut30'
WITHDRAWAL_ICA_ADDR='cosmos1x5p8er7e2ne8l54tx33l560l8djuyapny55pksctuguzdc00dj7saqcw2l'
REVENUE_EOA_ADDR='cosmos1wdplq6qjh2xruc7qqagma9ya665q6qhcwju3ng'
FEE_ICA_ADDR='cosmos1lkgt5sfshn9shm7hd7chtytkq4mvwvswgmyl0hkacd4rmusu9wwq60cezx'
GAIA_DELEGATE_VAL='cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne'
GAIA_DELEGATE_VAL_2='cosmosvaloper133lfs9gcpxqj6er3kx605e3v9lqp2pg5syhvsz'
GAIA_RECEIVER_ACCT='cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf'

JUNO_DELEGATION_ICA_ADDR='juno1xan7vt4nurz6c7x0lnqnvpmuc0lljz7rycqmuz2kk6wxv4k69d0sfats35'
JUNO_REDEMPTION_ICA_ADDR='juno1y6haxdt03cgkc7aedxrlaleeteel7fgc0nvtu2kggee3hnrlvnvs4kw2v9'
JUNO_WITHDRAWAL_ICA_ADDR='juno104n6h822n6n7psqjgjl7emd2uz67lptggp5cargh6mw0gxpch2gsk53qk5'
JUNO_FEE_ICA_ADDR='juno1rp8qgfq64wmjg7exyhjqrehnvww0t9ev3f3p2ls82umz2fxgylqsz3vl9h'

OSMO_DELEGATION_ICA_ADDR='osmo1cx04p5974f8hzh2lqev48kjrjugdxsxy7mzrd0eyweycpr90vk8q8d6f3h'
OSMO_REDEMPTION_ICA_ADDR='osmo1uy9p9g609676rflkjnnelaxatv8e4sd245snze7qsxzlk7dk7s8qrcjaez'
OSMO_WITHDRAWAL_ICA_ADDR='osmo10arcf5r89cdmppntzkvulc7gfmw5lr66y2m25c937t6ccfzk0cqqz2l6xv'
OSMO_FEE_ICA_ADDR='osmo1n4r77qsmu9chvchtmuqy9cv3s539q87r398l6ugf7dd2q5wgyg9su3wd4g'

CSLEEP() {
  for i in $(seq $1); do
    sleep 1
    printf "\r\t$(($1 - $i))s left..."
  done
  printf "\n"
}

BLOCK_SLEEP() {
  for i in $(seq $1); do
    sleep $BLOCK_TIME_SECONDS
    printf "\r\t$(($1 - $i)) blocks left..."
  done
  printf "\n"
}

# define JUNO vars
JUNO_CHAIN=JUNO
JUNO_PEER_PORT=24656
JUNO_NODE_NAME=JUNO
JUNO_VAL_ACCT=jval1
JUNO_REV_ACCT=jrev1
JUNO_VAL_MNEMONIC="move next relief spatial resemble onion exhibit fitness major toss where square wrong exact infant skate dragon shift region over you gospel absorb double"
JUNO_REV_MNEMONIC="tonight bonus finish chaos orchard plastic view nurse salad regret pause awake link bacon process core talent whale million hope luggage sauce card weasel"
JUNO_VAL_ADDR="juno1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2"
JUNO_HOME="$STATE/juno"
JUNO_CMD="$SCRIPT_DIR/../build/junod --home $JUNO_HOME"
JUNO_DELEGATE_VAL='junovaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuued3knlr0'
JUNO_RECEIVER_ACCT='juno1sy0q0jpaw4t3hnf6k5wdd4384g0syzlp7rrtsg'

# define OSMO vars
OSMO_CHAIN=OSMO
OSMO_PEER_PORT=23656
OSMO_NODE_NAME=OSMO
OSMO_VAL_ACCT=oval1
OSMO_REV_ACCT=orev1
OSMO_VAL_MNEMONIC="badge thumb upper scrap gift prosper milk whale journey term indicate risk acquire afford awake margin venture penalty simple fancy fluid review enrich ozone"
OSMO_REV_MNEMONIC="furnace spell ring dinosaur paper thank sketch social mystery tissue upgrade voice advice peasant quote surge meat december level broom clock hurdle portion predict"
OSMO_VAL_ADDR="osmo12ffkl30v0ghtyaezvedazquhtsf4q5ngapllmj"
OSMO_HOME="$STATE/osmo"
OSMO_CMD="$SCRIPT_DIR/../build/osmosisd --home $OSMO_HOME"
OSMO_DELEGATE_VAL='osmovaloper12ffkl30v0ghtyaezvedazquhtsf4q5ng8khuv4'
OSMO_RECEIVER_ACCT='osmo1w6wdc2684g9h3xl8nhgwr282tcxx4kl06n4sjl'

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
RLY_GAIA_LOGS=$SCRIPT_DIR/logs/rly/rly_gaia.log
RLY_OSMO_LOGS=$SCRIPT_DIR/logs/rly/rly_osmo.log
RLY_JUNO_LOGS=$SCRIPT_DIR/logs/rly/rly_juno.log
ICQ_LOGS=$SCRIPT_DIR/logs/rly/icq.log
JUNO_LOGS=$SCRIPT_DIR/logs/juno.log
OSMO_LOGS=$SCRIPT_DIR/logs/osmo.log
TX_LOGS=$SCRIPT_DIR/logs/tx.log
KEYS_LOGS=$SCRIPT_DIR/logs/keys.log
