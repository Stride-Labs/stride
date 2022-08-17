#!/bin/bash
# clean up logs one by one before creation (allows auto-updating logs with the command `while true; do make init build=logs ; sleep 5 ; done`)

set -eu
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source ${SCRIPT_DIR}/vars.sh

LOGS_DIR=$SCRIPT_DIR/logs
TEMP_LOGS_DIR=$LOGS_DIR/temp_osmo
mkdir -p $TEMP_LOGS_DIR
OSMO_LOGS_DIR=$LOGS_DIR/osmo
mkdir -p $OSMO_LOGS_DIR

while true; do
    # transactions logs
    $STRIDE_CMD q txs --events message.module=interchainquery --limit=100000 >$TEMP_LOGS_DIR/icq-events_osmo.log
    $STRIDE_CMD q txs --events message.module=stakeibc --limit=100000 >$TEMP_LOGS_DIR/stakeibc-events_osmo.log

    # accounts
    OSMO_DELEGATE="osmo1cx04p5974f8hzh2lqev48kjrjugdxsxy7mzrd0eyweycpr90vk8q8d6f3h"
    OSMO_WITHDRAWAL="osmo10arcf5r89cdmppntzkvulc7gfmw5lr66y2m25c937t6ccfzk0cqqz2l6xv"
    OSMO_REDEMPTION="osmo1uy9p9g609676rflkjnnelaxatv8e4sd245snze7qsxzlk7dk7s8qrcjaez"
    OSMO_REV="osmo1n4r77qsmu9chvchtmuqy9cv3s539q87r398l6ugf7dd2q5wgyg9su3wd4g"
    STRIDE_ADDRESS="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

    N_VALIDATORS_STRIDE=$($STRIDE_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
    N_VALIDATORS_OSMO=$($OSMO_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
    echo "STRIDE @ $($STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_STRIDE VALS" >$TEMP_LOGS_DIR/accounts_osmo.log
    echo "OSMO   @ $($OSMO_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_OSMO VALS" >>$TEMP_LOGS_DIR/accounts_osmo.log

    printf '\n%s\n' "BALANCES STRIDE" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $STRIDE_CMD q bank balances $STRIDE_ADDRESS >>$TEMP_LOGS_DIR/accounts_osmo.log
    printf '\n%s\n' "BALANCES OSMO (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $OSMO_CMD q bank balances $OSMO_DELEGATE >>$TEMP_LOGS_DIR/accounts_osmo.log
    printf '\n%s\n' "DELEGATIONS OSMO (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $OSMO_CMD q staking delegations $OSMO_DELEGATE >>$TEMP_LOGS_DIR/accounts_osmo.log
    printf '\n%s\n' "UNBONDING-DELEGATIONS OSMO (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $OSMO_CMD q staking unbonding-delegations $OSMO_DELEGATE >>$TEMP_LOGS_DIR/accounts_osmo.log

    printf '\n%s\n' "BALANCES OSMO (REDEMPTION ACCT)" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $OSMO_CMD q bank balances $OSMO_REDEMPTION >>$TEMP_LOGS_DIR/accounts_osmo.log
    printf '\n%s\n' "BALANCES OSMO (REVENUE ACCT)" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $OSMO_CMD q bank balances $OSMO_REV >>$TEMP_LOGS_DIR/accounts_osmo.log
    printf '\n%s\n' "BALANCES OSMO (WITHDRAWAL ACCT)" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $OSMO_CMD q bank balances $OSMO_WITHDRAWAL >>$TEMP_LOGS_DIR/accounts_osmo.log

    printf '\n%s\n' "LIST-HOST-ZONES STRIDE" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $STRIDE_CMD q stakeibc list-host-zone >>$TEMP_LOGS_DIR/accounts_osmo.log
    printf '\n%s\n' "LIST-DEPOSIT-RECORDS" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $STRIDE_CMD q records list-deposit-record  >> $TEMP_LOGS_DIR/accounts_osmo.log
    printf '\n%s\n' "LIST-EPOCH-UNBONDING-RECORDS" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $STRIDE_CMD q records list-epoch-unbonding-record  >> $TEMP_LOGS_DIR/accounts_osmo.log

    printf '\n%s\n' "LIST-USER-REDEMPTION-RECORDS" >>$TEMP_LOGS_DIR/accounts_osmo.log
    $STRIDE_CMD q records list-user-redemption-record >> $TEMP_LOGS_DIR/accounts_osmo.log
    # printf '\n%s\n' "LIST-PENDING-CLAIMS" >>$TEMP_LOGS_DIR/accounts_osmo.log
    # $STRIDE_CMD q records list-user-redemption-record >> $TEMP_LOGS_DIR/accounts_osmo.log
    
    mv $TEMP_LOGS_DIR/*.log $OSMO_LOGS_DIR
    sleep 3
done
