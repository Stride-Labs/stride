#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

deposit_ica_account=$(GET_ICA_ADDR GAIA community_pool_deposit)
proposal_file=${STATE}/${GAIA_NODE_PREFIX}1/proposal.json
cat << EOF > $proposal_file
{
  "title": "Community Pool Liquid Stake",
  "description": "Community Pool Liquid Stake",
  "recipient": "${deposit_ica_account}",
  "amount": "1000000uatom",
  "deposit": "10000000uatom"
}
EOF

echo ">>> Submitting proposal to spend community pool tokens..."
$GAIA_MAIN_CMD tx gov submit-proposal community-pool-spend $proposal_file --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
sleep 5

echo ">>> Voting on proposal..."
proposal_id=$($GAIA_MAIN_CMD q gov proposals | grep 'id:' | tail -1 | awk '{printf $2}' | tr -d '"')
$GAIA_MAIN_CMD tx gov vote $proposal_id yes --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX

echo ">>> Waiting for proposal to pass..."
printf "\nPROPOSAL STATUS\n"
while true; do
    status=$($GAIA_MAIN_CMD query gov proposal $proposal_id | grep "status" | awk '{printf $2}')
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