#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

## Create Accounts
echo ">>> Registering Accounts:"
mnemonic="match blade slide sort seven width degree february garden hospital valve odor scan exhaust bird chuckle age ozone timber claim office hurdle dance roast"
echo $mnemonic | $STRIDE_MAIN_CMD keys add staker --recover --keyring-backend test
sleep 5
echo $mnemonic | $GAIA_MAIN_CMD keys add staker --recover --keyring-backend test
sleep 5

# Staker Address on Stride: stride1x92tnm6pfkl3gsfy0rfaez5myq5zh99a6a2w0p
# Staker Address on GAIA:   cosmos1x92tnm6pfkl3gsfy0rfaez5myq5zh99aek2jmd
# Validator Address:        cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p
staker_gaia_address=$($GAIA_MAIN_CMD keys show staker -a) 
staker_stride_address=$($STRIDE_MAIN_CMD keys show staker -a) 
validator_address_1=$(GET_VAL_ADDR GAIA 1)
validator_address_2=$(GET_VAL_ADDR GAIA 2)
validator_address_3=$(GET_VAL_ADDR GAIA 3)
validator_address_4=$(GET_VAL_ADDR GAIA 4)

## Fund accounts
echo ">>> Fund staking accounts:"
$GAIA_MAIN_CMD tx bank send $($GAIA_MAIN_CMD keys show gval1 -a) $staker_gaia_address 2000000000000uatom --from rly8 -y | TRIM_TX 
sleep 5 && echo ""
$STRIDE_MAIN_CMD tx bank send $($STRIDE_MAIN_CMD keys show val1 -a) $staker_stride_address 10000000ustrd --from val1 -y | TRIM_TX 
sleep 5 && echo ""

echo "Bank balance:"
$GAIA_MAIN_CMD q bank balances $staker_gaia_address 

## Delegate, Tokenize and Transfer
echo ">>> Delegate to Val 1:"
$GAIA_MAIN_CMD tx staking delegate $validator_address_1 10000000uatom --from staker -y | TRIM_TX && echo ""
sleep 5
echo ">>> Delegate to Val 2:"
$GAIA_MAIN_CMD tx staking delegate $validator_address_2 10000000uatom --from staker -y | TRIM_TX && echo ""
sleep 5
echo ">>> Delegate to Val 3:"
$GAIA_MAIN_CMD tx staking delegate $validator_address_3 10000000uatom --from staker -y | TRIM_TX && echo ""
sleep 5
echo ">>> Delegate to Val 4:"
$GAIA_MAIN_CMD tx staking delegate $validator_address_4 10000000uatom --from staker -y | TRIM_TX && echo ""
sleep 5

echo "Delegations:"
$GAIA_MAIN_CMD q staking delegations $staker_gaia_address && echo ""
sleep 2

echo ">>> Tokenize to liquid staker:"
$GAIA_MAIN_CMD tx staking tokenize-share $validator_address_2 10000000uatom $staker_gaia_address --from staker -y \
    --gas auto --gas-adjustment 1.3 | TRIM_TX && echo ""
sleep 5

echo "Balance on GAIA:"
$GAIA_MAIN_CMD q bank balances $staker_gaia_address && echo ""
sleep 2

echo "Tokenized shares:"
$GAIA_MAIN_CMD q distribution tokenize-share-record-rewards $staker_gaia_address && echo ""
sleep 2

echo ">>> Transfer to Stride:"
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker_stride_address 10000000${validator_address_2}/1 --from staker -y | TRIM_TX && echo ""
sleep 10

echo "Balance on STRIDE:"
$STRIDE_MAIN_CMD q bank balances $staker_stride_address && echo ""
sleep 2
