### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

$UMEE_MAIN_CMD q bank balances umee1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6fh0uq
$STRIDE_MAIN_CMD q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
$STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 umee1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6fh0uq 1ustrd --from val1 -y 
sleep 10
$UMEE_MAIN_CMD q bank balances umee1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6fh0uq
$STRIDE_MAIN_CMD q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7