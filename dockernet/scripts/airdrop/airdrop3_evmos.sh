### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

### AIRDROP TESTING FLOW Pt 3 (EVMOS)

# This script tests claiming an evmos airdrop via autopilot with ibc-go v5+
# The claim is initiated by sending an IBC transfer with the stride address in the memo

# To run:
# 1. Enable EVMOS as the only dockernet host chain
# 2. Start the network with `make start-docker`
# 3. Run this script with `bash dockernet/scripts/airdrop/airdrop3_evmos.sh`

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

AIRDROP_NAME="evmos"
AIRDROP_CHAIN_ID="evmos_9001-2"

# The STRIDE/EVMOS recipient addresses represent the actual addresses owned by the claimer
# The mechanical address is a transient address derived by converting the evmos address 
#  directly to a stride address, without taking the coin type into consideration
# The mechanical address can be thought of as the "key" to this user's airdrop
AIRDROP_RECIPIENT_1_STRIDE="stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep"
AIRDROP_RECIPIENT_1_EVMOS="evmos1nmwp5uh5a3g08668c5eynes0hyfaw94dfnj796"
AIRDROP_RECIPIENT_1_MECHANICAL="stride1nmwp5uh5a3g08668c5eynes0hyfaw94dgervt7"

AIRDROP_RECIPIENT_2="stride17kht2x2ped6qytr2kklevtvmxpw7wq9rmuc3ca"
AIRDROP_RECIPIENT_3="stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv"
AIRDROP_RECIPIENT_4="stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm"

AIRDROP_DISTRIBUTOR_1="stride1qs6c3jgk7fcazrz328sqxqdv9d5lu5qqqgqsvj"

# airdrop recipient 1 key
# add airdrop recipient 1 on Stride
echo "prosper vivid sign donkey involve flee behind save satoshi reason girl cable ranch can arrive unable coyote race model disagree buzz peasant mechanic position" | \
    $STRIDE_MAIN_CMD keys add airdrop-recipient-1 --recover
# add airdrop recipient 1 on Evmos
echo "prosper vivid sign donkey involve flee behind save satoshi reason girl cable ranch can arrive unable coyote race model disagree buzz peasant mechanic position" | \
    $EVMOS_MAIN_CMD keys add airdrop-recipient-1 --recover
# add the airdrop distributor account
echo "rebel tank crop gesture focus frozen essay taxi prison lesson prefer smile chaos summer attack boat abandon school average ginger rib struggle drum drop" | \
    $STRIDE_MAIN_CMD keys add distributor --recover

## AIRDROP SETUP
echo "Funding accounts..."
# Fund the distributor account
$STRIDE_MAIN_CMD tx bank send val1 $AIRDROP_DISTRIBUTOR_1 100000000ustrd --from val1 -y | TRIM_TX
sleep 5
# Fund the evmos airdrop-recipient-1 account
$EVMOS_MAIN_CMD tx bank send eval1 evmos1nmwp5uh5a3g08668c5eynes0hyfaw94dfnj796 1000000000000000000aevmos --from val1 -y | TRIM_TX
sleep 5
# Fund the stride airdrop-recipient-1 account
$STRIDE_MAIN_CMD tx bank send val1 $AIRDROP_RECIPIENT_1_STRIDE 1000000ustrd --from val1 -y | TRIM_TX
sleep 5

# Verify initial balances
echo -e "\n>>> Initial Balances:"
# Distributor account
echo -e "\n> Distributor Account [100000000ustrd expected]:"
$STRIDE_MAIN_CMD q bank balances $AIRDROP_DISTRIBUTOR_1 --denom ustrd
# Airdrop recipient evmos account
echo -e "\n> Airdrop Recipient Account (on Evmos) [1000000000000000000aevmos expected]:"
$EVMOS_MAIN_CMD q bank balances $AIRDROP_RECIPIENT_1_EVMOS --denom aevmos
# Airdrop recipient stride account
echo -e "\n> Airdrop Recipient Stride (on Stride) [1000000ustrd expected]:"
$STRIDE_MAIN_CMD q bank balances $AIRDROP_RECIPIENT_1_STRIDE --denom ustrd

# ### Set up the airdrop
# create airdrop 1 
echo -e "\n\n>>> Creating Evmos airdrop..."
$STRIDE_MAIN_CMD tx claim create-airdrop $AIRDROP_NAME $AIRDROP_CHAIN_ID ustrd $(date +%s) 40000000 true --from distributor -y | TRIM_TX
sleep 5

echo -e "\n>>> Setting up airdrop allocations across 4 recipients..."
# set allocations to each recipient
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_1_MECHANICAL 1 --from distributor -y | TRIM_TX 
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_2 2 --from distributor -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_3 3 --from distributor -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_4 4 --from distributor -y | TRIM_TX
sleep 5

echo -e "\n>>> Checking airdrop eligibility..."
echo -e "\n >Checking the mechanical address. This should show 10000000ustrd since the address has not been overwritten yet."
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_1_MECHANICAL true

echo -e "\n >Checking for recipient 1's actual stride address. This should show no coins since we have not overwritten the mechanical address yet."
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_1_STRIDE true

echo -e "\n >Checking all other recipients. If they're non-zero, the airdrop is setup properly! :)"
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_2 true
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_3 true
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_4 true

echo -e "\n\n>>> Overwriting airdrop elibibility for recipient 1 (i.e. overriding the mechanical address with the true address)" 
#         b. ibc-transfer from Osmo to Stride to change the airdrop account to stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep
#              Memo: {
#                "autopilot": {
#                     "claim": {
#                       "stride_address": "stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep",
#                      },
#                 },
#              }
#              Receiver: "xxx"
# Note: autopilot will look at the sender of the packet (evmos1nmwp5uh5a3g08668c5eynes0hyfaw94dfnj796) and convert this address to the mechanical
# stride address (stride1nmwp5uh5a3g08668c5eynes0hyfaw94dgervt7) which will act as the key to lookup the claim record.
# Then the record will get set to the true stride address (stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep) 
MEMO='{ "autopilot": { "receiver": "'"$AIRDROP_RECIPIENT_1_STRIDE"'",  "claim": { "stride_address": "'"$AIRDROP_RECIPIENT_1_STRIDE"'" } } }'
$EVMOS_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $AIRDROP_RECIPIENT_1_STRIDE 1aevmos --memo "$MEMO" --from airdrop-recipient-1 -y | TRIM_TX
sleep 15

echo -e "\n>>> Verify the new stride address is eligible [10000000ustrd expected]:"
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_1_STRIDE true

echo -e "\n>>> Verify the old mechanical address is no longer eligible [empty array expected]:"
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_1_MECHANICAL true

echo -e "\n>>> Claiming the airdrop from the new stride address"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from airdrop-recipient-1 -y | TRIM_TX
sleep 5

echo "\n> After claiming, check that an action was complete"
$STRIDE_MAIN_CMD q claim claim-record $AIRDROP_NAME $AIRDROP_RECIPIENT_1_STRIDE | grep claim_record -A 4