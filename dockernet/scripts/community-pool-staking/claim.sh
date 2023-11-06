### CLAIM
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

community_pool_return_address=$(GET_ICA_ADDR GAIA community_pool_return)
community_pool_holding_address=$(GET_HOST_ZONE_FIELD GAIA community_pool_redeem_holding_address)

# check balances before claiming redeemed stake
echo ">>> Balances before claim..."
$GAIA_MAIN_CMD q bank balances $community_pool_return_address

#claim stake
echo -e "\n>>> Claiming redeemed tokens..."
epoch=$($STRIDE_MAIN_CMD q records list-user-redemption-record  | grep -B 3 -m 1 "receiver: $community_pool_return_address" | grep "epoch_number"| NUMBERS_ONLY)

$STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens GAIA $epoch $community_pool_holding_address --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
sleep 5

# check balances after claiming redeemed stake
echo -e "\n>>> Balances after claim..."
$GAIA_MAIN_CMD q bank balances $community_pool_return_address
