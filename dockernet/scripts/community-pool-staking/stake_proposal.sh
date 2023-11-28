#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

deposit_ica_account=$(GET_ICA_ADDR DYDX community_pool_deposit)
proposal_file=${STATE}/${DYDX_NODE_PREFIX}1/pool.json
cat << EOF > $proposal_file
{
  "title": "Community Spend: Liquid stake",
  "metadata": "Community Spend: Liquid stake",
  "summary": "Community Spend: Liquid stake",
  "messages": [
    {
      "@type": "/cosmos.distribution.v1beta1.MsgCommunityPoolSpend",
      "authority": "dydx10d07y265gmmuvt4z0w9aw880jnsr700jnmapky",
      "recipient": "$deposit_ica_account",
      "amount": [
        {
            "amount": "1000",
            "denom": "udydx"
        }
      ]
    }
  ],
  "deposit": "2000000000udydx"
}
EOF

echo ">>> Submitting proposal to spend community pool tokens..."
$DYDX_MAIN_CMD tx gov submit-proposal $proposal_file --from ${DYDX_VAL_PREFIX}1 -y | TRIM_TX
sleep 5

echo -e "\n>>> Voting on proposal..."
proposal_id=$($DYDX_MAIN_CMD q gov proposals | grep 'id:' | tail -1 | awk '{printf $2}' | tr -d '"')
$DYDX_MAIN_CMD tx gov vote $proposal_id yes --from ${DYDX_VAL_PREFIX}1 -y | TRIM_TX

echo -e "\n>>> Waiting for proposal to pass..."
printf "\nPROPOSAL STATUS\n"
while true; do
    status=$($DYDX_MAIN_CMD query gov proposal $proposal_id | grep "status" | awk '{printf $2}')
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