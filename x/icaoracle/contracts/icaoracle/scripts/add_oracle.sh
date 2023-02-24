set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

echo "Adding oracle..."
echo ">>> strided tx icaoracle add-oracle connection-0"
tx_hash=$($STRIDED tx icaoracle add-oracle connection-0 --from admin -y | grep -E "txhash:" | awk '{print $2}') 

echo "Tx Hash: $tx_hash"
echo $tx_hash > $METADATA/add_oracle_tx_hash.txt

sleep 10
echo ""

echo "Instantiating contract..."
code_id=$(cat $METADATA/code_id.txt)
echo ">>> strided tx icaoracle instantiate-oracle JUNO $code_id"
tx_hash=$($STRIDED tx icaoracle instantiate-oracle JUNO $code_id --from admin -y | grep -E "txhash:" | awk '{print $2}') 

echo "Tx Hash: $tx_hash"
echo $tx_hash > $METADATA/instantiate_oracle_tx_hash.txt

sleep 10

contract_address=$($STRIDED q icaoracle oracles | grep contract_address | awk '{print $2}')
echo "Contract Address: $contract_address"
echo $contract_address > $METADATA/contract_address.txt

