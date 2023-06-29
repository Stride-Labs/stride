#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

## Create Accounts
echo ">>> Registering Accounts:"
mnemonic1="match blade slide sort seven width degree february garden hospital valve odor scan exhaust bird chuckle age ozone timber claim office hurdle dance roast"
echo $mnemonic1 | $GAIA_MAIN_CMD keys add staker1 --recover --keyring-backend test
sleep 2
echo $mnemonic1 | $STRIDE_MAIN_CMD keys add staker1 --recover --keyring-backend test
sleep 2

# Staker #1 Address on Stride: stride1x92tnm6pfkl3gsfy0rfaez5myq5zh99a6a2w0p
# Staker #1 Address on GAIA:   cosmos1x92tnm6pfkl3gsfy0rfaez5myq5zh99aek2jmd

staker_gaia_address=$($GAIA_MAIN_CMD keys show staker1 -a) 
staker_stride_address=$($STRIDE_MAIN_CMD keys show staker1 -a) 

## Fund accounts
echo ">>> Fund staking accounts:"
$GAIA_MAIN_CMD tx bank send $($GAIA_MAIN_CMD keys show $RELAYER_GAIA_ACCT -a) $staker_gaia_address 200000000uatom --from $RELAYER_GAIA_ACCT -y | TRIM_TX 
sleep 10 && echo ""

# IBC transfer some tokens to Stride side to initialize the address
echo ">>> IBC Transfer to init Stride address:"
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $staker_stride_address 1000000uatom --from $RELAYER_GAIA_ACCT -y | TRIM_TX
sleep 12 && echo ""


echo "Bank balances:"
$GAIA_MAIN_CMD q bank balances $staker_gaia_address 
$STRIDE_MAIN_CMD q bank balances $staker_stride_address
sleep 3 && echo ""




# Setup a new proposal on the hub
echo ">>> Creating proposals and funding them with deposits:"
$GAIA_MAIN_CMD tx gov submit-proposal --proposal="dockernet/scripts/liquidgov/prop1.json" --from staker1 -y | TRIM_TX
sleep 6 && echo ""
$GAIA_MAIN_CMD tx gov submit-proposal --proposal="dockernet/scripts/liquidgov/prop2.json" --from staker1 -y | TRIM_TX
sleep 6 && echo ""

# Deposit enough to make them active proposals
$GAIA_MAIN_CMD tx gov deposit 1 10000000uatom --from staker1 -y | TRIM_TX
sleep 10 && echo ""

$GAIA_MAIN_CMD query gov proposals
sleep 2 && echo ""

