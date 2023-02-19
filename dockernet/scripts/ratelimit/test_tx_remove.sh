CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../../config.sh
source ${CURRENT_DIR}/common.sh

test_remove_rate_limit() {
    print_header "TESTING TX REMOVE RATE LIMIT - ATOM FROM GAIA -> STRIDE"

    wait_until_epoch_end

    start_gaia_balance=$(get_balance GAIA uatom)
    start_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)

    # Remove the rate limit
    printf "\n>>> Removing rate limit...\n"
    submit_proposal_and_vote remove-rate-limit remove_uatom.json
    sleep 30

    # Then successfully transfer a large amount the removal 
    printf "\n>>> Transferring $INITIAL_CHANNEL_VALUE uatom...\n"
    $GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
    sleep 15

    # Confirm balance was updated appropriately
    end_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)
    expected_stride_balance=$((start_stride_balance+INITIAL_CHANNEL_VALUE))

    print_expectation $expected_stride_balance $end_stride_balance "Balance on Stride" 
}
