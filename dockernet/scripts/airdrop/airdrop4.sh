### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

printf $STRIDE_MAIN_CMD
# CLEANUP if running tests twice, clear out and re-fund accounts
$STRIDE_MAIN_CMD keys delete dtest1 -y &> /dev/null || true 
$STRIDE_MAIN_CMD keys delete dtest2 -y &> /dev/null || true 
$GAIA_MAIN_CMD keys delete hosttest -y &> /dev/null || true 
$STRIDE_MAIN_CMD keys delete airdrop-test -y &> /dev/null || true 
$OSMO_MAIN_CMD keys delete host-address-test -y &> /dev/null || true 

# First, start the network with `make start-docker`
# Then, run this script with `bash dockernet/scripts/airdrop.sh`

# stride1qs6c3jgk7fcazrz328sqxqdv9d5lu5qqqgqsvj
printf "rebel tank crop gesture focus frozen essay taxi prison lesson prefer smile chaos summer attack boat abandon school average ginger rib struggle drum drop" | \
    $STRIDE_MAIN_CMD keys add dtest1 --recover

# stride1fc5k5lvjkt7qp5hn6kwey439p2x7ymm6w5wphk
printf "ocean spike awake world armor gossip harbor expand draft must scan give brass oval portion enhance stadium seven iron destroy ski frame loyal beyond" | \
    $STRIDE_MAIN_CMD keys add dtest2 --recover

## AIRDROP SETUP
printf "Funding accounts..."
# Fund the dtest1 and dtest2 accounts
$STRIDE_MAIN_CMD tx bank send val1 stride1qs6c3jgk7fcazrz328sqxqdv9d5lu5qqqgqsvj 100ustrd --from val1 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx bank send val1 stride1fc5k5lvjkt7qp5hn6kwey439p2x7ymm6w5wphk 100ustrd --from val1 -y | TRIM_TX
sleep 5
# query the balance of the dtest1 account
$STRIDE_MAIN_CMD q bank balances stride1qs6c3jgk7fcazrz328sqxqdv9d5lu5qqqgqsvj
$STRIDE_MAIN_CMD q bank balances stride1fc5k5lvjkt7qp5hn6kwey439p2x7ymm6w5wphk



# ### Test staggered airdrops
#  airdrop1 is ustrd; airdrop2 is ibc/ATOM. this simplifies telling them apart after testing a reset of airdrop1 before airdrop 2 has a chance to reset.

# create airdrop 1 
printf "\n\nSetting up first airdrop allocations and claiming"
$STRIDE_MAIN_CMD tx claim create-airdrop airdrop1 $(date +%s) 240 ustrd --from dtest1 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from dtest1 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
echo "\n Querying airdrop 1 claim-record"
$STRIDE_MAIN_CMD q claim claim-record airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
echo "\n Querying airdrop 1 total-claimable"
$STRIDE_MAIN_CMD q claim total-claimable airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z true
printf "...Sleeping 60 sec before setting up airdrop 2..."
sleep 60

# create airdrop 2 
printf "\n\n Setting up airdrop 2"
$STRIDE_MAIN_CMD tx claim create-airdrop airdrop2 $(date +%s) 240 ustrd --from dtest2 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from dtest2 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 5
echo "\n Querying airdrop 2 claim-record"
$STRIDE_MAIN_CMD q claim claim-record airdrop2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
echo "\n Querying airdrop 2 total-claimable"
$STRIDE_MAIN_CMD q claim total-claimable airdrop2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z true
printf "\n...Sleeping 60 more sec to wait for airdro1 reset to complete..."
sleep 60

# airdrop 1 resets
printf "\n\n\n> Query how many funds are vesting and in the balance before the airdrop1 season2 claim"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
printf "\n> Claiming more funds after reset"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 5
printf "\n> Verify more funds are vesting after the reset"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z

# airdrop 2 resets before airdrop 1 has a chance to reset again
printf "\n...Sleeping 60 more sec to wait for airdro2 reset to complete..."
sleep 120
printf "\n\n\n> Query how many funds are vesting and in the balance before the airdrop2 season2 claim"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
printf "\n> Claiming more funds after reset"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 5
printf "\n> Verify more funds are vesting after the airdrop2 reset"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z

