### REDEEM
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh

$STRIDE_MAIN_CMD tx stakeibc redeem-stake 89 CELESTIA $CELESTIA_RECEIVER_ADDRESS --from ${STRIDE_VAL_PREFIX}1 -y
