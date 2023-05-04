#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

liquid_staked_address="cosmos1x92tnm6pfkl3gsfy0rfaez5myq5zh99aek2jmd"
validator_address="cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p"

###### Setup
mnemonic="match blade slide sort seven width degree february garden hospital valve odor scan exhaust bird chuckle age ozone timber claim office hurdle dance roast"
echo $mnemonic | $LSM_MAIN_CMD keys add hot --recover --keyring-backend test
sleep 5

echo ">>> Adding new staking account"
$LSM_MAIN_CMD tx bank send $($LSM_MAIN_CMD keys show rly8 -a) $liquid_staked_address 10000000stake --from rly8 -y | TRIM_TX 
sleep 5 && echo ""

echo "Bank balance:"
$LSM_MAIN_CMD q bank balances $liquid_staked_address 

