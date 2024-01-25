#!/bin/bash
# clean up logs one by one before creation (allows auto-updating logs with the command `while true; do make init build=logs ; sleep 5 ; done`)

set -eu
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source ${SCRIPT_DIR}/../config.sh

LOGS_DIR=$DOCKERNET_HOME/logs
TEMP_LOGS_DIR=$LOGS_DIR/temp

STATE_LOG=state.log
BALANCES_LOG=balances.log
CHANNELS_LOG=channels.log

mkdir -p $TEMP_LOGS_DIR

# Temp interediate files
state=$TEMP_LOGS_DIR/$STATE_LOG
balances=$TEMP_LOGS_DIR/$BALANCES_LOG
channels=$TEMP_LOGS_DIR/$CHANNELS_LOG

print_separator() {
    header="$1"
    file_name="$2"
    printf '\n%s\n' "==========================  $header  =============================" >> $file_name
}

print_header() {
    header="$1"
    file_name="$2"
    printf '\n%s\n' "$header" >> $file_name
}

print_stride_balance() {
    address="$1"
    name="$2"

    print_header "$name" $balances
    $STRIDE_MAIN_CMD q bank balances $address | grep -vE "pagination|total|next_key" >> $balances
}

print_host_balance() {
    chain="$1"
    address="$2"
    name="$3"

    print_header "$name" $balances
    main_cmd=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
    $main_cmd q bank balances $address | grep -vE "pagination|total|next_key" >> $balances
}

while true; do
    N_VALIDATORS_STRIDE=$($STRIDE_MAIN_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
    echo "STRIDE @ $($STRIDE_MAIN_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_STRIDE VALS" > $state
    echo "STRIDE @ $($STRIDE_MAIN_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_STRIDE VALS" >  $balances
    echo "STRIDE @ $($STRIDE_MAIN_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_STRIDE VALS" > $channels

    for chain in ${HOST_CHAINS[@]}; do
        HOST_MAIN_CMD=$(GET_VAR_VALUE ${chain}_MAIN_CMD)

        N_VALIDATORS_HOST=$($HOST_MAIN_CMD q tendermint-validator-set | grep -o address | wc -l | tr -dc '0-9')
        echo "$chain   @ $($HOST_MAIN_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_HOST VALS" >> $state
        echo "$chain   @ $($HOST_MAIN_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9') | $N_VALIDATORS_HOST VALS" >> $balances
    done

    # Log host zone, records and other state
    print_separator "STAKEIBC" $state

    print_header "HOST ZONES" $state
    $STRIDE_MAIN_CMD q stakeibc list-host-zone >> $state
    print_header "DEPOSIT RECORDS" $state                           
    $STRIDE_MAIN_CMD q records list-deposit-record >> $state
    print_header "EPOCH UNBONDING RECORDS" $state
    $STRIDE_MAIN_CMD q records list-epoch-unbonding-record >> $state
    print_header "USER REDEMPTION RECORDS" $state
    $STRIDE_MAIN_CMD q records list-user-redemption-record >> $state
    print_header "LSM TOKEN DEPOSIT RECORDS" $state
    $STRIDE_MAIN_CMD q records lsm-deposits >> $state
    print_header "TRADE ROUTES" $state
    $STRIDE_MAIN_CMD q stakeibc list-trade-routes >> $state

    print_separator "STAKETIA" $state

    print_header "HOST ZONE" $state
    $STRIDE_MAIN_CMD q staketia host-zone >> $state
    print_header "DELEGATION RECORDS" $state
    $STRIDE_MAIN_CMD q staketia delegation-records >> $state
    print_header "UNBONDING RECORDS" $state
    $STRIDE_MAIN_CMD q staketia unbonding-records >> $state
    print_header "REDEMPTION RECORDS" $state
    $STRIDE_MAIN_CMD q staketia redemption-records >> $state
    print_header "SLASH RECORDS" $state
    $STRIDE_MAIN_CMD q staketia slash-records >> $state

    # Log stride stakeibc balances
    print_separator "VALIDATORS" $balances
    host_chain="${HOST_CHAINS[0]}"
    host_val_address="$(${host_chain}_ADDRESS)"
    print_stride_balance $(STRIDE_ADDRESS) "STRIDE" 
    print_host_balance $host_chain $host_val_address $host_chain

    # Log stride staketia balances
    print_separator "STAKETIA STRIDE" $balances

    deposit_address=$($STRIDE_MAIN_CMD keys show -a deposit)
    redemption_address=$($STRIDE_MAIN_CMD keys show -a redemption)
    claim_address=$($STRIDE_MAIN_CMD keys show -a claim)
    fee_address=$($STRIDE_MAIN_CMD q staketia host-zone | grep fee_address | awk '{print $2}')

    print_stride_balance $deposit_address    "DEPOSIT" 
    print_stride_balance $redemption_address "REDEMPTION" 
    print_stride_balance $claim_address      "CLAIM" 
    print_stride_balance $fee_address        "FEE" 

    # Log staketia balance on host chain
    print_separator "STAKETIA HOST" $balances
    print_host_balance "${HOST_CHAINS[0]}" $DELEGATION_ADDRESS  "DELEGATION CONTROLLER" 
    print_host_balance "${HOST_CHAINS[0]}" $REWARD_ADDRESS "REWARD CONTROLLER" 

    # Log staketia delegations/undelegations
    print_separator "STAKETIA STAKING" $balances
    delegation_address=$($STRIDE_MAIN_CMD q staketia host-zone | grep "delegation_address" | awk '{print $2}')

    print_header "DELEGATIONS $chain" $balances
    $HOST_MAIN_CMD q staking delegations $delegation_address | grep -vE "pagination|total|next_key" >> $balances
    print_header "UNBONDING-DELEGATIONS $chain" $balances
    $HOST_MAIN_CMD q staking unbonding-delegations $delegation_address | grep -vE "pagination|total|next_key" >> $balances

    # Log stride channels
    print_separator "STRIDE" $channels
    $STRIDE_MAIN_CMD q ibc channel channels | grep -E "channel_id|port|state" >> $channels || true

    for chain in ${HOST_CHAINS[@]}; do
        HOST_CHAIN_ID=$(GET_VAR_VALUE ${chain}_CHAIN_ID)
        HOST_MAIN_CMD=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
        
        delegation_ica_address=$(GET_ICA_ADDR $HOST_CHAIN_ID delegation)
        redemption_ica_address=$(GET_ICA_ADDR $HOST_CHAIN_ID redemption)
        withdrawal_ica_address=$(GET_ICA_ADDR $HOST_CHAIN_ID withdrawal)
        fee_ica_address=$(GET_ICA_ADDR $HOST_CHAIN_ID fee)

        community_pool_deposit_address=$(GET_ICA_ADDR $HOST_CHAIN_ID community_pool_deposit)
        community_pool_return_address=$(GET_ICA_ADDR $HOST_CHAIN_ID community_pool_return)

        community_pool_stake_address=$(GET_HOST_ZONE_FIELD $HOST_CHAIN_ID community_pool_stake_holding_address)
        community_pool_redeem_address=$(GET_HOST_ZONE_FIELD $HOST_CHAIN_ID community_pool_redeem_holding_address)

        # Log delegations/undelegations
        print_separator "STAKEIBC STAKING" $balances

        print_header "DELEGATIONS $chain" $balances
        $HOST_MAIN_CMD q staking delegations $delegation_ica_address | grep -vE "pagination|total|next_key" >> $balances
        print_header "UNBONDING-DELEGATIONS $chain" $balances
        $HOST_MAIN_CMD q staking unbonding-delegations $delegation_ica_address | grep -vE "pagination|total|next_key" >> $balances

        # Log ICA balances
        print_separator "ICA BALANCES" $balances
        print_host_balance $chain $delegation_ica_address "DELEGATION-ICA" 
        print_host_balance $chain $redemption_ica_address "REDEMPTION-ICA" 
        print_host_balance $chain $fee_ica_address        "FEE-ICA"        
        print_host_balance $chain $withdrawal_ica_address "WITHDRAWAL-ICA" 

        # Log balances for community pool liquid staking
        print_separator "COMMUNITY POOL BALANCES" $balances

        print_header "COMMUNITY POOL BALANCE" $balances
        $HOST_MAIN_CMD q distribution community-pool >> $balances
 
        print_host_balance $chain $community_pool_deposit_address "COMMUNITY POOL DEPOSIT ACCT BALANCE" 
        print_host_balance $chain $community_pool_return_address "COMMUNITY POOL RETURN ACCT BALANCE" 

        print_stride_balance $community_pool_stake_address "COMMUNITY POOL STAKE HOLDING ACCT BALANCE" 
        print_stride_balance $community_pool_redeem_address "COMMUNITY POOL REDEEM HOLDING ACCT BALANCE" 

        # Log host channels
        print_separator "$chain" $channels
        $HOST_MAIN_CMD q ibc channel channels | grep -E "channel_id|port|state" >> $channels || true
    done


    TRADE_ICA_ADDR=$($STRIDE_MAIN_CMD q stakeibc list-trade-routes | grep trade_account -A 2 | grep address | awk '{print $2}')
    if [[ "$TRADE_ICA_ADDR" == "$OSMO_ADDRESS_PREFIX"* ]]; then
        print_header "TRADE ACCT BALANCE" >> $balances
        $OSMO_MAIN_CMD q bank balances $TRADE_ICA_ADDR >> $balances
    fi

    for chain in ${ACCESSORY_CHAINS[@]:-}; do
        ACCESSORY_MAIN_CMD=$(GET_VAR_VALUE ${chain}_MAIN_CMD)
        print_header "==========================  $chain  =============================" >> $channels
        $ACCESSORY_MAIN_CMD q ibc channel channels | grep -E "channel_id|port|state" >> $channels || true
    done

    mv $TEMP_LOGS_DIR/*.log $LOGS_DIR
    sleep 3
done