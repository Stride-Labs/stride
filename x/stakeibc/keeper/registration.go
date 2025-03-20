package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"

	"github.com/Stride-Labs/stride/v26/utils"
	epochtypes "github.com/Stride-Labs/stride/v26/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v26/x/records/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

var (
	CommunityPoolStakeHoldingAddressKey  = "community-pool-stake"
	CommunityPoolRedeemHoldingAddressKey = "community-pool-redeem"

	DefaultMaxMessagesPerIcaTx = uint64(32)
)

func (k Keeper) RegisterHostZone(ctx sdk.Context, msg *types.MsgRegisterHostZone) (*types.MsgRegisterHostZoneResponse, error) {
	// Get ConnectionEnd (for counterparty connection)
	connectionEnd, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, msg.ConnectionId)
	if !found {
		return nil, errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection-id %s does not exist", msg.ConnectionId)
	}
	counterpartyConnection := connectionEnd.Counterparty

	// Get chain id from connection
	chainId, err := k.GetChainIdFromConnectionId(ctx, msg.ConnectionId)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to obtain chain id from connection %s", msg.ConnectionId)
	}

	// get zone
	_, found = k.GetHostZone(ctx, chainId)
	if found {
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, "host zone already registered for chain-id %s", chainId)
	}

	// check the denom is not already registered
	hostZones := k.GetAllHostZone(ctx)
	for _, hostZone := range hostZones {
		if hostZone.HostDenom == msg.HostDenom {
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, "host denom %s already registered", msg.HostDenom)
		}
		if hostZone.ConnectionId == msg.ConnectionId {
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, "connection-id %s already registered", msg.ConnectionId)
		}
		if hostZone.TransferChannelId == msg.TransferChannelId {
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, "transfer channel %s already registered", msg.TransferChannelId)
		}
		if hostZone.Bech32Prefix == msg.Bech32Prefix {
			return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, "bech32 prefix %s already registered", msg.Bech32Prefix)
		}
	}

	// create and save the zones's module account
	depositAddress := types.NewHostZoneDepositAddress(chainId)
	if err := utils.CreateModuleAccount(ctx, k.AccountKeeper, depositAddress); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to create deposit account for host zone %s", chainId)
	}

	// Create the host zone's community pool holding accounts
	communityPoolStakeAddress := types.NewHostZoneModuleAddress(chainId, CommunityPoolStakeHoldingAddressKey)
	communityPoolRedeemAddress := types.NewHostZoneModuleAddress(chainId, CommunityPoolRedeemHoldingAddressKey)
	if err := utils.CreateModuleAccount(ctx, k.AccountKeeper, communityPoolStakeAddress); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to create community pool stake account for host zone %s", chainId)
	}
	if err := utils.CreateModuleAccount(ctx, k.AccountKeeper, communityPoolRedeemAddress); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to create community pool redeem account for host zone %s", chainId)
	}

	// Validate the community pool treasury address if it's non-empty
	if msg.CommunityPoolTreasuryAddress != "" {
		_, err := utils.AccAddressFromBech32(msg.CommunityPoolTreasuryAddress, msg.Bech32Prefix)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "invalid community pool treasury address (%s)", msg.CommunityPoolTreasuryAddress)
		}
	}

	params := k.GetParams(ctx)
	if msg.MinRedemptionRate.IsNil() || msg.MinRedemptionRate.IsZero() {
		msg.MinRedemptionRate = sdk.NewDecWithPrec(utils.UintToInt(params.DefaultMinRedemptionRateThreshold), 2)
	}
	if msg.MaxRedemptionRate.IsNil() || msg.MaxRedemptionRate.IsZero() {
		msg.MaxRedemptionRate = sdk.NewDecWithPrec(utils.UintToInt(params.DefaultMaxRedemptionRateThreshold), 2)
	}

	// Set the max messages per ICA tx to the default value if it's not specified
	maxMessagesPerIcaTx := msg.MaxMessagesPerIcaTx
	if maxMessagesPerIcaTx == 0 {
		maxMessagesPerIcaTx = DefaultMaxMessagesPerIcaTx
	}

	// set the zone
	zone := types.HostZone{
		ChainId:           chainId,
		ConnectionId:      msg.ConnectionId,
		Bech32Prefix:      msg.Bech32Prefix,
		IbcDenom:          msg.IbcDenom,
		HostDenom:         msg.HostDenom,
		TransferChannelId: msg.TransferChannelId,
		// Start sharesToTokens rate at 1 upon registration
		RedemptionRate:                    sdk.NewDec(1),
		LastRedemptionRate:                sdk.NewDec(1),
		UnbondingPeriod:                   msg.UnbondingPeriod,
		DepositAddress:                    depositAddress.String(),
		CommunityPoolStakeHoldingAddress:  communityPoolStakeAddress.String(),
		CommunityPoolRedeemHoldingAddress: communityPoolRedeemAddress.String(),
		MinRedemptionRate:                 msg.MinRedemptionRate,
		MaxRedemptionRate:                 msg.MaxRedemptionRate,
		// Default the inner bounds to the outer bounds
		MinInnerRedemptionRate:       msg.MinRedemptionRate,
		MaxInnerRedemptionRate:       msg.MaxRedemptionRate,
		LsmLiquidStakeEnabled:        msg.LsmLiquidStakeEnabled,
		CommunityPoolTreasuryAddress: msg.CommunityPoolTreasuryAddress,
		MaxMessagesPerIcaTx:          maxMessagesPerIcaTx,
		RedemptionsEnabled:           true,
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
	delegateAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_DELEGATION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, delegateAccount, appVersion); err != nil {
		return nil, errorsmod.Wrap(err, "failed to register delegation ICA")
	}

	// generate fee account
	feeAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_FEE)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, feeAccount, appVersion); err != nil {
		return nil, errorsmod.Wrap(err, "failed to register fee ICA")
	}

	// generate withdrawal account
	withdrawalAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_WITHDRAWAL)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, withdrawalAccount, appVersion); err != nil {
		return nil, errorsmod.Wrap(err, "failed to register withdrawal ICA")
	}

	// generate redemption account
	redemptionAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_REDEMPTION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, redemptionAccount, appVersion); err != nil {
		return nil, errorsmod.Wrap(err, "failed to register redemption ICA")
	}

	// create community pool deposit account
	communityPoolDepositAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, communityPoolDepositAccount, appVersion); err != nil {
		return nil, errorsmod.Wrap(err, "failed to register community pool deposit ICA")
	}

	// create community pool return account
	communityPoolReturnAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_RETURN)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, communityPoolReturnAccount, appVersion); err != nil {
		return nil, errorsmod.Wrap(err, "failed to register community pool return ICA")
	}

	// add this host zone to unbonding hostZones, otherwise users won't be able to unbond
	// for this host zone until the following day
	dayEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.DAY_EPOCH)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, "epoch tracker (%s) not found", epochtypes.DAY_EPOCH)
	}
	epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, dayEpochTracker.EpochNumber)
	if !found {
		return nil, errorsmod.Wrapf(recordstypes.ErrEpochUnbondingRecordNotFound,
			"epoch unbonding record not found for epoch %d", dayEpochTracker.EpochNumber)
	}
	hostZoneUnbonding := recordstypes.HostZoneUnbonding{
		NativeTokenAmount: sdkmath.ZeroInt(),
		StTokenAmount:     sdkmath.ZeroInt(),
		Denom:             zone.HostDenom,
		HostZoneId:        zone.ChainId,
		Status:            recordstypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	err = k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding)
	if err != nil {
		return nil, err
	}

	// create an empty deposit record for the host zone
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, "epoch tracker (%s) not found", epochtypes.STRIDE_EPOCH)
	}
	depositRecord := recordstypes.DepositRecord{
		Id:                      0,
		Amount:                  sdkmath.ZeroInt(),
		Denom:                   zone.HostDenom,
		HostZoneId:              zone.ChainId,
		Status:                  recordstypes.DepositRecord_TRANSFER_QUEUE,
		DepositEpochNumber:      strideEpochTracker.EpochNumber,
		DelegationTxsInProgress: 0,
	}
	k.RecordsKeeper.AppendDepositRecord(ctx, depositRecord)

	// register stToken to consumer reward denom whitelist so that
	// stToken rewards can be distributed to provider validators
	err = k.RegisterStTokenDenomsToWhitelist(ctx, []string{types.StAssetDenomFromHostZoneDenom(zone.HostDenom)})
	if err != nil {
		return nil, errorsmod.Wrap(err, "unable to register stToken as ICS reward denom")
	}

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
