### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/local_vars.sh

## GAIA
#  ibc over atoms to stride
$GAIA_CMD tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 100000uatom --from $GAIA_VAL_ACCT --chain-id GAIA -y --keyring-backend test
CSLEEP 10
$STRIDE_CMD q bank balances $STRIDE_ADDRESS
