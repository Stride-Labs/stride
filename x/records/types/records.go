package types

import sdkmath "cosmossdk.io/math"

// Helper function to evaluate if a host zone unbonding record still needs
// the unbonding to be initiated
// This includes records either in the normal queue or the retry queue
func (r HostZoneUnbonding) ShouldInitiateUnbonding() bool {
	notYetUnbonding := r.Status == HostZoneUnbonding_UNBONDING_QUEUE
	hasFailedUnbonding := r.Status == HostZoneUnbonding_UNBONDING_RETRY_QUEUE
	hasAtLeastOneRecord := r.NativeTokenAmount.GT(sdkmath.ZeroInt())
	return (notYetUnbonding || hasFailedUnbonding) && hasAtLeastOneRecord
}
