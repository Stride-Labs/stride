#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# Create Accounts
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
rly_2_gaia_address=$($GAIA_MAIN_CMD keys show rly2 -a) 
echo $(GET_VAL_ADDR GAIA 2)

validator_address_3=$(GET_VAL_ADDR GAIA 3)
validator_address_4=$(GET_VAL_ADDR GAIA 4)

# Fund accounts
echo ">>> Fund staking accounts:"
$GAIA_MAIN_CMD tx bank send $($GAIA_MAIN_CMD keys show gval1 -a) $staker_gaia_address 200000000000uatom --from rly8 -y | TRIM_TX 
sleep 5 && echo ""
$STRIDE_MAIN_CMD tx bank send $($STRIDE_MAIN_CMD keys show val1 -a) $staker_stride_address 100000000ustrd --from val1 -y | TRIM_TX 
sleep 5 && echo ""

echo "Bank balance:"
$GAIA_MAIN_CMD q bank balances $staker_gaia_address 

## Delegate, Tokenize and Transfer
echo "\n\nDelegate from staker to Validator 1 and Validator 2 (10 ATOM each)"
echo ">>> Delegate to Val 1:"
$GAIA_MAIN_CMD tx staking delegate $validator_address_1 10000000uatom --from staker -y | TRIM_TX && echo ""
sleep 5
echo ">>> Delegate to Val 2:"
$GAIA_MAIN_CMD tx staking delegate $validator_address_2 10000000uatom --from staker -y | TRIM_TX && echo ""
sleep 5



echo "\n\nValidator Bond to Validator 2 (small amount 1 ATOM) and to Validator 3 (large amount 10 ATOM)"
## Delegate, Tokenize and Transfer
echo ">>> Delegate to Val 2 from rly2"
$GAIA_MAIN_CMD tx staking delegate $validator_address_2 1000000uatom --from rly2 -y | TRIM_TX
sleep 5
echo ">>> Validator Bond to Val 2 from rly2"
$GAIA_MAIN_CMD tx staking validator-bond $validator_address_2 --from rly2 -y | TRIM_TX
sleep 5

echo ">>> Delegate to Val 3 from rly2"
$GAIA_MAIN_CMD tx staking delegate $validator_address_3 10000000uatom --from rly2 -y | TRIM_TX
sleep 5
echo ">>> Validator Bond to Val 3 from rly2"
$GAIA_MAIN_CMD tx staking validator-bond $validator_address_3 --from rly2 -y | TRIM_TX

echo "\n\n\nCHECKING THE STATE BEFORE FRONTEND TESTING\n"
echo "\n\nValidator bonded delegations:"
$GAIA_MAIN_CMD q staking delegations $rly_2_gaia_address
echo "\n\nDelegations from user:"
$GAIA_MAIN_CMD q staking delegations $staker_gaia_address
