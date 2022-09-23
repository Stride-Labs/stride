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
TMP_MNEMONICS=${SCRIPT_DIR}/state/mnemonic.txt 
for chain_id in ${CHAINS[@]}; do
    account_name=$(GET_VAR_VALUE HERMES_${chain_id}_ACCT)
    mnemonic=$(GET_VAR_VALUE     HERMES_${chain_id}_MNEMONIC)

    echo "$mnemonic" > $TMP_MNEMONICS
    $HERMES_CMD keys add --key-name $account_name --chain $chain_id --mnemonic-file $TMP_MNEMONICS --overwrite
done
rm -f $TMP_MNEMONICS

echo "Adding Relayer keys"
for chain_id in ${CHAINS[@]}; do
    account_name=$(GET_VAR_VALUE RELAYER_${chain_id}_ACCT)
    mnemonic=$(GET_VAR_VALUE     RELAYER_${chain_id}_MNEMONIC)
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')

    $RELAYER_CMD keys restore $chain_name $account_name "$mnemonic" 
done

echo "Adding ICQ keys"
for chain_id in ${CHAINS[@]}; do
    account_name=$(GET_VAR_VALUE ICQ_${chain_id}_ACCT)
    mnemonic=$(GET_VAR_VALUE     ICQ_${chain_id}_MNEMONIC)
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')

    echo $mnemonic | $ICQ_CMD keys restore $account_name --chain $chain_name 
done