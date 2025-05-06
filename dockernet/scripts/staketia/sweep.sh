#!/bin/bash
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN="${HOST_CHAINS[0]}"
HOST_MAIN_CMD=$(GET_VAR_VALUE   ${HOST_CHAIN}_MAIN_CMD)
HOST_DENOM=$(GET_VAR_VALUE      ${HOST_CHAIN}_DENOM)

echo ">>> Querying action from records..."
$STRIDE_MAIN_CMD q staketia unbonding-records
unbond_amount=$($STRIDE_MAIN_CMD q staketia unbonding-records | grep -B 2 "UNBONDED" | grep "native_amount" | NUMBERS_ONLY)
record_id=$($STRIDE_MAIN_CMD q staketia unbonding-records | grep -B 4 "UNBONDED" | grep "id" | NUMBERS_ONLY)
sleep 1

echo -e "\n>>> Sweeping ${unbond_amount}${HOST_DENOM} for record $record_id..."
claim_address=$(GET_ADDRESS STRIDE claim)
output=$($HOST_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $claim_address ${unbond_amount}${HOST_DENOM} \
    --from delegation -y | TRIM_TX)
echo $output
sleep 10

echo -e "\n>>> Submitting confirm-sweep tx for record $record_id on Stride..."
tx_hash=$(echo $output | awk '{print $4}')
$STRIDE_MAIN_CMD tx staketia confirm-sweep $record_id $tx_hash --from operator -y | TRIM_TX
