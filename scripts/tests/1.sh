### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

## GAIA
#  ibc over atoms to stride
$GAIA1_EXEC tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS_1 100000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
CSLEEP 20
$STR1_EXEC q bank balances $STRIDE_ADDRESS_1
