#!/bin/bash

# NOTE: This script should be run locally (outside of the k8s pod)

set -eu

NAMESPACE=integration
UPGRADE_BUFFER=45 # blocks

# Helper to invoke strided in a specific validator pod.
strided_in() {
    pod_index="$1"; shift
    kubectl exec -it stride-validator-$pod_index -c validator -n $NAMESPACE -- strided "$@"
}

trim_tx() {
    grep -E "code:|txhash:" | sed 's/^[[:space:]]*//'
}

upgrade_name=$(kubectl exec stride-validator-0 -c validator -n $NAMESPACE -- printenv UPGRADE_NAME)
latest_height=$(strided_in 0 status | jq -r 'if .SyncInfo then .SyncInfo.latest_block_height else .sync_info.latest_block_height end')
upgrade_height=$((latest_height+UPGRADE_BUFFER))

echo -e "\nSubmitting proposal for $upgrade_name at height $upgrade_height...\n"
kubectl exec -it stride-validator-0 -c validator -n $NAMESPACE -- \
    bash scripts/propose_upgrade.sh $upgrade_name $upgrade_height | trim_tx

sleep 5
echo -e "\nProposal:\n"
proposal_id=$(strided_in 0 q gov proposals --output json | jq -r '.proposals | max_by(.id | tonumber).id')
strided_in 0 query gov proposal $proposal_id

sleep 1
echo -e "\nVoting on proposal #$proposal_id...\n"
for i in 0 1 2 3 4; do
    val_name="val$((i + 1))"
    echo "${val_name}:"
    strided_in "$i" tx gov vote $proposal_id yes --from "$val_name" -y | trim_tx
done

sleep 5
echo -e "\nVote confirmation:\n"
strided_in 0 query gov tally $proposal_id

echo -e "\nProposal Status:\n"
while true; do
    status=$(strided_in 0 query gov proposal $proposal_id --output json | jq -r '.proposal.status')
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
