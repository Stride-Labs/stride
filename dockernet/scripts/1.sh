### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

## IBC ATOM from GAIA to STRIDE
$UMEE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 1000000uumee --from ${UMEE_VAL_PREFIX}1 -y 
sleep 7
$STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS)
