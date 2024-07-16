#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

AIRDROP_NAME="sttia"
OS=$(uname)

date_with_offset() {
	days=$1
	if [[ "$OS" == "Darwin" ]]; then
		date -u -v+"${days}d" +"%Y-%m-%dT%H:%M:%S"
	else
		date -u +"%Y-%m-%dT%H:%M:%S" -d "$days days"
	fi
}

echo ">>> Creating airdrop..."
$STRIDE_MAIN_CMD tx airdrop create-airdrop $AIRDROP_NAME \
  	--distribution-start-date  $(date_with_offset 0) \
	--distribution-end-date    $(date_with_offset 20) \
	--clawback-date            $(date_with_offset 30) \
	--claim-type-deadline-date $(date_with_offset 10) \
	--early-claim-penalty      0.5 \
	--distribution-address     $($STRIDE_MAIN_CMD keys show admin -a) \
	--from admin -y | TRIM_TX
sleep 5

echo -e "\n>>> Adding allocations..."
$STRIDE_MAIN_CMD tx airdrop add-allocations $AIRDROP_NAME ${SCRIPT_DIR}/allocations.csv \
    --from admin -y --gas 1000000 | TRIM_TX 
sleep 5

echo -e "\n>>> Airdrops:"
$STRIDE_MAIN_CMD q airdrop airdrops

echo -e "\n>>> Allocations:"
$STRIDE_MAIN_CMD q airdrop all-allocations $AIRDROP_NAME