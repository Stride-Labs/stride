package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateTradeRoute(goCtx context.Context, msg *types.MsgCreateTradeRoute) (*types.MsgCreateTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("create trade route: %s", msg.String()))

	// get our addresses, make sure they're valid
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "creator address is invalid: %s. err: %s", msg.Creator, err.Error())
	}

	hostICA := types.ICAAccount{
		ChainId:      msg.HostChainId,
		Type:         types.ICAAccountType_WITHDRAWAL,
		ConnectionId: msg.HostConnectionId,
		Address:      msg.HostIcaAddress,
	}

	rewardICA := types.ICAAccount{
		ChainId:      msg.RewardChainId,
		Type:         types.ICAAccountType_UNWIND,
		ConnectionId: msg.RewardConnectionId,
		Address:      msg.RewardIcaAddress,
	}

	tradeICA := types.ICAAccount{
		ChainId:      msg.TradeChainId,
		Type:         types.ICAAccountType_TRADE,
		ConnectionId: msg.TradeConnectionId,
		Address:      msg.TradeIcaAddress,
	}

	hostRewardHop := types.TradeHop{
		TransferChannelId: msg.HostRewardTransferChannelId,
		FromAccount:       hostICA,
		ToAccount:         rewardICA,
	}

	rewardTradeHop := types.TradeHop{
		TransferChannelId: msg.RewardTradeTransferChannelId,
		FromAccount:       rewardICA,
		ToAccount:         tradeICA,
	}

	tradeHostHop := types.TradeHop{
		TransferChannelId: msg.TradeHostTransferChannelId,
		FromAccount:       tradeICA,
		ToAccount:         hostICA,
	}

	newTradeRoute := types.TradeRoute{
		RewardDenomOnHostZone:   msg.RewardDenomOnHost,
		RewardDenomOnRewardZone: msg.RewardDenomOnReward,
		RewardDenomOnTradeZone:  msg.RewardDenomOnTrade,
		TargetDenomOnTradeZone:  msg.TargetDenomOnTrade,
		TargetDenomOnHostZone:   msg.TargetDenomOnHost,
		HostToRewardHop:         hostRewardHop,
		RewardToTradeHop:        rewardTradeHop,
		TradeToHostHop:          tradeHostHop,
		PoolId:                  msg.PoolId,
		SpotPrice:               "", // this should only ever be set by ICQ so initialize to blank
		MinSwapAmount:           msg.MinSwapAmount,
		MaxSwapAmount:           msg.MaxSwapAmount,
	}

	k.SetTradeRoute(ctx, newTradeRoute)

	k.Logger(ctx).Info(fmt.Sprintf("create trade route: %s", msg.String()))
	return &types.MsgCreateTradeRouteResponse{}, nil
}
