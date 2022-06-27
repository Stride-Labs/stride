### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/local_vars.sh

# $STRIDE_CMD q stakeibc list-host-zone
# $GAIA_CMD q staking validators
$STRIDE_CMD tx stakeibc add-validator GAIA gval1 $GAIA_DELEGATE_VAL 0 10 
