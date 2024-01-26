#!/bin/bash
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN="${HOST_CHAINS[0]}"
HOST_MAIN_CMD=$(GET_VAR_VALUE   ${HOST_CHAIN}_MAIN_CMD)
HOST_DENOM=$(GET_VAR_VALUE      ${HOST_CHAIN}_DENOM)

echo ">>> Querying action from records..."
$STRIDE_MAIN_CMD q staketia delegation-records
delegation_amount=$($STRIDE_MAIN_CMD q staketia delegation-records | grep -B 2 "DELEGATION_QUEUE" | grep "native_amount" | NUMBERS_ONLY)
record_id=$($STRIDE_MAIN_CMD q staketia delegation-records | grep -B 3 "DELEGATION_QUEUE" | grep "id" | NUMBERS_ONLY)
sleep 1

echo -e "\n>>> Delegating ${delegation_amount}${HOST_DENOM} for record $record_id..."
validator_address=$(GET_VAL_ADDR $HOST_CHAIN 1)
output=$($HOST_MAIN_CMD tx staking delegate $validator_address ${delegation_amount}${HOST_DENOM} \
    --from delegation -y | TRIM_TX)
echo $output
sleep 1

echo -e "\n>>> Submitting confirm-delegation tx for record $record_id on Stride..."
tx_hash=$(echo $output | awk '{print $4}')
$STRIDE_MAIN_CMD tx staketia confirm-delegation $record_id $tx_hash --from operator -y | TRIM_TX
