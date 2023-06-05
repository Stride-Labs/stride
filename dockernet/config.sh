#!/bin/bash

set -eu
DOCKERNET_HOME=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$DOCKERNET_HOME/state
LOGS=$DOCKERNET_HOME/logs
UPGRADES=$DOCKERNET_HOME/upgrades
SRC=$DOCKERNET_HOME/src
PEER_PORT=26656
DOCKER_COMPOSE="docker compose -f $DOCKERNET_HOME/docker-compose.yml"

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
# UPGRADE_NAME=""
# UPGRADE_OLD_COMMIT_HASH=""

UPGRADE_NAME="v07-Theta"
UPGRADE_OLD_COMMIT_HASH="3d757b019da20889ef44f7fae65ecab1b5aa1d69"

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
UNBONDING_TIME="600s"
MAX_DEPOSIT_PERIOD="30s"
VOTING_PERIOD="30s"
INITIAL_ANNUAL_PROVISIONS="10000000000000.000000000000000000"

# Tokens are denominated in the macro-unit 
# (e.g. 5000000STRD implies 5000000000000ustrd)
VAL_TOKENS=5000000
STAKE_TOKENS=5000
ADMIN_TOKENS=1000

# CHAIN MNEMONICS
VAL_MNEMONIC_1="birth supply wait payment history venture fame usual still orange champion bless scale time option match asset skin apology release add legend glimpse gym"
VAL_MNEMONIC_2="step shrimp beyond pistol caught girl become dismiss unfair bottom cement excuse bulb below venue mouse panda assist cloth diet defense absent afford account"
VAL_MNEMONIC_3="dose liar thunder exit neck reveal trade future soup city hurry mechanic fantasy used dentist spray provide license gather nice lemon crawl crack harsh"
VAL_MNEMONIC_4="lecture radar energy hard exchange grunt cover math group raise palm immune truth spin faculty leave cup mutual can wealth focus achieve strategy dice"
VAL_MNEMONIC_5="toy image verify insect acquire attract pottery mercy govern clock lock worth spin agree tonight tattoo help process rich they maple emerge social pelican"
VAL_MNEMONIC_6="dismiss chief ski write bus stairs medal file maximum alter nominee forum talent fee tooth typical flush mechanic farm run awesome face strategy local"
VAL_MNEMONIC_7="blue they sausage fall wife shy enact aim pizza march energy knife push close squirrel leaf balcony trim actor clarify island arrange despair planet"
VAL_MNEMONIC_8="quit certain act disease gossip enact base pill frame chronic reduce predict elephant urge toward office country move rug coil indicate apple where win"
VAL_MNEMONIC_9="fine shield staff imitate mushroom code effort huge pull trade remind creek guard height forget grab dice slender always news intact wheel know casual"
VAL_MNEMONIC_10="address endorse ensure diet slam spike donate chest cable cupboard trade two kitchen suffer robust art boat stairs liquid scrub tail yellow will refuse"
VAL_MNEMONIC_11="unhappy light gadget drink confirm resemble furnace blue drill build such upgrade tennis before zone immense wine hidden vapor expand lunar mansion wage wonder"
VAL_MNEMONIC_12="tower decade van wild shield tent scene shuffle inform hub sport paper raise pattern acid next life must paper nuclear escape tobacco color crane"
VAL_MNEMONIC_13="thought matter keep liberty grocery hurry solar super pulp bacon base program cost scatter question group inflict scissors jazz quiz mention nominee gravity parade"
VAL_MNEMONIC_14="lunch evoke busy convince siren enforce erupt please major present stove nest orchard hurt lumber upon sail parrot lesson lake steak aware hospital peanut"
VAL_MNEMONIC_15="fatal bracket mansion album vanish tower coin range cross humor series road prize coffee move result copy champion concert sorry inherit antenna enforce better"
VAL_MNEMONIC_16="coyote dune elite main regular dinner truly lucky current fall zebra robot caution render visual fuel pass amazing main ivory journey upon box vivid"
VAL_MNEMONIC_17="only session lend whisper best cover clock castle hurry orange income honey impact pledge output panel disease topic eager affair icon flee floor glimpse"
VAL_MNEMONIC_18="apart wool father fortune keep craft pulse series kidney bless anchor fatal rare bicycle thought draft clown mixed glass topic ensure reveal catalog dumb"
VAL_MNEMONIC_19="cycle document plunge fatal soon switch tree moment grief theme output trip cloud weapon harvest tape banner perfect inject city enemy foil honey just"
VAL_MNEMONIC_20="despair ladder banner icon outer debate exhibit announce total hood donor spy sail minor cream enhance cute grain guitar rally tunnel body post smart"
VAL_MNEMONIC_21="exotic brother horse brass age lunch crime aunt sweet pretty hotel orient license ordinary rely wine lazy feature bamboo use into defy bounce steak"
VAL_MNEMONIC_22="fringe tired labor clever gather fork bitter clarify salmon alcohol inhale seat brother below brain enable message bomb female chaos slush hint answer desert"
VAL_MNEMONIC_23="corn mistake egg region random isolate view wife name member cabin output mind service tree column common mom resist lawn stable engine ski typical"
VAL_MNEMONIC_24="fog orange problem machine hand diagram mind cloth rose need climb color gown damage business cup inquiry style pepper urge hello edit rifle disease"
VAL_MNEMONIC_25="horse opinion exclude insect arrow guilt sound food poverty ritual orient pen bench first elegant gorilla detail impulse usage foam wheel energy juice hole"
VAL_MNEMONIC_26="impact whisper bring address damp spend accuse moment hover deer release depart hockey mammal discover category town unveil please faint capable hold cause obtain"
VAL_MNEMONIC_27="collect since slot crumble mutual argue jazz property green month obscure knee involve barely pull earth safe fragile project ignore unlock stadium hammer heavy"
VAL_MNEMONIC_28="genre develop estate miracle hope blood word neutral potato blast wing guard purchase decide globe describe route advance eye replace supply speak idle sweet"
VAL_MNEMONIC_29="danger scorpion giggle rose treat olympic cart spice bitter other copper meadow company salad ticket also offer ostrich cook rebuild energy people glue wagon"
VAL_MNEMONIC_30="nuclear athlete jeans depend maid switch limit twice orphan veteran pupil item ankle hire van have high young symbol pool inherit gospel unfold empower"
VAL_MNEMONIC_31="vivid rather nuclear major reflect silk device health session airport burst isolate mosquito weekend section various spy bulk grant photo behave capital foam fitness"
VAL_MNEMONIC_32="piano palm top truth imitate heart destroy void trip moment copper load know trust harsh chronic symptom sleep bracket sense divert onion cabbage high"
VAL_MNEMONIC_33="ranch empty shell kiss floor habit toss attitude inmate stool glad ticket scheme pass future clever push neutral tray rebel cattle swamp expose census"
VAL_MNEMONIC_34="erosion diary ask trash magic begin decorate cheap gloom dilemma category project alien off model bomb uniform clay banner demise display edge slab evil"
VAL_MNEMONIC_35="other net syrup walnut wonder dilemma horn join olive cargo cricket shed napkin fresh swamp grow voyage picture one invest convince wife garage salute"
VAL_MNEMONIC_36="coral cycle gloom title swallow advance belt monitor duck apology velvet million trip tape veteran extend uncle whale edit festival snack alert make mountain"
VAL_MNEMONIC_37="resource hockey portion empty stool inquiry embrace dawn corn betray wing tiger organ number piece impose shove hold fee number old turtle zero fiber"
VAL_MNEMONIC_38="begin april pact weekend eyebrow fun couple urban anxiety remind lunar engage endorse pride congress artwork warm cube float fame wonder marriage float pause"
VAL_MNEMONIC_39="come puzzle round forest unable inhale valley assist axis dash reunion ignore bomb ocean stereo annual weird coin cram pave collect fine polar cabbage"
VAL_MNEMONIC_40="cattle unique destroy holiday used lecture ocean title coyote wolf task wisdom forget sleep fog sort door expand behind pen bonus feel iron grant"
VAL_MNEMONIC_41="web resist also you album recycle replace early seek situate lucky biology index parent parent true attend hair layer allow talk health busy enjoy"
VAL_MNEMONIC_42="general reunion guard exercise cruise coral soldier cotton accident say spray canal provide surge input assist music best trigger side pumpkin rail document there"
VAL_MNEMONIC_43="destroy pupil design disorder office worry vessel soap april truth stuff put quote regular tomorrow scare clip relief coach eagle wrist wise flat such"
VAL_MNEMONIC_44="black scout open twenty gadget man account spawn ceiling where globe rate exclude wing garage tourist response team tackle tongue below science atom sting"
VAL_MNEMONIC_45="rude twice oval wise hotel relax logic dutch joy noodle energy bone crop happy maid focus frost consider video drastic music annual alone question"
VAL_MNEMONIC_46="tank exit pupil lock legal roast chat story crawl street buddy pepper acquire tilt genre provide radio emotion label bottom aspect fiscal dose hockey"
VAL_MNEMONIC_47="camp oxygen draw wasp method top special thank economy shadow gold trap debate enhance goose arch float assist ship cement employ you great battle"
VAL_MNEMONIC_48="benefit this negative refuse negative flip wet card rocket what comic plate palm noise butter market inflict lottery dial pulse rich armed merry employ"
VAL_MNEMONIC_49="clump milk supreme long grit silk opinion oyster path craft what divert hole tip amount taste dentist tooth blur small depth skate bonus mouse"
VAL_MNEMONIC_50="guilt shrug tonight ethics brass acquire crazy caught history mule exile buddy secret initial chaos venture recall aerobic doctor write lunar fit gauge leopard"
VAL_MNEMONIC_51="start equal merit warrior exit also degree green attitude disorder prosper address mansion sketch gift baby warfare tourist fire cousin artwork note plug gas"
VAL_MNEMONIC_52="safe weather rebel help build pistol nasty document melt plate people flight sustain click own knock enough secret gun mountain carbon fatigue wrong broken"
VAL_MNEMONIC_53="valve bleak couple forum tape portion vacuum scout victory another pole steel muscle figure good disease before lawsuit sudden forest apple abstract rubber gift"
VAL_MNEMONIC_54="diesel slush question group steel drama wife educate sorry rice climb panda diagram resemble pattern master play safe skin select neglect wage captain carbon"
VAL_MNEMONIC_55="crouch again this purity derive recipe bike cannon wolf endorse endless refuse join arrive merge boy slight aware judge attract language cross scorpion surge"
VAL_MNEMONIC_56="sauce glance grocery razor thunder robust frame repeat dust drastic example cargo appear return process number produce great found off trash tissue leave dumb"
VAL_MNEMONIC_57="blossom stage earn dune later visa urge season derive cash shield bicycle bench uncover life entry clip autumn car robot arena path attitude glance"
VAL_MNEMONIC_58="slim derive parade where pipe receive vague cry off matter vendor smile hand soft frequent ring again multiply rotate bracket stand purpose nominee bird"
VAL_MNEMONIC_59="nature piece repair canoe planet bridge wild street cactus lake female pill tonight fatal cost census believe hour execute dolphin normal stove floor invite"
VAL_MNEMONIC_60="misery denial news hope rigid trumpet secret upon broccoli denial champion tilt crater extend tourist defense nuclear diesel pluck orbit trend slight ripple rich"
VAL_MNEMONIC_61="mention weapon since wrong visual mansion federal fuel blast resemble sunny echo enforce ecology season receive illegal nuclear usage apart seat dilemma place absurd"
VAL_MNEMONIC_62="market seminar direct era surface doll source entry abuse ugly august jar valley acoustic joke involve tomato about usual pluck glance never moral edit"
VAL_MNEMONIC_63="fatigue layer twin ghost vehicle flower betray clever stadium unveil near panic sock rival whisper liar apology walnut wool stairs idle grow mountain input"
VAL_MNEMONIC_64="sick among deal apology loyal gesture clutch become hundred route sauce wrist wool coil police era execute when okay sudden sorry velvet erode axis"
VAL_MNEMONIC_65="debris exotic car pass please absent sauce must tunnel salon heart learn stamp move saddle weird advance cruise awesome multiply net ranch drift neck"
VAL_MNEMONIC_66="volume judge chronic bundle survey early physical fire legal draft arctic garage harsh exhaust hockey size awkward panic rifle pioneer uncle skirt ice vendor"
VAL_MNEMONIC_67="party room duty spoil rebel resemble obvious merry recipe pink remove fame dress noodle beyond fish twin chat receive squirrel steak digital expire garage"
VAL_MNEMONIC_68="select jungle pulp zone garlic soldier quality food ostrich utility stereo orphan parrot purchase primary fitness car awkward candy picnic doll card body sight"
VAL_MNEMONIC_69="target clinic trophy once clock pumpkin afford film blouse vapor office muscle there label script dawn notable team safe foam foster plunge steak kingdom"
VAL_MNEMONIC_70="spatial victory pattern example assume next boss during light review notice rather thing task aunt spot taste nasty help receive conduct wash level like"
VAL_MNEMONIC_71="cram pistol gain increase kick toss ocean diary pepper pull roast denial image glue board horn strike random palace secret latin salad because turn"
VAL_MNEMONIC_72="erase original soon actor little canvas dish resemble gesture crack cinnamon state puppy pill almost lawsuit forum call inch error screen better news smoke"
VAL_MNEMONIC_73="crush good online coconut crane margin position grab remain evolve hello ankle onion isolate vibrant hint like captain cash cute suggest rhythm answer silver"
VAL_MNEMONIC_74="zone fire soldier box blade soft color blood gadget agent second airport enemy odor dirt today inmate tomorrow female trash cereal depend stereo music"
VAL_MNEMONIC_75="exist frame under news inmate umbrella near cage barrel pond scan jungle issue pottery rhythm bamboo mushroom business object twist canvas connect response twin"
VAL_MNEMONIC_76="rail myth can cancel regular space tired view idea sketch exercise danger wood desert usage armed shadow unfair copy social angry flush vivid blanket"
VAL_MNEMONIC_77="kangaroo attend cluster fire rose kingdom borrow autumn blush because define fatal token peanut discover air today dolphin machine keen state often ski repair"
VAL_MNEMONIC_78="luggage enjoy trap hand rebuild equal identify cart expect road elegant give marble inhale drastic come title jelly plug today other unknown forest again"
VAL_MNEMONIC_79="horn observe grunt excuse indoor leisure render sunny pretty car real area exotic rate faint ride around forward adjust reflect mother mad mass bleak"
VAL_MNEMONIC_80="music pizza ladder fun detect large tennis loud caution insect job core supreme clutch seat crucial forest avoid enrich jar spoil scorpion second ugly"
VAL_MNEMONIC_81="rally dumb reflect document suspect expand hand vibrant stumble near unable buddy apology veteran glory bulk actual very charge lab offer render mammal maze"
VAL_MNEMONIC_82="wrong miss viable owner pattern middle shoot valve helmet success hair frown join such length ridge expire whisper cave daring trip primary filter canyon"
VAL_MNEMONIC_83="uncover coconut brown call tongue couple embody meadow doctor foil sweet kangaroo section unit city team monitor visual gallery receive elegant lobster sound sport"
VAL_MNEMONIC_84="hour swift genius once silk connect moon cage avoid hotel slender response flush imitate moon adult face shed job liquid join undo crunch into"
VAL_MNEMONIC_85="peanut dynamic sort struggle shallow kitchen drum card pet author liar session quarter script damp quit lunch voyage eagle clinic script smart fashion illegal"
VAL_MNEMONIC_86="tent rigid afraid layer route trick certain wish divide initial slush sauce appear orient decline nut cliff symbol jelly battle boat ski elevator addict"
VAL_MNEMONIC_87="potato term proud crumble vibrant attitude child cushion deny dash wall trash travel train quick biology again turkey ready glass reason voice prepare spot"
VAL_MNEMONIC_88="path barely vault mercy accuse answer report fish flip steak audit final balcony decline give despair fat jungle quarter notable check sentence blouse exotic"
VAL_MNEMONIC_89="bag midnight thumb have exile suggest imitate festival border insane possible card warfare beyond toddler dad capable since craft confirm want broccoli sphere scissors"
VAL_MNEMONIC_90="blue bike expire album battle hen can benefit remove sight other sketch avocado useful sleep want oyster siren volcano outer plastic thought truck kid"
VAL_MNEMONIC_91="where tribe grid torch gate clarify ready initial leave ginger panther east grant strike crush sentence tray lava help swear normal one off drive"
VAL_MNEMONIC_92="skirt labor infant again rookie fame dish crawl manage prepare tonight endorse furnace tent then note video entry juice know base twice jar betray"
VAL_MNEMONIC_93="express meat alley leg possible quarter banana inject enforce clever carpet river follow grain science neck reform genuine physical force cycle frame rocket furnace"
VAL_MNEMONIC_94="brave mimic grass hand awesome border afraid mask bird winter unfold betray turtle gold once quarter void machine ancient gate turkey check owner opera"
VAL_MNEMONIC_95="goose race deputy sister party often lazy then inject magic boss crumble silent powder armed live zero spike consider journey tired regular stable mixture"
VAL_MNEMONIC_96="discover initial live man coffee survey skirt favorite bone candy electric essay travel couch onion job shrug scatter vehicle limb picnic cross jeans embody"
VAL_MNEMONIC_97="fashion brain defense vintage undo oven cart spring glide radio team end almost pulse trip city hen vault fatigue script wheel melt bronze cupboard"
VAL_MNEMONIC_98="wrong special coin will lamp until cruise steak afford annual behave similar design boss later best absurd tube hard winter have foot script kangaroo"
VAL_MNEMONIC_99="bulk flavor ghost model fatal duck dragon panic ripple excite often they enlist letter spin upgrade dirt twice nasty monster wagon super mercy wide"
VAL_MNEMONIC_100="lucky buzz pond team lucky again dinner weapon razor race fancy crumble eyebrow cable bone sail nephew impact clutch candy arm fire romance gospel"
VAL_MNEMONICS=(
  "$VAL_MNEMONIC_1"
  "$VAL_MNEMONIC_2"
  "$VAL_MNEMONIC_3"
  "$VAL_MNEMONIC_4"
  "$VAL_MNEMONIC_5"
  "$VAL_MNEMONIC_6"
  "$VAL_MNEMONIC_7"
  "$VAL_MNEMONIC_8"
  "$VAL_MNEMONIC_9"
  "$VAL_MNEMONIC_10"
  "$VAL_MNEMONIC_11"
  "$VAL_MNEMONIC_12"
  "$VAL_MNEMONIC_13"
  "$VAL_MNEMONIC_14"
  "$VAL_MNEMONIC_15"
  "$VAL_MNEMONIC_16"
  "$VAL_MNEMONIC_17"
  "$VAL_MNEMONIC_18"
  "$VAL_MNEMONIC_19"
  "$VAL_MNEMONIC_20"
  "$VAL_MNEMONIC_21"
  "$VAL_MNEMONIC_22"
  "$VAL_MNEMONIC_23"
  "$VAL_MNEMONIC_24"
  "$VAL_MNEMONIC_25"
  "$VAL_MNEMONIC_26"
  "$VAL_MNEMONIC_27"
  "$VAL_MNEMONIC_28"
  "$VAL_MNEMONIC_29"
  "$VAL_MNEMONIC_30"
  "$VAL_MNEMONIC_31"
  "$VAL_MNEMONIC_32"
  "$VAL_MNEMONIC_33"
  "$VAL_MNEMONIC_34"
  "$VAL_MNEMONIC_35"
  "$VAL_MNEMONIC_36"
  "$VAL_MNEMONIC_37"
  "$VAL_MNEMONIC_38"
  "$VAL_MNEMONIC_39"
  "$VAL_MNEMONIC_40"
  "$VAL_MNEMONIC_41"
  "$VAL_MNEMONIC_42"
  "$VAL_MNEMONIC_43"
  "$VAL_MNEMONIC_44"
  "$VAL_MNEMONIC_45"
  "$VAL_MNEMONIC_46"
  "$VAL_MNEMONIC_47"
  "$VAL_MNEMONIC_48"
  "$VAL_MNEMONIC_49"
  "$VAL_MNEMONIC_50"
  "$VAL_MNEMONIC_51"
  "$VAL_MNEMONIC_52"
  "$VAL_MNEMONIC_53"
  "$VAL_MNEMONIC_54"
  "$VAL_MNEMONIC_55"
  "$VAL_MNEMONIC_56"
  "$VAL_MNEMONIC_57"
  "$VAL_MNEMONIC_58"
  "$VAL_MNEMONIC_59"
  "$VAL_MNEMONIC_60"
  "$VAL_MNEMONIC_61"
  "$VAL_MNEMONIC_62"
  "$VAL_MNEMONIC_63"
  "$VAL_MNEMONIC_64"
  "$VAL_MNEMONIC_65"
  "$VAL_MNEMONIC_66"
  "$VAL_MNEMONIC_67"
  "$VAL_MNEMONIC_68"
  "$VAL_MNEMONIC_69"
  "$VAL_MNEMONIC_70"
  "$VAL_MNEMONIC_71"
  "$VAL_MNEMONIC_72"
  "$VAL_MNEMONIC_73"
  "$VAL_MNEMONIC_74"
  "$VAL_MNEMONIC_75"
  "$VAL_MNEMONIC_76"
  "$VAL_MNEMONIC_77"
  "$VAL_MNEMONIC_78"
  "$VAL_MNEMONIC_79"
  "$VAL_MNEMONIC_80"
  "$VAL_MNEMONIC_81"
  "$VAL_MNEMONIC_82"
  "$VAL_MNEMONIC_83"
  "$VAL_MNEMONIC_84"
  "$VAL_MNEMONIC_85"
  "$VAL_MNEMONIC_86"
  "$VAL_MNEMONIC_87"
  "$VAL_MNEMONIC_88"
  "$VAL_MNEMONIC_89"
  "$VAL_MNEMONIC_90"
  "$VAL_MNEMONIC_91"
  "$VAL_MNEMONIC_92"
  "$VAL_MNEMONIC_93"
  "$VAL_MNEMONIC_94"
  "$VAL_MNEMONIC_95"
  "$VAL_MNEMONIC_96"
  "$VAL_MNEMONIC_97"
  "$VAL_MNEMONIC_98"
  "$VAL_MNEMONIC_99"
  "$VAL_MNEMONIC_100"
)
REV_MNEMONIC="tonight bonus finish chaos orchard plastic view nurse salad regret pause awake link bacon process core talent whale million hope luggage sauce card weasel"

# STRIDE 
STRIDE_CHAIN_ID=STRIDE
STRIDE_NODE_PREFIX=stride
STRIDE_NUM_NODES=20
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
GAIA_NUM_NODES=20
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
RELAYER_GAIA_ICS_EXEC="$DOCKER_COMPOSE run --rm relayer-gaia-ics"
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
RELAYER_STRIDE_ICS_ACCT=rly11
RELAYER_GAIA_ICS_ACCT=rly12
RELAYER_ACCTS=(
  $RELAYER_GAIA_ACCT 
  $RELAYER_JUNO_ACCT 
  $RELAYER_OSMO_ACCT 
  $RELAYER_STARS_ACCT 
  $RELAYER_HOST_ACCT 
  $RELAYER_EVMOS_ACCT
  $RELAYER_GAIA_ICS_ACCT
)

RELAYER_GAIA_MNEMONIC="fiction perfect rapid steel bundle giant blade grain eagle wing cannon fever must humble dance kitchen lazy episode museum faith off notable rate flavor"
RELAYER_JUNO_MNEMONIC="kiwi betray topple van vapor flag decorate cement crystal fee family clown cry story gain frost strong year blanket remain grass pig hen empower"
RELAYER_OSMO_MNEMONIC="unaware wine ramp february bring trust leaf beyond fever inside option dilemma save know captain endless salute radio humble chicken property culture foil taxi"
RELAYER_STARS_MNEMONIC="deposit dawn erosion talent old broom flip recipe pill hammer animal hill nice ten target metal gas shoe visual nephew soda harbor child simple"
RELAYER_HOST_MNEMONIC="renew umbrella teach spoon have razor knee sock divert inner nut between immense library inhale dog truly return run remain dune virus diamond clinic"
RELAYER_GAIA_ICS_MNEMONIC="size chimney clog job robot thunder gaze vapor economy smooth kit denial alter merit produce front force eager outside mansion believe fan tonight detect"
RELAYER_EVMOS_MNEMONIC="science depart where tell bus ski laptop follow child bronze rebel recall brief plug razor ship degree labor human series today embody fury harvest"
RELAYER_MNEMONICS=(
  "$RELAYER_GAIA_MNEMONIC"
  "$RELAYER_JUNO_MNEMONIC"
  "$RELAYER_OSMO_MNEMONIC"
  "$RELAYER_STARS_MNEMONIC"
  "$RELAYER_HOST_MNEMONIC"
  "$RELAYER_GAIA_ICS_MNEMONIC"
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
  $MAIN_CMD q staking validators | grep ${chain}_${val_index} -A 5 | tail -n 7 | grep operator | awk '{print $2}'
}

GET_ICA_ADDR() {
  chain_id="$1"
  ica_type="$2" #delegation, fee, redemption, or withdrawal

  $STRIDE_MAIN_CMD q stakeibc show-host-zone $chain_id | grep ${ica_type}_account -A 1 | grep address | awk '{print $2}'
}

TRIM_TX() {
  grep -E "code:|txhash:" | sed 's/^/  /'
}