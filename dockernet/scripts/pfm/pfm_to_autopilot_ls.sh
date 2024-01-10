#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

########################################################################################
#     PFM (a -> b -> autopilot-LS-on-a: stride -> osmosis -> stride (autopilot LS)     #
########################################################################################
# docs: https://github.com/cosmos/ibc-apps/tree/main/middleware/packet-forward-middleware#full-example---chain-forward-a-b-c-d-with-retry-on-timeout

# to verify pfm works:
# - open two terminal windows as described below:
    # - terminal 1 monitors the sender on stride: `while true; do stridedl q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 && echo "--------------" && sleep 1; done`
    # - terminal 2 monitors the final receiver on gaia: `while true; do gaiadl q bank balances cosmos1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgl2scj && echo "--------------" && sleep 1; done`
# - watch the terminal windows to see the receiver's stuatom balance increase after the sender's ibc/atom bal decreases.
# - run `sh dockernet/scripts/pfm/pfm_to_autopilot_ls.sh`

## Prep: fund account by ibc'ing atom from gaia to stride
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 100uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 10
$STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS)


STRIDE_SENDER=$($STRIDE_MAIN_CMD keys show ${STRIDE_VAL_PREFIX}1 --keyring-backend test -a | grep $STRIDE_ADDRESS_PREFIX)
STRIDE_LIQUID_STAKER=$($STRIDE_MAIN_CMD keys show ${STRIDE_VAL_PREFIX}2 --keyring-backend test -a | grep $STRIDE_ADDRESS_PREFIX)
GAIA_RECEIVER=$(GAIA_ADDRESS)
echo $STRIDE_SENDER, $STRIDE_LIQUID_STAKER, $GAIA_RECEIVER

CHANNEL_A_TO_B="channel-1" # channel 1: stride -> osmosis
CHANNEL_B_TO_C="channel-0" # channel 0: osmosis -> stride
INVALID_RECEIVER_CHAINB="pfm" # purposely using invalid bech32 here
TIMEOUT="10m"
RETRIES="2"

autopilot_memo="{\"autopilot\":{\"receiver\":\"$STRIDE_LIQUID_STAKER\",\"stakeibc\":{\"action\":\"LiquidStake\",\"ibc_receiver\":\"$GAIA_RECEIVER\"}}}"

memo="{\"forward\":{\"receiver\":\"$STRIDE_LIQUID_STAKER\",\"port\":\"transfer\",\"channel\":\"$CHANNEL_B_TO_C\",\"timeout\":\"$TIMEOUT\",\"retries\":$RETRIES,\"next\":\"$autopilot_memo\"}}"

echo $memo

# pfm transfer
IBC_VOUCHER_ATOM_ON_STRIDE="ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
$STRIDE_MAIN_CMD tx ibc-transfer transfer transfer $CHANNEL_A_TO_B "pfm" 1$IBC_VOUCHER_ATOM_ON_STRIDE --memo "$memo" --from ${STRIDE_VAL_PREFIX}1 -y