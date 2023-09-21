### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# CLEANUP if running tests twice, clear out and re-fund accounts
$STRIDE_MAIN_CMD keys delete distributor-test3 -y &> /dev/null || true 
$GAIA_MAIN_CMD keys delete hosttest -y &> /dev/null || true 
$STRIDE_MAIN_CMD keys delete airdrop-test -y &> /dev/null || true 
$OSMO_MAIN_CMD keys delete host-address-test -y &> /dev/null || true 

# First, start the network with `make start-docker`
# Then, run this script with `bash dockernet/scripts/airdrop.sh`

# NOTE: First, store the keys using the following mnemonics
# distributor address: stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl
# distributor mnemonic: barrel salmon half click confirm crunch sense defy salute process cart fiscal sport clump weasel render private manage picture spell wreck hill frozen before
echo "person pelican purchase boring theme eagle jaguar screen frame attract mad link ribbon ball poverty valley cross cradle real idea payment ramp nature anchor" | \
    $STRIDE_MAIN_CMD keys add distributor-test3 --recover

# airdrop-test address: stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# airdrop claimer mnemonic: royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate
echo "royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate" | \
    $STRIDE_MAIN_CMD keys add airdrop-test --recover

## AIRDROP SETUP
echo "Funding accounts..."
# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
sleep 5
# Fund the distributor account
$STRIDE_MAIN_CMD tx bank send val1 stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl 600000ustrd --from val1 -y | TRIM_TX
sleep 5
# Fund the airdrop account
$STRIDE_MAIN_CMD tx bank send val1 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000000ustrd --from val1 -y | TRIM_TX
sleep 5

### Test airdrop reset and multiple claims flow
    #   The Stride airdrop occurs in batches. We need to test three batches. 

    # SETUP
    # 1. Create a new airdrop that rolls into its next batch in just 30 seconds
    #    - include the add'l param that makes each batch 30 seconds long (after the first batch) 
    # 2. Set the airdrop allocations

# Create the airdrop, so that the airdrop account can claim tokens
echo ">>> Testing multiple airdrop reset and claims flow..."
$STRIDE_MAIN_CMD tx claim create-airdrop stride2 $(date +%s) 30 ustrd --from distributor-test3 -y | TRIM_TX
sleep 5
# # Set airdrop allocations
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations stride2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from distributor-test3 -y | TRIM_TX
sleep 5

#     # BATCH 1
#     # 3. Check eligibility and claim the airdrop
echo "> Checking claim record elibility"
$STRIDE_MAIN_CMD q claim claim-record stride stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
echo "> Claiming airdrop"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 5

#     # 5. Query to check airdrop vesting account was created (w/ correct amount)
echo "Verifying funds are vesting, should be 1."
$STRIDE_MAIN_CMD q claim user-vestings stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl


    # BATCH 2
    # 6. Wait 30 seconds
echo "> Waiting 30 seconds for next batch..."
sleep 30
    # 7. Claim the airdrop
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX

#     # 8. Query to check airdrop vesting account was created (w/ correct amount)
echo "> Verifying more funds are vesting, should be 2."
$STRIDE_MAIN_CMD q claim user-vestings stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl

#     # BATCH 3
#     # 10. Wait 30 seconds
echo "> Waiting 30 seconds for next batch..."
sleep 30
#     # 11. Claim the airdrop
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX

#     # 12. Query to check airdrop vesting account was created (w/ correct amount)
echo "> Verifying more funds are vesting, should be 3."
$STRIDE_MAIN_CMD q claim user-vestings stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl




# ### Test staggered airdrops

# # create airdrop 1 with a 60 day start window, 60 sec reset, claim, sleep 35
# # $STRIDE_MAIN_CMD tx claim create-airdrop airdrop1 $(date +%s) 60 ustrd --from distributor-test3 -y
# # sleep 5
# # $STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from distributor-test3 -y
# # sleep 5
# # $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# # sleep 35

# # # create airdrop 2 with a 60 day start window, 60 sec reset, claim, sleep 35
# # $STRIDE_MAIN_CMD tx claim create-airdrop airdrop1 $(date +%s) 60 stuatom --from distributor-test3 -y
# # sleep 5
# # $STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop1 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from distributor-test3 -y
# # sleep 5
# # $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# # sleep 35

# # # airdrop 1 resets
# # $STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# # $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
# # $STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z