CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../../config.sh
source ${CURRENT_DIR}/common.sh

test_tx_reset_atom_from_gaia_to_stride() {
    print_header "TESTING TX QUOTA RESET - UNIDIRECTIONAL FLOW - ATOM FROM GAIA -> STRIDE"

    wait_until_epoch_end

    start_gaia_balance=$(get_balance GAIA uatom)
    start_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)
    start_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM channel-0)

    # Transfer successfully
    check_transfer_status GAIA STRIDE channel-0 channel-0 80000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true

    # Attempt to transfer but should fail because it gets rate limited
    check_transfer_status GAIA STRIDE channel-0 channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM false 

    # Reset the rate limit manually
    printf "\n>>> Resetting rate limit...\n"
    submit_proposal_and_vote reset_uatom.json
    sleep 30

    # Then successfully transfer after the reset 
    check_transfer_status GAIA STRIDE channel-0 channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true 

    # Channel value should go up since the ibc denom is minted
    wait_until_epoch_end
    expected_channel_value=$((start_channel_value+120000000))
    end_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM channel-0)

    print_expectation $expected_channel_value $end_channel_value "Channel Value" 

    # Confirm balance was updated appropriately
    end_gaia_balance=$(get_balance GAIA uatom)
    end_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)

    expected_stride_balance=$((start_stride_balance+120000000))
    expected_gaia_balance=$((start_gaia_balance-120000000))

    print_expectation $expected_stride_balance $end_stride_balance "Balance on Stride" 
    print_expectation $expected_gaia_balance $end_gaia_balance "Balance on Gaia" 
}

test_tx_reset_atom_from_stride_to_gaia() {
    print_header "TESTING TX QUOTA RESET - UNIDIRECTIONAL FLOW - ATOM FROM STRIDE -> GAIA"

    wait_until_epoch_end

    start_gaia_balance=$(get_balance GAIA uatom)
    start_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)
    start_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM channel-0)

    # Transfer successfully
    check_transfer_status STRIDE GAIA channel-0 channel-0 80000000 $IBC_GAIA_CHANNEL_0_DENOM $IBC_GAIA_CHANNEL_0_DENOM true

    # Attempt to transfer but should fail because it gets rate limited
    check_transfer_status STRIDE GAIA channel-0 channel-0 40000000 $IBC_GAIA_CHANNEL_0_DENOM $IBC_GAIA_CHANNEL_0_DENOM false 

    # Reset the rate limit manually
    printf "\n>>> Resetting rate limit...\n"
    submit_proposal_and_vote reset_uatom.json
    sleep 30

    # Then successfully transfer after the reset 
    check_transfer_status STRIDE GAIA channel-0 channel-0 40000000 $IBC_GAIA_CHANNEL_0_DENOM $IBC_GAIA_CHANNEL_0_DENOM true 

    # Channel value should go down since the ibc denom will be burned
    wait_until_epoch_end
    expected_channel_value=$((start_channel_value-120000000))
    end_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM channel-0)

    print_expectation $expected_channel_value $end_channel_value "Channel Value" 

    # Wait a few seconds for the ack error to refund the failed tokens on gaia
    sleep 15

    # Confirm balance was updated appropriately
    end_gaia_balance=$(get_balance GAIA uatom)
    end_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)

    expected_stride_balance=$((start_stride_balance-120000000))
    expected_gaia_balance=$((start_gaia_balance+120000000))

    print_expectation $expected_stride_balance $end_stride_balance "Balance on Stride" 
    print_expectation $expected_gaia_balance $end_gaia_balance "Balance on Gaia" 
}
