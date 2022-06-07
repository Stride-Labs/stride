package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
func (k Keeper) RegisterHostZone(goCtx context.Context, msg *types.MsgRegisterHostZone) (*types.MsgRegisterHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_ = ctx

	// Get chain id from connection
	chainId, err := k.GetChainID(ctx, msg.ConnectionId)
	if err != nil {
		return nil, fmt.Errorf("unable to obtain chain id: %w", err)
	}

	// get zone
	_, found := k.GetHostZone(ctx, chainId)
	if found {
		return nil, fmt.Errorf("invalid chain id, zone for \"%s\" already registered", chainId)
	}

	// set the zone
	zone := types.HostZone{
		ChainId: chainId,
		ConnectionId: msg.ConnectionId,
		LocalDenom: msg.LocalDenom,
		BaseDenom: msg.BaseDenom,
		// Start exchange rate at 1 upon registration
		RedemptionRate: sdk.NewDec(1),
		LastRedemptionRate: sdk.NewDec(1),
	}
	// write the zone back to the store
	k.SetHostZone(ctx, zone)

	// generate delegate account
	// NOTE: in the future, if we implement proxy governance, we'll need many more delegate accounts
	delegateAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_DELEGATION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, delegateAccount); err != nil {
		return nil, err
	}

	// generate fee account
	feeAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_FEE)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, feeAccount); err != nil {
		return nil, err
	}

	// generate withdrawal account
	withdrawalAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_WITHDRAWAL)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, withdrawalAccount); err != nil {
		return nil, err
	}

	// TODO(TEST-39): TODO(TEST-42): Set validators on the host zone, either using ICQ + intents or a WL 

	// emit events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
		sdk.NewEvent(
			types.EventTypeRegisterZone,
			sdk.NewAttribute(types.AttributeKeyConnectionId, msg.ConnectionId),
			sdk.NewAttribute(types.AttributeKeyRecipientChain, chainId),
		),
	})

	return &types.MsgRegisterHostZoneResponse{}, nil
}
