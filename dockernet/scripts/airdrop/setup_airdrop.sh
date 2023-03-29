### AIRDROP SETUP SCRIPT
#
#  Instructions: 
#   1. First, start the network with `make start-docker`
#   2. Then, run this script with `bash dockernet/scripts/airdrop/setup_airdrop.sh`
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

AIRDROP_NAME="evmos"
AIRDROP_RECIPIENT_1="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"
AIRDROP_RECIPIENT_2="stride17kht2x2ped6qytr2kklevtvmxpw7wq9rmuc3ca"
AIRDROP_RECIPIENT_3="stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv"
AIRDROP_RECIPIENT_4="stride1py0fvhdtq4au3d9l88rec6vyda3e0wtt9szext"
AIRDROP_RECIPIENT_5="stride1c5jnf370kaxnv009yhc3jt27f549l5u36chzem"
AIRDROP_DISTRIBUTOR_1="stride1qs6c3jgk7fcazrz328sqxqdv9d5lu5qqqgqsvj"

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


# ### Set up the airdrop

# create airdrop 1 
printf "\n\nCreating first airdrop, should last 1 hour and reset every 60 seconds to allow for new claims every 60 seconds..."
$STRIDE_MAIN_CMD tx claim create-airdrop $AIRDROP_NAME $(date +%s) 3600 ustrd --from d1 -y | TRIM_TX
sleep 5

printf "\nSetting up first airdrop allocations...\n"
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_1 1 --from d1 -y | TRIM_TX 
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_2 2 --from d1 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_3 3 --from d1 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations $AIRDROP_NAME $AIRDROP_RECIPIENT_4 4 --from d1 -y | TRIM_TX
sleep 5

echo "\n Querying airdrop eligibilities. The results of the query show the total claimable amount for each account. If they're non-zero, the airdrop is live! :)"
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_1 true
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_2 true
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_3 true
$STRIDE_MAIN_CMD q claim total-claimable $AIRDROP_NAME $AIRDROP_RECIPIENT_4 true

