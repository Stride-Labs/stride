package types

import sdkmath "cosmossdk.io/math"

// Helper function to evaluate if a host zone unbonding record still needs to be initiated
func (r HostZoneUnbonding) ShouldInitiateUnbonding() bool {
	return r.Status == HostZoneUnbonding_UNBONDING_QUEUE && r.NativeTokenAmount.GT(sdkmath.ZeroInt())
}
