set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh
source ${SCRIPT_DIR}/denoms.sh

################################################################
# Liquid stake failure by moving tokens while query is in flight
################################################################

# LSM Liquid stake twice to set checkpoint
STAKER_1_ADDRESS=$($STRIDE_MAIN_CMD keys show staker1 -a)
STAKER_2_ADDRESS=$($STRIDE_MAIN_CMD keys show staker2 -a)

STAKER_1_LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 2 1) # channel-0, validator 2, recordId 1
STAKER_2_LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 2 2) # channel-0, validator 2, recordId 2

echo ">>> LSM Liquid stake 1000000 from staker1"
$STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $STAKER_1_LSM_TOKEN_DENOM --from staker1 --gas auto -y | TRIM_TX
echo -e "\n>>> Sleeping 30 seconds..."
sleep 10 

echo -e "\n>>> LSM Liquid stake 1000000 again from staker1"
$STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $STAKER_1_LSM_TOKEN_DENOM --from staker1 --gas auto -y | TRIM_TX
sleep 5

# LSM Liquid stake from user 2 to trip the query
echo -e "\nStaker #2 Initial Balance:"
$STRIDE_MAIN_CMD q bank balances $STAKER_2_ADDRESS

echo -e "\n>>> LSM Liquid stake 1000000 from staker2 - should submit query"
$STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $STAKER_2_LSM_TOKEN_DENOM --from staker2 --gas auto -y | TRIM_TX
sleep 2

echo -e "\n>>> Moving LSM tokens out of account"
$STRIDE_MAIN_CMD tx bank send $STAKER_2_ADDRESS $STAKER_1_ADDRESS 10000000${STAKER_2_LSM_TOKEN_DENOM} --from staker2 -y | TRIM_TX
sleep 4

echo -e "\nStaker #2 Final Balance (should be empty):"
$STRIDE_MAIN_CMD q bank balances $STAKER_2_ADDRESS
sleep 2

echo -e "\n>>> Querying liquid stake errors from events [insufficient funds error expected]:"
$STRIDE_MAIN_CMD q txs --events 'lsm_liquid_stake_failed.module=stakeibc' | grep "key: error" -A 3