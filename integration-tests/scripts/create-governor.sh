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

create_governor() {
    echo "Creating governor..."
    pub_key=$($BINARY tendermint show-validator)
    $BINARY tx staking create-validator \
        --amount ${VALIDATOR_STAKE}${DENOM} \
        --pubkey=$pub_key \
        --commission-rate="0.10" \
        --commission-max-rate="0.20" \
        --commission-max-change-rate="0.01" \
        --min-self-delegation="1" \
        --from ${VALIDATOR_NAME} -y
}

main() {
    wait_for_startup
    add_keys
    create_governor
    echo "Done"
}

main >> governor.log 2>&1 &