package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

// todo rename for clarity (this is the validator query : 1st step of daisy chain)
func (k msgServer) QueryDelegation(goCtx context.Context, msg *types.MsgQueryDelegation) (*types.MsgQueryDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, err := k.GetHostZoneFromHostDenom(ctx, msg.Hostdenom)
	if err != nil {
		k.Logger(ctx).Error("Host Zone not found for denom (%s)", msg.Hostdenom)
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "no host zone found for denom (%s)", msg.Hostdenom)
	}

	_, valAddr, _ := bech32.DecodeAndConvert(msg.Valoper)
	data := stakingtypes.GetValidatorKey(valAddr)

	key := "store/staking/key"
	k.Logger(ctx).Info(fmt.Sprintf("Querying validator %v key %v denom %v", valAddr, key, hostZone.HostDenom))
	err = k.InterchainQueryKeeper.MakeRequest(
		ctx,
		hostZone.ConnectionId,
		hostZone.ChainId,
		key,
		data,
		sdk.NewInt(-1),
		types.ModuleName,
		"validator",
		0, // ttl
		0, // height always 0 (which means current height)
	)
	if err != nil {
		k.Logger(ctx).Error("Error querying for validator", "error", err)
		return nil, err
	}
	return &types.MsgQueryDelegationResponse{}, nil
}
