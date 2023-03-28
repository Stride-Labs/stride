### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# CLEANUP if running tests twice, clear out and re-fund accounts
# $STRIDE_MAIN_CMD keys delete distributor-test3 -y &> /dev/null || true 
# $GAIA_MAIN_CMD keys delete hosttest -y &> /dev/null || true 
# $STRIDE_MAIN_CMD keys delete airdrop-test -y &> /dev/null || true 
# $OSMO_MAIN_CMD keys delete host-address-test -y &> /dev/null || true 

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
sleep 8
# query the distribution-test3 account
echo "query the distribution-test3 account"
$STRIDE_MAIN_CMD q bank balances stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl
# Fund the distributor-test3 account
$STRIDE_MAIN_CMD tx bank send val1 stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl 100ustrd --from val1 -y | TRIM_TX
sleep 8
# Fund the airdrop account
$STRIDE_MAIN_CMD tx bank send val1 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000000ustrd --from val1 -y | TRIM_TX
sleep 8

echo "DISTRIBUTOR"
$STRIDE_MAIN_CMD q bank balances stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl
echo "AIRDROP"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z

### Test airdrop reset and multiple claims flow
    #   The Stride airdrop occurs in batches. We need to test three batches. 

    # SETUP
    # 1. Create a new airdrop that rolls into its next batch in just 30 seconds
    #    - include the add'l param that makes each batch 30 seconds long (after the first batch) 
    # 2. Set the airdrop allocations

# Create the airdrop, so that the airdrop account can claim tokens
echo ">>> Testing multiple airdrop reset and claims flow..."
$STRIDE_MAIN_CMD tx claim create-airdrop stride2 $(date +%s) 240 ustrd --from distributor-test3 -y | TRIM_TX
sleep 8
# # Set airdrop allocations
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations stride2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from distributor-test3 -y | TRIM_TX
sleep 8

#     # BATCH 1
#     # 3. Check eligibility and claim the airdrop
echo "> Checking claim record elibility"
$STRIDE_MAIN_CMD q claim claim-record stride2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q claim claim-record stride2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
echo "> Claiming airdrop"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 8

#     # 5. Query to check airdrop vesting account was created (w/ correct amount)
echo "Verifying funds are vesting, should be 1."
$STRIDE_MAIN_CMD q claim claim-record stride2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z


    # BATCH 2
    # 6. Wait 60 seconds
printf "\n\n\n> Waiting 60 seconds for next batch..."
sleep 120
$STRIDE_MAIN_CMD q claim claim-record stride2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
    # 7. Claim the airdrop
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 8
#     # 8. Query to check airdrop vesting account was created (w/ correct amount)
echo "> Verifying more funds are vesting, should be 2."
$STRIDE_MAIN_CMD q claim claim-record stride2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z

#     # BATCH 3
#     # 10. Wait 60 seconds
printf "\n\n\n> Waiting 60 seconds for next batch..."
sleep 65
$STRIDE_MAIN_CMD q claim claim-record stride2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
#     # 11. Claim the airdrop
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 8
#     # 12. Query to check airdrop vesting account was created (w/ correct amount)
echo "> Verifying more funds are vesting, should be 3."
# $STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q claim claim-record stride2 stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z
