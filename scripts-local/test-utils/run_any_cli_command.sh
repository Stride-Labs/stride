### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

echo $STRIDE_VAL_MNEMONIC | $STRIDE_CMD keys add $STRIDE_VAL_ACCT --recover --keyring-backend=test 
exit
$STRIDE_CMD keys show $STRIDE_VAL_ACCT --keyring-backend test -a $DELEGATION_ICA_ADDR cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne
