#!/bin/bash
# clean up logs one by one before creation (allows auto-updating logs with the command `while true; do make init build=logs ; sleep 5 ; done`)

set -eu
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source ${SCRIPT_DIR}/../config.sh

LOGS_DIR=$DOCKERNET_HOME/logs
TEMP_LOGS_DIR=$LOGS_DIR/temp

STATE_LOG=state.log
BALANCES_LOG=balances.log

mkdir -p $TEMP_LOGS_DIR

while true; do
    N_VALIDATORS_STRIDE=$($STRIDE_MAIN_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
    echo "STRIDE @ $($STRIDE_MAIN_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_STRIDE VALS" >$TEMP_LOGS_DIR/$STATE_LOG
    echo "STRIDE @ $($STRIDE_MAIN_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_STRIDE VALS" >$TEMP_LOGS_DIR/$BALANCES_LOG

    for chain in ${HOST_CHAINS[@]}; do
        HOST_MAIN_CMD=$(GET_VAR_VALUE ${chain}_MAIN_CMD)

        N_VALIDATORS_HOST=$($HOST_MAIN_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
        echo "$chain   @ $($HOST_MAIN_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_HOST VALS" >>$TEMP_LOGS_DIR/$STATE_LOG
        echo "$chain   @ $($HOST_MAIN_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_HOST VALS" >>$TEMP_LOGS_DIR/$BALANCES_LOG
    done

    printf '\n%s\n' "LIST-HOST-ZONES STRIDE" >>$TEMP_LOGS_DIR/$STATE_LOG
    $STRIDE_MAIN_CMD q stakeibc list-host-zone >>$TEMP_LOGS_DIR/$STATE_LOG
    printf '\n%s\n' "LIST-DEPOSIT-RECORDS" >>$TEMP_LOGS_DIR/$STATE_LOG
    $STRIDE_MAIN_CMD q records list-deposit-record  >> $TEMP_LOGS_DIR/$STATE_LOG
    printf '\n%s\n' "LIST-EPOCH-UNBONDING-RECORDS" >>$TEMP_LOGS_DIR/$STATE_LOG
    $STRIDE_MAIN_CMD q records list-epoch-unbonding-record  >> $TEMP_LOGS_DIR/$STATE_LOG
    printf '\n%s\n' "LIST-USER-REDEMPTION-RECORDS" >>$TEMP_LOGS_DIR/$STATE_LOG
    $STRIDE_MAIN_CMD q records list-user-redemption-record >> $TEMP_LOGS_DIR/$STATE_LOG
    printf '\n%s\n' "LIST-LSM-TOKEN-DEPOSIT-RECORDS" >>$TEMP_LOGS_DIR/$STATE_LOG
    $STRIDE_MAIN_CMD q records lsm-deposits >> $TEMP_LOGS_DIR/$STATE_LOG

    printf '\n%s\n' "BALANCES STRIDE" >>$TEMP_LOGS_DIR/$BALANCES_LOG
    $STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) >>$TEMP_LOGS_DIR/$BALANCES_LOG

    for chain in ${HOST_CHAINS[@]}; do
        HOST_CHAIN_ID=$(GET_VAR_VALUE ${chain}_CHAIN_ID)
        HOST_MAIN_CMD=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
        
        DELEGATION_ICA_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID delegation)
        REDEMPTION_ICA_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID redemption)
        WITHDRAWAL_ICA_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID withdrawal)
        FEE_ICA_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID fee)

        COMMUNITY_POOL_DEPOSIT_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID community_pool_deposit)
        COMMUNITY_POOL_RETURN_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID community_pool_return)

        COMMUNITY_POOL_STAKE_ADDR=$(GET_HOST_ZONE_FIELD $HOST_CHAIN_ID community_pool_stake_holding_address)
        COMMUNITY_POOL_REDEEM_ADDR=$(GET_HOST_ZONE_FIELD $HOST_CHAIN_ID community_pool_redeem_holding_address)

        printf '\n%s\n' "==========================  $chain  =============================" >> $TEMP_LOGS_DIR/$BALANCES_LOG

        printf '\n%s\n' "DELEGATIONS $chain" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $HOST_MAIN_CMD q staking delegations $DELEGATION_ICA_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG
        printf '\n%s\n' "UNBONDING-DELEGATIONS $chain" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $HOST_MAIN_CMD q staking unbonding-delegations $DELEGATION_ICA_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG

        printf '\n%s\n' "DELEGATION ACCT BALANCE" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $HOST_MAIN_CMD q bank balances $DELEGATION_ICA_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG
        printf '\n%s\n' "REDEMPTION ACCT BALANCE" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $HOST_MAIN_CMD q bank balances $REDEMPTION_ICA_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG
        printf '\n%s\n' "FEE ACCT BALANCE" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $HOST_MAIN_CMD q bank balances $FEE_ICA_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG
        printf '\n%s\n' "WITHDRAWAL ACCT BALANCE" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $HOST_MAIN_CMD q bank balances $WITHDRAWAL_ICA_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG

        printf '\n%s\n' "COMMUNITY POOL BALANCE" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $HOST_MAIN_CMD q distribution community-pool >> $TEMP_LOGS_DIR/$BALANCES_LOG

        printf '\n%s\n' "COMMUNITY POOL DEPOSIT ACCT BALANCE" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $HOST_MAIN_CMD q bank balances $COMMUNITY_POOL_DEPOSIT_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG
        printf '\n%s\n' "COMMUNITY POOL RETURN ACCT BALANCE" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $HOST_MAIN_CMD q bank balances $COMMUNITY_POOL_RETURN_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG

        printf '\n%s\n' "COMMUNITY POOL STAKE HOLDING ACCT BALANCE" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $STRIDE_MAIN_CMD q bank balances $COMMUNITY_POOL_STAKE_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG
        printf '\n%s\n' "COMMUNITY POOL REDEEM HOLDING ACCT BALANCE" >> $TEMP_LOGS_DIR/$BALANCES_LOG
        $STRIDE_MAIN_CMD q bank balances $COMMUNITY_POOL_REDEEM_ADDR >> $TEMP_LOGS_DIR/$BALANCES_LOG
    done

    mv $TEMP_LOGS_DIR/*.log $LOGS_DIR
    sleep 3
done