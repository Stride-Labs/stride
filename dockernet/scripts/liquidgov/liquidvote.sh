#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

staker_gaia_address=$($GAIA_MAIN_CMD keys show staker1 -a) 
staker_stride_address=$($STRIDE_MAIN_CMD keys show staker1 -a)



# Pull in proposal id1 from GAIA hub to stride (using ICQ)
echo ">>> Trigger ICQ to pull in proposal from GAIA to Stride:"
$STRIDE_MAIN_CMD tx liquidgov update-proposal GAIA 1 --from staker1 -y | TRIM_TX
sleep 10 && echo ""


# setup.sh should have prepared staker1 to have 1000000 st${ATOM_DENOM} in stride balance

# Escrow 500000 stTokens from user on stride for voting
echo ">>> Deposit 500000 stuatom for voting escrow:"
$STRIDE_MAIN_CMD tx liquidgov deposit-voting-stake st$ATOM_DENOM 500000 --from staker1 -y | TRIM_TX
sleep 5 && echo ""
echo "STRIDE balances:"
$STRIDE_MAIN_CMD q bank balances $staker_stride_address
sleep 3 && echo ""

# Cast a vote on stride using 200000 of escrowed stTokens


# Withdraw extra 300000 stTokens not being used
echo ">>> Withdraw 300000 stuatom from voting escrow:"
$STRIDE_MAIN_CMD tx liquidgov withdraw-voting-stake st$ATOM_DENOM 300000 --from staker1 -y | TRIM_TX
sleep 5 && echo ""
echo "STRIDE balances:"
$STRIDE_MAIN_CMD q bank balances $staker_stride_address
sleep 3 && echo ""
