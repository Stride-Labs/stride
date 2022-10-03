#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

HOST=JUNO
HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log
STRIDE_LOGS=$SCRIPT_DIR/logs/stride.log

HOST_CHAIN_ID=$(GET_VAR_VALUE ${HOST}_CHAIN_ID)
HERMES_HOST_ACCT=$(GET_VAR_VALUE HERMES_${HOST}_ACCT)
RELAYER_HOST_ACCT=$(GET_VAR_VALUE RELAYER_${HOST}_ACCT)
HOST_CHAIN_NAME=$(printf "$HOST" | awk '{ print tolower($0) }')

# cleanup any stale state
docker-compose down
rm -rf $SCRIPT_DIR/state $SCRIPT_DIR/logs/*.log $SCRIPT_DIR/logs/temp
mkdir -p $SCRIPT_DIR/logs

# Start stride
sh ${SCRIPT_DIR}/init_chain.sh STRIDE

docker-compose up -d stride1
docker-compose logs -f stride1 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS 2>&1 &

printf "Waiting for Stride to start..."
( tail -f -n0 $STRIDE_LOGS & ) | grep -q "finalizing commit of block"
echo "Done"

# Setup relayers
mkdir -p $STATE/hermes
mkdir -p $STATE/relayer/config
cp ${SCRIPT_DIR}/config/hermes_config.toml $STATE/hermes/config.toml
cp ${SCRIPT_DIR}/config/relayer_config.yaml $STATE/relayer/config/config.yaml

echo "Adding Hermes keys"
TMP_MNEMONICS=${SCRIPT_DIR}/state/mnemonic.txt 
echo "$HERMES_STRIDE_MNEMONIC" > $TMP_MNEMONICS
$HERMES_CMD keys add --key-name $HERMES_STRIDE_ACCT --chain $STRIDE_CHAIN_ID --mnemonic-file $TMP_MNEMONICS --overwrite
echo "$HOT_RELAYER_HOST" > $TMP_MNEMONICS
$HERMES_CMD keys add --key-name $HERMES_HOST_ACCT --chain $HOST_CHAIN_ID --mnemonic-file $TMP_MNEMONICS --overwrite
rm -f $TMP_MNEMONICS

echo "Adding Relayer keys"
$RELAYER_CMD keys restore stride $RELAYER_STRIDE_ACCT "$HOT_RELAYER_STRIDE" 
$RELAYER_CMD keys restore $HOST_CHAIN_NAME $RELAYER_HOST_ACCT "$HOT_RELAYER_HOST" 

echo "Done"
