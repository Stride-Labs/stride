### AIRDROP SETUP SCRIPT
#
#  Instructions: 
#   1. First, start the network with `make start-docker`
#   2. Then, run this script with `bash dockernet/scripts/airdrop/evmos_airdrop.sh`
#   3. If the final stdout print lines from the script match what's below, the airdrop is live!
#    
#      \n Querying airdrop eligibilities
#         coins:
#         - amount: "22222255"
#         denom: ustrd
#         coins:
#         - amount: "44444511"
#         denom: ustrd
#         coins:
#         - amount: "111111279"
#         denom: ustrd
#         coins:
#         - amount: "222222555"
#         denom: ustrd
#

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

$STRIDE_MAIN_CMD keys delete airdrop-recipient-1 -y &> /dev/null || true 
$EVMOS_MAIN_CMD keys delete airdrop-recipient-1 -y &> /dev/null || true 

AIRDROP_NAME="evmos"

# airdrop recipient 1 key
# add airdrop recipient 1
echo "prosper vivid sign donkey involve flee behind save satoshi reason girl cable ranch can arrive unable coyote race model disagree buzz peasant mechanic position" | \
    $STRIDE_MAIN_CMD keys add airdrop-recipient-1 --recover

echo "prosper vivid sign donkey involve flee behind save satoshi reason girl cable ranch can arrive unable coyote race model disagree buzz peasant mechanic position" | \
    $EVMOS_MAIN_CMD keys add airdrop-recipient-1 --recover

AIRDROP_RECIPIENT_1_STRIDE="stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep"
AIRDROP_RECIPIENT_1_EVMOS="evmos1nmwp5uh5a3g08668c5eynes0hyfaw94dfnj796"
AIRDROP_RECIPIENT_1_MECHANICAL="stride1nmwp5uh5a3g08668c5eynes0hyfaw94dgervt7"

AIRDROP_RECIPIENT_2="stride17kht2x2ped6qytr2kklevtvmxpw7wq9rmuc3ca"
AIRDROP_RECIPIENT_3="stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv"
AIRDROP_RECIPIENT_4_TO_BE_REPLACED="stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm"
AIRDROP_DISTRIBUTOR_1="stride1qs6c3jgk7fcazrz328sqxqdv9d5lu5qqqgqsvj"


echo ">>> Querying the claims module to verify that the new address is eligible"
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_1_STRIDE true

# cleanup: clear out and re-fund accounts
$STRIDE_MAIN_CMD keys delete d1 -y &> /dev/null || true 
# add the airdrop distributor account
echo "rebel tank crop gesture focus frozen essay taxi prison lesson prefer smile chaos summer attack boat abandon school average ginger rib struggle drum drop" | \
    $STRIDE_MAIN_CMD keys add d1 --recover

## AIRDROP SETUP
printf "Funding accounts..."
# Fund the d1 account
$STRIDE_MAIN_CMD tx bank send val1 $AIRDROP_DISTRIBUTOR_1 100000000ustrd --from val1 -y | TRIM_TX
sleep 5
# query the balance of the d1 account to make sure it was funded
$STRIDE_MAIN_CMD q bank balances $AIRDROP_DISTRIBUTOR_1

# fund the evmos account
$EVMOS_MAIN_CMD tx bank send nval1 evmos1nmwp5uh5a3g08668c5eynes0hyfaw94dfnj796 1000aevmos --from val1 -y
sleep 5
# query the balance of the airdrop-recipient-1 account to make sure it was funded
$EVMOS_MAIN_CMD q bank balances $AIRDROP_RECIPIENT_1_EVMOS



# ### Set up the airdrop

# create airdrop 1 
printf "\n\nCreating first airdrop, should last 1 hour and reset every 60 seconds to allow for new claims every 60 seconds..."
$STRIDE_MAIN_CMD tx claim create-airdrop $AIRDROP_NAME $(date +%s) 3600 ustrd --from d1 -y | TRIM_TX
sleep 5

printf "\nSetting up first airdrop allocations...\n"
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_1_MECHANICAL 1 --from d1 -y | TRIM_TX 
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_2 2 --from d1 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_3 3 --from d1 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_4_TO_BE_REPLACED 4 --from d1 -y | TRIM_TX
sleep 5

echo "\n Querying airdrop eligibilities. The results of the query show the total claimable amount for each account. If they're non-zero, the airdrop is live! :)"
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_1_MECHANICAL true
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_2 true
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_3 true
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_4_TO_BE_REPLACED true

echo "Sleeping 2 minutes before linking the evmos address to its stride address..."
sleep 10
echo "\n Overwrite airdrop elibibility for recipient 4. They should no longer be eligible." 
#         b. ibc-transfer from Osmo to Stride to change the airdrop account to stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep
#              Memo: {
#                "autopilot": {
#                     "stakeibc": {
#                       "stride_address": "stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep",
#                       },
#                         "claim": {
#                         }
#                    },
#                }
#               Receiver: "xxx"
# Note: autopilot will look at the sender of the packet (evmos1nmwp5uh5a3g08668c5eynes0hyfaw94dfnj796) and convert this address to the mechanical
# stride address (stride1nmwp5uh5a3g08668c5eynes0hyfaw94dgervt7), then reset it to the true stride address (stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep)
MEMO='{"autopilot": {"receiver": "stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep","claim": { "stride_address": "stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep", "airdrop_id": "evmos" } }}'
# $EVMOS_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$MEMO" 1aevmos --from airdrop-recipient-1 -y | TRIM_TX
echo "tx ibc-transfer transfer transfer channel-0 stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep 1aevmos --note "$MEMO" --from airdrop-recipient-1 -y | TRIM_TX"
$EVMOS_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep 1aevmos --note "$MEMO" --from airdrop-recipient-1 -y | TRIM_TX
echo ">>> Waiting for 15 seconds to allow the IBC transfer to complete..."
sleep 5
$EVMOS_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1qlly03ar5ll85ww4usvkv09832vv5tkhtnnaep 1aevmos --memo "$MEMO" --from airdrop-recipient-1 -y | TRIM_TX
echo ">>> Waiting for 15 seconds to allow the IBC transfer to complete..."
sleep 15

echo ">>> Querying the claims module to verify that the new address is eligible"
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_1_STRIDE true