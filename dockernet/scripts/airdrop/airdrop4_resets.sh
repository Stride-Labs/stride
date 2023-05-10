### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

### AIRDROP TESTING FLOW Pt 4 (RESETS)

# This script tests that the the airdrop resets properly

# To run:
#   1. Update the following in `x/claim/types/params.go`
#      * `DefaultEpochDuration` to `time.Second * 60`
#      * `DefaultVestingInitialPeriod` to `time.Second * 120`
#   2. Start the network with `make start-docker`
#   3. Run this script with `bash dockernet/scripts/airdrop/airdrop4_resets.sh`

# NOTE: First, store the keys using the following mnemonics
echo "Registering distributor account..."
# distributor address: stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl
# distributor mnemonic: barrel salmon half click confirm crunch sense defy salute process cart fiscal sport clump weasel render private manage picture spell wreck hill frozen before
echo "person pelican purchase boring theme eagle jaguar screen frame attract mad link ribbon ball poverty valley cross cradle real idea payment ramp nature anchor" | \
    $STRIDE_MAIN_CMD keys add distributor-test --recover

## AIRDROP SETUP
echo "Funding accounts..."
# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
sleep 15
# Fund the distributor account
$STRIDE_MAIN_CMD tx bank send val1 stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl 100ustrd --from val1 -y | TRIM_TX
sleep 5

# Confirm initial balance setup
echo -e "\n>>> Initial Balances:"
echo "> Distributor Account [100ustrd expected]:"
$STRIDE_MAIN_CMD q bank balances stride12lw3587g97lgrwr2fjtr8gg5q6sku33e5yq9wl --denom ustrd 

echo "> Claim Account [5000000000000ustrd expected]:"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z --denom ustrd

### Test airdrop reset and multiple claims flow
    #   The Stride airdrop occurs in batches. We need to test three batches. 

    # SETUP
    # 1. Create a new airdrop that rolls into its next batch in just 30 seconds
    #    - include the add'l param that makes each batch 30 seconds long (after the first batch) 
    # 2. Set the airdrop allocations

# Create the airdrop, so that the airdrop account can claim tokens
echo -e "\n>>> Creating airdrop and setting allocations..."
$STRIDE_MAIN_CMD tx claim create-airdrop gaia GAIA ustrd $(date +%s) 40000000 false --from distributor-test -y | TRIM_TX
sleep 5
# Set airdrop allocations
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations gaia stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z 1 --from distributor-test -y | TRIM_TX
sleep 5
# Check eligibility
echo "> Checking claim elibility, should return 1 claim record:"
$STRIDE_MAIN_CMD q claim claim-record gaia stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z

#     # BATCH 1
#     # 3. Check eligibility and claim the airdrop
echo -e "\n>>> Claiming airdrop"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 5

#     # 5. Query to check airdrop vesting account was created (w/ correct amount)
echo -e "\n>>> Claim verification..."
# Check actions
echo "> Checking claim record actions [expected: 1 action complete]:"
$STRIDE_MAIN_CMD q claim claim-record gaia stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z | grep claim_record -A 4
# Check vesting
echo -e "\n> Verifying funds are vesting [expected: 20ustrd]:"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z | grep spendable_coins -A 2
# Check balance
echo -e "\n> Verifying balance [expected: 5000000000020ustrd]:"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z --denom ustrd


#    # BATCH 2
#    # 6. Wait 120 seconds
echo -e "\n>>> Waiting 120 seconds for next batch..."
sleep 120
echo -e "\n>>> Verify claim was reset [expected: no actions complete]:"
$STRIDE_MAIN_CMD q claim claim-record gaia stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z | grep claim_record -A 4

    # 7. Claim the airdrop
echo -e "\n>>> Claim airdrop"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 5

#     # 8. Query to check airdrop vesting account was created (w/ correct amount)
echo -e "\n>>> Claim verification..."
# Check actions
echo "> Checking claim record actions [expected: 1 action complete]:"
$STRIDE_MAIN_CMD q claim claim-record gaia stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z  | grep claim_record -A 4
# Check vesting
echo -e "\n> Verifying the vesting tokens have not changed [expected: 20ustrd]:"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z | grep spendable_coins -A 2
# Check balance
echo -e "\n> Verifying balance [expected: 5000000000036ustrd]:"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z --denom ustrd

#     # BATCH 3
#     # 10. Wait 65 seconds
echo -e ">>> Waiting 65 seconds for next batch..."
sleep 65
echo -e "\n>>> Verify claim was reset [expected: no actions complete]:"
$STRIDE_MAIN_CMD q claim claim-record gaia stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z | grep claim_record -A 4

#     # 11. Claim the airdrop
echo -e "\n>>> Claim airdrop"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z -y | TRIM_TX
sleep 5

#     # 12. Query to check airdrop vesting account was created (w/ correct amount)
echo -e "\n>>> Claim verification..."
# Check actions
echo "> Checking claim record actions [expected: 1 action complete]:"
$STRIDE_MAIN_CMD q claim claim-record gaia stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z  | grep claim_record -A 4
# Check vesting
echo -e "\n> Verifying the vesting tokens have not changed [expected: 20ustrd]:"
$STRIDE_MAIN_CMD q claim user-vestings stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z | grep spendable_coins -A 2
# Check balance
echo -e "\n> Verifying balance [expected: 5000000000049ustrd]:"
$STRIDE_MAIN_CMD q bank balances stride1kwll0uet4mkj867s4q8dgskp03txgjnswc2u4z --denom ustrd


