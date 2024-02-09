#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh


for chain in ${HOST_CHAINS[@]}; do
    chain_id=$(GET_VAR_VALUE     ${chain}_CHAIN_ID)
    relayer_exec=$(GET_VAR_VALUE RELAYER_${chain}_EXEC)
    chain_name=$(printf "$chain" | awk '{ print tolower($0) }')
    account_name=$(GET_VAR_VALUE RELAYER_${chain}_ACCT)
    mnemonic=$(GET_VAR_VALUE     RELAYER_${chain}_MNEMONIC)

    relayer_logs=${LOGS}/relayer-${chain_name}.log
    relayer_config=$STATE/relayer-${chain_name}/config

    mkdir -p $relayer_config
    chmod -R 777 $STATE/relayer-${chain_name}
    cp ${DOCKERNET_HOME}/config/relayer_config_stride.yaml $relayer_config/config.yaml

    echo "STRIDE <> $chain - Adding relayer keys..."
    printf "STRIDE relayer key... "
    relayer_address_1=$($relayer_exec rly keys restore stride $RELAYER_STRIDE_ACCT "$mnemonic")
    echo $relayer_address_1 >> $relayer_logs 2>&1
    echo $relayer_address_1

    printf  "$chain_name relayer key... "
    relayer_address_2=$($relayer_exec rly keys restore $chain_name $account_name "$mnemonic")
    echo $relayer_address_2 >> $relayer_logs 2>&1
    echo $relayer_address_2

    echo "Done"

    printf "STRIDE <> $chain - Creating client, connection, and transfer channel... " | tee -a $relayer_logs
    $relayer_exec rly transact link stride-${chain_name} >> $relayer_logs 2>&1
    echo "Done"

    $DOCKER_COMPOSE up -d relayer-${chain_name}
    SAVE_DOCKER_LOGS relayer-${chain_name} $relayer_logs
done

