# #!/bin/bash

set -eu

STRIDED="strided --home ${HOME}/.stride-localstride"
TX_ARGS="--keyring-backend test --chain-id stride-test-1 --fees 10000ustrd"
UPGRADE_BUFFER=45 # blocks

if [[ "$UPGRADE_NAME" == "" ]]; then
    echo "ERROR: Please specfiy UPGRADE_NAME variable"
    exit 1
fi

trim_tx() {
    grep -E "code:|txhash:" | sed 's/^[[:space:]]*//'
}

upgrade_name=$UPGRADE_NAME
latest_height=$($STRIDED status | jq -r 'if .SyncInfo then .SyncInfo.latest_block_height else .sync_info.latest_block_height end')
upgrade_height=$((latest_height+UPGRADE_BUFFER))

echo -e "\nSubmitting proposal for $upgrade_name at height $upgrade_height...\n"
cat > /tmp/proposal.json << EOF
{
  "title": "Upgrade $upgrade_name",
  "summary": "Upgrade $upgrade_name",
  "metadata": "",
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
      "plan": {
        "name": "$upgrade_name",
        "height": "$upgrade_height"
      }
    }
  ],
  "deposit": "2000000000ustrd"
}
EOF

strided tx gov submit-proposal /tmp/proposal.json --from val -y $TX_ARGS | trim_tx

sleep 5
echo -e "\nProposal:\n"
proposal_id=$($STRIDED q gov proposals --output json | jq -r '.proposals | max_by(.id | tonumber).id')
$STRIDED query gov proposal $proposal_id

sleep 1
echo -e "\nVoting on proposal #$proposal_id...\n"
$STRIDED tx gov vote $proposal_id yes --from val -y $TX_ARGS | trim_tx

sleep 5
echo -e "\nVote confirmation:\n"
$STRIDED query gov tally $proposal_id

echo -e "\nProposal Status:\n"
while true; do
    status=$($STRIDED query gov proposal $proposal_id --output json | jq -r '.proposal.status')
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

