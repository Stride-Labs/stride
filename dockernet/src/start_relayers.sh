#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh

for (( row=0; row<${#RELAYER_PATHS[@]}; row+=RELAYER_PATH_NUM_COLUMNS )); do  
    # Loooping though the path array, and save down the chain names at either end of the relayed path
    path=${RELAYER_PATHS[row+RELAYER_PATH_NAME_COLUMN]}
    relayer_exec="$DOCKER_COMPOSE run --rm relayer-$path"

    src_chain=${RELAYER_PATHS[row+RELAYER_SRC_CHAIN_COLUMN]}
    dst_chain=${RELAYER_PATHS[row+RELAYER_DST_CHAIN_COLUMN]}

    # Only continue if both chains are currently enabled
    chains_in_path_enabled=0
    for enabled_chain in STRIDE ${HOST_CHAINS[@]} ${ACCESSORY_CHAINS[@]:-}; do
        if [[ "$enabled_chain" == "$src_chain" || "$enabled_chain" == "$dst_chain" ]]; then
           chains_in_path_enabled=$((chains_in_path_enabled+1))
        fi 
    done
    if [[ "$chains_in_path_enabled" != 2 ]]; then
        continue
    fi

    # Save down the lower-case name (for the relayer config), as well as the account names and mnemonics
    src_chain_name=$(printf "$src_chain" | awk '{ print tolower($0) }')
    dst_chain_name=$(printf "$dst_chain" | awk '{ print tolower($0) }')

    src_account=${RELAYER_PATHS[row+RELAYER_SRC_ACCT_COLUMN]}
    dst_account=${RELAYER_PATHS[row+RELAYER_DST_ACCT_COLUMN]}

    src_mnemonic_var=${RELAYER_PATHS[row+RELAYER_SRC_MNEMONIC_COLUMN]}
    dst_mnemonic_var=${RELAYER_PATHS[row+RELAYER_DST_MNEMONIC_COLUMN]}
    src_mnemonic=$(GET_VAR_VALUE ${src_mnemonic_var})
    dst_mnemonic=$(GET_VAR_VALUE ${dst_mnemonic_var})

    # Setup the logs and config file
    # The relevant config file is located by checking the account name
    relayer_logs=${LOGS}/relayer-${path}.log
    relayer_config=$STATE/relayer-${path}/config

    mkdir -p $relayer_config
    chmod -R 777 $STATE/relayer-${path}

    config_file=${DOCKERNET_HOME}/config/relayer_config_stride.yaml 
    if [[ "$src_account" == "stride-ics" ]]; then
        config_file=${DOCKERNET_HOME}/config/relayer_config_ics.yaml
    elif [[ "$src_account" == "dydx-noble" ]]; then
        config_file=${DOCKERNET_HOME}/config/relayer_config_dydx_noble.yaml
    fi
    cp $config_file $relayer_config/config.yaml

    # Share the same account name for the stride side of each path
    if [[ "$src_account" == "${src_chain_name}-${dst_chain_name}" && "$src_chain_name" == "stride" ]]; then
        src_account=stride
    fi

    # Restore relayer keys and create the transfer channel
    printf "$src_chain <> $dst_chain - Adding relayer keys..."
    $relayer_exec rly keys restore $src_chain_name $src_account "$src_mnemonic" >> $relayer_logs 2>&1
    $relayer_exec rly keys restore $dst_chain_name $dst_account "$dst_mnemonic" >> $relayer_logs 2>&1
    echo "Done"

    printf "$src_chain <> $dst_chain - Creating client, connection, and transfer channel..." | tee -a $relayer_logs
    $relayer_exec rly transact link ${path} >> $relayer_logs 2>&1
    echo "Done"

    # Pipe the logs to a file
    $DOCKER_COMPOSE up -d relayer-${path}
    SAVE_DOCKER_LOGS relayer-${path} $relayer_logs
done
