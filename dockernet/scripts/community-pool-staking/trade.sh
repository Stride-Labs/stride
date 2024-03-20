#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# Simulates a trade by sending the native token to the trade account
# We'll send the amount that should have been sent from the ICA, which has the rebate excluded
trade_account=$($STRIDE_MAIN_CMD q stakeibc list-trade-routes | grep trade_account -A 3 | grep address | awk '{print $2}')
host_denom_on_trade=$($STRIDE_MAIN_CMD q stakeibc list-trade-routes | grep host_denom_on_trade | awk '{print $2}')
$OSMO_MAIN_CMD tx bank send ${OSMO_VAL_PREFIX}1 $trade_account 997500${host_denom_on_trade} --from ${OSMO_VAL_PREFIX}1 -y