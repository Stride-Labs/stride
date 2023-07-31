#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source $SCRIPT_DIR/../config.sh

INCLUDE_HOST="$1"
STRIDE_STAKE_TOKENS=${STAKE_TOKENS}000000ustrd
HOST_STAKE_TOKENS=${STAKE_TOKENS}000000uwalk

create_validators() {
  CHAIN_ID=$1
  NUM_NODES=$2
  NODE_PREFIX=$3
  BINARY=$4
  VAL_PREFIX=$5
  STAKE_TOKENS=$6

  echo "Creating $CHAIN_ID governors.."
  for (( i=1; i <= $NUM_NODES; i++ )); do
    NODE_NAME="${NODE_PREFIX}${i}"
    MAIN_CMD="$BINARY --home ${STATE}/$NODE_NAME"
    VAL_ACCT="${VAL_PREFIX}${i}"
    PUB_KEY=$($MAIN_CMD tendermint show-validator)

    $MAIN_CMD tx staking create-validator --amount $STAKE_TOKENS --from $VAL_ACCT \
      --pubkey=$PUB_KEY --commission-rate="0.10" --commission-max-rate="0.20" \
      --commission-max-change-rate="0.01" --min-self-delegation="1" -y | TRIM_TX
    sleep 2
  done
}

create_validators stride $STRIDE_NUM_NODES $STRIDE_NODE_PREFIX $STRIDE_BINARY $STRIDE_VAL_PREFIX $STRIDE_STAKE_TOKENS

if [[ $INCLUDE_HOST == true ]]; then
  create_validators HOST $HOST_NUM_NODES $HOST_NODE_PREFIX $HOST_BINARY $HOST_VAL_PREFIX $HOST_STAKE_TOKENS
fi
echo "Done"