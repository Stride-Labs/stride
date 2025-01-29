#!/bin/bash

# Constants - fill in these values
CELESTIA_HOST_ZONE="mocha-4"
STTIA_DENOM="stutia"
TIA_DENOM="utia"

staketia_state() {
    DEPOSIT_ADDRESS="$(strided q staketia host-zone | jq -r .host_zone.deposit_address)"
    REDEMPTION_ADDRESS="$(strided q staketia host-zone | jq -r .host_zone.redemption_address)"
    CLAIM_ADDRESS="$(strided q staketia host-zone | jq -r .host_zone.claim_address)"
    FEE_ADDRESS="$(strided q auth module-account staketia_fee_address | jq -r .account.base_account.address)"
    DELEGATION_ADDRESS="$(strided q staketia host-zone | jq -r .host_zone.delegation_address)"
    REWARD_ADDRESS="$(strided q staketia host-zone | jq -r .host_zone.reward_address)"

    # Host Zone
    echo "Host Zone:"
    strided q staketia host-zone | jq .host_zone
    echo "----------------------------------------"

    # Delegation records
    echo "Delegation Records:"
    strided q staketia delegation-records | jq .delegation_records
    echo "----------------------------------------"

    # Unbonding records
    echo "Unbonding Records:"
    strided q staketia unbonding-records | jq .unbonding_records
    echo "----------------------------------------"

    # Redemption records
    echo "Redemption Records:"
    strided q staketia redemption-records | jq .redemption_record_responses
    echo "----------------------------------------"

    # Host zone delegated balance
    echo "Host Zone Delegated Balance:"
    strided q staketia host-zone | jq -r '.host_zone | if has("delegated_balance") then .delegated_balance else .remaining_delegated_balance end'
    echo "----------------------------------------"

    # Redemption rate
    echo "Redemption Rate:"
    strided q staketia host-zone | jq -r .host_zone.redemption_rate
    echo "----------------------------------------"

    # stTIA supply
    echo "stTIA Supply:"
    strided q bank total --denom ${STTIA_DENOM} | jq -r .amount
    echo "----------------------------------------"

    # Account balances on Stride
    echo "Deposit Account Balance:"
    strided q bank balances ${DEPOSIT_ADDRESS} | jq .balances
    echo "----------------------------------------"

    echo "Redemption Account Balance:"
    strided q bank balances ${REDEMPTION_ADDRESS} | jq .balances
    echo "----------------------------------------"

    echo "Claim Account Balance:"
    strided q bank balances ${CLAIM_ADDRESS} | jq .balances
    echo "----------------------------------------"

    echo "Fee Account Balance:"
    strided q bank balances ${FEE_ADDRESS} | jq .balances
    echo "----------------------------------------"

    # Celestia account balances
    echo "Delegation Account Balance:"
    celestia-appd q bank balances ${DELEGATION_ADDRESS} | jq .balances
    echo "----------------------------------------"

    echo "Delegation Account Staked Balance:"
    celestia-appd q staking delegations ${DELEGATION_ADDRESS} | jq .delegation_responses
    echo "----------------------------------------"

    echo "Delegation Account Unbondings"
    celestia-appd q staking unbonding-delegations ${DELEGATION_ADDRESS} | jq .unbonding_responses
    echo "----------------------------------------"

    echo "Reward Account Balance:"
    celestia-appd q bank balances ${REWARD_ADDRESS} | jq .balances
    echo "----------------------------------------"
}

stakeibc_state() {
    echo "Host Zone:"
    strided q stakeibc show-host-zone ${CELESTIA_HOST_ZONE} | jq .host_zone
    echo "----------------------------------------"

    echo "Host Zone Staked Balance:"
    strided q stakeibc show-host-zone ${CELESTIA_HOST_ZONE} | jq -r .host_zone.total_delegations
    echo "----------------------------------------"

    echo "Deposit Records:"
    strided q records list-deposit-record ${CELESTIA_HOST_ZONE} | jq .deposit_record
    echo "----------------------------------------"

    echo "Unbonding Records:"
    strided q records list-epoch-unbonding-record ${CELESTIA_HOST_ZONE} | jq .epoch_unbonding_record
    echo "----------------------------------------"

    echo "Redemption Records:"
    strided q records list-user-redemption-record ${CELESTIA_HOST_ZONE} | jq .user_redemption_record
    echo "----------------------------------------"

    echo "Deposit Account Balance:"
    DEPOSIT_ADDRESS="$(strided q stakeibc show-host-zone ${CELESTIA_HOST_ZONE} | jq -r .host_zone.deposit_address)"
    strided q bank balances ${DEPOSIT_ADDRESS}
    echo "----------------------------------------"

    echo "Reward Collector Balance:"
    REWARD_COLLECTOR_ADDRESS="$(strided q auth module-account reward_collector | jq -r .account.base_account.address)"
    strided q bank balances ${REWARD_COLLECTOR_ADDRESS}
    echo "----------------------------------------"

    echo "ICA Account Balances:"
    echo "Delegation ICA Balance:"
    DELEGATION_ICA=$(strided q stakeibc show-host-zone ${CELESTIA_HOST_ZONE} | jq -r .host_zone.delegation_ica_address)
    celestia-appd q bank balances ${DELEGATION_ICA} | jq .balances
    echo "----------------------------------------"

    echo "Fee ICA Balance:"
    FEE_ICA=$(strided q stakeibc show-host-zone ${CELESTIA_HOST_ZONE} | jq -r .host_zone.fee_ica_address)
    celestia-appd q bank balances ${FEE_ICA} | jq .balances
    echo "----------------------------------------"

    echo "Withdrawal ICA Balance:"
    WITHDRAWAL_ICA=$(strided q stakeibc show-host-zone ${CELESTIA_HOST_ZONE} | jq -r .host_zone.withdrawal_ica_address)
    celestia-appd q bank balances ${WITHDRAWAL_ICA} | jq .balances
    echo "----------------------------------------"

    echo "Redemption ICA Balance:"
    REDEMPTION_ICA=$(strided q stakeibc show-host-zone ${CELESTIA_HOST_ZONE} | jq -r .host_zone.redemption_ica_address)
    celestia-appd q bank balances ${REDEMPTION_ICA} | jq .balances
    echo "----------------------------------------"
}

# Function to capture state before upgrade
capture_pre_upgrade_state() {
    echo "Capturing pre-upgrade state..."
    echo "================================"

    # Create timestamp for file
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S%z")
    OUTPUT_FILE="stride_pre_upgrade_state_${TIMESTAMP}.txt"

    {
        echo "Pre-upgrade State Capture - ${TIMESTAMP}"
        echo "========================================"

        staketia_state

    } | tee "$OUTPUT_FILE"

    echo "Pre-upgrade state captured in $OUTPUT_FILE"
}

# Function to capture state after upgrade
capture_post_upgrade_state() {
    echo "Capturing post-upgrade state..."
    echo "================================"

    # Create timestamp for file
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S%z")
    OUTPUT_FILE="stride_post_upgrade_state_${TIMESTAMP}.txt"

    {
        echo "Post-upgrade State Capture - ${TIMESTAMP}"
        echo "========================================"

        # staketia state
        staketia_state

        # stakeibc checks
        stakeibc_state

    } | tee "$OUTPUT_FILE"

    echo "Post-upgrade state captured in $OUTPUT_FILE"
}

# Main execution
case "$1" in
    "pre")
        capture_pre_upgrade_state
        ;;
    "post")
        capture_post_upgrade_state
        ;;
    *)
        echo "Usage: $0 {pre|post}"
        echo "  pre  - Capture pre-upgrade state"
        echo "  post - Capture post-upgrade state"
        exit 1
        ;;
esac