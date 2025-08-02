# #!/bin/bash

set -eu
source scripts/config.sh

wait_for_startup() {
    echo "Waiting for node to start..."

    while true; do
        if $BINARY status &> /dev/null; then
            latest_height=$($BINARY status | jq -r 'if .SyncInfo then .SyncInfo.latest_block_height else .sync_info.latest_block_height end // "0"')
            if [[ "$latest_height" -gt "0" ]]; then
                echo "Node synced"
                break
            fi
        fi

        echo "Node still syncing..."
        sleep 10
    done
}

add_keys() {
    echo "Adding validator keys..."

    validator_config=$(jq -r '.validators[$index]' --argjson index "$POD_INDEX" ${KEYS_FILE})
    mnemonic=$(echo $validator_config | jq -r '.mnemonic')
    name=$(echo $validator_config | jq -r '.name')

    if ! $BINARY keys show $name -a &> /dev/null ; then
        echo "$mnemonic" | $BINARY keys add $name --recover 
    fi
}

create_validator() {
    echo "Creating validator..."
    pub_key=$($BINARY tendermint show-validator)

    # For sdk 50, use validator.json file
    if $BINARY tx staking create-validator --help | grep -q validator.json; then 
        cat > validator.json << EOF
{
    "pubkey": $pub_key,
    "amount": "1000000000${DENOM}",
    "moniker": "${VALIDATOR_NAME}",
    "commission-rate": "0.10",
    "commission-max-rate": "0.20",
    "commission-max-change-rate": "0.01",
    "min-self-delegation": "1"
}
EOF
        $BINARY tx staking create-validator validator.json --from ${VALIDATOR_NAME} -y
    else 
        # For sdk 47, use cli command
        min_self_delegation=""
        if [[ $($BINARY tx staking create-validator --help | grep -c "min-self-delegation") -gt 0 ]]; then
            min_self_delegation='--min-self-delegation=1000000'
        fi

        $BINARY tx staking create-validator \
            --amount ${VALIDATOR_STAKE}${DENOM} \
            --pubkey=$pub_key \
            --commission-rate="0.10" \
            --commission-max-rate="0.20" \
            --commission-max-change-rate="0.01" \
            $min_self_delegation \
            --fees 300000$DENOM \
            --gas auto \
            --gas-adjustment 1.2 \
            --from ${VALIDATOR_NAME} -y 
    fi
}

main() {
    echo "Adding validator..."
    wait_for_startup
    add_keys
    create_validator
    echo "Done"
}

main >> logs/startup.log 2>&1 &