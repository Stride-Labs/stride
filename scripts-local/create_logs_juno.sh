#!/bin/bash
# clean up logs one by one before creation (allows auto-updating logs with the command `while true; do make init build=logs ; sleep 5 ; done`)

set -eu
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source ${SCRIPT_DIR}/vars.sh

LOGS_DIR=$SCRIPT_DIR/logs
TEMP_LOGS_DIR=$LOGS_DIR/temp_juno
mkdir -p $TEMP_LOGS_DIR
JUNO_LOGS_DIR=$LOGS_DIR/juno
mkdir -p $JUNO_LOGS_DIR

while true; do

    # accounts
    JUNO_DELEGATE="juno1xan7vt4nurz6c7x0lnqnvpmuc0lljz7rycqmuz2kk6wxv4k69d0sfats35"
    JUNO_WITHDRAWAL="juno104n6h822n6n7psqjgjl7emd2uz67lptggp5cargh6mw0gxpch2gsk53qk5"
    JUNO_REDEMPTION="juno1y6haxdt03cgkc7aedxrlaleeteel7fgc0nvtu2kggee3hnrlvnvs4kw2v9"
    JUNO_REV="juno1rp8qgfq64wmjg7exyhjqrehnvww0t9ev3f3p2ls82umz2fxgylqsz3vl9h"
    STRIDE_ADDRESS="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

    N_VALIDATORS_STRIDE=$($STRIDE_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
    N_VALIDATORS_JUNO=$($JUNO_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
    echo "STRIDE @ $($STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_STRIDE VALS" >$TEMP_LOGS_DIR/accounts_juno.log
    echo "JUNO   @ $($JUNO_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_JUNO VALS" >>$TEMP_LOGS_DIR/accounts_juno.log

    printf '\n%s\n' "BALANCES STRIDE" >>$TEMP_LOGS_DIR/accounts_juno.log
    $STRIDE_CMD q bank balances $STRIDE_ADDRESS >>$TEMP_LOGS_DIR/accounts_juno.log
    printf '\n%s\n' "BALANCES JUNO (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts_juno.log
    $JUNO_CMD q bank balances $JUNO_DELEGATE >>$TEMP_LOGS_DIR/accounts_juno.log
    printf '\n%s\n' "DELEGATIONS JUNO (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts_juno.log
    $JUNO_CMD q staking delegations $JUNO_DELEGATE >>$TEMP_LOGS_DIR/accounts_juno.log
    printf '\n%s\n' "UNBONDING-DELEGATIONS JUNO (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts_juno.log
    $JUNO_CMD q staking unbonding-delegations $JUNO_DELEGATE >>$TEMP_LOGS_DIR/accounts_juno.log

    printf '\n%s\n' "BALANCES JUNO (REDEMPTION ACCT)" >>$TEMP_LOGS_DIR/accounts_juno.log
    $JUNO_CMD q bank balances $JUNO_REDEMPTION >>$TEMP_LOGS_DIR/accounts_juno.log
    printf '\n%s\n' "BALANCES JUNO (REVENUE ACCT)" >>$TEMP_LOGS_DIR/accounts_juno.log
    $JUNO_CMD q bank balances $JUNO_REV >>$TEMP_LOGS_DIR/accounts_juno.log
    printf '\n%s\n' "BALANCES JUNO (WITHDRAWAL ACCT)" >>$TEMP_LOGS_DIR/accounts_juno.log
    $JUNO_CMD q bank balances $JUNO_WITHDRAWAL >>$TEMP_LOGS_DIR/accounts_juno.log

    printf '\n%s\n' "LIST-HOST-ZONES STRIDE" >>$TEMP_LOGS_DIR/accounts_juno.log
    $STRIDE_CMD q stakeibc list-host-zone >>$TEMP_LOGS_DIR/accounts_juno.log
    printf '\n%s\n' "LIST-DEPOSIT-RECORDS" >>$TEMP_LOGS_DIR/accounts_juno.log
    $STRIDE_CMD q records list-deposit-record  >> $TEMP_LOGS_DIR/accounts_juno.log
    printf '\n%s\n' "LIST-EPOCH-UNBONDING-RECORDS" >>$TEMP_LOGS_DIR/accounts_juno.log
    $STRIDE_CMD q records list-epoch-unbonding-record  >> $TEMP_LOGS_DIR/accounts_juno.log

    printf '\n%s\n' "LIST-USER-REDEMPTION-RECORDS" >>$TEMP_LOGS_DIR/accounts_juno.log
    $STRIDE_CMD q records list-user-redemption-record >> $TEMP_LOGS_DIR/accounts_juno.log
    # printf '\n%s\n' "LIST-PENDING-CLAIMS" >>$TEMP_LOGS_DIR/accounts_juno.log
    # $STRIDE_CMD q records list-user-redemption-record >> $TEMP_LOGS_DIR/accounts_juno.log
    
    mv $TEMP_LOGS_DIR/*.log $JUNO_LOGS_DIR
    sleep 3
done
