#!/bin/bash

GET_ADDRESS() {
  grep -i -A 10 "\- name: val1" /gaia/keys.txt | sed -n 3p | awk '{printf $2}'
}

STRIDE_ADDRESS=$(GET_ADDRESS val1)

printf "\n>>> gaiad tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 100000uatom... \n"
gaiad tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 100000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
sleep 5

printf "\n>>> gaiad q staking validators \n"
gaiad q staking validators 

echo "Validator Addresses:"
gaiad q staking validators | grep operator_address | awk '{print $2}'

#
#    1. Get stride address, replace $STRIDE_ADDRESS above with that
#    2. Run the above command `ibc-transfer`
#    3. Run the `q staking validators` command and grab the GAIA validator address
#         this should start with "cosmosvaloper"
#    4. Move to verify_testnet_stride.sh
#

