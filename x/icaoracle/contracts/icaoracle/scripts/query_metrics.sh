set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

contract_address=$(cat $METADATA/contract_address.txt)

echo "ALL_LATEST_METRICS"
msg='{ "all_latest_metrics" : { } }'
echo ">>> junod q wasm contract-state smart $contract_address $msg"
$JUNOD q wasm contract-state smart $contract_address "$msg"
sleep 1

printf "\nHISTORICAL_METRICS\n"
msg='{ "historical_metrics" : { "key": "stujuno_redemption_rate" } }'
echo ">>> junod q wasm contract-state smart $contract_address $msg"
$JUNOD q wasm contract-state smart $contract_address "$msg"
sleep 1

printf "\nPRICES\n"
msg='{ "price" : { "denom": "stujuno", "base_denom": "ibc/04F5F501207C3626A2C14BFEF654D51C2E0B8F7CA578AB8ED272A66FE4E48097" } }'
echo ">>> junod q wasm contract-state smart $contract_address $msg"
$JUNOD q wasm contract-state smart $contract_address "$msg"
sleep 1