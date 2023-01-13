### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

## IBC ATOM from GAIA to STRIDE
# $GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
# sleep 10
# $STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS)

CHAIN_NAME=EVMOS TRANSFER_CHANNEL_NUMBER=0
HOST_MAIN_CMD=$(GET_VAR_VALUE  ${CHAIN_NAME}_MAIN_CMD)
HOST_CHAIN_ID=$(GET_VAR_VALUE  ${CHAIN_NAME}_CHAIN_ID)

GETSTAKE() {
    tail -n 2 | head -n 1 | grep -o -E '[0-9]+' | head -n 1
}
# echo "ONE"
# $HOST_MAIN_CMD q
# echo "TWO"
# # works
# ica=$(GET_ICA_ADDR $HOST_CHAIN_ID delegation)
# echo $ica
# echo "THREE"
# echo $(GET_VAL_ADDR EVMOS 1)
# $($HOST_MAIN_CMD q staking delegation $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) $(GET_VAL_ADDR $HOST_CHAIN_ID 1) | GETSTAKE)

# NEW_STAKE=$($HOST_MAIN_CMD q staking delegation $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) $(GET_VAL_ADDR $HOST_CHAIN_ID 1) | GETSTAKE)

$STRIDE_CMD q stakeibc show-host-zone