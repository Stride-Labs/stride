set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh
source ${SCRIPT_DIR}/denoms.sh

################################################################
# Liquid stake failure by moving tokens while query is in flight
################################################################

# # LSM Liquid stake twice to set checkpoint
# STAKER_1_ADDRESS=$($STRIDE_MAIN_CMD keys show staker1 -a)
# STAKER_2_ADDRESS=$($STRIDE_MAIN_CMD keys show staker2 -a)

# STAKER_1_LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 2 1) # channel-0, validator 2, recordId 1
# STAKER_2_LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 2 2) # channel-0, validator 2, recordId 2

# echo ">>> LSM Liquid stake 1000000 from staker1"
# $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $STAKER_1_LSM_TOKEN_DENOM --from staker1 --gas auto -y | TRIM_TX
# echo -e "\n>>> Sleeping 30 seconds..."
# sleep 10 

# echo -e "\n>>> LSM Liquid stake 1000000 again from staker1"
# $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $STAKER_1_LSM_TOKEN_DENOM --from staker1 --gas auto -y | TRIM_TX
# sleep 5

# # LSM Liquid stake from user 2 to trip the query
# echo -e "\nStaker #2 Initial Balance:"
# $STRIDE_MAIN_CMD q bank balances $STAKER_2_ADDRESS

# echo -e "\n>>> LSM Liquid stake 1000000 from staker2 - should submit query"
# $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $STAKER_2_LSM_TOKEN_DENOM --from staker2 --gas auto -y | TRIM_TX
# sleep 2

# echo -e "\n>>> Moving LSM tokens out of account"
# $STRIDE_MAIN_CMD tx bank send $STAKER_2_ADDRESS $STAKER_1_ADDRESS 10000000${STAKER_2_LSM_TOKEN_DENOM} --from staker2 -y | TRIM_TX
# sleep 4

# echo -e "\nStaker #2 Final Balance (should be empty):"
# $STRIDE_MAIN_CMD q bank balances $STAKER_2_ADDRESS
# sleep 2

# echo -e "\n>>> Querying liquid stake errors from events [insufficient funds error expected]:"
# $STRIDE_MAIN_CMD q txs --events 'lsm_liquid_stake_failed.module=stakeibc' | grep "key: error" -A 3


#####################
### Rebalance - Setup
#####################
# Total: 11,429,950
#        Weight  Starting    Target
# Val 1:     5    128,000 -> 672,350   (+544,350)
# Val 2:    10  3,097,000 -> 1,344,700 (-1,752,300)
# Val 3:    10  1,344,700 -> 1,344,700 (+0)
# Val 4:    25  1,875,000 -> 3,361,750 (+1,486,750)
# Val 5:    35  4,985,250 -> 4,706,450 (-278,800)

# #### Setup
# echo ">>> Registering Accounts:"
# mnemonic1="match blade slide sort seven width degree february garden hospital valve odor scan exhaust bird chuckle age ozone timber claim office hurdle dance roast"
# echo $mnemonic1 | $LSM_MAIN_CMD keys add staker1 --recover --keyring-backend test
# sleep 2
# echo $mnemonic1 | $STRIDE_MAIN_CMD keys add staker1 --recover --keyring-backend test
# sleep 2

# mnemonic2="supply follow sudden machine pledge primary maple head turkey young prefer virus output kind corn horse concert claw chronic pear repeat salad night caution"
# echo $mnemonic2 | $LSM_MAIN_CMD keys add staker2 --recover --keyring-backend test
# sleep 2
# echo $mnemonic2 | $STRIDE_MAIN_CMD keys add staker2 --recover --keyring-backend test
# sleep 2

# staker1_lsm_address=$($LSM_MAIN_CMD keys show staker1 -a) 
# staker2_lsm_address=$($LSM_MAIN_CMD keys show staker2 -a) 

# staker1_stride_address=$($STRIDE_MAIN_CMD keys show staker1 -a) 
# staker2_stride_address=$($STRIDE_MAIN_CMD keys show staker2 -a) 

# echo ">>> Fund staking accounts:"
# $LSM_MAIN_CMD tx bank send $($LSM_MAIN_CMD keys show rly8 -a) $staker1_lsm_address 100000000stake --from rly8 -y | TRIM_TX 
# sleep 10
# $LSM_MAIN_CMD tx bank send $($LSM_MAIN_CMD keys show rly8 -a) $staker2_lsm_address 100000000stake --from rly8 -y | TRIM_TX 
# sleep 5 && echo ""

# echo ">>> Delegate:"
# $LSM_MAIN_CMD tx staking delegate $(GET_VAL_ADDR LSM 1) 10000000stake --from staker1 -y | TRIM_TX 
# sleep 2
# $LSM_MAIN_CMD tx staking delegate $(GET_VAL_ADDR LSM 2) 10000000stake --from staker2 -y | TRIM_TX 
# sleep 5
# $LSM_MAIN_CMD tx staking delegate $(GET_VAL_ADDR LSM 3) 10000000stake --from staker1 -y | TRIM_TX 
# sleep 2
# $LSM_MAIN_CMD tx staking delegate $(GET_VAL_ADDR LSM 4) 10000000stake --from staker2 -y | TRIM_TX 
# sleep 5
# $LSM_MAIN_CMD tx staking delegate $(GET_VAL_ADDR LSM 5) 10000000stake --from staker1 -y | TRIM_TX 
# sleep 5 && echo ""

# echo ">>> Tokenize to liquid staker:"
# $LSM_MAIN_CMD tx staking tokenize-share $(GET_VAL_ADDR LSM 1) 10000000stake $staker1_lsm_address --from staker1 -y --gas auto | TRIM_TX 
# sleep 2
# $LSM_MAIN_CMD tx staking tokenize-share $(GET_VAL_ADDR LSM 2) 10000000stake $staker2_lsm_address --from staker2 -y --gas auto | TRIM_TX 
# sleep 5
# $LSM_MAIN_CMD tx staking tokenize-share $(GET_VAL_ADDR LSM 3) 10000000stake $staker1_lsm_address --from staker1 -y --gas auto | TRIM_TX 
# sleep 2
# $LSM_MAIN_CMD tx staking tokenize-share $(GET_VAL_ADDR LSM 4) 10000000stake $staker2_lsm_address --from staker2 -y --gas auto | TRIM_TX 
# sleep 5
# $LSM_MAIN_CMD tx staking tokenize-share $(GET_VAL_ADDR LSM 5) 10000000stake $staker1_lsm_address --from staker1 -y --gas auto | TRIM_TX 
# sleep 5 && echo ""

# echo ">>> Transfer to Stride:"
# $LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker1_stride_address 10000000$(GET_VAL_ADDR LSM 1)/1 --from staker1 -y | TRIM_TX 
# sleep 2
# $LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker2_stride_address 10000000$(GET_VAL_ADDR LSM 2)/2 --from staker2 -y | TRIM_TX 
# sleep 5
# $LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker1_stride_address 10000000$(GET_VAL_ADDR LSM 3)/3 --from staker1 -y | TRIM_TX 
# sleep 2
# $LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker2_stride_address 10000000$(GET_VAL_ADDR LSM 4)/4 --from staker2 -y | TRIM_TX 
# sleep 5
# $LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker1_stride_address 10000000$(GET_VAL_ADDR LSM 5)/5 --from staker1 -y | TRIM_TX 
# sleep 5 && echo ""

# echo ">>> Balance on STRIDE:"
# echo -e "\n> Staker #1:"
# $STRIDE_MAIN_CMD q bank balances $staker1_stride_address 
# echo -e "\n> Staker #2:"
# $STRIDE_MAIN_CMD q bank balances $staker2_stride_address 
# sleep 2 && echo ""


#####################
### Rebalance - Test
#####################
# LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 1 1) # channel-0, validator 1, recordId 1
# $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 128000 $LSM_TOKEN_DENOM --from staker1 --gas auto -y | TRIM_TX
# sleep 5

# LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 2 2) # channel-0, validator 2, recordId 2
# $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 3097000 $LSM_TOKEN_DENOM --from staker2 --gas auto -y | TRIM_TX
# sleep 5

# LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 3 3) # channel-0, validator 3, recordId 3
# $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1344700 $LSM_TOKEN_DENOM --from staker1 --gas auto -y | TRIM_TX
# sleep 5

# LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 4 4) # channel-0, validator 4, recordId 4
# $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1875000 $LSM_TOKEN_DENOM --from staker2 --gas auto -y | TRIM_TX
# sleep 5

# LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 5 5) # channel-0, validator 5, recordId 5
# $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 4985250 $LSM_TOKEN_DENOM --from staker1 --gas auto -y | TRIM_TX

