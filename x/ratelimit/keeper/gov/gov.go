package gov

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"

	"github.com/Stride-Labs/stride/v9/x/ratelimit/keeper"
	"github.com/Stride-Labs/stride/v9/x/ratelimit/types"
)

// Adds a new rate limit. Fails if the rate limit already exists or the channel value is 0
func AddRateLimit(ctx sdk.Context, k keeper.Keeper, channelKeeper channelkeeper.Keeper, p *types.AddRateLimitProposal) error {
	// Confirm the channel value is not zero
	channelValue := k.GetChannelValue(ctx, p.Denom)
	if channelValue.IsZero() {
		return types.ErrZeroChannelValue
	}

	// Confirm the rate limit does not already exist
	_, found := k.GetRateLimit(ctx, p.Denom, p.ChannelId)
	if found {
		return types.ErrRateLimitAlreadyExists
	}

	// Confirm the channel exists
	_, found = channelKeeper.GetChannel(ctx, transfertypes.PortID, p.ChannelId)
	if !found {
		return types.ErrChannelNotFound
	}

	// Create and store the rate limit object
	path := types.Path{
		Denom:     p.Denom,
		ChannelId: p.ChannelId,
	}
	quota := types.Quota{
		MaxPercentSend: p.MaxPercentSend,
		MaxPercentRecv: p.MaxPercentRecv,
		DurationHours:  p.DurationHours,
	}
	flow := types.Flow{
		Inflow:       sdkmath.ZeroInt(),
		Outflow:      sdkmath.ZeroInt(),
		ChannelValue: channelValue,
	}

	k.SetRateLimit(ctx, types.RateLimit{
		Path:  &path,
		Quota: &quota,
		Flow:  &flow,
	})

	return nil
}

// Updates an existing rate limit. Fails if the rate limit doesn't exist
func UpdateRateLimit(ctx sdk.Context, k keeper.Keeper, p *types.UpdateRateLimitProposal) error {
	// Confirm the rate limit exists
	_, found := k.GetRateLimit(ctx, p.Denom, p.ChannelId)
	if !found {
		return types.ErrRateLimitNotFound
	}

	// Update the rate limit object with the new quota information
	// The flow should also get reset to 0
	path := types.Path{
		Denom:     p.Denom,
		ChannelId: p.ChannelId,
	}
	quota := types.Quota{
		MaxPercentSend: p.MaxPercentSend,
		MaxPercentRecv: p.MaxPercentRecv,
		DurationHours:  p.DurationHours,
	}
	flow := types.Flow{
		Inflow:       sdkmath.ZeroInt(),
		Outflow:      sdkmath.ZeroInt(),
		ChannelValue: k.GetChannelValue(ctx, p.Denom),
	}

	k.SetRateLimit(ctx, types.RateLimit{
		Path:  &path,
		Quota: &quota,
		Flow:  &flow,
	})

	return nil
}

// Removes a rate limit. Fails if the rate limit doesn't exist
func RemoveRateLimit(ctx sdk.Context, k keeper.Keeper, msg *types.RemoveRateLimitProposal) error {
	_, found := k.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if !found {
		return types.ErrRateLimitNotFound
	}

	k.RemoveRateLimit(ctx, msg.Denom, msg.ChannelId)
	return nil
}

// Resets the flow on a rate limit. Fails if the rate limit doesn't exist
func ResetRateLimit(ctx sdk.Context, k keeper.Keeper, msg *types.ResetRateLimitProposal) error {
	return k.ResetRateLimit(ctx, msg.Denom, msg.ChannelId)
}
