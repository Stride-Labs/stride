## SUBMIT MANY SMALL LIQ STAKES

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../account_vars.sh

# liquid stake many times in sequence to flood the deposit record queue, in order to test the max staking ICA calls limit
for I in {1..50}
do
    $STRIDE_CMD tx stakeibc liquid-stake 1 $ATOM_DENOM --keyring-backend test --from $STRIDE_VAL_ACCT -y --chain-id $STRIDE_CHAIN -y
    echo "Waiting for liquid staked tokens to be delegated..."
    CSLEEP 10
done
