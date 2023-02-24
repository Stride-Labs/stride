set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

CONTRACT=./artifacts/ica_oracle.wasm

echo "Storing contract..."

echo ">>> junod tx wasm store $CONTRACT"
tx_hash=$($JUNOD tx wasm store $CONTRACT $GAS --from jval1 -y | grep -E "txhash:" | awk '{print $2}') 

echo "Tx Hash: $tx_hash"
echo $tx_hash > $METADATA/store_tx_hash.txt

sleep 3

code_id=$($JUNOD q tx $tx_hash | grep code_id -m 1 -A 1 | tail -1 | awk '{print $2}' | tr -d '"')
echo "Code ID: $code_id"
echo $code_id > $METADATA/code_id.txt