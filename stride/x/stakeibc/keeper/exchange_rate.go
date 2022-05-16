package keeper

import (
	"context"

	"github.com/Stride-labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// is this the right keeper?
func (k Keeper) getStTokenExchRate(goCtx context.Context, hostZone types.HostZone, inclOutstandingRewards bool) (sdk.Dec, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatedBalOfVirtualPoolOnHost := k.Interchain.Staking.Delegations()
	// Read stAsset supply
	supplyOfStTokens := k.BankKeeper.Supply(stDenom(inCoin.Denom))
	k.Logger(ctx).Info("stAsset outstanding supply:", supplyOfStTokens)

	// ICQ accrued rewards
	if inclOutstandingRewards {
		outstandingRewardsOfVirtualPoolOnHost := k.Interchain.Distribution.OutstandingRewards()
		balOfVirtualPoolOnHost := delegatedBalOfVirtualPoolOnHost + outstandingRewardsOfVirtualPoolOnHost
	} else {
		balOfVirtualPoolOnHost := delegatedBalOfVirtualPoolOnHost
	}

	exchRate := balOfVirtualPoolOnHost.toDec() / supplyOfStTokens.toDec()

	return exchRate
}
