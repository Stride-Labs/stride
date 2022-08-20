#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

# Pass chain IDs as arguments
CHAINS="$@"
if [[ "$CHAINS" == "" ]]; then
    echo "ERROR: Please specify at least chain IDs that require a relayer connection"
    exit 1
fi

mkdir -p $STATE/hermes
mkdir -p $STATE/icq
mkdir -p $STATE/relayer/config

cp ${SCRIPT_DIR}/config/icq_config.yaml $STATE/icq/config.yaml
cp ${SCRIPT_DIR}/config/hermes_config.toml $STATE/hermes/config.toml
cp ${SCRIPT_DIR}/config/relayer_config.yaml $STATE/relayer/config/config.yaml

echo "Adding Hermes keys"
for chain_id in ${CHAINS[@]}; do
    __account_name_var=HERMES_${chain_id}_ACCT
    __mnemonic_var=HERMES_${chain_id}_MNEMONIC

    account_name=${!__account_name_var}
    mnemonic=${!__mnemonic_var}

    $HERMES_CMD keys restore --name $account_name --mnemonic "$mnemonic" $chain_id 
done

# echo "Adding Relayer keys"
# for chain_id in ${chains[@]}; do
#     __account_name_var=RELAYER_${chain_id}_ACCT
#     __mnemonic_var=RELAYER_${chain_id}_MNEMONIC

#     account_name=${!__account_name_var}
#     mnemonic=${!__mnemonic_var}
#     chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')

#     $RELAYER_CMD keys restore $chain_name $account_name "$mnemonic" 
# done

echo "Adding ICQ keys"
for chain_id in ${CHAINS[@]}; do
    __account_name_var=ICQ_${chain_id}_ACCT
    __mnemonic_var=ICQ_${chain_id}_MNEMONIC

    account_name=${!__account_name_var}
    mnemonic=${!__mnemonic_var}
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')

    echo $mnemonic | $ICQ_CMD keys restore $account_name --chain $chain_name 
done
