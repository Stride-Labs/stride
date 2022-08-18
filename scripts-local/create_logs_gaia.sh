#!/bin/bash
# clean up logs one by one before creation (allows auto-updating logs with the command `while true; do make init build=logs ; sleep 5 ; done`)

set -eu
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source ${SCRIPT_DIR}/vars.sh

LOGS_DIR=$SCRIPT_DIR/logs
TEMP_LOGS_DIR=$LOGS_DIR/temp_gaia
mkdir -p $TEMP_LOGS_DIR
GAIA_LOGS_DIR=$LOGS_DIR/gaia
mkdir -p $GAIA_LOGS_DIR

while true; do
    # transactions logs
    $STRIDE_CMD q txs --events message.module=interchainquery --limit=100000 >$TEMP_LOGS_DIR/icq-events.log
    $STRIDE_CMD q txs --events message.module=stakeibc --limit=100000 >$TEMP_LOGS_DIR/stakeibc-events.log

    # accounts
    GAIA_DELEGATE="cosmos1sy63lffevueudvvlvh2lf6s387xh9xq72n3fsy6n2gr5hm6u2szs2v0ujm"
    GAIA_WITHDRAWAL="cosmos1x5p8er7e2ne8l54tx33l560l8djuyapny55pksctuguzdc00dj7saqcw2l"
    GAIA_REDEMPTION="cosmos1xmcwu75s8v7s54k79390wc5gwtgkeqhvzegpj0nm2tdwacv47tmqg9ut30"
    GAIA_REV="cosmos1wdplq6qjh2xruc7qqagma9ya665q6qhcwju3ng"
    STRIDE_ADDRESS="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

    N_VALIDATORS_STRIDE=$($STRIDE_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
    N_VALIDATORS_GAIA=$($GAIA_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
    echo "STRIDE @ $($STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_STRIDE VALS" >$TEMP_LOGS_DIR/accounts.log
    echo "GAIA   @ $($GAIA_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_GAIA VALS" >>$TEMP_LOGS_DIR/accounts.log

    printf '\n%s\n' "BALANCES STRIDE" >>$TEMP_LOGS_DIR/accounts.log
    $STRIDE_CMD q bank balances $STRIDE_ADDRESS >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "BALANCES GAIA (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $GAIA_CMD q bank balances $GAIA_DELEGATE >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "DELEGATIONS GAIA (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $GAIA_CMD q staking delegations $GAIA_DELEGATE >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "UNBONDING-DELEGATIONS GAIA (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $GAIA_CMD q staking unbonding-delegations $GAIA_DELEGATE >>$TEMP_LOGS_DIR/accounts.log

    printf '\n%s\n' "BALANCES GAIA (REDEMPTION ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $GAIA_CMD q bank balances $GAIA_REDEMPTION >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "BALANCES GAIA (REVENUE ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $GAIA_CMD q bank balances $GAIA_REV >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "BALANCES GAIA (WITHDRAWAL ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $GAIA_CMD q bank balances $GAIA_WITHDRAWAL >>$TEMP_LOGS_DIR/accounts.log

    printf '\n%s\n' "LIST-HOST-ZONES STRIDE" >>$TEMP_LOGS_DIR/accounts.log
    $STRIDE_CMD q stakeibc list-host-zone >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "LIST-DEPOSIT-RECORDS" >>$TEMP_LOGS_DIR/accounts.log
    $STRIDE_CMD q records list-deposit-record  >> $TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "LIST-EPOCH-UNBONDING-RECORDS" >>$TEMP_LOGS_DIR/accounts.log
    $STRIDE_CMD q records list-epoch-unbonding-record  >> $TEMP_LOGS_DIR/accounts.log

    printf '\n%s\n' "LIST-USER-REDEMPTION-RECORDS" >>$TEMP_LOGS_DIR/accounts.log
    $STRIDE_CMD q records list-user-redemption-record >> $TEMP_LOGS_DIR/accounts.log
    # printf '\n%s\n' "LIST-PENDING-CLAIMS" >>$TEMP_LOGS_DIR/accounts.log
    # $STRIDE_CMD q records list-user-redemption-record >> $TEMP_LOGS_DIR/accounts.log

    # ================ OSMO =============================================================================
    # accounts
    OSMO_DELEGATE="osmo1cx04p5974f8hzh2lqev48kjrjugdxsxy7mzrd0eyweycpr90vk8q8d6f3h"
    OSMO_WITHDRAWAL="osmo10arcf5r89cdmppntzkvulc7gfmw5lr66y2m25c937t6ccfzk0cqqz2l6xv"
    OSMO_REDEMPTION="osmo1uy9p9g609676rflkjnnelaxatv8e4sd245snze7qsxzlk7dk7s8qrcjaez"
    OSMO_REV="osmo19uvw0azm9u0k6vqe4e22cga6kteskdqq6vv7c7"

    N_VALIDATORS_OSMO=$($OSMO_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
    echo "OSMO   @ $($OSMO_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_OSMO VALS" >>$TEMP_LOGS_DIR/accounts.log

    printf '\n%s\n' "BALANCES STRIDE" >>$TEMP_LOGS_DIR/accounts.log
    $STRIDE_CMD q bank balances $STRIDE_ADDRESS >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "BALANCES OSMO (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $OSMO_CMD q bank balances $OSMO_DELEGATE >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "DELEGATIONS OSMO (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $OSMO_CMD q staking delegations $OSMO_DELEGATE >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "UNBONDING-DELEGATIONS OSMO (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $OSMO_CMD q staking unbonding-delegations $OSMO_DELEGATE >>$TEMP_LOGS_DIR/accounts.log

    printf '\n%s\n' "BALANCES OSMO (REDEMPTION ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $OSMO_CMD q bank balances $OSMO_REDEMPTION >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "BALANCES OSMO (REVENUE ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $OSMO_CMD q bank balances $OSMO_REV >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "BALANCES OSMO (WITHDRAWAL ACCT)" >>$TEMP_LOGS_DIR/accounts.log
    $OSMO_CMD q bank balances $OSMO_WITHDRAWAL >>$TEMP_LOGS_DIR/accounts.log

    printf '\n%s\n' "LIST-HOST-ZONES STRIDE" >>$TEMP_LOGS_DIR/accounts.log
    $STRIDE_CMD q stakeibc show-host-zone GAIA >>$TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "LIST-DEPOSIT-RECORDS" >>$TEMP_LOGS_DIR/accounts.log
    $STRIDE_CMD q records list-deposit-record  >> $TEMP_LOGS_DIR/accounts.log
    printf '\n%s\n' "LIST-EPOCH-UNBONDING-RECORDS" >>$TEMP_LOGS_DIR/accounts.log
    $STRIDE_CMD q records list-epoch-unbonding-record  >> $TEMP_LOGS_DIR/accounts.log
    
    mv $TEMP_LOGS_DIR/*.log $GAIA_LOGS_DIR
    sleep 3
done
