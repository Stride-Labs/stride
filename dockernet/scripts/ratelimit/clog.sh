CURRENT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${CURRENT_DIR}/../../config.sh
source ${CURRENT_DIR}/common.sh

# Sends a tx that will clog 15% capacity, leaving only 5% remaining
echo "Clogging capacity for stuatom"
$STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(GAIA_ADDRESS) 15000000stuatom --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
