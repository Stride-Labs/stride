#!/bin/bash

convertsecs() {
 h=$(bc <<< "${1}/3600")
 m=$(bc <<< "(${1}%3600)/60")
 s=$(bc <<< "${1}%60")
 printf "%02dD:%02dM:%05.2fS\n" $h $m $s
}

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/../vars.sh

# vesting params
CLIFF=120
DURATION=300
VESTING_START_TIME=$(($(date +%s)+$CLIFF)) # <= unix time start of vesting period (2 minutes from now)
VESTING_END_TIME=$((VESTING_START_TIME+$DURATION)) # <= unix time end of vesting period (7 minutes from now)

echo "VESTING ACCOUNTS:"
while IFS=, read -r addr amt
do
    echo "Vesting $amt to $addr, with a $(convertsecs $CLIFF) cliff, then linearly over $(convertsecs $DURATION)."
    $STRIDE_CMD add-genesis-account ${addr} $amt --vesting-start-time $VESTING_START_TIME --vesting-end-time $VESTING_END_TIME --vesting-amount $amt # actually set this account as a validator
done < $SCRIPT_DIR/vesting/vesting_accts.csv

