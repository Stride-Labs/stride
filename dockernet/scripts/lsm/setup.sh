#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

## Create Accounts
echo ">>> Registering Accounts:"
mnemonic1="match blade slide sort seven width degree february garden hospital valve odor scan exhaust bird chuckle age ozone timber claim office hurdle dance roast"
echo $mnemonic1 | $LSM_MAIN_CMD keys add staker1 --recover --keyring-backend test
sleep 2
echo $mnemonic1 | $STRIDE_MAIN_CMD keys add staker1 --recover --keyring-backend test
sleep 2

mnemonic2="supply follow sudden machine pledge primary maple head turkey young prefer virus output kind corn horse concert claw chronic pear repeat salad night caution"
echo $mnemonic2 | $LSM_MAIN_CMD keys add staker2 --recover --keyring-backend test
sleep 2
echo $mnemonic2 | $STRIDE_MAIN_CMD keys add staker2 --recover --keyring-backend test
sleep 2

# Staker #1 Address on Stride: stride1x92tnm6pfkl3gsfy0rfaez5myq5zh99a6a2w0p
# Staker #1 Address on LSM:    cosmos1x92tnm6pfkl3gsfy0rfaez5myq5zh99aek2jmd
# Staker #2 Address on Stride: stride14tvx4ee0v6cs3j04u7hqckcxxt9kax0tcfamw9
# Staker #2 Address on LSM:    cosmos14tvx4ee0v6cs3j04u7hqckcxxt9kax0tmza86f
# Validator #1 Address:        cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p
# Validator #2 Address:        cosmosvaloper17kht2x2ped6qytr2kklevtvmxpw7wq9rarvcqz
# Validator #3 Address:        cosmosvaloper1nnurja9zt97huqvsfuartetyjx63tc5zxcyn3n
# Validator #4 Address:        cosmosvaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttr0ks75
staker1_lsm_address=$($LSM_MAIN_CMD keys show staker1 -a) 
staker2_lsm_address=$($LSM_MAIN_CMD keys show staker2 -a) 
staker1_stride_address=$($STRIDE_MAIN_CMD keys show staker1 -a) 
staker2_stride_address=$($STRIDE_MAIN_CMD keys show staker2 -a) 
validator_address=$(GET_VAL_ADDR LSM 2) 

## Fund accounts
echo ">>> Fund staking accounts:"
$LSM_MAIN_CMD tx bank send $($LSM_MAIN_CMD keys show rly8 -a) $staker1_lsm_address 10000000stake --from rly8 -y | TRIM_TX 
sleep 10
$LSM_MAIN_CMD tx bank send $($LSM_MAIN_CMD keys show rly8 -a) $staker2_lsm_address 10000000stake --from rly8 -y | TRIM_TX 
sleep 5 && echo ""

echo "Bank balances:"
echo -e "\n> Staker #1:"
$LSM_MAIN_CMD q bank balances $staker1_lsm_address 
echo -e "\n> Staker #2:"
$LSM_MAIN_CMD q bank balances $staker2_lsm_address 
echo ""

## Delegate, Tokenize and Transfer
echo ">>> Delegate:"
$LSM_MAIN_CMD tx staking delegate $validator_address 10000000stake --from staker1 -y | TRIM_TX 
sleep 2
$LSM_MAIN_CMD tx staking delegate $validator_address 10000000stake --from staker2 -y | TRIM_TX 
sleep 5 && echo ""

echo "Delegations:"
echo -e "\n> Staker #1:"
$LSM_MAIN_CMD q staking delegations $staker1_lsm_address 
echo -e "\n> Staker #2:"
$LSM_MAIN_CMD q staking delegations $staker2_lsm_address 
sleep 2 && echo ""

echo ">>> Tokenize to liquid staker:"
$LSM_MAIN_CMD tx staking tokenize-share $validator_address 10000000stake $staker1_lsm_address --from staker1 -y --gas auto | TRIM_TX 
sleep 2
$LSM_MAIN_CMD tx staking tokenize-share $validator_address 10000000stake $staker2_lsm_address --from staker2 -y --gas auto | TRIM_TX 
sleep 5 && echo ""

echo "Balance on LSM:"
echo -e "\n> Staker #1:"
$LSM_MAIN_CMD q bank balances $staker1_lsm_address
echo -e "\n> Staker #2:"
$LSM_MAIN_CMD q bank balances $staker2_lsm_address
sleep 2 && echo ""

echo "Tokenized shares:"
echo -e "\n> Staker #1:"
$LSM_MAIN_CMD q distribution tokenize-share-record-rewards $staker1_lsm_address 
echo -e "\n> Staker #2:"
$LSM_MAIN_CMD q distribution tokenize-share-record-rewards $staker2_lsm_address 
sleep 2 && echo ""

echo ">>> Transfer to Stride:"
$LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker1_stride_address 10000000${validator_address}/1 --from staker1 -y | TRIM_TX 
sleep 2
$LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker2_stride_address 10000000${validator_address}/2 --from staker2 -y | TRIM_TX 
sleep 5 && echo ""

echo "Balance on STRIDE:"
echo -e "\n> Staker #1:"
$STRIDE_MAIN_CMD q bank balances $staker1_stride_address 
echo -e "\n> Staker #2:"
$STRIDE_MAIN_CMD q bank balances $staker2_stride_address 
sleep 2 && echo ""
