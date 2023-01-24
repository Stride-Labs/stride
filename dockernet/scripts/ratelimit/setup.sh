CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../../config.sh
source ${CURRENT_DIR}/common.sh

setup_juno_osmo_channel() {
    print_header "CREATING JUNO <> OSMO CHANNEL"

    relayer_exec="$DOCKER_COMPOSE run --rm relayer-juno-osmo"
    path="juno-osmo"

    relayer_logs=$LOGS/relayer-${path}.log
    relayer_config=$STATE/relayer-${path}/config

    mkdir -p $relayer_config
    cp ${DOCKERNET_HOME}/config/relayer_config_juno_osmo.yaml $relayer_config/config.yaml

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

setup_channel_value() {
    print_header "INITIALIZING CHANNEL VALUE"

    # IBC Transfer
    echo "Transfering for channel value..."
    echo ">>> uatom"
    $GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> ujuno" # second transfer is for stujuno
    $JUNO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}ujuno --from ${JUNO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3
    $JUNO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}ujuno --from ${JUNO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    echo ">>> uosmo"
    $OSMO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}uosmo --from ${OSMO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 10
    
    echo ">>> traveler-ujuno"
    juno_on_osmo='ibc/448C1061CE97D86CC5E86374CD914870FB8EBA16C58661B5F1D3F46729A2422D'
    $JUNO_MAIN_CMD tx ibc-transfer transfer transfer channel-5 $(OSMO_ADDRESS) ${INITIAL_CHANNEL_VALUE}ujuno --from ${JUNO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 10
    $OSMO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) ${INITIAL_CHANNEL_VALUE}${juno_on_osmo} --from ${OSMO_VAL_PREFIX}1 -y | TRIM_TX
    sleep 3

    # Liquid Stake
    printf "\nLiquid staking juno...\n"
    $STRIDE_MAIN_CMD tx stakeibc liquid-stake ${INITIAL_CHANNEL_VALUE} ujuno --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
    sleep 5
}

setup_rate_limits() {
    print_header "ADDING RATE LIMITS"
        
    # ustrd channel-2
    echo "ustrd on Stride <> Osmo Channel:"
    submit_proposal_and_vote add-rate-limit add_ustrd.json
    sleep 10

    # ibc/uatom channel-0
    echo "uatom on Stride <> Gaia Channel:"
    submit_proposal_and_vote add-rate-limit add_uatom.json
    sleep 10

    # ibc/ujuno channel-1
    echo "ujuno on Stride <> Juno Channel:"
    submit_proposal_and_vote add-rate-limit add_ujuno.json
    sleep 10

    # ibc/uosmo channel-2
    echo "uosmo on Stride <> Osmo Channel:"
    submit_proposal_and_vote add-rate-limit add_uosmo.json
    sleep 10

    # stujuno channel-2
    echo "stujuno on Stride <> Osmo Channel:"
    submit_proposal_and_vote add-rate-limit add_stujuno.json
    sleep 10

    # traveler juno channel-1
    echo "traveler-ujuno on Stride <> Juno Channel:"
    submit_proposal_and_vote add-rate-limit add_traveler_ujuno_on_juno.json
    sleep 10

    echo "traveler-ujuno on Stride <> Osmo Channel:"
    # traveler juno channel-2
    submit_proposal_and_vote add-rate-limit add_traveler_ujuno_on_osmo.json
    sleep 40

    # Confirm all rate limits were added
    num_rate_limits=$($STRIDE_MAIN_CMD q ratelimit list-rate-limits | grep path | wc -l | xargs)
    if [[ "$num_rate_limits" != "7" ]]; then 
        echo "ERROR: Not all rate limits were added. Exiting."
        exit 1
    fi

    # Confirm there are 4 rate limits on osmo (this is to test out the rate-limits-by-chain query)
    num_rate_limits=$($STRIDE_MAIN_CMD q ratelimit rate-limits-by-chain OSMO | grep path | wc -l | xargs)
    if [[ "$num_rate_limits" != "4" ]]; then 
        echo "ERROR: OSMO should have 4 rate limits (it had: $num_rate_limits)"
        exit 1
    fi
}
