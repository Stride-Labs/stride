#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/../vars.sh

# vesting params
VESTING_START_TIME=$(($(date +%s)+120)) # <= unix time start of vesting period (2 minutes from now)
VESTING_END_TIME=$((VESTING_START_TIME+300)) # <= unix time end of vesting period (7 minutes from now)

while IFS=, read -r addr amt
do
    echo "Vesting $amt to $addr."
    $STRIDE_CMD add-genesis-account ${addr} $amt --vesting-start-time $VESTING_START_TIME --vesting-end-time $VESTING_END_TIME --vesting-amount $amt # actually set this account as a validator
done < $SCRIPT_DIR/vesting/vesting_accts.csv

