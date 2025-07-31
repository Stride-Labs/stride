package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ccvtypes "github.com/cosmos/interchain-security/v4/x/ccv/consumer/types"

	auctiontypes "github.com/Stride-Labs/stride/v27/x/auction/types"

	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

// AuctionOffRewardCollectorBalance distributes rewards from the reward collector:
// Sends 15% to ICS, and the remainder to the auction module
// ConsumerRedistributionFraction = what Stride keeps = 0.85 on mainnet
// ICS Portion = 1 - ConsumerRedistributionFraction = 0.15
// Fees arrive in the reward collector account as native tokens
func (k Keeper) AuctionOffRewardCollectorBalance(ctx sdk.Context) {
	rewardCollectorAddress := k.AccountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()

	// Get consumer redistribution fraction from CCV params
	consumerRedistributionFracStr := k.ConsumerKeeper.GetConsumerParams(ctx).ConsumerRedistributionFraction
	strideKeepRate, err := sdk.NewDecFromStr(consumerRedistributionFracStr)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Invalid strideKeepRate, cannot send stTokens to ICS provider: %s", err))
		return
	}

	// Calculate Hub's keep rate (1 - strideKeepRate)
	hubKeepRate := sdk.OneDec().Sub(strideKeepRate)

	// Get all host zones and process their tokens in reward collector balance
	for _, hz := range k.GetAllHostZone(ctx) {
		// Check if reward collector has this host zone's IBC denom
		// These are the fees collected from liquid staking rewards
		tokenBalance := k.bankKeeper.GetBalance(ctx, rewardCollectorAddress, hz.IbcDenom)
		if tokenBalance.IsZero() {
			continue
		}

		// Calculate the ICS portion to liquid stake
		tokensToLiquidStake := sdk.NewDecCoinsFromCoins(tokenBalance).MulDec(hubKeepRate).AmountOf(hz.IbcDenom).TruncateInt()
		if tokensToLiquidStake.IsZero() {
			continue
		}

		// Liquid stake the ICS portion
		msg := types.NewMsgLiquidStake(rewardCollectorAddress.String(), tokensToLiquidStake, hz.HostDenom)
		if err := msg.ValidateBasic(); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Liquid stake from reward collector failed validation: %s", err.Error()))
			continue
		}
		liquidStakeResp, err := NewMsgServerImpl(k).LiquidStake(ctx, msg)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to liquid stake %s for hostzone %s: %s", sdk.NewCoin(hz.IbcDenom, tokensToLiquidStake).String(), hz.ChainId, err.Error()))
			continue
		}

		// Get the resulting stToken balance for this hostzone
		if liquidStakeResp.StToken.IsZero() {
			continue
		}
		icsProviderStTokens := sdk.NewCoins(liquidStakeResp.StToken)

		// Send stTokens to ConsumerToSendToProvider module
		err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.RewardCollectorName, ccvtypes.ConsumerToSendToProviderName, icsProviderStTokens)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Cannot send stTokens from RewardCollector to ConsumerToSendToProvider: %s", err))
			continue
		}

		// Send remaining native tokens to auction module
		rewardCollectorBalances := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, rewardCollectorAddress, hz.IbcDenom))
		if rewardCollectorBalances.Empty() {
			continue
		}

		err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.RewardCollectorName, auctiontypes.ModuleName, rewardCollectorBalances)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Cannot send rewards from RewardCollector to Auction module: %s", err))
			continue
		}
	}
}
