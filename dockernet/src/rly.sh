#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh

# create localhost connection + relayer
relayer_exec=$(GET_VAR_VALUE RELAYER_LOCAL_EXEC)
# echo $relayer_exec
# chain_name=$(printf "$chain" | awk '{ print tolower($0) }')
account_name=$(GET_VAR_VALUE RELAYER_LOCAL_ACCT)
# echo $account_name
mnemonic=$(GET_VAR_VALUE     RELAYER_LOCAL_MNEMONIC)
# echo $mnemonic

relayer_logs=${LOGS}/relayer-localhost.log
relayer_config=$STATE/relayer-localhost/config

# 1
mkdir -p $relayer_config
chmod -R 777 $STATE/relayer-localhost
cp ${DOCKERNET_HOME}/config/relayer_config.yaml $relayer_config/config.yaml

printf "STRIDE <> STRIDE - Adding relayer keys..."
$relayer_exec rly keys restore stride $RELAYER_STRIDE_ACCT "$mnemonic" >> $relayer_logs 2>&1
# $relayer_exec rly keys restore local $account_name "$mnemonic" >> $relayer_logs 2>&1
echo "Done"

# 2 
$relayer_exec rly paths add STRIDE STRIDE stride-stride

# 3
$relayer_exec rly transact channel stride-stride --src-port transfer --dst-port transfer --order unordered --version ics20-1

# $relayer_exec rly keys list stride # channel stride-stride --src-port transfer --dst-port transfer --order unordered --version ics20-1

# $relayer_exec rly transact channel stride-localhost --src-port transfer --dst-port transfer --order unordered --version ics20


# printf "STRIDE <> STRIDE - Creating client, connection, and transfer channel..." | tee -a $relayer_logs
# $relayer_exec rly transact link stride-stride >> $relayer_logs 2>&1
# echo "Done"
$DOCKER_COMPOSE up -d relayer-localhost
$DOCKER_COMPOSE logs -f relayer-localhost | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $relayer_logs 2>&1 &
