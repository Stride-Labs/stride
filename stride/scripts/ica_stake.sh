#!/bin/bash

echo "Register zone"
# setup logic on controller zone
$STR1_EXEC tx stakeibc register-host-zone \
    connection-0 uatom statom --chain-id \
    STRIDE_1 --home /stride/.strided --keyring-backend test \
    --from val1 --gas 500000 -y

echo "Sleeping for 30s"
# wait 10 blocks
sleep 30

echo "Host zones"
# store the delegate account
$STR1_EXEC q stakeibc list-host-zone

# host zone accounts
# gaiad keys list --home /gaia/.gaiad --keyring-backend test
VAL_KEY=cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2
DELEGATION_ADDR=cosmos10ltqave0ml70h9ynfsp6py2pv925xuzys7ypmffr8ud92sj09dzs6xtq8e

echo "Transferring tokens from $VAL_KEY to $DELEGATION_ADDR"
# transfer tokens to delegate account on the host zone
$GAIA1_EXEC tx bank send $VAL_KEY $DELEGATION_ADDR 100uatom --chain-id GAIA_1 --home /gaia/.gaiad --keyring-backend test -y

echo "Sleeping for 7s"
# wait 1 blocks
sleep 14

echo "Balance before staking"
# check balances
$GAIA1_EXEC q bank balances cosmos10ltqave0ml70h9ynfsp6py2pv925xuzys7ypmffr8ud92sj09dzs6xtq8e --home /gaia/.gaiad

echo "Sleeping for 80s"
# wait 10 blocks
sleep 100

echo "Balance after staking"
# check balances
$GAIA1_EXEC q bank balances cosmos10ltqave0ml70h9ynfsp6py2pv925xuzys7ypmffr8ud92sj09dzs6xtq8e --home /gaia/.gaiad

# check that tokens were staked after 10 blocks
$GAIA1_EXEC q staking delegations cosmos10ltqave0ml70h9ynfsp6py2pv925xuzys7ypmffr8ud92sj09dzs6xtq8e
