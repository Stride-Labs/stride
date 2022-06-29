SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/local_vars.sh

# check that the right validators are added, with the correct weight
# $STRIDE_CMD tx stakeibc change-validator-weight GAIA $GAIA_DELEGATE_VAL 15 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
# sleep 1
$STRIDE_CMD q stakeibc show-host-zone GAIA