package keeper

import (
	"context"
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

	// check the denom is not already registered
	hostZones := k.GetAllHostZone(ctx)
	for _, hostZone := range hostZones {
		if hostZone.HostDenom == msg.HostDenom {
			return nil, fmt.Errorf("host denom \"%s\" already registered", msg.HostDenom)
		}
	}

	// set the zone
	zone := types.HostZone{
		ChainId:           chainId,
		ConnectionId:      msg.ConnectionId,
		Bech32Prefix:      msg.Bech32Prefix,
		IBCDenom:          msg.IbcDenom,
		HostDenom:         msg.HostDenom,
		TransferChannelId: msg.TransferChannelId,
		// Start exchange rate at 1 upon registration
		RedemptionRate:     sdk.NewDec(1),
		LastRedemptionRate: sdk.NewDec(1),
		UnbondingFrequency: msg.UnbondingFrequency,
	}
	// write the zone back to the store
	k.SetHostZone(ctx, zone)

	// generate delegate account
	// NOTE: in the future, if we implement proxy governance, we'll need many more delegate accounts
	delegateAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_DELEGATION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, delegateAccount); err != nil {
		k.Logger(ctx).Error("unable to register delegate account", "error", err)
		return nil, err
	}

	// generate fee account
	feeAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_FEE)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, feeAccount); err != nil {
		k.Logger(ctx).Error("unable to register fee account", "error", err)
		return nil, err
	}

	// generate withdrawal account
	withdrawalAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_WITHDRAWAL)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, withdrawalAccount); err != nil {
		k.Logger(ctx).Error("unable to register withdrawal account", "error", err)
		return nil, err
	}

	// generate redemption account
	redemptionAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_REDEMPTION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, redemptionAccount); err != nil {
		k.Logger(ctx).Error("unable to register redemption account", "error", err)
		return nil, err
	}

	// add this host zone to unbonding hostZones, otherwise users won't be able to unbond
	// for this host zone until the following day
	epochUnbondingRecord, found := k.RecordsKeeper.GetLatestEpochUnbondingRecord(ctx)
	if !found {
		errMsg := "unable to add host zone to latest epoch unbonding record"
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(recordstypes.ErrEpochUnbondingRecordNotFound, errMsg)
	}
	hostZoneUnbondings := epochUnbondingRecord.GetHostZoneUnbondings()
	if len(hostZoneUnbondings) == 0 {
		hostZoneUnbondings = make(map[string]*recordstypes.HostZoneUnbonding)
	}
	hostZoneUnbondings[zone.ChainId] = &recordstypes.HostZoneUnbonding{
		Amount:     0,
		Denom:      zone.HostDenom,
		HostZoneId: zone.ChainId,
		Status:     recordstypes.HostZoneUnbonding_BONDED,
	}
	epochUnbondingRecord.HostZoneUnbondings = hostZoneUnbondings
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	epochUnbondingRecordNew, found := k.RecordsKeeper.GetLatestEpochUnbondingRecord(ctx)
	if !found {
		errMsg := "unable to add host zone to latest epoch unbonding record"
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(recordstypes.ErrEpochUnbondingRecordNotFound, errMsg)
	}
	k.Logger(ctx).Info(fmt.Sprintf("hostZoneUnbondings after register host zone %v", epochUnbondingRecordNew.GetHostZoneUnbondings()))

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
