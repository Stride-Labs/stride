package stakeibc

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker of stakeibc module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper, ak types.AccountKeeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// Iterate over all host zones and verify redemption rate
	for _, hz := range k.GetAllHostZone(ctx) {
		rrSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, hz)
		if !rrSafe {
			panic(fmt.Sprintf("[INVARIANT BROKEN!!!] %s's RR is %s. ERR: %v", hz.GetChainId(), hz.RedemptionRate.String(), err.Error()))
		}
	}
}

// BeginBlocker of stakeibc module
func EndBlocker(ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper, ak types.AccountKeeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	rewardCollectorAddress := ak.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()
	fmt.Println("rewardCollectorAddress", rewardCollectorAddress)
	rewardedTokens := bk.GetAllBalances(ctx, rewardCollectorAddress)
	k.Logger(ctx).Info("rewardedTokens", rewardedTokens)
	if rewardedTokens.IsEqual(sdk.Coins{}) {
		return []abci.ValidatorUpdate{}
	}

	msgSvr := keeper.NewMsgServerImpl(k)
	for _, token := range rewardedTokens {
		// get hostzone by reward token (in ibc denom format)
		hz, err := k.GetHostZoneFromIBCDenom(ctx, token.Denom)
		if err != nil {
			panic(fmt.Sprintf("Can't get host zone from ibc token %s", token.Denom))
		}

		// liquid stake all tokens
		msg := types.NewMsgLiquidStake(rewardCollectorAddress.String(), token.Amount, hz.HostDenom)
		_, err = msgSvr.LiquidStake(ctx, msg)
		fmt.Println("err", err)
		if err != nil {
			panic(fmt.Sprintf("Can't liquid stake %s for hostzone %s", token.String(), hz.ChainId))
		}
	}

	// After liquid stake all tokens, reward collector receive stTokens
	// Send all stTokens to fee collector to distribute to delegator later
	stTokens := bk.GetAllBalances(ctx, rewardCollectorAddress)
	err := bk.SendCoinsFromModuleToModule(ctx, types.RewardCollectorName, authtypes.FeeCollectorName, stTokens)
	if err != nil {
		panic(fmt.Sprintf("Can't send coins from module %s to module %s", types.RewardCollectorName, authtypes.FeeCollectorName))
	}
	return []abci.ValidatorUpdate{}
}
