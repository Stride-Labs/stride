#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/../config.sh

# register host zone
# $STRIDE_MAIN_CMD tx stakeibc register-host-zone connection-localhost ustrd stride ibc/F4EF72139CDCF526FFD520402C53FFB7A927B528BDF0C5FF60A4D8B4780D2C6B channel-5 1 --gas 1000000 --from admin


echo "Stride - Registering validators..."
validator_json=$DOCKERNET_HOME/src/stride_vals.json

# Add host zone validators to Stride's host zone struct
$STRIDE_MAIN_CMD tx stakeibc add-validators "stride-1" $validator_json \
    --from $STRIDE_ADMIN_ACCT -y | TRIM_TX





