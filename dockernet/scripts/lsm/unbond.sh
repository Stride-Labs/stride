#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh
source ${SCRIPT_DIR}/denoms.sh

## Create Accounts
echo ">>> Registering Accounts:"
mnemonic1="match blade slide sort seven width degree february garden hospital valve odor scan exhaust bird chuckle age ozone timber claim office hurdle dance roast"
echo $mnemonic1 | $LSM_MAIN_CMD keys add staker1 --recover --keyring-backend test
sleep 2
echo $mnemonic1 | $STRIDE_MAIN_CMD keys add staker1 --recover --keyring-backend test
sleep 2

echo ">>> Verify Validators have addresses and weights as expected:"
$STRIDE_MAIN_CMD query stakeibc show-validators LSM
sleep 10

# Staker #1 Address on Stride: stride1x92tnm6pfkl3gsfy0rfaez5myq5zh99a6a2w0p
# Staker #1 Address on LSM:    cosmos1x92tnm6pfkl3gsfy0rfaez5myq5zh99aek2jmd

# Before these tests edit the weights in src/register_host.sh to be (5, 10, 0, 10)
# Validator #1 Address:        cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p     weight 5
# Validator #2 Address:        cosmosvaloper17kht2x2ped6qytr2kklevtvmxpw7wq9rarvcqz     weight 10
# Validator #3 Address:        cosmosvaloper1nnurja9zt97huqvsfuartetyjx63tc5zxcyn3n     weight 0
# Validator #4 Address:        cosmosvaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttr0ks75     weight 10

staker_lsm_address=$($LSM_MAIN_CMD keys show staker1 -a) 
staker_stride_address=$($STRIDE_MAIN_CMD keys show staker1 -a) 
weighted_validator_address=$(GET_VAL_ADDR LSM 2) 
zero_weight_validator_address=$(GET_VAL_ADDR LSM 3)


## Fund accounts
echo ">>> Fund staking accounts:"
$LSM_MAIN_CMD tx bank send $($LSM_MAIN_CMD keys show rly8 -a) $staker_lsm_address 20000000stake --from rly8 -y | TRIM_TX 
sleep 10 && echo ""

echo "Bank balances:"
$LSM_MAIN_CMD q bank balances $staker_lsm_address 
echo ""

## Delegate, Tokenize and Transfer
echo ">>> Delegate:"
$LSM_MAIN_CMD tx staking delegate $weighted_validator_address 5000000stake --from staker1 -y | TRIM_TX 
sleep 2
$LSM_MAIN_CMD tx staking delegate $zero_weight_validator_address 10000000stake --from staker1 -y | TRIM_TX 
sleep 5 && echo ""

## staker1 is going to send (5m) as a normal stake, as LSM token on val2 with weight 10, and 10m as LSM token on val3 with 0 weight

echo "Delegations:"
$LSM_MAIN_CMD q staking delegations $staker_lsm_address 
sleep 2 && echo ""

echo ">>> Tokenize to liquid stake:"
$LSM_MAIN_CMD tx staking tokenize-share $weighted_validator_address 5000000stake $staker_lsm_address --from staker1 -y --gas auto | TRIM_TX 
sleep 2
$LSM_MAIN_CMD tx staking tokenize-share $zero_weight_validator_address 10000000stake $staker_lsm_address --from staker1 -y --gas auto | TRIM_TX 
sleep 5 && echo ""

echo "Balance on LSM:"
$LSM_MAIN_CMD q bank balances $staker_lsm_address
sleep 2 && echo ""

echo "Tokenized shares:"
$LSM_MAIN_CMD q distribution tokenize-share-record-rewards $staker_lsm_address 
sleep 2 && echo ""

echo ">>> Transfer LSM Tokens to Stride:"
$LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker_stride_address 5000000${weighted_validator_address}/1 --from staker1 -y | TRIM_TX 
sleep 2
$LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker_stride_address 10000000${zero_weight_validator_address}/2 --from staker1 -y | TRIM_TX 
sleep 5 && echo ""

echo ">>> Transfer normal native state to Stride:"
$LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker_stride_address 5000000stake --from staker1 -y | TRIM_TX 
sleep 5 && echo ""

echo "Balance on STRIDE:"
$STRIDE_MAIN_CMD q bank balances $staker_stride_address 
sleep 2 && echo ""

# Liquid Stake the LSM tokens -- should stay on their respective validators
echo "LSM Liquid Staking the NFTs:"
staker1_lsm_ibc_denom1="ibc/19825915130745DAC8CBC51A6DBE4FC7644463CF4254CD46D49B15AABEE73FB8"
staker1_lsm_ibc_denom2="ibc/46D92AD8242504CFE5E339C5D720D46981C732D4EFD640B0444429AFEF09A1BE"
lsm_ibc_denom1=$(GET_LSM_IBC_TOKEN_DENOM 0 2 1) # channel-0, validator 2, recordId 1
lsm_ibc_denom2=$(GET_LSM_IBC_TOKEN_DENOM 0 3 2) # channel-0, validator 3, recordId 1
echo lsm_ibc_denom1
echo lsm_ibc_denom2

$STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 5000000 $lsm_ibc_denom1 --from staker1 -y --gas auto | TRIM_TX
sleep 3
$STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 10000000 $lsm_ibc_denom2 --from staker1 -y --gas auto | TRIM_TX
sleep 3 && echo ""

# Liquid Stake the native tokens -- should be distributed according to the weights, just a normal liquid stake
echo "Normal Liquid Staking the remaining native tokens:"
#staker1_native_ibc_denom="ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
$STRIDE_MAIN_CMD tx stakeibc liquid-stake 5000000 stake --from staker1 -y --gas auto | TRIM_TX
sleep 180 && echo "" # need very long sleep to make sure it finished

echo "Balance on STRIDE:"
$STRIDE_MAIN_CMD q bank balances $staker_stride_address 
echo "Balance on LSM HUB:"
$LSM_MAIN_CMD q bank balances $staker_lsm_address
sleep 5 && echo ""

echo "Validator state before unbonding:"
$STRIDE_MAIN_CMD query stakeibc show-validators LSM
sleep 10 && echo ""

# Before unbonding: 5m LSM on val2, 10m LSM on val3, 5m split according to the weights

#        Weight     Starting   
# Val 1:     5      1,000,000 
# Val 2:    10      7,000,000 
# Val 3:     0     10,000,000 
# Val 4:    10      2,000,000




##################################################################################################################
# Use only 1 of the 4 unbonding commands below
##################################################################################################################


echo "Unbond less than the zero weight LSM total which is 10m:"
$STRIDE_MAIN_CMD tx stakeibc redeem-stake 6000000 LSM $staker_lsm_address --from staker1 -y --gas auto | TRIM_TX
sleep 180 && echo ""

# Unbonding 6m which is less than the sum of the LSM stakes on 0-weighted validators 10m
#        Weight     Starting    Target
# Val 1:     5      1,000,000 ->  1,000,000 (+0)
# Val 2:    10      7,000,000 ->  7,000,000 (+0)
# Val 3:     0     10,000,000 ->  4,000,000 (-6,000,000)
# Val 4:    10      2,000,000 ->  2,000,000 (+0)
# All 6m unbonded should come from the LSM amount on the 0-weighted validator 3 



#echo "Unbond more than the zero weight sum but less than total LSM stake:"
#$STRIDE_MAIN_CMD tx stakeibc redeem-stake 12000000 LSM $staker_lsm_address --from staker1 -y --gas auto | TRIM_TX
#sleep 180 && echo ""

# Unbonding 12m which is less than the LSM stakes total 15m, but more than exists on only 0-weighted vals total 10m
#        Weight     Starting    Target
# Val 1:     5      1,000,000 ->  1,000,000 (+0)
# Val 2:    10      7,000,000 ->  5,000,000 (-2,000,000)
# Val 3:     0     10,000,000 ->          0 (-10,000,000)
# Val 4:    10      2,000,000 ->  2,000,000 (+0)
# All 12m unbonded should come from the LSM stakes where:
# - all 10m available coming from 0-weighted validators first
# - remaining 2m coming from non-0-weighted validators with LSM



#echo "Unbond more than the total LSM stake:"
#$STRIDE_MAIN_CMD tx stakeibc redeem-stake 18000000 LSM $staker_lsm_address --from staker1 -y --gas auto | TRIM_TX
#sleep 180 && echo ""

# Unbonding 18m which is more than the LSM stakes total which is 15m
#        Weight     Starting    Target
# Val 1:     5      1,000,000 ->  400,000 (-600,000)
# Val 2:    10      7,000,000 ->  800,000 (-6,200,000)
# Val 3:     0     10,000,000 ->        0 (-10,000,000)
# Val 4:    10      2,000,000 ->  800,000 (-1,200,000)
# All 12m unbonded should come from the LSM stakes where:
# - all 10m available coming from 0-weighted validators first
# - all 5m coming from non-0-weighted validators with LSM
# - 3m taken from normal stake on weighted validators
# - remaining 2m split across validators in proportion to the weights


#echo "Unbond more than the total possible stake (should error):"
#$STRIDE_MAIN_CMD tx stakeibc redeem-stake 24000000 LSM $staker_lsm_address --from staker1 -y --gas auto | TRIM_TX
#sleep 180 && echo ""

##################################################################################################################
# Use only 1 of the 4 unbonding commands above
##################################################################################################################



echo "Balance on STRIDE:"
$STRIDE_MAIN_CMD q bank balances $staker_stride_address 
echo "Balance on LSM HUB:"
$LSM_MAIN_CMD q bank balances $staker_lsm_address
sleep 5 && echo ""

echo "Validator state after unbonding:"
$STRIDE_MAIN_CMD query stakeibc show-validators LSM
echo ""
