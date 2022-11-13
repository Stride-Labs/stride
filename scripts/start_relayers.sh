#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh
chmod -R a+rwx $STATE/relayer

HOST_CHAINS=($@)

for i in ${!HOST_CHAINS[@]}; do
    chain_id=${HOST_CHAINS[$i]}
    relayer_exec=$(GET_VAR_VALUE RELAYER_${chain_id}_EXEC)
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')
    relayer_logs=${LOGS}/relayer-${chain_name}.log

    printf "STRIDE <> $chain_id Creating client, connection, and transfer channel..." | tee -a $relayer_logs
    $relayer_exec rly transact link stride-${chain_name} >> $relayer_logs 2>&1
    echo "Done"

    docker-compose up -d relayer-${chain_name}
    docker-compose logs -f relayer-${chain_name} | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $relayer_logs 2>&1 &

    bash $SCRIPT_DIR/register_host.sh ${HOST_CHAINS[$i]} $i
done
