#!/bin/bash
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN="${HOST_CHAINS[0]}"
HOST_MAIN_CMD=$(GET_VAR_VALUE   ${HOST_CHAIN}_MAIN_CMD)
HOST_DENOM=$(GET_VAR_VALUE      ${HOST_CHAIN}_DENOM)

reward_address=$($HOST_MAIN_CMD keys show -a reward)
deposit_address=$($STRIDE_MAIN_CMD keys show -a deposit)
fee_address=$($STRIDE_MAIN_CMD q auth module-account staketia_fee_address | grep "address:" | awk '{print $2}')

echo ">>> Claiming outstanding rewards records..."
$HOST_MAIN_CMD tx distribution withdraw-all-rewards --from delegation -y | TRIM_TX
sleep 5

echo -e "\n>>> Querying rewards balance..."
output=$($HOST_MAIN_CMD q bank balances $reward_address --denom $HOST_DENOM)
echo $output
reward_amount=$(echo $output | NUMBERS_ONLY)
sleep 1

reinvest_amount=$(echo "scale=0; $reward_amount * 90 / 100" | bc -l)
fee_amount=$(echo "scale=0; $reward_amount * 10 / 100" | bc -l)

echo -e "\n>>> Sweeping ${reinvest_amount}${HOST_DENOM} for reinvestment..."
output=$($HOST_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $deposit_address ${reinvest_amount}${HOST_DENOM} \
    --from delegation -y | TRIM_TX)
echo $output
sleep 10

echo -e "\n>>> Sweeping ${fee_amount}${HOST_DENOM} for fee collection..."
output=$($HOST_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $fee_address ${fee_amount}${HOST_DENOM} \
    --from delegation -y | TRIM_TX)
echo $output
sleep 10
