#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

###############################################
#   Minimal Example - Chain forward A->B->C   #
###############################################
# docs: https://github.com/cosmos/ibc-apps/tree/main/middleware/packet-forward-middleware#minimal-example---chain-forward-a-b-c

# to verify pfm works:
# - open two terminal windows as described below
    # - terminal 1 monitors the sender on A: `while true; do gaiadl q bank balances $(gaiadl keys show gval1 -a)  && echo "--------------" && sleep 1; done`
    # - terminal 2 monitors the receiver on C: `while true; do osmosisdl q bank balances osmo1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrqyeqwq  && echo "--------------" && sleep 1; done`
# - run `sh dockernet/scripts/pfm/pfm_abc_default.sh`
# - watch the terminal windows to see C's balance increase after A's descreases.

FINAL_RECEIVER=$(OSMO_ADDRESS)
CHANNEL_B_TO_C="channel-1"
INVALID_RECEIVER_CHAINB="pfm" # purposely using invalid bech32 here
TIMEOUT="10m"
RETRIES="2"

CHANNEL_BETWEEN_A_AND_B="channel-0"

# note that the memo must be escaped as follows to work from this script.
memo="{\"forward\":{\"receiver\":\"$FINAL_RECEIVER\",\"port\":\"transfer\",\"channel\":\"$CHANNEL_B_TO_C\",\"timeout\":\"$TIMEOUT\",\"retries\":$RETRIES}}"
echo $memo

# pfm transfer
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer $CHANNEL_A_TO_B "pfm" 1uatom --memo "$memo" --from ${GAIA_VAL_PREFIX}1 -y