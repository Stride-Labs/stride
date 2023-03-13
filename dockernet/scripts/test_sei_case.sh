### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

## IBC ATOM from OSMO to STRIDE
$OSMO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 1000000uosmo --from ${OSMO_VAL_PREFIX}1 -y 
sleep 3
## LIQUID STAKE IT
$STRIDE_MAIN_CMD tx stakeibc liquid-stake 10000 $OSMO_DENOM --from sval1_test -y --keyring-backend os
sleep 3
## CHECK WE GOT stOSMO
$STRIDE_MAIN_CMD q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7