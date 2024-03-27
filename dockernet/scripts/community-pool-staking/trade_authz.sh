#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

TRADE_AMOUNT=1000000

trade_account=$($STRIDE_MAIN_CMD q stakeibc list-trade-routes | grep trade_account -A 3 | grep address | awk '{print $2}')
host_denom_on_trade=$($STRIDE_MAIN_CMD q stakeibc list-trade-routes | grep host_denom_on_trade | awk '{print $2}')
reward_denom_on_trade=$($STRIDE_MAIN_CMD q stakeibc list-trade-routes | grep reward_denom_on_trade | awk '{print $2}')

echo "Granting authz permissions..."
$STRIDE_MAIN_CMD tx stakeibc toggle-trade-controller $OSMO_CHAIN_ID grant $(OSMO_ADDRESS) --from admin -y
sleep 15

tx_file=${STATE}/${OSMO_NODE_PREFIX}1/swap_tx.json
$OSMO_MAIN_CMD tx gamm swap-exact-amount-in ${TRADE_AMOUNT}${reward_denom_on_trade} 1 \
    --swap-route-pool-ids 1 --swap-route-denoms $host_denom_on_trade \
    --from $trade_account --generate-only > $tx_file
sleep 5

echo "Executing swap through authz..."
$OSMO_MAIN_CMD tx authz exec $tx_file --from ${OSMO_VAL_PREFIX}1 -y | TRIM_TX
sleep 1
rm -f $tx_file