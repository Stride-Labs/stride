package types

import sdkmath "cosmossdk.io/math"

// Helper function to evaluate if a host zone unbonding record still needs to be initiated
func (r HostZoneUnbonding) ShouldInitiateUnbonding() bool {
	notYetUnbonding := r.Status == HostZoneUnbonding_UNBONDING_QUEUE
	hasAtLeastOneRecord := r.NativeTokenAmount.GT(sdkmath.ZeroInt())
	return notYetUnbonding && hasAtLeastOneRecord
}
