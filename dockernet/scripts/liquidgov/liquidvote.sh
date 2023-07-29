#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh


# Pull in proposal from hub to stride (using ICQ)

# Escrow user stTokens on stride for voting

# Cast a vote on stride using part of escrowed stTokens

# Withdraw extra stTokens not being used
