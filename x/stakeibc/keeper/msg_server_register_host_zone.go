package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RegisterHostZone(goCtx context.Context, msg *types.MsgRegisterHostZone) (*types.MsgRegisterHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.Keeper.RegisterHostZone(
		ctx,
		msg.ConnectionId,
		msg.HostDenom,
		msg.TransferChannelId,
		msg.Bech32Prefix,
		msg.IbcDenom,
		msg.UnbondingFrequency,
		msg.MinRedemptionRate,
		msg.MaxRedemptionRate,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterHostZoneResponse{}, nil
}
