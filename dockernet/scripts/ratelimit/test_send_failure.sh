CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../../config.sh
source ${CURRENT_DIR}/common.sh

test_ack_timeout() {
    print_header "TESTING ACK TIMEOUT"

    # Capture the start flow
    start_flow=$(get_flow_amount ustrd channel-2 outflow)
    echo "Initial outflow for ustrd: $start_flow"

     # Send the transfer with a sub-second timeout timestamp to force a timeout
    echo "Transferring..."
    timeout=1000 # in nanoseconds
    $STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-2 $(OSMO_ADDRESS) 10000000ustrd \
        --from val1 -y --packet-timeout-timestamp $timeout | TRIM_TX
    sleep 15

    # Capture the outflow
    end_flow=$(get_flow_amount ustrd channel-2 outflow)
    echo "End outflow for ustrd: $end_flow"

    # The outflow should have been reset
    actual_flow_change=$((end_flow-start_flow))
    print_expectation 0 $actual_flow_change "Flow Change"
}

test_ack_failure() {
    print_header "TESTING ACK FAILURE"

    # Capture the start flow
    start_flow=$(get_flow_amount ustrd channel-2 outflow)
    echo "Initial outflow for ustrd: $start_flow"

    # Send the transfer with a stride destination address instead of an osmo address
    # to cause the transfer to fail
    echo "Transferring..."
    invalid_address=$(STRIDE_ADDRESS)
    $STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-2 $invalid_address 10000000ustrd --from val1 -y | TRIM_TX
    sleep 15

    # Capture the outflow
    end_flow=$(get_flow_amount ustrd channel-2 outflow)
    echo "End outflow for ustrd: $end_flow"

    # The outflow should have been reset
    actual_flow_change=$((end_flow-start_flow))
    print_expectation 0 $actual_flow_change "Flow Change"
}

test_send_failures() {
    wait_until_epoch_end

    test_ack_timeout
    test_ack_failure
}
