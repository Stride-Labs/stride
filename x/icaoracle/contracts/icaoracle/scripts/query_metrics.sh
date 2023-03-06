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
msg='{ "historical_metrics" : { "key": "statom_redemption_rate" } }'
echo ">>> junod q wasm contract-state smart $contract_address $msg"
$JUNOD q wasm contract-state smart $contract_address "$msg"
sleep 1

printf "\nPRICES\n"
msg='{ "price" : { "denom": "statom", "base_denom": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2" } }'
echo ">>> junod q wasm contract-state smart $contract_address $msg"
$JUNOD q wasm contract-state smart $contract_address "$msg"
sleep 1