#!/bin/bash

# NOTE: This script should be run locally (outside of the k8s pod)

set -eu

NAMESPACE=integration

EXEC0="kubectl exec -it stride-validator-0 -c validator -n $NAMESPACE -- "
STRIDED0="kubectl exec -it stride-validator-0 -c validator -n $NAMESPACE -- strided"
STRIDED1="kubectl exec -it stride-validator-1 -c validator -n $NAMESPACE -- strided"
STRIDED2="kubectl exec -it stride-validator-2 -c validator -n $NAMESPACE -- strided"

UPGRADE_BUFFER=45 # blocks

trim_tx() {
    grep -E "code:|txhash:" | sed 's/^[[:space:]]*//'
}

upgrade_name=$(kubectl exec stride-validator-0 -c validator -- printenv UPGRADE_NAME)
latest_height=$($STRIDED0 status | jq -r 'if .SyncInfo then .SyncInfo.latest_block_height else .sync_info.latest_block_height end')
upgrade_height=$((latest_height+UPGRADE_BUFFER))

echo -e "\nSubmitting proposal for $upgrade_name at height $upgrade_height...\n"
$EXEC0 bash scripts/propose_upgrade.sh $upgrade_name $upgrade_height | trim_tx

sleep 5
echo -e "\nProposal:\n"
proposal_id=$($STRIDED0 q gov proposals --output json | jq -r '.proposals | max_by(.id | tonumber).id')
$STRIDED0 query gov proposal $proposal_id

sleep 1
echo -e "\nVoting on proposal #$proposal_id...\n"
echo "Val1:"
$STRIDED0 tx gov vote $proposal_id yes --from val1 -y | trim_tx
echo "Val2:"
$STRIDED1 tx gov vote $proposal_id yes --from val2 -y | trim_tx
echo "Val3:"
$STRIDED2 tx gov vote $proposal_id yes --from val3 -y | trim_tx

sleep 5
echo -e "\nVote confirmation:\n"
$STRIDED0 query gov tally $proposal_id

echo -e "\nProposal Status:\n"
while true; do
    status=$($STRIDED0 query gov proposal $proposal_id --output json | jq -r '.proposal.status')
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
