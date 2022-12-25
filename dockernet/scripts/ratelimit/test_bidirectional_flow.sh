CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../../config.sh
source ${CURRENT_DIR}/common.sh

test_bidirectional() {
    print_header "TESTING BIDIRECTIONAL FLOW - ATOM"

    wait_until_epoch_end

    start_gaia_balance=$(get_balance GAIA uatom)
    start_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)
    start_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM channel-0)

    # Continuously transfer back and forth (the rate limit should never get hit)
    check_transfer_status GAIA STRIDE channel-0 channel-0 40000000 uatom                     $IBC_GAIA_CHANNEL_0_DENOM true 
    check_transfer_status STRIDE GAIA channel-0 channel-0 40000000 $IBC_GAIA_CHANNEL_0_DENOM $IBC_GAIA_CHANNEL_0_DENOM true

    check_transfer_status GAIA STRIDE channel-0 channel-0 40000000 uatom                     $IBC_GAIA_CHANNEL_0_DENOM true 
    check_transfer_status STRIDE GAIA channel-0 channel-0 40000000 $IBC_GAIA_CHANNEL_0_DENOM $IBC_GAIA_CHANNEL_0_DENOM true

    check_transfer_status GAIA STRIDE channel-0 channel-0 40000000 uatom                     $IBC_GAIA_CHANNEL_0_DENOM true 

    wait_until_epoch_end 

    check_transfer_status STRIDE GAIA channel-0 channel-0 40000000 $IBC_GAIA_CHANNEL_0_DENOM $IBC_GAIA_CHANNEL_0_DENOM true

    check_transfer_status GAIA STRIDE channel-0 channel-0 40000000 uatom                     $IBC_GAIA_CHANNEL_0_DENOM true 
    check_transfer_status STRIDE GAIA channel-0 channel-0 40000000 $IBC_GAIA_CHANNEL_0_DENOM $IBC_GAIA_CHANNEL_0_DENOM true

    check_transfer_status GAIA STRIDE channel-0 channel-0 40000000 uatom                     $IBC_GAIA_CHANNEL_0_DENOM true 
    check_transfer_status STRIDE GAIA channel-0 channel-0 40000000 $IBC_GAIA_CHANNEL_0_DENOM $IBC_GAIA_CHANNEL_0_DENOM true

    # Wait for the channel value to reset
    wait_until_epoch_end

    # Balances and channel value should be unchanged
    end_gaia_balance=$(get_balance GAIA uatom)
    end_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)
    end_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM channel-0)

    print_expectation $start_channel_value $end_channel_value "Channel Value" 
    print_expectation $start_stride_balance $end_stride_balance "Balance on Stride" 
    print_expectation $start_gaia_balance $end_gaia_balance "Balance on Gaia" 
}
