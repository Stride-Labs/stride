# #!/bin/bash

set -eu
source scripts/config.sh

wait_for_startup() {
    echo "Waiting for node to start..."
    while ! (($BINARY status &> /dev/null) && [[ "$($BINARY status | jq -r '.SyncInfo.latest_block_height')" -gt "0" ]]); do 
        echo "Node still syncing..."
        sleep 10
    done
    echo "Node synced"
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
    min_self_delegation=""
    if [[ $($BINARY tx staking create-validator --help | grep -c "min-self-delegation") -gt 0 ]]; then
        min_self_delegation='--min-self-delegation=1000000'
    fi

    pub_key=$($BINARY tendermint show-validator)
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
}

main() {
    echo "Creating validator..."
    wait_for_startup
    add_keys
    create_validator
    echo "Done"
}

main >> logs/startup.log 2>&1 &