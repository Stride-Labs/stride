### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

$STRIDE_CMD tx stakeibc liquid-stake 1000 $ATOM --keyring-backend test --from $STRIDE_VAL_ACCT -y --chain-id $STRIDE_CHAIN
