#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source $SCRIPT_DIR/../config.sh

STAKE_TOKENS=${STAKE_TOKENS}ustrd

echo "Creating stride governors.."
for (( i=1; i <= $STRIDE_NUM_NODES; i++ )); do
  NODE_NAME="${STRIDE_NODE_PREFIX}${i}"
  MAIN_CMD="$STRIDE_BINARY --home ${STATE}/$NODE_NAME"
  VAL_ACCT="${STRIDE_VAL_PREFIX}${i}"
  PUB_KEY=$($MAIN_CMD tendermint show-validator)
  echo "$PUB_KEY"

  $MAIN_CMD tx staking create-validator --amount $STAKE_TOKENS --from $VAL_ACCT \
    --pubkey=$PUB_KEY --commission-rate="0.10" --commission-max-rate="0.20" \
    --commission-max-change-rate="0.01" --min-self-delegation="1" -y | TRIM_TX
  sleep 2
done
echo "Done"