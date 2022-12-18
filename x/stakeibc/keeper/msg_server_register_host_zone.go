package keeper

import (
	"context"
	"fmt"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RegisterHostZone(goCtx context.Context, msg *types.MsgRegisterHostZone) (*types.MsgRegisterHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get chain id from connection
	chainId, err := k.GetChainID(ctx, msg.ConnectionId)
	if err != nil {
		errMsg := fmt.Sprintf("unable to obtain chain id from connection %s, err: %s", msg.ConnectionId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// get zone
	_, found := k.GetHostZone(ctx, chainId)
	if found {
		errMsg := fmt.Sprintf("invalid chain id, zone for %s already registered", chainId)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// check the denom is not already registered
	hostZones := k.GetAllHostZone(ctx)
	for _, hostZone := range hostZones {
		if hostZone.HostDenom == msg.HostDenom {
			errMsg := fmt.Sprintf("host denom %s already registered", msg.HostDenom)
			k.Logger(ctx).Error(errMsg)
			return nil, sdkerrors.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
		}
		if hostZone.ConnectionId == msg.ConnectionId {
			errMsg := fmt.Sprintf("connectionId %s already registered", msg.ConnectionId)
			k.Logger(ctx).Error(errMsg)
			return nil, sdkerrors.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
		}
		if hostZone.Bech32Prefix == msg.Bech32Prefix {
			errMsg := fmt.Sprintf("bech32prefix %s already registered", msg.Bech32Prefix)
			k.Logger(ctx).Error(errMsg)
			return nil, sdkerrors.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
		}
	}

	// create and save the zones's module account to the account keeper
	zoneAddress := types.NewZoneAddress(chainId)
	acc := k.accountKeeper.NewAccount(
		ctx,
		authtypes.NewModuleAccount(
			authtypes.NewBaseAccountWithAddress(zoneAddress),
			zoneAddress.String(),
		),
	)
	k.accountKeeper.SetAccount(ctx, acc)

	// set the zone
	zone := types.HostZone{
		ChainId:           chainId,
		ConnectionId:      msg.ConnectionId,
		Bech32Prefix:      msg.Bech32Prefix,
		IbcDenom:          msg.IbcDenom,
		HostDenom:         msg.HostDenom,
		TransferChannelId: msg.TransferChannelId,
		// Start exchange rate at 1 upon registration
		RedemptionRate:     sdk.NewDec(1),
		LastRedemptionRate: sdk.NewDec(1),
		UnbondingFrequency: msg.UnbondingFrequency,
		Address:            zoneAddress.String(),
	}
	// write the zone back to the store
	k.SetHostZone(ctx, zone)

	// generate delegate account
	// NOTE: in the future, if we implement proxy governance, we'll need many more delegate accounts
	delegateAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_DELEGATION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, delegateAccount); err != nil {
		errMsg := fmt.Sprintf("unable to register delegation account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// generate fee account
	feeAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_FEE)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, feeAccount); err != nil {
		errMsg := fmt.Sprintf("unable to register fee account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// generate withdrawal account
	withdrawalAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_WITHDRAWAL)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, withdrawalAccount); err != nil {
		errMsg := fmt.Sprintf("unable to register withdrawal account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// generate redemption account
	redemptionAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_REDEMPTION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, redemptionAccount); err != nil {
		errMsg := fmt.Sprintf("unable to register redemption account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// add this host zone to unbonding hostZones, otherwise users won't be able to unbond
	// for this host zone until the following day
	dayEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.DAY_EPOCH)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrEpochNotFound, "epoch tracker (%s) not found", epochtypes.DAY_EPOCH)
	}
	epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, dayEpochTracker.EpochNumber)
	if !found {
		errMsg := "unable to find latest epoch unbonding record"
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(recordstypes.ErrEpochUnbondingRecordNotFound, errMsg)
	}
	hostZoneUnbonding := &recordstypes.HostZoneUnbonding{
		NativeTokenAmount: sdk.ZeroInt(),
		StTokenAmount:     sdk.ZeroInt(),
		Denom:             zone.HostDenom,
		HostZoneId:        zone.ChainId,
		Status:            recordstypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	updatedEpochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding)
	if !success {
		errMsg := fmt.Sprintf("Failed to set host zone epoch unbonding record: epochNumber %d, chainId %s, hostZoneUnbonding %v. Err: %s",
			epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding, err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrEpochNotFound, errMsg)
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)

	// create an empty deposit record for the host zone
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrEpochNotFound, "epoch tracker (%s) not found", epochtypes.STRIDE_EPOCH)
	}
	depositRecord := recordstypes.DepositRecord{
		Id:                 0,
		Amount:             sdk.ZeroInt(),
		Denom:              zone.HostDenom,
		HostZoneId:         zone.ChainId,
		Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
		DepositEpochNumber: strideEpochTracker.EpochNumber,
	}
	k.RecordsKeeper.AppendDepositRecord(ctx, depositRecord)

	// emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRegisterZone,
			sdk.NewAttribute(types.AttributeKeyConnectionId, msg.ConnectionId),
			sdk.NewAttribute(types.AttributeKeyRecipientChain, chainId),
		),
	)

	return &types.MsgRegisterHostZoneResponse{}, nil
}
