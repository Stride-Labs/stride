#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

UPGRADE_HEIGHT="${UPGRADE_HEIGHT:-150}"

PROPOSAL_ID=1

printf "PROPOSAL\n"
$STRIDE_MAIN_CMD tx gov submit-legacy-proposal software-upgrade $UPGRADE_NAME \
    --title $UPGRADE_NAME --description "description" --no-validate \
    --upgrade-height $UPGRADE_HEIGHT --from val1 -y | TRIM_TX

sleep 5
printf "\nPROPOSAL CONFIRMATION\n"
$STRIDE_MAIN_CMD query gov proposals

sleep 5 
printf "\nDEPOSIT\n"
$STRIDE_MAIN_CMD tx gov deposit $PROPOSAL_ID 10000001ustrd --from val1 -y | TRIM_TX

sleep 5
printf "\nDEPOSIT CONFIRMATION\n"
$STRIDE_MAIN_CMD query gov deposits $PROPOSAL_ID

sleep 5
printf "\nVOTING\n"
$STRIDE_MAIN_CMD tx gov vote $PROPOSAL_ID yes --from val1 -y | TRIM_TX
$STRIDE_MAIN_CMD tx gov vote $PROPOSAL_ID yes --from val2 -y | TRIM_TX
$STRIDE_MAIN_CMD tx gov vote $PROPOSAL_ID yes --from val3 -y | TRIM_TX

sleep 5
printf "\nVOTE CONFIRMATION\n"
$STRIDE_MAIN_CMD query gov tally $PROPOSAL_ID

printf "\nPROPOSAL STATUS\n"
while true; do
    status=$($STRIDE_MAIN_CMD query gov proposal $PROPOSAL_ID | grep "status" | awk '{printf $2}')
    if [[ "$status" == "PROPOSAL_STATUS_VOTING_PERIOD" ]]; then
        echo "Proposal still in progress..."
        sleep 5
    elif [[ "$status" == "PROPOSAL_STATUS_PASSED" ]]; then
        echo "Proposal passed!"
        exit 0
    elif [[ "$status" == "PROPOSAL_STATUS_REJECTED" ]]; then
        echo "Proposal Failed!"
        exit 1
    else 
        echo "Unknown proposal status: $status"
        exit 1
    fi
done
