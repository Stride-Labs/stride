#/bin/bash
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

echo ">>> Redeeming stake..."
$STRIDE_MAIN_CMD tx staketia redeem-stake 1000000 --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
sleep 1