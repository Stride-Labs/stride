#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

AIRDROP_NAME="sttia"

echo ">>> Creating airdrop..."
$STRIDE_MAIN_CMD tx airdrop create-airdrop $AIRDROP_NAME \
  	--distribution-start-date  2024-07-01 \
	--distribution-end-date    2024-10-01 \
	--clawback-date            2024-11-01 \
	--claim-type-deadline-date 2024-08-01 \
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