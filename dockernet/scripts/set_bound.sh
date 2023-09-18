### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

## Set the redemption rate bounds
# Set bounds from the correct account (success)
# $STRIDE_MAIN_CMD tx stakeibc set-redemption-rate-bounds GAIA 1 1.4 --from $STRIDE_ADMIN_ACCT -y | TRIM_TX
# Set bounds from the wrong account (fail)
# $STRIDE_MAIN_CMD tx stakeibc set-redemption-rate-bounds GAIA 1.1 1.3 --from val1 -y | TRIM_TX
# Set tight bound and observe halt
$STRIDE_MAIN_CMD tx stakeibc set-redemption-rate-bounds GAIA 1 1.000001 --from $STRIDE_ADMIN_ACCT -y | TRIM_TX
