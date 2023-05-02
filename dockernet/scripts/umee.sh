SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

$STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 umee1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6fh0uq 777ustrd --from val1 -y