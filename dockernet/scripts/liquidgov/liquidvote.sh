#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

staker_gaia_address=$($GAIA_MAIN_CMD keys show staker1 -a) 
staker_stride_address=$($STRIDE_MAIN_CMD keys show staker1 -a)



# Pull in proposal from hub to stride (using ICQ)

# Escrow user stTokens on stride for voting

# Cast a vote on stride using part of escrowed stTokens

# Withdraw extra stTokens not being used
