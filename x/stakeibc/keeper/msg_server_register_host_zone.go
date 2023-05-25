package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/Stride-Labs/stride/v9/utils"
	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RegisterHostZone(goCtx context.Context, msg *types.MsgRegisterHostZone) (*types.MsgRegisterHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get ConnectionEnd (for counterparty connection)
	connectionEnd, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, msg.ConnectionId)
	if !found {
		errMsg := fmt.Sprintf("invalid connection id, %s not found", msg.ConnectionId)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}
	counterpartyConnection := connectionEnd.Counterparty

	// Get chain id from connection
	chainId, err := k.GetChainID(ctx, msg.ConnectionId)
	if err != nil {
		errMsg := fmt.Sprintf("unable to obtain chain id from connection %s, err: %s", msg.ConnectionId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// get zone
	_, found = k.GetHostZone(ctx, chainId)
	if found {
		errMsg := fmt.Sprintf("invalid chain id, zone for %s already registered", chainId)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// check the denom is not already registered
	hostZones := k.GetAllHostZone(ctx)
	for _, hostZone := range hostZones {
		if hostZone.HostDenom == msg.HostDenom {
			errMsg := fmt.Sprintf("host denom %s already registered", msg.HostDenom)
			k.Logger(ctx).Error(errMsg)
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
		}
		if hostZone.ConnectionId == msg.ConnectionId {
			errMsg := fmt.Sprintf("connectionId %s already registered", msg.ConnectionId)
			k.Logger(ctx).Error(errMsg)
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
		}
		if hostZone.TransferChannelId == msg.TransferChannelId {
			errMsg := fmt.Sprintf("transfer channel %s already registered", msg.TransferChannelId)
			k.Logger(ctx).Error(errMsg)
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
		}
		if hostZone.Bech32Prefix == msg.Bech32Prefix {
			errMsg := fmt.Sprintf("bech32prefix %s already registered", msg.Bech32Prefix)
			k.Logger(ctx).Error(errMsg)
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
		}
	}

	// create and save the zones's module account
	zoneAddress := types.NewZoneAddress(chainId)
	if err := utils.CreateModuleAccount(ctx, k.accountKeeper, zoneAddress); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to create module account for host zone %s", chainId)
	}

	params := k.GetParams(ctx)
	if msg.MinRedemptionRate.IsNil() || msg.MinRedemptionRate.IsZero() {
		msg.MinRedemptionRate = sdk.NewDecWithPrec(int64(params.DefaultMinRedemptionRateThreshold), 2)
	}
	if msg.MaxRedemptionRate.IsNil() || msg.MaxRedemptionRate.IsZero() {
		msg.MaxRedemptionRate = sdk.NewDecWithPrec(int64(params.DefaultMaxRedemptionRateThreshold), 2)
	}

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
		MinRedemptionRate:  msg.MinRedemptionRate,
		MaxRedemptionRate:  msg.MaxRedemptionRate,
	}
	// write the zone back to the store
	k.SetHostZone(ctx, zone)

	appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: zone.ConnectionId,
		HostConnectionId:       counterpartyConnection.ConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))

	// generate delegate account
	// NOTE: in the future, if we implement proxy governance, we'll need many more delegate accounts
	delegateAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_DELEGATION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, delegateAccount, appVersion); err != nil {
		errMsg := fmt.Sprintf("unable to register delegation account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// generate fee account
	feeAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_FEE)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, feeAccount, appVersion); err != nil {
		errMsg := fmt.Sprintf("unable to register fee account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// generate withdrawal account
	withdrawalAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_WITHDRAWAL)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, withdrawalAccount, appVersion); err != nil {
		errMsg := fmt.Sprintf("unable to register withdrawal account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// generate redemption account
	redemptionAccount := types.FormatICAAccountOwner(chainId, types.ICAAccountType_REDEMPTION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, redemptionAccount, appVersion); err != nil {
		errMsg := fmt.Sprintf("unable to register redemption account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// add this host zone to unbonding hostZones, otherwise users won't be able to unbond
	// for this host zone until the following day
	dayEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.DAY_EPOCH)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, "epoch tracker (%s) not found", epochtypes.DAY_EPOCH)
	}
	epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, dayEpochTracker.EpochNumber)
	if !found {
		errMsg := "unable to find latest epoch unbonding record"
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(recordstypes.ErrEpochUnbondingRecordNotFound, errMsg)
	}
	hostZoneUnbonding := &recordstypes.HostZoneUnbonding{
		NativeTokenAmount: sdkmath.ZeroInt(),
		StTokenAmount:     sdkmath.ZeroInt(),
		Denom:             zone.HostDenom,
		HostZoneId:        zone.ChainId,
		Status:            recordstypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	updatedEpochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding)
	if !success {
		errMsg := fmt.Sprintf("Failed to set host zone epoch unbonding record: epochNumber %d, chainId %s, hostZoneUnbonding %v. Err: %s",
			epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding, err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, errMsg)
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)

	// create an empty deposit record for the host zone
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, "epoch tracker (%s) not found", epochtypes.STRIDE_EPOCH)
	}
	depositRecord := recordstypes.DepositRecord{
		Id:                 0,
		Amount:             sdkmath.ZeroInt(),
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
