CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../../config.sh

PURPLE='\033[0;35m'
BOLD="\033[1m"
BLUE='\033[1;34m'
ITALIC="\033[3m"
NC="\033[0m"

INITIAL_CHANNEL_VALUE=1000000000
TRAVELER_JUNO=ibc/CD369927BBCE5198E0DC0D1A341C2F1DE51B1228BFD0633430055A39F58D229C

checkmark() {
    printf "$BLUE\xE2\x9C\x94$NC\n"
}

xmark() {
    printf "$PURPLE\xE2\x9C\x97$NC\n" 
}

print_header() {
    header=$1
    printf "\n\n$BLUE[$header]$NC\n"
    printf "$BLUE----------------------------------------------------------------------------$NC\n"
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

wait_until_epoch_end() {
    seconds_til_epoch_start=$($STRIDE_MAIN_CMD q epochs seconds-remaining hour)
    sleep_time=$((seconds_til_epoch_start+5))

    echo ">>> Sleeping $sleep_time seconds until start of epoch..."
    sleep $sleep_time
}

get_flow_amount() {
    denom=$1
    channel=$2
    flow_type=$3
    $STRIDE_MAIN_CMD q ratelimit rate-limit $channel --denom=$denom | grep $flow_type | awk '{printf $2}' | tr -d '"'
}

get_channel_value() {
    denom=$1
    channel=$2
    $STRIDE_MAIN_CMD q ratelimit rate-limit $channel --denom=$denom | grep "channel_value" | awk '{printf $2}' | tr -d '"'
}

get_balance() {
    chain=$1
    denom=$2

    cmd=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
    address=$(${chain}_ADDRESS)

    $cmd q bank balances $address --denom $denom | grep amount | awk '{printf $2}' | tr -d '"'
}

get_last_proposal_id() {
    $STRIDE_MAIN_CMD q gov proposals | grep " id:" | tail -1 | awk '{printf $2}' | tr -d '"'
}

check_transfer_status() {
    src_chain=$1 
    dst_chain=$2
    transfer_channel=$3 
    rate_limit_channel=$4
    amount=$5
    transfer_denom=$6
    rate_limit_denom=$7
    success=$8

    cmd=$(GET_VAR_VALUE        ${src_chain}_MAIN_CMD)
    val_prefix=$(GET_VAR_VALUE ${src_chain}_VAL_PREFIX)
    destination_address=$(${dst_chain}_ADDRESS)

    # Determine packet direction
    if [[ "$src_chain" == "STRIDE" ]]; then
        transfer_description="from STRIDE to $dst_chain"
        flow_type="outflow"
        transfer_delay=4
    else 
        transfer_description="from $src_chain to STRIDE"
        flow_type="inflow"
        transfer_delay=10
    fi

    # Determine expectation
    if [[ "$success" == "true" ]]; then 
        expected_flow_change=$amount
        success_description="SHOULD SUCCEED"
    else 
        expected_flow_change=0
        success_description="SHOULD FAIL"
    fi

    printf "\n>>> Transferring ${amount}${transfer_denom} $transfer_description - $success_description\n"

    # Capture the inflow
    start_flow=$(get_flow_amount $rate_limit_denom $rate_limit_channel $flow_type)
    echo "Initial $flow_type for $transfer_denom: $start_flow"

    # Send the transfer
    echo "Transferring..."
    $cmd tx ibc-transfer transfer transfer $transfer_channel $destination_address ${amount}${transfer_denom} --from ${val_prefix}1 -y | TRIM_TX
    sleep $transfer_delay

    # Capture the outflow
    end_flow=$(get_flow_amount $rate_limit_denom $rate_limit_channel $flow_type)
    echo "End $flow_type for $transfer_denom: $end_flow"

    # Determine if the flow change was a success
    actual_flow_change=$((end_flow-start_flow))
    print_expectation $expected_flow_change $actual_flow_change "Flow Change"
}

submit_proposal_and_vote() {
    proposal_type=$1
    proposal_file=$2

    echo ">>> Submitting proposal for: $proposal_file"
    $STRIDE_MAIN_CMD tx gov submit-legacy-proposal $proposal_type ${CURRENT_DIR}/proposals/${proposal_file} --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    proposal_id=$(get_last_proposal_id)
    echo ">>> Voting on proposal $proposal_id"
    $STRIDE_MAIN_CMD tx gov vote $proposal_id yes --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    $STRIDE_MAIN_CMD tx gov vote $proposal_id yes --from ${STRIDE_VAL_PREFIX}2 -y | TRIM_TX
    $STRIDE_MAIN_CMD tx gov vote $proposal_id yes --from ${STRIDE_VAL_PREFIX}3 -y | TRIM_TX

    echo ""
}
