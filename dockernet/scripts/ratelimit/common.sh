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
    sleep_time=$((seconds_til_epoch_start+10))

    echo ">>> Sleeping $sleep_time seconds until start of epoch..."
    sleep $sleep_time
}

get_flow_amount() {
    denom=$1
    channel=$2
    flow_type=$3
    $STRIDE_MAIN_CMD q ratelimit list-rate-limits | grep $denom -B 6 | grep $channel -B 5 | grep $flow_type | awk '{printf $2}' | tr -d '"'
}

get_channel_value() {
    denom=$1
    channel=$2
    $STRIDE_MAIN_CMD q ratelimit list-rate-limits | grep $denom -B 6 | grep $channel -B 5 | grep "channel_value" | awk '{printf $2}' | tr -d '"'
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
        transfer_delay=3
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

setup_juno_osmo_channel() {
    relayer_exec="$DOCKER_COMPOSE run --rm relayer-juno-osmo"
    path="juno-osmo"

    relayer_logs=$LOGS/relayer-${path}.log
    relayer_config=$STATE/relayer-${path}/config

    mkdir -p $relayer_config
    cp ${SCRIPT_DIR}/config/relayer_config_juno_osmo.yaml $relayer_config/config.yaml

    printf "JUNO <> OSMO - Adding relayer keys..."
    RELAYER_JUNO_MNEMONIC="awkward remind blanket around naive senior sock pigeon portion umbrella edit scheme middle supreme agent indoor duty sock conduct market ethics exchange phrase mirror"
    RELAYER_OSMO_MNEMONIC="solution simple collect warrior neither grain ethics dust guard high base hamster sail science valley organ mistake soon letter garden october morning correct hidden"
    JUNO_RELAYER_ADDRESS=$($relayer_exec rly keys restore juno juno-osmo-rly1 "$RELAYER_JUNO_MNEMONIC") 
    OSMO_RELAYER_ADDRESS=$($relayer_exec rly keys restore osmo juno-osmo-rly2 "$RELAYER_OSMO_MNEMONIC")
    echo "Done"

    printf "JUNO <> OSMO - Funding Relayers...\n" 
    $JUNO_MAIN_CMD tx bank send ${JUNO_VAL_PREFIX}1 $JUNO_RELAYER_ADDRESS 1000000ujuno --from ${JUNO_VAL_PREFIX}1 -y | TRIM_TX
    $OSMO_MAIN_CMD tx bank send ${OSMO_VAL_PREFIX}1 $OSMO_RELAYER_ADDRESS 1000000uosmo --from ${OSMO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3
    echo "Done"

    printf "JUNO <> OSMO - Creating client, connection, and transfer channel..." | tee -a $relayer_logs
    $relayer_exec rly transact link $path >> $relayer_logs 2>&1
    echo "Done"

    $DOCKER_COMPOSE up -d relayer-${path}
    $DOCKER_COMPOSE logs -f relayer-${path} | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $relayer_logs 2>&1 &
}

add_rate_limits() {
    printf "$BLUE[ADDING RATE LIMITS]$NC\n"
    printf "$BLUE---------------------------------------------------------------------$NC\n"

    echo "Transfering for channel value..."
    echo ">>> uatom"
    $GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
    echo ">>> ujuno"
    $JUNO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}ujuno --from ${JUNO_VAL_PREFIX}1 -y | TRIM_TX
    echo ">>> uosmo"
    $OSMO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}uosmo --from ${OSMO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3
    
    echo ">>> traveler-ujuno"
    juno_on_osmo='ibc/448C1061CE97D86CC5E86374CD914870FB8EBA16C58661B5F1D3F46729A2422D'
    $JUNO_MAIN_CMD tx ibc-transfer transfer transfer channel-5 $(OSMO_ADDRESS) ${INITIAL_CHANNEL_VALUE}ujuno --from ${JUNO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 10
    $OSMO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}${juno_on_osmo} --from ${OSMO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    printf "\nLiquid staking juno...\n"
    $STRIDE_MAIN_CMD tx stakeibc liquid-stake ${INITIAL_CHANNEL_VALUE} ujuno --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 5

    printf "\nAdding rate limits...\n"
    echo ">>> ustrd on Stride <> Osmo Channel:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit $STRD_DENOM               channel-2 10 10 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> uatom on Stride <> Gaia Channel:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit $IBC_GAIA_CHANNEL_0_DENOM channel-0 10 10 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> ujuno on Stride <> Juno Channel:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit $IBC_JUNO_CHANNEL_1_DENOM channel-1 10 10 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> uosmo on Stride <> Osmo Channel:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit $IBC_OSMO_CHANNEL_2_DENOM channel-2 10 10 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> stujuno on Stride <> Osmo Channel:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit stujuno                   channel-2 10 10 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> traveler-ujuno on Stride <> Juno Channel:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit $TRAVELER_JUNO            channel-1 10 10 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> traveler-ujuno on Stride <> Osmo Channel:"
    $STRIDE_MAIN_CMD tx ratelimit add-rate-limit $TRAVELER_JUNO            channel-2 10 10 1 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3
}
