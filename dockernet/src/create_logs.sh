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

SET_UP_COMPLETE=false
PRINT_SET_UP_MESSAGE=true
PRINT_SET_UP_WARNING=true

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

    printf '\n%s\n' "BALANCES STRIDE" >>$TEMP_LOGS_DIR/$BALANCES_LOG
    $STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) >>$TEMP_LOGS_DIR/$BALANCES_LOG

    if [ "$SET_UP_COMPLETE" = "false" ]; then
        ICA_SET_UP=()
    fi

    for chain in ${HOST_CHAINS[@]}; do
        HOST_CHAIN_ID=$(GET_VAR_VALUE ${chain}_CHAIN_ID)
        HOST_MAIN_CMD=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
        
        DELEGATION_ICA_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID delegation)
        REDEMPTION_ICA_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID redemption)
        WITHDRAWAL_ICA_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID withdrawal)
        FEE_ICA_ADDR=$(GET_ICA_ADDR $HOST_CHAIN_ID fee)

        # check if ICAs are available
        if [ "$SET_UP_COMPLETE" = "false" ]; then
            if [ -n "$DELEGATION_ICA_ADDR" ] && [ -n "$REDEMPTION_ICA_ADDR" ] && [ -n "$WITHDRAWAL_ICA_ADDR" ] && [ -n "$FEE_ICA_ADDR" ]; then
                ICA_SET_UP+=("true")
            else
                ICA_SET_UP+=("false")
            fi
        fi

        printf '\n%s\n' "==========================  $chain  =============================" >>$TEMP_LOGS_DIR/$BALANCES_LOG

        printf '\n%s\n' "BALANCES $chain (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/$BALANCES_LOG
        if [ -n "$DELEGATION_ICA_ADDR" ]; then 
            $HOST_MAIN_CMD q bank balances $DELEGATION_ICA_ADDR >>$TEMP_LOGS_DIR/$BALANCES_LOG
            printf '\n%s\n' "DELEGATIONS $chain (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/$BALANCES_LOG
            $HOST_MAIN_CMD q staking delegations $DELEGATION_ICA_ADDR >>$TEMP_LOGS_DIR/$BALANCES_LOG
            printf '\n%s\n' "UNBONDING-DELEGATIONS $chain (DELEGATION ACCT)" >>$TEMP_LOGS_DIR/$BALANCES_LOG
            $HOST_MAIN_CMD q staking unbonding-delegations $DELEGATION_ICA_ADDR >>$TEMP_LOGS_DIR/$BALANCES_LOG
        fi

        if [ -n "$REDEMPTION_ICA_ADDR" ]; then 
            printf '\n%s\n' "BALANCES $chain (REDEMPTION ACCT)" >>$TEMP_LOGS_DIR/$BALANCES_LOG
            $HOST_MAIN_CMD q bank balances $REDEMPTION_ICA_ADDR >>$TEMP_LOGS_DIR/$BALANCES_LOG
        fi

        if [ -n "$FEE_ICA_ADDR" ]; then 
            printf '\n%s\n' "BALANCES $chain (FEE ACCT)" >>$TEMP_LOGS_DIR/$BALANCES_LOG
            $HOST_MAIN_CMD q bank balances $FEE_ICA_ADDR >>$TEMP_LOGS_DIR/$BALANCES_LOG
        fi

        if [ -n "$WITHDRAWAL_ICA_ADDR" ]; then 
            printf '\n%s\n' "BALANCES $chain (WITHDRAWAL ACCT)" >>$TEMP_LOGS_DIR/$BALANCES_LOG
            $HOST_MAIN_CMD q bank balances $WITHDRAWAL_ICA_ADDR >>$TEMP_LOGS_DIR/$BALANCES_LOG
        fi

    done

    if [ "$SET_UP_COMPLETE" = "false" ]; then
        ALL_COMPLETE=true
        for i in "${ICA_SET_UP[@]}"; do
            if [[ "$i" != "true" ]]; then
                ALL_COMPLETE=false
                break
            fi
        done
        if [ "$ALL_COMPLETE" = "true" ]; then
            SET_UP_COMPLETE=true
        fi
    fi
    if [ "$PRINT_SET_UP_MESSAGE" = "true" ]; then
        if [ "$SET_UP_COMPLETE" = "false" ] && [ "$PRINT_SET_UP_WARNING" = "true" ]; then
            echo "Warning: ICAs not registered yet. Wait before running itegration tests."
            PRINT_SET_UP_WARNING=false
        else
            echo "ICAs registered. Ok to run itegration tests."
            echo "Logs in $LOGS_DIR"
            PRINT_SET_UP_MESSAGE=false
        fi
    fi

    mv $TEMP_LOGS_DIR/*.log $LOGS_DIR
    sleep 3
done