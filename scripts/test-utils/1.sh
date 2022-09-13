### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

## GAIA
#  ibc over atoms to stride
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 10
$STRIDE_MAIN_CMD q bank balances $STRIDE_ADDRESS
