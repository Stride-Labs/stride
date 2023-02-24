set -eu

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

METADATA=${SCRIPT_DIR}/metadata
mkdir -p $METADATA

JUNO_HOME=${STRIDE_HOME}/dockernet/state/juno1
JUNOD="${STRIDE_HOME}/build/junod --home ${STRIDE_HOME}/dockernet/state/juno1"
STRIDED="${STRIDE_HOME}/build/strided --home ${STRIDE_HOME}/dockernet/state/stride1"

GAS="--gas-prices 0.1ujuno --gas auto --gas-adjustment 1.3"
