#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

### AIRDROP TESTING FLOW Pt 1 (STANDARD)

# This script tests airdrop claiming directly on Stride
# This covers the case where the host zone has a coin type of 118 
# and autopilot is disabled for the claim

# To run:
#   1. Start the network with `make start-docker`
#   2. Run this script with `bash dockernet/scripts/airdrop/airdrop1_standard.sh`

# NOTE: First, store the keys using the following mnemonics
echo "Registering accounts..."
# distributor address: stride1z835j3j65nqr6ng257q0xkkc9gta72gf48txwl
# distributor mnemonic: barrel salmon half click confirm crunch sense defy salute process cart fiscal sport clump weasel render private manage picture spell wreck hill frozen before
echo "barrel salmon half click confirm crunch sense defy salute process cart fiscal sport clump weasel render private manage picture spell wreck hill frozen before" | \
    $STRIDE_MAIN_CMD keys add distributor-test --recover

# airdrop-test address: stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# airdrop claimer mnemonic: royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate
echo "royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate" | \
    $STRIDE_MAIN_CMD keys add airdrop-test --recover

## AIRDROP SETUP
echo "Funding accounts..."
# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
sleep 5
# Fund the distributor account
$STRIDE_MAIN_CMD tx bank send val1 stride1z835j3j65nqr6ng257q0xkkc9gta72gf48txwl 600000ustrd --from val1 -y | TRIM_TX
sleep 5
# Fund the airdrop account
$STRIDE_MAIN_CMD tx bank send val1 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000000ustrd --from val1 -y | TRIM_TX
sleep 5
# Create the airdrop, so that the airdrop account can claim tokens
$STRIDE_MAIN_CMD tx claim create-airdrop gaia GAIA ustrd 1679715340 40000000 false --from distributor-test -y | TRIM_TX
sleep 5
# Set airdrop allocations
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations gaia stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1 --from distributor-test -y | TRIM_TX
sleep 5

# AIRDROP CLAIMS
# Check balances before claims
echo -e "\nInitial balance before claim [1000000000ustrd expected]:"
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr --denom ustrd
# NOTE: You can claim here using the CLI, or from the frontend!
# Claim 20% of the free tokens
echo -e "\nClaiming free amount..."
$STRIDE_MAIN_CMD tx claim claim-free-amount --from airdrop-test --gas 400000 -y | TRIM_TX
sleep 5
echo -e "\nBalance after claim [1000120000ustrd expected]:" 
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr --denom ustrd
# Stake, to claim another 20%
echo -e "\nStaking..."
$STRIDE_MAIN_CMD tx staking delegate stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f 100ustrd --from airdrop-test --gas 400000 -y | TRIM_TX
sleep 5
echo -e "\nBalance after stake [1000239900ustrd expected]:" 
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr --denom ustrd
# Liquid stake, to claim the final 60% of tokens
echo -e "\nLiquid staking..."
$STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 uatom --from airdrop-test --gas 400000 -y | TRIM_TX
sleep 5
echo -e "\nBalance after liquid stake [1000599900ustrd expected]:" 
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr --denom ustrd
