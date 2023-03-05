set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

contract_address=$(cat $METADATA/contract_address.txt)

echo "Adding metric..."

key=${KEY:-key1}
value=${VALUE:-value1}
msg=$(cat ${SCRIPT_DIR}/post_metric.json)

echo ">>> junod tx wasm execute $contract_address $msg"
tx_hash=$($JUNOD tx wasm execute $contract_address "$msg" --from jval1 -y | grep -E "txhash:" | awk '{print $2}')

echo "Tx Hash: $tx_hash"
echo $tx_hash > $METADATA/store_tx_hash.txt