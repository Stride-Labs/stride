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
func (k msgServer) UpdateValidatorSharesExchRate(goCtx context.Context, msg *types.MsgUpdateValidatorSharesExchRate) (*types.MsgUpdateValidatorSharesExchRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// ensure ICQ can be issued now! else fail the callback
	valid, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return nil, err
	} else if !valid {
		return nil, sdkerrors.Wrapf(types.ErrOutsideIcqWindow, "no host zone found for denom (%s)", msg.ChainId)
	}

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		k.Logger(ctx).Error("Host Zone not found for denom (%s)", msg.ChainId)
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "no host zone found for denom (%s)", msg.ChainId)
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
	return &types.MsgUpdateValidatorSharesExchRateResponse{}, nil
}
