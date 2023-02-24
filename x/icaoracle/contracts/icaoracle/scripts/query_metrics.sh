set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

contract_address=$(cat $METADATA/contract_address.txt)
msg='{ "all_metrics" : {} }'

echo ">>> junod q wasm contract-state smart $contract_address $msg"
$JUNOD q wasm contract-state smart $contract_address "$msg"
