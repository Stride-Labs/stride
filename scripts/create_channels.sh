#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

HOST_CHAINS="$@"

for chain_id in ${HOST_CHAINS[@]}; do
    relayer_exec=$(GET_VAR_VALUE RELAYER_${chain_id}_EXEC)
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')

    printf "Creating client, connection, and transfer channel STRIDE <> $chain_id..." | tee -a ${LOGS}/relayer-${chain_name}.log
    $relayer_exec rly transact link stride-${chain_name} >> ${LOGS}/relayer-${chain_name}.log 2>&1

    echo "Done"
done
