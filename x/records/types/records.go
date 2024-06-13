package types

import sdkmath "cosmossdk.io/math"

// Helper function to evaluate if a host zone unbonding record should
// have it's unbonding initiated
// This is indicated by a record in status UNBONDING_QUEUE with a non-zero
// st token amount
func (r HostZoneUnbonding) ShouldInitiateUnbonding() bool {
	notYetUnbonding := r.Status == HostZoneUnbonding_UNBONDING_QUEUE
	hasAtLeastOneRedemption := r.StTokenAmount.GT(sdkmath.ZeroInt())
	return notYetUnbonding && hasAtLeastOneRedemption
}

// Helper function to evaluate if a host zone unbonding record should
// have it's unbonding retried
// This is indicated by a record in status UNBONDING_RETRY_QUEUE and
// 0 undelegations in progress
func (r HostZoneUnbonding) ShouldRetryUnbonding() bool {
	hasAtLeastOneRedemption := r.StTokenAmount.GT(sdkmath.ZeroInt())
	shouldRetryUnbonding := r.Status == HostZoneUnbonding_UNBONDING_RETRY_QUEUE
	hasNoPendingICAs := r.UndelegationTxsInProgress == 0
	return hasAtLeastOneRedemption && shouldRetryUnbonding && hasNoPendingICAs
}
