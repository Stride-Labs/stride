package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// Liquid Stake Reward Collector Balance
func (k Keeper) LiquidStakeRewardCollectorBalance(ctx sdk.Context, msgSvr types.MsgServer) bool {
	k.Logger(ctx).Info("Liquid Staking reward collector balance")
	rewardCollectorAddress := k.accountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()
	rewardedTokens := k.bankKeeper.GetAllBalances(ctx, rewardCollectorAddress)
	if rewardedTokens.IsEqual(sdk.Coins{}) {
		k.Logger(ctx).Info("No reward to allocate from RewardCollector")
		return false
	}

	rewardsAccrued := false
	for _, token := range rewardedTokens {
		// get hostzone by reward token (in ibc denom format)
		hz, err := k.GetHostZoneFromIBCDenom(ctx, token.Denom)
		if err != nil {
			k.Logger(ctx).Info("Token denom %s in module account is not from a supported host zone", token.Denom)
			continue
		}

		// liquid stake all tokens
		msg := types.NewMsgLiquidStake(rewardCollectorAddress.String(), token.Amount, hz.HostDenom)
		if err := msg.ValidateBasic(); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Liquid stake from reward collector address failed validate basic: %s", err.Error()))
			continue
		}
		_, err = msgSvr.LiquidStake(ctx, msg)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to liquid stake %s for hostzone %s: %s", token.String(), hz.ChainId, err.Error()))
			continue
		}
		k.Logger(ctx).Info(fmt.Sprintf("Liquid staked %s for hostzone %s's accrued rewards", token.String(), hz.ChainId))
		rewardsAccrued = true
	}
	return rewardsAccrued
}

// Sweep stTokens from Reward Collector to Fee Collector
func (k Keeper) SweepStTokensFromRewardCollToFeeColl(ctx sdk.Context) error {
	// Send all stTokens to fee collector to distribute to delegator later
	rewardCollectorAddress := k.accountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()

	rewardCollCoins := k.bankKeeper.GetAllBalances(ctx, rewardCollectorAddress)
	k.Logger(ctx).Info(fmt.Sprintf("Reward collector has %s", rewardCollCoins.String()))
	stTokens := sdk.NewCoins()
	for _, token := range rewardCollCoins {
		// get hostzone by reward token (in stToken denom format)
		isStToken := k.CheckIsStToken(ctx, token.Denom)
		k.Logger(ctx).Info(fmt.Sprintf("%s is stToken: %t", token.String(), isStToken))
		if isStToken {
			stTokens = append(stTokens, token)
		}
	}
	k.Logger(ctx).Info(fmt.Sprintf("Sending %s stTokens from %s to %s", stTokens.String(), types.RewardCollectorName, authtypes.FeeCollectorName))

	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.RewardCollectorName, authtypes.FeeCollectorName, stTokens)
	if err != nil {
		return errorsmod.Wrapf(err, fmt.Sprintf("Can't send coins from module %s to module %s, err %s", types.RewardCollectorName, authtypes.FeeCollectorName, err.Error()))
	}
	return nil
}

// (1) liquid stake reward collector balance, then (2) sweet stTokens from reward collector to fee collector
func (k Keeper) AllocateHostZoneReward(ctx sdk.Context) {
	msgSvr := NewMsgServerImpl(k)
	if rewardsFound := k.LiquidStakeRewardCollectorBalance(ctx, msgSvr); !rewardsFound {
		k.Logger(ctx).Info("No accrued rewards in the reward collector account")
		return
	}
	if err := k.SweepStTokensFromRewardCollToFeeColl(ctx); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Unable to allocate host zone reward, err: %s", err.Error()))
	}
}
