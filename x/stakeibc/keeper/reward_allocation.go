package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v31/utils"
	auctiontypes "github.com/Stride-Labs/stride/v31/x/auction/types"

	"github.com/Stride-Labs/stride/v31/x/stakeibc/types"
)

// AuctionOffRewardCollectorBalance distributes rewards from the reward collector:
// Sends 15% to PoA validators, and the remainder to the auction module
// ConsumerRedistributionFraction = what Stride keeps = 0.85 on mainnet
// ICS Portion = 1 - ConsumerRedistributionFraction = 0.15
// Fees arrive in the reward collector account as native tokens
func (k Keeper) AuctionOffRewardCollectorBalance(ctx sdk.Context) {
	rewardCollectorAddress := k.AccountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()

	// Get all host zones and process their tokens in reward collector balance
	for _, hz := range k.GetAllActiveHostZone(ctx) {
		// Check if reward collector has this host zone's IBC denom
		// These are the fees collected from liquid staking rewards
		if hz.IbcDenom == "" { // prevents panic in balance query if the denom field is not set
			continue
		}
		tokenBalance := k.bankKeeper.GetBalance(ctx, rewardCollectorAddress, hz.IbcDenom)
		if tokenBalance.IsZero() {
			continue
		}

		// Calculate the ICS portion to liquid stake
		tokensToLiquidStakeForVals := sdk.NewDecCoinsFromCoins(tokenBalance).MulDec(utils.PoaValPaymentRate).AmountOf(hz.IbcDenom).TruncateInt()
		if tokensToLiquidStakeForVals.IsZero() {
			continue
		}

		// Liquid stake the ICS portion
		msg := types.NewMsgLiquidStake(rewardCollectorAddress.String(), tokensToLiquidStakeForVals, hz.HostDenom)
		if err := msg.ValidateBasic(); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Liquid stake from reward collector failed validation: %s", err.Error()))
			continue
		}
		liquidStakeResp, err := NewMsgServerImpl(k).LiquidStake(ctx, msg)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to liquid stake %s for hostzone %s: %s",
				sdk.NewCoin(hz.IbcDenom, tokensToLiquidStakeForVals).String(), hz.ChainId, err.Error()))
			continue
		}

		// Get the resulting stToken balance for this hostzone
		if liquidStakeResp.StToken.IsZero() {
			continue
		}

		// Send stTokens to each validator in the set
		// Note: This ignores the remainder for simplicity
		totalValidatorStTokenAmount := liquidStakeResp.StToken.Amount
		numValidators := sdkmath.NewInt(int64(len(utils.PoaValidatorSet))).ToLegacyDec()
		perValidatorStTokenAmount := sdkmath.LegacyNewDecFromInt(totalValidatorStTokenAmount).Quo(numValidators).TruncateInt()
		perValidatorStToken := sdk.NewCoin(liquidStakeResp.StToken.Denom, perValidatorStTokenAmount)

		if perValidatorStToken.Amount.GT(sdkmath.ZeroInt()) {
			for _, validator := range utils.PoaValidatorSet {
				valAddress := sdk.MustAccAddressFromBech32(validator)
				err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.RewardCollectorName, valAddress, sdk.NewCoins(perValidatorStToken))
				if err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("Cannot send stTokens from RewardCollector to %s: %s",
						validator, err))
					continue
				}
			}
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
