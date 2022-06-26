### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

## GAIA
#  ibc over atoms to stride
STRIDE_ADDRESS=$($STRIDE_CMD keys show $STRIDE_VAL_ACCT --keyring-backend test -a)

$GAIA_CMD tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 100000uatom --from $GAIA_VAL_ACCT --chain-id $GAIA_CHAIN -y --keyring-backend test
CSLEEP 5
$STRIDE_CMD q bank balances $STRIDE_ADDRESS
