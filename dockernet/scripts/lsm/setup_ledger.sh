#!/bin/bash
# set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# Create Accounts
# echo ">>> Registering Accounts on Ledger:"
# echo ">>> This will error because it can't take ledger input in a script. Instead comment it out and add them to keyring manually. Then re-run the script."
# echo $STRIDE_MAIN_CMD keys add stakerl --ledger --keyring-backend test
# sleep 5
# echo $GAIA_MAIN_CMD keys add stakerl --ledger --keyring-backend test
# sleep 5

# NOTE: these accounts will differ for your ledger!
# Staker Address on Stride: stride1ngu6s0wdghc6dgy4jlz68e6zsn7u6ccn33myhr
# Staker Address on GAIA:   cosmos1ngu6s0wdghc6dgy4jlz68e6zsn7u6ccnj6mcr0
# Validator Address:        cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p
echo ">>> You will need to sign ledger transactions as this script runs!"
stakerl_lsm_address=$($GAIA_MAIN_CMD keys show stakerl -a) 
stakerl_stride_address=$($STRIDE_MAIN_CMD keys show stakerl -a) 
validator_address=$(GET_VAL_ADDR GAIA 2)
validator_address_2="cosmosvaloper1nnurja9zt97huqvsfuartetyjx63tc5zxcyn3n"
validator_address_4="cosmosvaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttr0ks75"

## Fund accounts
echo ">>> Fund staking account:"
$GAIA_MAIN_CMD tx bank send $($GAIA_MAIN_CMD keys show gval1 -a) $stakerl_lsm_address 2000000000000uatom --from rly8 -y | TRIM_TX 
sleep 5 && echo ""

echo "Bank balance:"
$GAIA_MAIN_CMD q bank balances $stakerl_lsm_address 

## Delegate, Tokenize and Transfer
echo ">>> Delegate to Coinbase:"
$GAIA_MAIN_CMD tx staking delegate $validator_address 1500000000000uatom --from stakerl -y | TRIM_TX && echo ""
sleep 5
echo ">>> Delegate to XXX:"
$GAIA_MAIN_CMD tx staking delegate $validator_address_2 10000000uatom --from stakerl -y | TRIM_TX && echo ""
sleep 5
echo ">>> Delegate to XXX:"
$GAIA_MAIN_CMD tx staking delegate $validator_address_4 10000000uatom --from stakerl -y | TRIM_TX && echo ""
sleep 5

echo "Delegations:"
$GAIA_MAIN_CMD q staking delegations $stakerl_lsm_address && echo ""
exit
sleep 2

echo ">>> Tokenize to liquid stakerl:"
$GAIA_MAIN_CMD tx staking tokenize-share $validator_address 10000000uatom $stakerl_lsm_address --from stakerl -y --gas auto | TRIM_TX && echo ""
sleep 5

echo "Balance on GAIA:"
$GAIA_MAIN_CMD q bank balances $stakerl_lsm_address && echo ""
sleep 2

echo "Tokenized shares:"
$GAIA_MAIN_CMD q distribution tokenize-share-record-rewards $stakerl_lsm_address && echo ""
sleep 2

echo ">>> Transfer to Stride:"
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $stakerl_stride_address 10000000${validator_address}/1 --from stakerl -y | TRIM_TX && echo ""
sleep 10

echo "Balance on STRIDE:"
$STRIDE_MAIN_CMD q bank balances $stakerl_stride_address && echo ""
sleep 2
