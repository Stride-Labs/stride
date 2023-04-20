### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

### AIRDROP TESTING FLOW Pt 4

# This script tests multiple staggered airdrops live in tandem

# First, start the network with `make start-docker`
# Then, run this script with `bash dockernet/scripts/airdrop/airdrop4.sh`

echo "Registering accounts..."
# distributor address: stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl
echo "person pelican purchase boring theme eagle jaguar screen frame attract mad link ribbon ball poverty valley cross cradle real idea payment ramp nature anchor" | \
    $STRIDE_MAIN_CMD keys add distributor-test1 --recover

# stride1wl22etyhepwmsmycnvt3ragjyv2r5ctrk4emv3
echo "skill essence buddy slot trim rich acid end ability sketch evoke volcano fantasy visit maze mouse sword squirrel weasel mandate main author zebra lunar" | \
    $STRIDE_MAIN_CMD keys add distributor-test2 --recover

## AIRDROP SETUP
echo "Funding accounts..."
# Fund the distributor1 account
$STRIDE_MAIN_CMD tx bank send val1 stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl 100ustrd --from val1 -y | TRIM_TX
sleep 5
# Fund the distributor2 account
$STRIDE_MAIN_CMD tx bank send val1 stride1wl22etyhepwmsmycnvt3ragjyv2r5ctrk4emv3 100ustrd --from val1 -y | TRIM_TX
sleep 5

echo -e "\n>>> Initial Balances:"
echo "> Distributor1 Account [100ustrd expected]:"
$STRIDE_MAIN_CMD q bank balances stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl --denom ustrd 

echo "> Distributor2 Account [100ustrd expected]:"
$STRIDE_MAIN_CMD q bank balances stride1wl22etyhepwmsmycnvt3ragjyv2r5ctrk4emv3 --denom ustrd

echo "> Claim Account [5000000000000ustrd expected]:"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z --denom ustrd

# ### Test staggered airdrops
#  airdrop1 is ustrd; airdrop2 is ibc/ATOM. this simplifies telling them apart after testing a reset of airdrop1 before airdrop 2 has a chance to reset.

# create airdrop 1 for ustrd
echo -e "\n>>> Creating airdrop1 and allocations..."
$STRIDE_MAIN_CMD tx claim create-airdrop airdrop1 GAIA ustrd $(date +%s) 240 false --from distributor-test1 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from distributor-test1 -y | TRIM_TX
sleep 5

# claim airdrop
echo -e "\n>>> Claiming airdrop1"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
# verify claim record
echo "> Checking claim eligibility for airdrop1, should return 1 claim-record:"
$STRIDE_MAIN_CMD q claim claim-record airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# verify total claimable
echo "> Checking total claimable for airdrop1 [expected: 100ustrd]:"
$STRIDE_MAIN_CMD q claim total-claimable airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z true

echo -e "\n...Sleeping 60 before creating airdrop2..."
sleep 60

# create airdrop 2 
echo -e "\n>>> Creating airdrop2 and setting allocations"
$STRIDE_MAIN_CMD tx claim create-airdrop airdrop2 GAIA2 ustrd $(date +%s) 60 false --from distributor-test2 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from distributor-test2 -y | TRIM_TX
sleep 5

# claim airdrop
echo -e "\n>>> Claiming airdrop2"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
# verify claim record
echo "> Checking claim eligibility for airdrop2, should return 1 claim-record:"
$STRIDE_MAIN_CMD q claim claim-record airdrop2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# verify total claimable
echo "> Checking total claimable for airdrop2 [expected: 100ustrd]:"
$STRIDE_MAIN_CMD q claim total-claimable airdrop2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z true

echo -e "\n...Sleeping 60 more sec to wait for airdrop1 to reset..."
sleep 35

### airdrop 1 resets - check state before claim
echo -e "\n>>> Airdrop1 Reset <<<"
echo -e "\n>>> Verifying the vesting funds and balance for airdrop1. AFTER the reset, but BEFORE season2 claim"
# Check vesting
echo "> AIRDROP 1 - Vesting funds AFTER reset, but BEFORE claim [expected: XXX]:"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# Check balance
echo "> AIRDROP 1 - Balance AFTER reset, but BEFORE claim [XXX expected]:"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z --denom ustrd

# Claim again after reset
echo -e "\n>>> Claiming airdrop1"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX

# check state after claim
echo -e "\n>>> Verifying the vesting funds and balance for airdrop1. AFTER the reset, and AFTER season2 claim"
# Check vesting
echo "> AIRDROP 1 - Vesting funds AFTER reset, and AFTER claim [expected: XXX]:"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# Check balance
echo "> AIRDROP 1 - Balance AFTER reset, and AFTER claim [XXX expected]:"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z --denom ustrd

echo -e "\n...Sleeping 60 more sec to wait for airdrop2 to reset..."

### airdrop 2 resets before airdrop 1 has a chance to reset again
echo -e "\n>>> Airdrop2 Reset <<<"
echo -e "\n>>> Verifying the vesting funds and balance for airdrop2. AFTER the reset, but BEFORE season2 claim"
# Check vesting
echo "> AIRDROP 2 - Vesting AFTER reset, but BEFORE claim [expected: XXX]:"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# Check balance
echo "> AIRDROP 2 - Balance AFTER reset, but BEFORE claim [expected: XXX]:"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z --denom ustrd

# Claim again after reset
echo -e "\n>>> Claiming airdrop2 after reset"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX

# check state after claim
echo -e "\n>>> Verifying the vesting funds and balance for airdrop2. AFTER the reset, and AFTER season2 claim"
# Check vesting
echo "> AIRDROP 2 - Vesting funds AFTER reset, and AFTER claim [expected: XXX]:"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# Check balance
echo "> AIRDROP 2 - Balance AFTER reset, and AFTER claim [XXX expected]:"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z --denom ustrd
