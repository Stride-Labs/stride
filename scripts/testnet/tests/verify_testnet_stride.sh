#!/bin/bash

GET_ADDRESS() {
  grep -i -A 10 "\- name: $1" /stride/keys.txt | sed -n 3p | awk '{printf $2}'
}

# SET GAIA ADDRESS TO THE DESIRED VALIDATOR
GAIA_VAL_ADDR="$1"
if [[ "$GAIA_VAL_ADDR" == "" ]]; then
    echo "Please pass the GAIA validator address as an arugment to this script."
    exit
fi

STRIDE_ACCT="stride"
STRIDE_ADDR=$(GET_ADDRESS $STRIDE_ACCT)
IBCATOM="ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"

while true; do
    printf "\n>>> strided q bank balances $STRIDE_ADDR \n"
    strided q bank balances $STRIDE_ADDR
    if strided q bank balances $STRIDE_ADDR | grep -q $IBCATOM; then 
        break
    fi
    sleep 5
done

printf "\n>>> strided tx stakeibc register-host-zone connection-0 uatom cosmos $IBCATOM channel-0 3... \n"
strided tx stakeibc register-host-zone connection-0 uatom cosmos $IBCATOM channel-0 2 --from $STRIDE_ACCT --gas 1000000 -y

sleep 5
printf "\n>>> strided tx stakeibc add-validator GAIA gval1 $GAIA_VAL_ADDR 10 5... \n"
strided tx stakeibc add-validator GAIA gval1 $GAIA_VAL_ADDR 10 5 --from $STRIDE_ACCT -y

sleep 5
while true; do
    printf "\n>>> strided q stakeibc list-host-zone \n"
    strided q stakeibc list-host-zone 
    if strided q stakeibc list-host-zone | grep -q $GAIA_VAL_ADDR; then 
        break
    fi
    sleep 5
done

#
#    0. Run the above command `q bank balances val1` to check that tokens were IBC'd over
#    1. Replace GAIA_VAL_ADDR with the cosmosvaloper address found in the other file
#    2. Run the above command register-host-zone
#    3. Run the above command add-validator
#    4. Run the `q stakeibc list-host-zone` function to see if the validator was added



