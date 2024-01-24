#!/bin/bash
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN="${HOST_CHAINS[0]}"
HOST_MAIN_CMD=$(GET_VAR_VALUE   ${HOST_CHAIN}_MAIN_CMD)
HOST_DENOM=$(GET_VAR_VALUE      ${HOST_CHAIN}_DENOM)

echo ">>> Querying action from records..."
$STRIDE_MAIN_CMD q staketia unbonding-records
unbond_amount=$($STRIDE_MAIN_CMD q staketia unbonding-records | grep -B 2 "UNBONDING_QUEUE" | grep "native_amount" | NUMBERS_ONLY)
record_id=$($STRIDE_MAIN_CMD q staketia unbonding-records | grep -B 4 "UNBONDING_QUEUE" | grep "id" | NUMBERS_ONLY)
sleep 1

echo -e "\n>>> Unbonding ${unbond_amount}${HOST_DENOM} for record $record_id..."
validator_address=$(GET_VAL_ADDR $HOST_CHAIN 1)
output=$($HOST_MAIN_CMD tx staking unbond $validator_address ${unbond_amount}${HOST_DENOM} \
    --from delegation -y | TRIM_TX)
echo $output
sleep 1

echo -e "\n>>> Submitting confirm-undelegation tx for record $record_id on Stride..."
tx_hash=$(echo $output | awk '{print $4}')
$STRIDE_MAIN_CMD tx staketia confirm-undelegation $record_id $tx_hash --from operator -y | TRIM_TX
