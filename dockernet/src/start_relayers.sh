#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh

for chain_id in ${HOST_CHAINS[@]}; do
    relayer_exec=$(GET_VAR_VALUE RELAYER_${chain_id}_EXEC)
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')
    account_name=$(GET_VAR_VALUE RELAYER_${chain_id}_ACCT)
    mnemonic=$(GET_VAR_VALUE     RELAYER_${chain_id}_MNEMONIC)

    relayer_logs=${LOGS}/relayer-${chain_name}.log
    relayer_config=$STATE/relayer-${chain_name}/config

    mkdir -p $relayer_config
    cp ${SCRIPT_DIR}/config/relayer_config.yaml $relayer_config/config.yaml

    printf "STRIDE <> $chain_id - Adding relayer keys..."
    $relayer_exec rly keys restore stride $RELAYER_STRIDE_ACCT "$mnemonic" >> $relayer_logs 2>&1
    $relayer_exec rly keys restore $chain_name $account_name "$mnemonic" >> $relayer_logs 2>&1
    echo "Done"

    printf "STRIDE <> $chain_id - Creating client, connection, and transfer channel..." | tee -a $relayer_logs
    $relayer_exec rly transact link stride-${chain_name} >> $relayer_logs 2>&1
    echo "Done"

    $DOCKER_COMPOSE up -d relayer-${chain_name}
    $DOCKER_COMPOSE logs -f relayer-${chain_name} | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $relayer_logs 2>&1 &
done
