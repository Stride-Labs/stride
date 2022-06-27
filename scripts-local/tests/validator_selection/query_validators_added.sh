SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/local_vars.sh

# check that the right validators are added, with the correct weight
$STRIDE_CMD q stakeibc show-host-zone GAIA