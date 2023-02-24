set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

echo ">>> junod q txs --events 'ics27_packet.module=interchainaccounts' and 'ics_27packet.error=*'"
$JUNOD q txs --events 'ics27_packet.module=interchainaccounts' and 'ics_27packet.error=*'
