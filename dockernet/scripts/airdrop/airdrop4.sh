### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

echo $STRIDE_MAIN_CMD
# CLEANUP if running tests twice, clear out and re-fund accounts
$STRIDE_MAIN_CMD keys delete distributor-test3 -y &> /dev/null || true 
$STRIDE_MAIN_CMD keys delete distributor-test4 -y &> /dev/null || true 
$GAIA_MAIN_CMD keys delete hosttest -y &> /dev/null || true 
$STRIDE_MAIN_CMD keys delete airdrop-test -y &> /dev/null || true 
$OSMO_MAIN_CMD keys delete host-address-test -y &> /dev/null || true 

# First, start the network with `make start-docker`
# Then, run this script with `bash dockernet/scripts/airdrop.sh`

echo "person pelican purchase boring theme eagle jaguar screen frame attract mad link ribbon ball poverty valley cross cradle real idea payment ramp nature anchor" | \
    $STRIDE_MAIN_CMD keys add distributor-test3 --recover

# stride1wl22etyhepwmsmycnvt3ragjyv2r5ctrk4emv3
echo "skill essence buddy slot trim rich acid end ability sketch evoke volcano fantasy visit maze mouse sword squirrel weasel mandate main author zebra lunar" | \
    $STRIDE_MAIN_CMD keys add distributor-test4 --recover

## AIRDROP SETUP
echo "Funding accounts..."
# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1wl22etyhepwmsmycnvt3ragjyv2r5ctrk4emv3 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
sleep 5
# Fund the distributor3 account
$STRIDE_MAIN_CMD tx bank send val1 stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl 600000ustrd --from val1 -y | TRIM_TX
sleep 5
# Fund the distributor4 account
$STRIDE_MAIN_CMD tx bank send val1 stride1wl22etyhepwmsmycnvt3ragjyv2r5ctrk4emv3 600000$ATOM_DENOM --from val1 -y | TRIM_TX
sleep 5


# ### Test staggered airdrops
#  airdrop1 is ustrd; airdrop2 is ibc/ATOM. this simplifies telling them apart after testing a reset of airdrop1 before airdrop 2 has a chance to reset.

# create airdrop 1 with a 60 day start window, 60 sec reset, claim, sleep 35
$STRIDE_MAIN_CMD tx claim create-airdrop airdrop1 $(date +%s) 60 ustrd --from distributor-test3 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from distributor-test3 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
echo "...Sleeping 35 more sec to wait for reset to complete..."
sleep 35

# create airdrop 2 with a 60 day start window, 60 sec reset, claim, sleep 35
$STRIDE_MAIN_CMD tx claim create-airdrop airdrop1 $(date +%s) 60 $ATOM_DENOM --from distributor-test4 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from distributor-test4 -y | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
echo "...Sleeping 35 more sec to wait for reset to complete..."
sleep 35

# airdrop 1 resets
echo "> Query how many funds are vesting before the reset + re-claim"
$STRIDE_MAIN_CMD q claim user-vestings stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl
echo "> Claiming more funds after reset"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 5
echo "> Verify more funds are vesting before the reset"
$STRIDE_MAIN_CMD q claim user-vestings stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl

