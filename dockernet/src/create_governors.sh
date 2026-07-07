#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/../config.sh

CHAIN="$1"

CHAIN_ID=$(GET_VAR_VALUE    ${CHAIN}_CHAIN_ID)
BINARY=$(GET_VAR_VALUE      ${CHAIN}_BINARY)
DENOM=$(GET_VAR_VALUE       ${CHAIN}_DENOM)
NUM_NODES=$(GET_VAR_VALUE   ${CHAIN}_NUM_NODES)
NODE_PREFIX=$(GET_VAR_VALUE ${CHAIN}_NODE_PREFIX)
VAL_PREFIX=$(GET_VAR_VALUE  ${CHAIN}_VAL_PREFIX)

STAKE_TOKENS=${STAKE_TOKENS}000000

echo "Creating $CHAIN_ID governors.."
for (( i=1; i <= $NUM_NODES; i++ )); do
  node_name="${NODE_PREFIX}${i}"
  cmd="$BINARY --home ${STATE}/$node_name"
  val_acct="${VAL_PREFIX}${i}"
  pub_key=$($cmd tendermint show-validator)

  # SDK v0.50+ create-validator takes a json file instead of flags
  validator_json=${STATE}/$node_name/governor.json
  cat << EOF > $validator_json
{
  "pubkey": $pub_key,
  "amount": "${STAKE_TOKENS}${DENOM}",
  "moniker": "$val_acct",
  "commission-rate": "0.10",
  "commission-max-rate": "0.20",
  "commission-max-change-rate": "0.01",
  "min-self-delegation": "1"
}
EOF

  $cmd tx staking create-validator $validator_json --from $val_acct --keyring-backend test -y | TRIM_TX
  sleep 2
done

echo "Done"