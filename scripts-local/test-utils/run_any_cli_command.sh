### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

$GAIA_CMD q staking delegation $DELEGATION_ICA_ADDR $GAIA_DELEGATE_VAL