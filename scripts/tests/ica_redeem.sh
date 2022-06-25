#!/bin/bash

echo "Register zone"
# setup logic on controller zone
$STR1_EXEC tx stakeibc register-host-zone \
    connection-0 uatom ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9 channel-0 --chain-id \
    STRIDE --home /stride/.strided --keyring-backend test \
    --from val1 --gas 500000 -y

echo "Sleeping for 30s"
sleep 30

echo "Host zones"
# store the delegate account
$STR1_EXEC q stakeibc list-host-zone

# host zone accounts
# gaiad keys list --home /gaia/.gaiad --keyring-backend test
# TODO(TEST-58): make this dynamic
VAL_KEY=cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2
DELEGATION_ADDR=cosmos10ltqave0ml70h9ynfsp6py2pv925xuzys7ypmffr8ud92sj09dzs6xtq8e

echo "Transferring tokens from $VAL_KEY to $DELEGATION_ADDR"
# transfer tokens to delegate account on the host zone
$GAIA1_EXEC tx bank send $VAL_KEY $DELEGATION_ADDR 100uatom --chain-id GAIA_1 --home /gaia/.gaiad --keyring-backend test -y

echo "Sleeping for 14s"
# wait 2 blocks
sleep 14

echo "Balance before staking"
# check balances
$GAIA1_EXEC q bank balances cosmos10ltqave0ml70h9ynfsp6py2pv925xuzys7ypmffr8ud92sj09dzs6xtq8e --home /gaia/.gaiad

echo "Sleeping for 14s"
# wait 4 blocks
sleep 30

echo "Balance after staking"
# check balances
$GAIA1_EXEC q bank balances cosmos10ltqave0ml70h9ynfsp6py2pv925xuzys7ypmffr8ud92sj09dzs6xtq8e --home /gaia/.gaiad

# check that tokens were staked after 10 blocks
$GAIA1_EXEC q staking delegations cosmos10ltqave0ml70h9ynfsp6py2pv925xuzys7ypmffr8ud92sj09dzs6xtq8e


# redeem stAssets for native tokens
STR1_EXEC tx stakeibc redeem-stake 1 statom \
    --chain-id STRIDE_1 --home /stride/.strided --keyring-backend test \
    --from val1 --gas 500000 -y

# query unbonding delegations, should reflect 1uatom
gaiad q staking unbonding-delegations $DELEGATION_ADDR
