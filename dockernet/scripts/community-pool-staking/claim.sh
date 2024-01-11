### CLAIM
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN=$REWARD_CONVERTER_HOST_ZONE
HOST_CHAIN_ID=$(GET_VAR_VALUE ${HOST_CHAIN}_CHAIN_ID)
HOST_MAIN_CMD=$(GET_VAR_VALUE ${HOST_CHAIN}_MAIN_CMD)
HOST_DENOM=$(GET_VAR_VALUE    ${HOST_CHAIN}_DENOM)

community_pool_return_address=$(GET_ICA_ADDR $HOST_CHAIN community_pool_return)

# check balances before claiming redeemed stake
echo ">>> Balances before claim..."
$HOST_MAIN_CMD q bank balances $community_pool_return_address

#claim stake
echo -e "\n>>> Claiming redeemed tokens..."
epoch=$($STRIDE_MAIN_CMD q records list-user-redemption-record  | grep -B 5 -m 1 "receiver: $community_pool_return_address" | grep "epoch_number"| NUMBERS_ONLY)
$STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens $HOST_CHAIN_ID $epoch $community_pool_return_address --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
sleep 5

# check balances after claiming redeemed stake
echo -e "\n>>> Balances after claim..."
$HOST_MAIN_CMD q bank balances $community_pool_return_address
