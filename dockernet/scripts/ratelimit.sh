SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

PURPLE='\033[0;35m'
BOLD="\033[1m"
BLUE='\033[1;34m'
ITALIC="\033[3m"
NC="\033[0m"

INITIAL_CHANNEL_VALUE=1000000000

get_current_time_unix() {
    date +%s000 | rev | cut -c3- | rev
}

checkmark() {
    printf "$BLUE\xE2\x9C\x94$NC\n"
}

xmark() {
    printf "$PURPLE\xE2\x9C\x97$NC\n" 
}

get_test_indicator() {
    expected=$1
    actual=$2
    if [[ "$expected" == "$actual" ]]; then
        echo $(checkmark)
    else 
        echo $(xmark)
    fi
}

print_expectation() {
    expected=$1
    actual=$2
    description=$3
    
    indicator=$(get_test_indicator $expected $actual)
    printf "\n$indicator $indicator $indicator Expected $description: $expected | Actual $description: $actual $indicator $indicator $indicator\n"
}

wait_until_unix_time() {
    target_time=$1
    while true; do
        current_time=$(get_current_time_unix)
        if (( current_time > target_time )); then 
            break
        fi 
        sleep 1
    done
}

wait_until_epoch_end() {
    seconds_til_epoch_start=$($STRIDE_MAIN_CMD q epochs seconds-remaining hour)
    sleep_time=$((seconds_til_epoch_start+10))

    echo ">>> Sleeping $sleep_time seconds until start of epoch..."
    sleep $sleep_time
}

get_flow_amount() {
    denom=$1
    flow_type=$2
    $STRIDE_MAIN_CMD q ratelimit list-rate-limits | grep $denom -B 6 | grep $flow_type | awk '{printf $2}' | tr -d '"'
}

get_channel_value() {
    denom=$1
    $STRIDE_MAIN_CMD q ratelimit list-rate-limits | grep $denom -B 6 | grep "channel_value" | awk '{printf $2}' | tr -d '"'
}

get_balance() {
    chain=$1
    denom=$2

    cmd=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
    address=$(${chain}_ADDRESS)

    $cmd q bank balances $address --denom $denom | grep amount | awk '{printf $2}' | tr -d '"'
}

check_transfer_status() {
    src_chain=$1 
    dst_chain=$2
    src_channel=$3 
    amount=$4
    base_denom=$5
    rate_limit_denom=$6
    success=$7

    cmd=$(GET_VAR_VALUE        ${src_chain}_MAIN_CMD)
    val_prefix=$(GET_VAR_VALUE ${src_chain}_VAL_PREFIX)
    destination_address=$(${dst_chain}_ADDRESS)

    # Determine packet direction
    if [[ "$src_chain" != "STRIDE" ]]; then
        transfer_description="from $src_chain to STRIDE"
        flow_type="inflow"
        transfer_delay=10
        transfer_denom=$base_denom
    else 
        description="from STRIDE to $dst_chain"
        flow_type="outflow"
        transfer_delay=3
        transfer_denom=$rate_limit_denom
    fi

    # Determine expectation
    if [[ "$success" == "true" ]]; then 
        expected_flow_change=$amount
        success_description="SHOULD SUCCEED"
    else 
        expected_flow_change=0
        success_description="SHOULD FAIL"
    fi

    printf "\n>>> Transferring ${amount}${base_denom} $transfer_description - $success_description\n"

    # Capture the inflow
    start_flow=$(get_flow_amount $rate_limit_denom $flow_type)
    echo "Initial $flow_type for $base_denom: $start_flow"

    # Send the transfer
    echo "Transferring..."
    $cmd tx ibc-transfer transfer transfer $src_channel $destination_address ${amount}${transfer_denom} --from ${val_prefix}1 -y | TRIM_TX
    sleep $transfer_delay

    # Capture the outflow
    end_flow=$(get_flow_amount $rate_limit_denom $flow_type)
    echo "End $flow_type for $base_denom: $end_flow"

    # Determine if the flow change was a success
    actual_flow_change=$((end_flow-start_flow))
    print_expectation $expected_flow_change $actual_flow_change "Flow Change"
}

setup() {
    printf "$BLUE[SETTING UP TESTS]$NC\n"
    printf "$BLUE---------------------------------------------------------------------$NC\n"

    echo "Transfering..."
    echo ">>> uatom"
    $GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
    echo ">>> ujuno"
    $JUNO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}ujuno --from ${JUNO_VAL_PREFIX}1 -y | TRIM_TX
    echo ">>> usomo"
    $OSMO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}uosmo --from ${OSMO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    printf "\nLiquid staking juno...\n"
    $STRIDE_MAIN_CMD tx stakeibc liquid-stake ${INITIAL_CHANNEL_VALUE} ujuno --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 5

    printf "\nAdding rate limits...\n"
    echo ">>> uatom:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit $IBC_GAIA_CHANNEL_0_DENOM channel-0 10 10 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> stujuno:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit stujuno                   channel-1 20 20 2 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> uosmo:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit $IBC_OSMO_CHANNEL_2_DENOM channel-2 50 50 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> ustrd:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit $STRD_DENOM               channel-2 10 10 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3
}

test_epoch_reset_atom_from_gaia_to_stride() {
    printf "\n\n$BLUE[TESTING EPOCHLY QUOTA RESET - UNIDIRECTIONAL FLOW - ATOM FROM GAIA -> STRIDE]$NC\n"
    printf "$BLUE---------------------------------------------------------------------$NC\n"

    wait_until_epoch_end

    start_gaia_balance=$(get_balance GAIA uatom)
    start_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)
    start_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM)

    # Transfer 2 times successfully
    check_transfer_status GAIA STRIDE channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true
    check_transfer_status GAIA STRIDE channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true

    # Attempt to transfer but should fail because it gets rate limited
    check_transfer_status GAIA STRIDE channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM false 

    # Wait for rate limit to reset and then transfer successfully again
    wait_until_epoch_end
    check_transfer_status GAIA STRIDE channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true 

    # Channel value should go up since the ibc denom is minted
    expected_channel_value=$((start_channel_value+80000000))
    end_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM)

    print_expectation $expected_channel_value $end_channel_value "Channel Value" 

    # Confirm balance was updated appropriately
    end_gaia_balance=$(get_balance GAIA uatom)
    end_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)

    expected_stride_balance=$((start_stride_balance+120000000))
    expected_gaia_balance=$((start_gaia_balance-120000000))

    print_expectation $expected_stride_balance $end_stride_balance "Balance on Stride" 
    print_expectation $expected_gaia_balance $end_gaia_balance "Balance on Gaia" 
}

test_epoch_reset_atom_from_stride_to_gaia() {
    printf "\n\n$BLUE[TESTING EPOCHLY QUOTA RESET - UNIDIRECTIONAL FLOW - ATOM FROM STRIDE -> GAIA]$NC\n"
    printf "$BLUE---------------------------------------------------------------------$NC\n"

    wait_until_epoch_end

    start_gaia_balance=$(get_balance GAIA uatom)
    start_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)
    start_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM)

    # Transfer 2 times successfully
    check_transfer_status STRIDE GAIA channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true
    check_transfer_status STRIDE GAIA channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true

    # Attempt to transfer but should fail because it gets rate limited
    check_transfer_status STRIDE GAIA channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM false 

    # Wait for rate limit to reset and then transfer successfully again
    wait_until_epoch_end 
    check_transfer_status STRIDE GAIA channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true 

    # Channel value should go down since the ibc denom will be burned
    expected_channel_value=$((start_channel_value-80000000))
    end_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM)

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

test_bidirectional_atom() {
    printf "\n\n$BLUE[TESTING BIDIRECTIONAL FLOW - ATOM]$NC\n"
    printf "$BLUE---------------------------------------------------------------------$NC\n"

    wait_until_epoch_end

    start_gaia_balance=$(get_balance GAIA uatom)
    start_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)
    start_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM)

    # Continuously transfer back and forth (the rate limit should never get hit)
    check_transfer_status GAIA STRIDE channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true 
    check_transfer_status STRIDE GAIA channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true

    check_transfer_status GAIA STRIDE channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true 
    check_transfer_status STRIDE GAIA channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true

    check_transfer_status GAIA STRIDE channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true 

    wait_until_epoch_end 

    check_transfer_status STRIDE GAIA channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true

    check_transfer_status GAIA STRIDE channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true 
    check_transfer_status STRIDE GAIA channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true

    check_transfer_status GAIA STRIDE channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true 
    check_transfer_status STRIDE GAIA channel-0 40000000 uatom $IBC_GAIA_CHANNEL_0_DENOM true

    # Wait for the channel value to reset
    wait_until_epoch_end

    # Balances and channel value should be unchanged
    end_gaia_balance=$(get_balance GAIA uatom)
    end_stride_balance=$(get_balance STRIDE $IBC_GAIA_CHANNEL_0_DENOM)
    end_channel_value=$(get_channel_value $IBC_GAIA_CHANNEL_0_DENOM)

    print_expectation $start_channel_value $end_channel_value "Channel Value" 
    print_expectation $start_stride_balance $end_stride_balance "Balance on Stride" 
    print_expectation $start_gaia_balance $end_gaia_balance "Balance on Gaia" 
}

setup
test_epoch_reset_atom_from_gaia_to_stride
test_epoch_reset_atom_from_stride_to_gaia
test_bidirectional_atom
