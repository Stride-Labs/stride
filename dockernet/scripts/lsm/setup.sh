#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

## Create Accounts
echo ">>> Registering Accounts:"
mnemonic="match blade slide sort seven width degree february garden hospital valve odor scan exhaust bird chuckle age ozone timber claim office hurdle dance roast"
echo $mnemonic | $STRIDE_MAIN_CMD keys add staker --recover --keyring-backend test
sleep 5
echo $mnemonic | $LSM_MAIN_CMD keys add staker --recover --keyring-backend test
sleep 5

# Staker Address on Stride: stride1x92tnm6pfkl3gsfy0rfaez5myq5zh99a6a2w0p
# Staker Address on LSM:    cosmos1x92tnm6pfkl3gsfy0rfaez5myq5zh99aek2jmd
# Validator Address:        cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p
staker_lsm_address=$($LSM_MAIN_CMD keys show staker -a) 
staker_stride_address=$($STRIDE_MAIN_CMD keys show staker -a) 
validator_address=$(GET_VAL_ADDR LSM 1) 

## Fund accounts
echo ">>> Fund staking account:"
$LSM_MAIN_CMD tx bank send $($LSM_MAIN_CMD keys show lval1 -a) $staker_lsm_address 10000000stake --from rly8 -y | TRIM_TX 
sleep 5 && echo ""

echo "Bank balance:"
$LSM_MAIN_CMD q bank balances $staker_lsm_address 

## Delegate, Tokenize and Transfer
echo ">>> Delegate:"
$LSM_MAIN_CMD tx staking delegate $validator_address 10000000stake --from staker -y | TRIM_TX && echo ""
sleep 5

echo "Delegations:"
$LSM_MAIN_CMD q staking delegations $staker_lsm_address && echo ""
sleep 2

echo ">>> Tokenize to liquid staker:"
$LSM_MAIN_CMD tx staking tokenize-share $validator_address 1000000stake $staker_lsm_address --from staker -y --gas auto | TRIM_TX && echo ""
sleep 5

echo "Balance on LSM:"
$LSM_MAIN_CMD q bank balances $staker_lsm_address && echo ""
sleep 2

echo "Tokenized shares:"
$LSM_MAIN_CMD q distribution tokenize-share-record-rewards $staker_lsm_address && echo ""
sleep 2

echo ">>> Transfer to Stride:"
$LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker_stride_address 1000000${validator_address}/1 --from staker -y | TRIM_TX && echo ""
sleep 5

echo "Balance on STRIDE:"
$STRIDE_MAIN_CMD q bank balances $staker_stride_address && echo ""
sleep 2
