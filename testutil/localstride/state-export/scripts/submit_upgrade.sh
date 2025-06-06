#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Confirm an upgrade name and height were provided
USAGE_INSTRUCTION="  Please provide 'upgrade_name' as an argument\n  e.g. 'make localnet-state-export-upgrade upgrade_name=v5'\n"
if [ -z "${upgrade_name:-}" ]; then
    echo "ERROR: 'upgrade_name' not provided."
    printf "$USAGE_INSTRUCTION"
    exit 1
fi

# Helper function to clean logs after a transaction
TRIM_TX() {
  grep -E "code:|txhash:" | sed 's/^/  /'
}

STRIDE_MAIN_CMD="docker compose -f ${SCRIPT_DIR}/../docker-compose.yml exec -it stride1 strided"

current_height=$($STRIDE_MAIN_CMD status | jq -r '.SyncInfo.latest_block_height')
upgrade_height=$((current + 75))

printf "PROPOSAL\n"
$STRIDE_MAIN_CMD tx gov submit-legacy-proposal software-upgrade $upgrade_name \
    --title $upgrade_name --description "upgrade" --upgrade-info "test" --no-validate \
    --upgrade-height $upgrade_height --from val --deposit 1000000000ustrd -y | TRIM_TX

sleep 5
printf "\nPROPOSAL CONFIRMATION\n"
proposal_id=$($STRIDE_MAIN_CMD q gov proposals | grep 'id:' | tail -1 | awk '{printf $2}' | tr -d '"')
$STRIDE_MAIN_CMD query gov proposal $proposal_id
echo "Proposal ID: $proposal_id"
sleep 10

sleep 5 
printf "\nDEPOSIT\n"
$STRIDE_MAIN_CMD tx gov deposit $proposal_id 20000000001ustrd --from val -y | TRIM_TX

sleep 5
printf "\nDEPOSIT CONFIRMATION\n"
$STRIDE_MAIN_CMD query gov deposits $proposal_id

sleep 5
printf "\nVOTING\n"
$STRIDE_MAIN_CMD tx gov vote $proposal_id yes --from val -y | TRIM_TX

sleep 5
printf "\nVOTE CONFIRMATION\n"
$STRIDE_MAIN_CMD query gov tally $proposal_id

printf "\nPROPOSAL STATUS\n"
while true; do
    status=$($STRIDE_MAIN_CMD query gov proposal $proposal_id | grep "status" | awk '{printf $2}')
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
