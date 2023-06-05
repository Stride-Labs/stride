#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh

HERMES_LOGS=${LOGS}/hermes.log
HERMES_CONFIG=$STATE/hermes/config.toml
TMP_MNEMONICS=mnemonic.txt 

mkdir -p $STATE/hermes
chmod -R 777 $STATE/hermes
cp ${DOCKERNET_HOME}/config/hermes_config.toml $HERMES_CONFIG

HERMES_CMD="$DOCKERNET_HOME/../build/hermes/release/hermes --config $HERMES_CONFIG"

added_stride_account="false"

for chain in ${HOST_CHAINS[@]}; do
    chain_id=$(GET_VAR_VALUE     ${chain}_CHAIN_ID)
    relayer_exec=$(GET_VAR_VALUE RELAYER_${chain}_EXEC)
    chain_name=$(printf "$chain" | awk '{ print tolower($0) }')
    account_name=$(GET_VAR_VALUE RELAYER_${chain}_ACCT)
    mnemonic=$(GET_VAR_VALUE     RELAYER_${chain}_MNEMONIC)
    coin_type=$(GET_VAR_VALUE    ${chain}_COIN_TYPE)

    relayer_logs=${LOGS}/relayer-${chain_name}.log
    relayer_config=$STATE/relayer-${chain_name}/config

    mkdir -p $relayer_config
    chmod -R 777 $STATE/relayer-${chain_name}
    cp ${DOCKERNET_HOME}/config/relayer_config.yaml $relayer_config/config.yaml

    printf "STRIDE <> $chain - Adding relayer keys..."
    sudo chmod -R 777 /home/stride/stride/dockernet/state
    $relayer_exec rly keys restore stride $RELAYER_STRIDE_ACCT "$mnemonic" >> $relayer_logs 2>&1
    sudo chmod -R 777 /home/stride/stride/dockernet/state
    $relayer_exec rly keys restore $chain_name $account_name "$mnemonic" --coin-type $coin_type >> $relayer_logs 2>&1
    echo "Done restoring relayer keys"

    printf "STRIDE <> $chain - Creating client, connection, and transfer channel..." | tee -a $relayer_logs
    sudo rm -f /home/stride/stride/dockernet/state/relayer-gaia/config/config.lock
    sudo chmod -R 777 /home/stride/stride/dockernet/state
    $relayer_exec rly transact link stride-${chain_name} >> $relayer_logs 2>&1
    echo "Done."

    printf "STRIDE <> $chain - Adding hermes host key..."
    echo "$mnemonic" > $TMP_MNEMONICS
    sudo chmod -R 777 /home/stride/stride/dockernet/state
    $HERMES_CMD keys add --key-name $account_name --chain $chain_id --mnemonic-file $TMP_MNEMONICS --overwrite
    echo "Done"

    if [[ "$added_stride_account" == "false" ]]; then 
        printf "STRIDE <> $chain - Adding hermes Stride key..."
        echo "$mnemonic" > $TMP_MNEMONICS
        sudo chmod -R 777 /home/stride/stride/dockernet/state
        $HERMES_CMD keys add --key-name $RELAYER_STRIDE_ACCT --chain $STRIDE_CHAIN_ID --mnemonic-file $TMP_MNEMONICS --overwrite
        echo "Done"

        added_stride_account="true"
    fi

    # $DOCKER_COMPOSE up -d relayer-${chain_name}
    # $DOCKER_COMPOSE logs -f relayer-${chain_name} | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $relayer_logs 2>&1 &
done

$DOCKER_COMPOSE up -d hermes 
$DOCKER_COMPOSE logs -f hermes | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 &

rm $TMP_MNEMONICS