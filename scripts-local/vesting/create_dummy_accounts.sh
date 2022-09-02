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

NUM_ACCTS=20
for i in $(seq 1 $NUM_ACCTS); do
    $STRIDE_CMD keys add vester$i --keyring-backend test --output json >> $SCRIPT_DIR/vesting/dummy_accts.csv
    $STRIDE_CMD keys show vester$i -a
    $STRIDE_CMD keys delete vester$i -y &> /dev/null
    # echo "Vesting $amt to $addr, with a $(convertsecs $CLIFF) cliff, then linearly over $(convertsecs $DURATION)." >> $TX_LOGS 2>&1
    # $STRIDE_CMD add-genesis-account ${addr} $amt --vesting-start-time $VESTING_START_TIME --vesting-end-time $VESTING_END_TIME --vesting-amount $amt # actually set this account as a validator
done
# done < $SCRIPT_DIR/vesting/vesting_accts.csv

