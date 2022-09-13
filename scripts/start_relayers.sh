#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

# Pass chain IDs as arguments
CHAINS="$@"
if [[ "$CHAINS" == "" ]]; then
    echo "ERROR: Please specify chain IDs that require a relayer connection"
    exit 1
fi

mkdir -p $STATE/relayer/config
mkdir -p $STATE/refresh-clients/config
cp ${SCRIPT_DIR}/config/relayer_config.yaml $STATE/relayer/config/config.yaml
cp ${SCRIPT_DIR}/config/relayer_config.yaml $STATE/refresh-clients/config/config.yaml 

echo "Adding Relayer keys"
$RELAYER_CMD keys restore stride $RELAYER_STRIDE_ACCT "$RELAYER_STRIDE_MNEMONIC" 
$REFRESH_CMD keys restore stride $RELAYER_STRIDE_ACCT "$REFRESH_STRIDE_MNEMONIC" 
for chain_id in ${CHAINS[@]}; do
    account_name=$(GET_VAR_VALUE RELAYER_${chain_id}_ACCT)
    relayer_mnemonic=$(GET_VAR_VALUE     RELAYER_${chain_id}_MNEMONIC)
    refresh_mnemonic=$(GET_VAR_VALUE     REFRESH_${chain_id}_MNEMONIC)
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')

    $RELAYER_CMD keys restore $chain_name $account_name "$relayer_mnemonic" 
    $REFRESH_CMD keys restore $chain_name $account_name "$refresh_mnemonic" 
done

echo "Creating clients and connections..."
for chain_id in ${CHAINS[@]}; do
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')
    relayer_exec=$(GET_VAR_VALUE     RELAYER_${chain_id}_EXEC)

    printf "\t$chain_id\n"
    $relayer_exec transact link stride-${chain_name} > ${LOGS}/relayer-${chain_name}.log 2>&1
done

echo "Starting relayers..."
for chain_id in ${CHAINS[@]}; do
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')

    docker-compose up -d relayer-${chain_name}
    docker-compose logs -f relayer-${chain_name} | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> ${LOGS}/relayer-${chain_name}.log 2>&1 &
done

# update refresh-client config after channels were created (to include clients, connections, and channels)
cp $STATE/relayer/config/config.yaml $STATE/refresh-clients/config/config.yaml 
echo "Starting light client refresher..."
docker-compose up -d refresh-clients

