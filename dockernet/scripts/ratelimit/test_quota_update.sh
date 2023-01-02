CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../../config.sh
source ${CURRENT_DIR}/common.sh

test_quota_update() {
    print_header "TESTING QUOTA UPDATE - OSMO FROM OSMOSIS -> STRIDE"

    wait_until_epoch_end

    start_osmo_balance=$(get_balance OSMO uosmo)
    start_stride_balance=$(get_balance STRIDE $IBC_OSMO_CHANNEL_2_DENOM)

    # Transfer once successfully (just below the limit)
    check_transfer_status OSMO STRIDE channel-0 channel-2 99999999 uosmo $IBC_OSMO_CHANNEL_2_DENOM true

    # Attempt to transfer but should fail because it gets rate limited
    check_transfer_status OSMO STRIDE channel-0 channel-2 2 uosmo $IBC_OSMO_CHANNEL_2_DENOM false 

    # Transfer in the other direction so the channel value stays is unchanged
    check_transfer_status STRIDE OSMO channel-2 channel-2 99999999 $IBC_OSMO_CHANNEL_2_DENOM $IBC_OSMO_CHANNEL_2_DENOM true

    # Relax the send quota threshold (this will reset the flow)
    printf "\n>>> Updating rate limit...\n"
    submit_proposal_and_vote update-rate-limit update_uosmo.json
    sleep 30

    # Try the two transfers again, this time the second one should succeed
    check_transfer_status OSMO STRIDE channel-0 channel-2 99999999 uosmo $IBC_OSMO_CHANNEL_2_DENOM true
    check_transfer_status OSMO STRIDE channel-0 channel-2 2 uosmo $IBC_OSMO_CHANNEL_2_DENOM true 

    # Confirm balance was updated appropriately
    end_osmo_balance=$(get_balance OSMO uosmo)
    end_stride_balance=$(get_balance STRIDE $IBC_OSMO_CHANNEL_2_DENOM)

    expected_stride_balance=$((start_stride_balance+100000001))
    expected_osmo_balance=$((start_osmo_balance-100000001))

    print_expectation $expected_stride_balance $end_stride_balance "Balance on Stride" 
    print_expectation $expected_osmo_balance $end_osmo_balance "Balance on Osmo" 
}
