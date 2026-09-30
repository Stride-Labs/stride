package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// Test shims for the unexported sweep helpers
type SweepDestinationForTest = sweepDestination

func ResolveSweepDestinationForTest(k Keeper, ctx sdk.Context, denom string) (SweepDestinationForTest, error) {
	return k.resolveSweepDestination(ctx, denom)
}

func TransferEscrowAddressesForTest(k Keeper, ctx sdk.Context) map[string]bool {
	return k.transferEscrowAddresses(ctx)
}

func SweepProtocolAddressSetForTest() map[string]bool {
	return sweepProtocolAddressSet()
}

func SweepSkipReasonForTest(
	k Keeper,
	ctx sdk.Context,
	address sdk.AccAddress,
	escrows map[string]bool,
	protocol map[string]bool,
) (string, bool) {
	return k.sweepSkipReason(ctx, address, escrows, protocol)
}
