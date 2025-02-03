package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	auctiontypes "github.com/Stride-Labs/stride/v25/x/auction/types"
	"github.com/Stride-Labs/stride/v25/x/stakeibc/types"
)

// AuctionOffRewardCollectorBalance transfers all balances from the reward collector module account
// to the auction module account. If the reward collector has no balance, it does nothing.
func (k Keeper) AuctionOffRewardCollectorBalance(ctx sdk.Context) {
	k.Logger(ctx).Info("Auctioning reward collector balance")

	rewardCollectorAddress := k.AccountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()

	rewardCollectorBalances := k.bankKeeper.GetAllBalances(ctx, rewardCollectorAddress)
	if rewardCollectorBalances.Empty() {
		k.Logger(ctx).Info("No rewards to auction from RewardCollector")
		return
	}

	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.RewardCollectorName, auctiontypes.ModuleName, rewardCollectorBalances)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Cannot send rewards from RewardCollector to Auction module: %s", err))
	}
}
