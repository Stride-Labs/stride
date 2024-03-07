package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proto "github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v19/utils"
	epochtypes "github.com/Stride-Labs/stride/v19/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v19/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v19/x/records/types"
	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

var (
	CommunityPoolStakeHoldingAddressKey  = "community-pool-stake"
	CommunityPoolRedeemHoldingAddressKey = "community-pool-redeem"

	DefaultMaxAllowedSwapLossRate = "0.05"
	DefaultMaxSwapAmount          = sdkmath.NewIntWithDecimal(10, 24) // 10e24
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

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
	chainId, err := k.GetChainIdFromConnectionId(ctx, msg.ConnectionId)
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
		MinInnerRedemptionRate: msg.MinRedemptionRate,
		MaxInnerRedemptionRate: msg.MaxRedemptionRate,
		LsmLiquidStakeEnabled:  msg.LsmLiquidStakeEnabled,
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
		errMsg := fmt.Sprintf("unable to register delegation account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// generate fee account
	feeAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_FEE)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, feeAccount, appVersion); err != nil {
		errMsg := fmt.Sprintf("unable to register fee account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// generate withdrawal account
	withdrawalAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_WITHDRAWAL)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, withdrawalAccount, appVersion); err != nil {
		errMsg := fmt.Sprintf("unable to register withdrawal account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// generate redemption account
	redemptionAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_REDEMPTION)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, redemptionAccount, appVersion); err != nil {
		errMsg := fmt.Sprintf("unable to register redemption account, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
	}

	// create community pool deposit account
	communityPoolDepositAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, communityPoolDepositAccount, appVersion); err != nil {
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, "failed to register community pool deposit ICA")
	}

	// create community pool return account
	communityPoolReturnAccount := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_RETURN)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, zone.ConnectionId, communityPoolReturnAccount, appVersion); err != nil {
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, "failed to register community pool return ICA")
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

	// register stToken to consumer reward denom whitelist so that
	// stToken rewards can be distributed to provider validators
	err = k.RegisterStTokenDenomsToWhitelist(ctx, []string{types.StAssetDenomFromHostZoneDenom(zone.HostDenom)})
	if err != nil {
		errMsg := fmt.Sprintf("unable to register reward denom, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrFailedToRegisterHostZone, errMsg)
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

func (k msgServer) AddValidators(goCtx context.Context, msg *types.MsgAddValidators) (*types.MsgAddValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, validator := range msg.Validators {
		if err := k.AddValidatorToHostZone(ctx, msg.HostZone, *validator, false); err != nil {
			return nil, err
		}

		// Query and store the validator's sharesToTokens rate
		if err := k.QueryValidatorSharesToTokensRate(ctx, msg.HostZone, validator.Address); err != nil {
			return nil, err
		}
	}

	return &types.MsgAddValidatorsResponse{}, nil
}

func (k msgServer) DeleteValidator(goCtx context.Context, msg *types.MsgDeleteValidator) (*types.MsgDeleteValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.RemoveValidatorFromHostZone(ctx, msg.HostZone, msg.ValAddr)
	if err != nil {
		errMsg := fmt.Sprintf("Validator (%s) not removed from host zone (%s) | err: %s", msg.ValAddr, msg.HostZone, err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrValidatorNotRemoved, errMsg)
	}

	return &types.MsgDeleteValidatorResponse{}, nil
}

func (k msgServer) ChangeValidatorWeight(goCtx context.Context, msg *types.MsgChangeValidatorWeights) (*types.MsgChangeValidatorWeightsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, types.ErrInvalidHostZone
	}

	for _, weightChange := range msg.ValidatorWeights {

		validatorFound := false
		for _, validator := range hostZone.Validators {
			if validator.Address == weightChange.Address {
				validator.Weight = weightChange.Weight
				k.SetHostZone(ctx, hostZone)

				validatorFound = true
				break
			}
		}

		if !validatorFound {
			return nil, types.ErrValidatorNotFound
		}
	}

	// Confirm the new weights wouldn't cause any validator to exceed the weight cap
	if err := k.CheckValidatorWeightsBelowCap(ctx, hostZone.Validators); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to change validator weight")
	}

	return &types.MsgChangeValidatorWeightsResponse{}, nil
}

func (k msgServer) RebalanceValidators(goCtx context.Context, msg *types.MsgRebalanceValidators) (*types.MsgRebalanceValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("RebalanceValidators executing %v", msg))

	if err := k.RebalanceDelegationsForHostZone(ctx, msg.HostZone); err != nil {
		return nil, err
	}
	return &types.MsgRebalanceValidatorsResponse{}, nil
}

func (k msgServer) ClearBalance(goCtx context.Context, msg *types.MsgClearBalance) (*types.MsgClearBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	zone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInvalidHostZone, "chainId: %s", msg.ChainId)
	}
	if zone.FeeIcaAddress == "" {
		return nil, errorsmod.Wrapf(types.ErrICAAccountNotFound, "fee acount not found for chainId: %s", msg.ChainId)
	}

	sourcePort := ibctransfertypes.PortID
	// Should this be a param?
	// I think as long as we have a timeout on this, it should be hard to attack (even if someone send a tx on a bad channel, it would be reverted relatively quickly)
	sourceChannel := msg.Channel
	coinString := cast.ToString(msg.Amount) + zone.GetHostDenom()
	tokens, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to parse coin (%s)", coinString))
		return nil, errorsmod.Wrapf(err, "failed to parse coin (%s)", coinString)
	}
	// KeyICATimeoutNanos are for our Stride ICA calls, KeyFeeTransferTimeoutNanos is for the IBC transfer
	feeTransferTimeoutNanos := k.GetParam(ctx, types.KeyFeeTransferTimeoutNanos)
	timeoutTimestamp := cast.ToUint64(ctx.BlockTime().UnixNano()) + feeTransferTimeoutNanos
	msgs := []proto.Message{
		&ibctransfertypes.MsgTransfer{
			SourcePort:       sourcePort,
			SourceChannel:    sourceChannel,
			Token:            tokens,
			Sender:           zone.FeeIcaAddress, // fee account on the host zone
			Receiver:         types.FeeAccount,   // fee account on stride
			TimeoutTimestamp: timeoutTimestamp,
		},
	}

	connectionId := zone.GetConnectionId()

	icaTimeoutNanos := k.GetParam(ctx, types.KeyICATimeoutNanos)
	icaTimeoutNanos = cast.ToUint64(ctx.BlockTime().UnixNano()) + icaTimeoutNanos

	_, err = k.SubmitTxs(ctx, connectionId, msgs, types.ICAAccountType_FEE, icaTimeoutNanos, "", nil)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to submit txs")
	}
	return &types.MsgClearBalanceResponse{}, nil
}

// Exchanges a user's native tokens for stTokens using the current redemption rate
// The native tokens must live on Stride with an IBC denomination before this function is called
// The typical flow consists, first, of a transfer of native tokens from the host zone to Stride,
//
//	and then the invocation of this LiquidStake function
//
// WARNING: This function is invoked from the begin/end blocker in a way that does not revert partial state when
//
//	an error is thrown (i.e. the execution is non-atomic).
//	As a result, it is important that the validation steps are positioned at the top of the function,
//	and logic that creates state changes (e.g. bank sends, mint) appear towards the end of the function
func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the host zone from the base denom in the message (e.g. uatom)
	hostZone, err := k.GetHostZoneFromHostDenom(ctx, msg.HostDenom)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidToken, "no host zone found for denom (%s)", msg.HostDenom)
	}

	// Error immediately if the host zone is halted
	if hostZone.Halted {
		return nil, errorsmod.Wrapf(types.ErrHaltedHostZone, "halted host zone found for denom (%s)", msg.HostDenom)
	}

	// Get user and module account addresses
	liquidStakerAddress, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "user's address is invalid")
	}
	hostZoneDepositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "host zone address is invalid")
	}

	// Safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, *hostZone)
	if !rateIsSafe || (err != nil) {
		return nil, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, "HostZone: %s, err: %s", hostZone.ChainId, err.Error())
	}

	// Grab the deposit record that will be used for record keeping
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	depositRecord, found := k.RecordsKeeper.GetTransferDepositRecordByEpochAndChain(ctx, strideEpochTracker.EpochNumber, hostZone.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "no deposit record for epoch (%d)", strideEpochTracker.EpochNumber)
	}

	// The tokens that are sent to the protocol are denominated in the ibc hash of the native token on stride (e.g. ibc/xxx)
	nativeDenom := hostZone.IbcDenom
	nativeCoin := sdk.NewCoin(nativeDenom, msg.Amount)
	if !types.IsIBCToken(nativeDenom) {
		return nil, errorsmod.Wrapf(types.ErrInvalidToken, "denom is not an IBC token (%s)", nativeDenom)
	}

	// Confirm the user has a sufficient balance to execute the liquid stake
	balance := k.bankKeeper.GetBalance(ctx, liquidStakerAddress, nativeDenom)
	if balance.IsLT(nativeCoin) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "balance is lower than staking amount. staking amount: %v, balance: %v", msg.Amount, balance.Amount)
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stAmount := (sdk.NewDecFromInt(msg.Amount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return nil, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}

	// Transfer the native tokens from the user to module account
	if err := k.bankKeeper.SendCoins(ctx, liquidStakerAddress, hostZoneDepositAddress, sdk.NewCoins(nativeCoin)); err != nil {
		return nil, errorsmod.Wrap(err, "failed to send tokens from Account to Module")
	}

	// Mint the stTokens and transfer them to the user
	stDenom := types.StAssetDenomFromHostZoneDenom(msg.HostDenom)
	stCoin := sdk.NewCoin(stDenom, stAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to mint coins")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStakerAddress, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to send %s from module to account", stCoin.String())
	}

	// Update the liquid staked amount on the deposit record
	depositRecord.Amount = depositRecord.Amount.Add(msg.Amount)
	k.RecordsKeeper.SetDepositRecord(ctx, *depositRecord)

	// Emit liquid stake event
	EmitSuccessfulLiquidStakeEvent(ctx, msg, *hostZone, stAmount)

	k.hooks.AfterLiquidStake(ctx, liquidStakerAddress)
	return &types.MsgLiquidStakeResponse{StToken: stCoin}, nil
}

func (k msgServer) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("redeem stake: %s", msg.String()))

	// ----------------- PRELIMINARY CHECKS -----------------
	// get our addresses, make sure they're valid
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "creator address is invalid: %s. err: %s", msg.Creator, err.Error())
	}
	// then make sure host zone is valid
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInvalidHostZone, "host zone is invalid: %s", msg.HostZone)
	}

	if hostZone.Halted {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone halted for zone (%s)", msg.HostZone))
		return nil, errorsmod.Wrapf(types.ErrHaltedHostZone, "halted host zone found for zone (%s)", msg.HostZone)
	}

	// first construct a user redemption record
	epochTracker, found := k.GetEpochTracker(ctx, "day")
	if !found {
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, "epoch tracker found: %s", "day")
	}

	// ensure the recipient address is a valid bech32 address on the hostZone
	_, err = utils.AccAddressFromBech32(msg.Receiver, hostZone.Bech32Prefix)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	// construct desired unstaking amount from host zone
	// TODO [cleanup]: Consider changing to truncate int
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	nativeAmount := sdk.NewDecFromInt(msg.Amount).Mul(hostZone.RedemptionRate).RoundInt()

	if nativeAmount.GT(hostZone.TotalDelegations) {
		return nil, errorsmod.Wrapf(types.ErrInvalidAmount, "cannot unstake an amount g.t. staked balance on host zone: %v", msg.Amount)
	}

	// safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, hostZone)
	if !rateIsSafe || (err != nil) {
		errMsg := fmt.Sprintf("IsRedemptionRateWithinSafetyBounds check failed. hostZone: %s, err: %s", hostZone.String(), err.Error())
		return nil, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, errMsg)
	}

	// safety checks on the coin
	// 	- Redemption amount must be positive
	if !nativeAmount.IsPositive() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %v", msg.Amount)
	}
	// 	- Creator owns at least "amount" stAssets
	balance := k.bankKeeper.GetBalance(ctx, sender, stDenom)
	if balance.Amount.LT(msg.Amount) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "balance is lower than redemption amount. redemption amount: %v, balance %v: ", msg.Amount, balance.Amount)
	}

	// ----------------- UNBONDING RECORD KEEPING -----------------
	// Fetch the record
	redemptionId := recordstypes.UserRedemptionRecordKeyFormatter(hostZone.ChainId, epochTracker.EpochNumber, msg.Receiver)
	userRedemptionRecord, userHasRedeemedThisEpoch := k.RecordsKeeper.GetUserRedemptionRecord(ctx, redemptionId)
	if userHasRedeemedThisEpoch {
		k.Logger(ctx).Info(fmt.Sprintf("UserRedemptionRecord found for %s", redemptionId))
		// Add the unbonded amount to the UserRedemptionRecord
		// The record is set below
		userRedemptionRecord.StTokenAmount = userRedemptionRecord.StTokenAmount.Add(msg.Amount)
		userRedemptionRecord.NativeTokenAmount = userRedemptionRecord.NativeTokenAmount.Add(nativeAmount)
	} else {
		// First time a user is redeeming this epoch
		userRedemptionRecord = recordstypes.UserRedemptionRecord{
			Id:                redemptionId,
			Receiver:          msg.Receiver,
			NativeTokenAmount: nativeAmount,
			Denom:             hostZone.HostDenom,
			HostZoneId:        hostZone.ChainId,
			EpochNumber:       epochTracker.EpochNumber,
			StTokenAmount:     msg.Amount,
			// claimIsPending represents whether a redemption is currently being claimed,
			// contingent on the host zone unbonding having status CLAIMABLE
			ClaimIsPending: false,
		}
		k.Logger(ctx).Info(fmt.Sprintf("UserRedemptionRecord not found - creating for %s", redemptionId))
	}

	// then add undelegation amount to epoch unbonding records
	epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochTracker.EpochNumber)
	if !found {
		k.Logger(ctx).Error("latest epoch unbonding record not found")
		return nil, errorsmod.Wrapf(recordstypes.ErrEpochUnbondingRecordNotFound, "latest epoch unbonding record not found")
	}
	// get relevant host zone on this epoch unbonding record
	hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInvalidHostZone, "host zone not found in unbondings: %s", hostZone.ChainId)
	}
	hostZoneUnbonding.NativeTokenAmount = hostZoneUnbonding.NativeTokenAmount.Add(nativeAmount)
	if !userHasRedeemedThisEpoch {
		// Only append a UserRedemptionRecord to the HZU if it wasn't previously appended
		hostZoneUnbonding.UserRedemptionRecords = append(hostZoneUnbonding.UserRedemptionRecords, userRedemptionRecord.Id)
	}

	// Escrow user's balance
	redeemCoin := sdk.NewCoins(sdk.NewCoin(stDenom, msg.Amount))
	depositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return nil, fmt.Errorf("could not bech32 decode address %s of zone with id: %s", hostZone.DepositAddress, hostZone.ChainId)
	}
	err = k.bankKeeper.SendCoins(ctx, sender, depositAddress, redeemCoin)
	if err != nil {
		k.Logger(ctx).Error("Failed to send sdk.NewCoins(inCoins) from account to module")
		return nil, errorsmod.Wrapf(types.ErrInsufficientFunds, "couldn't send %v derivative %s tokens to module account. err: %s", msg.Amount, hostZone.HostDenom, err.Error())
	}

	// record the number of stAssets that should be burned after unbonding
	hostZoneUnbonding.StTokenAmount = hostZoneUnbonding.StTokenAmount.Add(msg.Amount)

	// Actually set the records, we wait until now to prevent any errors
	k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)

	// Set the UserUnbondingRecords on the proper HostZoneUnbondingRecord
	hostZoneUnbondings := epochUnbondingRecord.GetHostZoneUnbondings()
	if hostZoneUnbondings == nil {
		hostZoneUnbondings = []*recordstypes.HostZoneUnbonding{}
		epochUnbondingRecord.HostZoneUnbondings = hostZoneUnbondings
	}
	updatedEpochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId, hostZoneUnbonding)
	if !success {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to set host zone epoch unbonding record: epochNumber %d, chainId %s, hostZoneUnbonding %v", epochUnbondingRecord.EpochNumber, hostZone.ChainId, hostZoneUnbonding))
		return nil, errorsmod.Wrapf(types.ErrEpochNotFound, "couldn't set host zone epoch unbonding record. err: %s", err.Error())
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)

	k.Logger(ctx).Info(fmt.Sprintf("executed redeem stake: %s", msg.String()))
	return &types.MsgRedeemStakeResponse{}, nil
}

// Exchanges a user's LSM tokenized shares for stTokens using the current redemption rate
// The LSM tokens must live on Stride as an IBC voucher (whose denomtrace we recognize)
// before this function is called
//
// The typical flow:
//   - A staker tokenizes their delegation on the host zone
//   - The staker IBC transfers their tokenized shares to Stride
//   - They then call LSMLiquidStake
//   - - The staker's LSM Tokens are sent to the Stride module account
//   - - The staker recieves stTokens
//
// As a safety measure, at period checkpoints, the validator's sharesToTokens rate is queried and the transaction
// is not settled until the query returns
// As a result, this transaction has been split up into a (1) Start and (2) Finish function
//   - If no query is needed, (2) is called immediately after (1)
//   - If a query is needed, (2) is called in the query callback
//
// The transaction response indicates if the query occurred by returning an attribute `TransactionComplete` set to false
func (k msgServer) LSMLiquidStake(goCtx context.Context, msg *types.MsgLSMLiquidStake) (*types.MsgLSMLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lsmLiquidStake, err := k.StartLSMLiquidStake(ctx, *msg)
	if err != nil {
		return nil, err
	}

	if k.ShouldCheckIfValidatorWasSlashed(ctx, *lsmLiquidStake.Validator, msg.Amount) {
		if err := k.SubmitValidatorSlashQuery(ctx, lsmLiquidStake); err != nil {
			return nil, err
		}

		EmitPendingLSMLiquidStakeEvent(ctx, *lsmLiquidStake.HostZone, *lsmLiquidStake.Deposit)

		return &types.MsgLSMLiquidStakeResponse{TransactionComplete: false}, nil
	}

	async := false
	if err := k.FinishLSMLiquidStake(ctx, lsmLiquidStake, async); err != nil {
		return nil, err
	}

	return &types.MsgLSMLiquidStakeResponse{TransactionComplete: true}, nil
}

// Gov tx to register a trade route that swaps reward tokens for a different denom
//
// Example proposal:
//
//		{
//		   "title": "Create a new trade route for host chain X",
//		   "metadata": "Create a new trade route for host chain X",
//		   "summary": "Create a new trade route for host chain X",
//		   "messages":[
//		      {
//		         "@type": "/stride.stakeibc.MsgCreateTradeRoute",
//		         "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
//
//		         "stride_to_host_connection_id": "connection-0",
//				 "stride_to_reward_connection_id": "connection-1",
//			     "stride_to_trade_connection_id": "connection-2",
//
//				 "host_to_reward_transfer_channel_id": "channel-0",
//				 "reward_to_trade_transfer_channel_id": "channel-1",
//			     "trade_to_host_transfer_channel_id": "channel-2",
//
//				 "reward_denom_on_host": "ibc/rewardTokenXXX",
//				 "reward_denom_on_reward": "rewardToken",
//				 "reward_denom_on_trade": "ibc/rewardTokenYYY",
//				 "host_denom_on_trade": "ibc/hostTokenZZZ",
//				 "host_denom_on_host": "hostToken",
//
//				 "pool_id": 1,
//				 "max_allowed_swap_loss_rate": "0.05"
//				 "min_swap_amount": "10000000",
//				 "max_swap_amount": "1000000000"
//			  }
//		   ],
//		   "deposit": "2000000000ustrd"
//	   }
//
// >>> strided tx gov submit-proposal {proposal_file.json} --from wallet
func (ms msgServer) CreateTradeRoute(goCtx context.Context, msg *types.MsgCreateTradeRoute) (*types.MsgCreateTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	// Validate trade route does not already exist for this denom
	_, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenomOnReward, msg.HostDenomOnHost)
	if found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteAlreadyExists,
			"trade route already exists for rewardDenom %s, hostDenom %s", msg.RewardDenomOnReward, msg.HostDenomOnHost)
	}

	// Confirm the host chain exists and the withdrawal address has been initialized
	hostZone, err := ms.Keeper.GetActiveHostZone(ctx, msg.HostChainId)
	if err != nil {
		return nil, err
	}
	if hostZone.WithdrawalIcaAddress == "" {
		return nil, errorsmod.Wrapf(types.ErrICAAccountNotFound, "withdrawal account not initialized on host zone")
	}

	// Register the new ICA accounts
	tradeRouteId := types.GetTradeRouteId(msg.RewardDenomOnReward, msg.HostDenomOnHost)
	hostICA := types.ICAAccount{
		ChainId:      msg.HostChainId,
		Type:         types.ICAAccountType_WITHDRAWAL,
		ConnectionId: hostZone.ConnectionId,
		Address:      hostZone.WithdrawalIcaAddress,
	}

	unwindConnectionId := msg.StrideToRewardConnectionId
	unwindICAType := types.ICAAccountType_CONVERTER_UNWIND
	unwindICA, err := ms.Keeper.RegisterTradeRouteICAAccount(ctx, tradeRouteId, unwindConnectionId, unwindICAType)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register the unwind ICA account")
	}

	tradeConnectionId := msg.StrideToTradeConnectionId
	tradeICAType := types.ICAAccountType_CONVERTER_TRADE
	tradeICA, err := ms.Keeper.RegisterTradeRouteICAAccount(ctx, tradeRouteId, tradeConnectionId, tradeICAType)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register the trade ICA account")
	}

	// If a max allowed swap loss is not provided, use the default
	maxAllowedSwapLossRate := msg.MaxAllowedSwapLossRate
	if maxAllowedSwapLossRate == "" {
		maxAllowedSwapLossRate = DefaultMaxAllowedSwapLossRate
	}
	maxSwapAmount := msg.MaxSwapAmount
	if maxSwapAmount.IsZero() {
		maxSwapAmount = DefaultMaxSwapAmount
	}

	// Create the trade config to specify parameters needed for the swap
	tradeConfig := types.TradeConfig{
		PoolId:               msg.PoolId,
		SwapPrice:            sdk.ZeroDec(), // this should only ever be set by ICQ so initialize to blank
		PriceUpdateTimestamp: 0,

		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr(maxAllowedSwapLossRate),
		MinSwapAmount:          msg.MinSwapAmount,
		MaxSwapAmount:          maxSwapAmount,
	}

	// Finally build and store the main trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnHostZone:   msg.RewardDenomOnHost,
		RewardDenomOnRewardZone: msg.RewardDenomOnReward,
		RewardDenomOnTradeZone:  msg.RewardDenomOnTrade,
		HostDenomOnTradeZone:    msg.HostDenomOnTrade,
		HostDenomOnHostZone:     msg.HostDenomOnHost,

		HostAccount:   hostICA,
		RewardAccount: unwindICA,
		TradeAccount:  tradeICA,

		HostToRewardChannelId:  msg.HostToRewardTransferChannelId,
		RewardToTradeChannelId: msg.RewardToTradeTransferChannelId,
		TradeToHostChannelId:   msg.TradeToHostTransferChannelId,

		TradeConfig: tradeConfig,
	}

	ms.Keeper.SetTradeRoute(ctx, tradeRoute)

	return &types.MsgCreateTradeRouteResponse{}, nil
}

// Gov tx to remove a trade route
//
// Example proposal:
//
//		{
//		   "title": "Remove a new trade route for host chain X",
//		   "metadata": "Remove a new trade route for host chain X",
//		   "summary": "Remove a new trade route for host chain X",
//		   "messages":[
//		      {
//		         "@type": "/stride.stakeibc.MsgDeleteTradeRoute",
//		         "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
//				 "reward_denom": "rewardToken",
//				 "host_denom": "hostToken
//			  }
//		   ],
//		   "deposit": "2000000000ustrd"
//	   }
//
// >>> strided tx gov submit-proposal {proposal_file.json} --from wallet
func (ms msgServer) DeleteTradeRoute(goCtx context.Context, msg *types.MsgDeleteTradeRoute) (*types.MsgDeleteTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	_, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenom, msg.HostDenom)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteNotFound,
			"no trade route for rewardDenom %s and hostDenom %s", msg.RewardDenom, msg.HostDenom)
	}

	ms.Keeper.RemoveTradeRoute(ctx, msg.RewardDenom, msg.HostDenom)

	return &types.MsgDeleteTradeRouteResponse{}, nil
}

// Gov tx to update the trade config of a trade route
//
// Example proposal:
//
//		{
//		   "title": "Update a the trade config for host chain X",
//		   "metadata": "Update a the trade config for host chain X",
//		   "summary": "Update a the trade config for host chain X",
//		   "messages":[
//		      {
//		         "@type": "/stride.stakeibc.MsgUpdateTradeRoute",
//		         "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
//
//				 "pool_id": 1,
//				 "max_allowed_swap_loss_rate": "0.05",
//				 "min_swap_amount": "10000000",
//				 "max_swap_amount": "1000000000"
//			  }
//		   ],
//		   "deposit": "2000000000ustrd"
//	   }
//
// >>> strided tx gov submit-proposal {proposal_file.json} --from wallet
func (ms msgServer) UpdateTradeRoute(goCtx context.Context, msg *types.MsgUpdateTradeRoute) (*types.MsgUpdateTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	route, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenom, msg.HostDenom)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteNotFound,
			"no trade route for rewardDenom %s and hostDenom %s", msg.RewardDenom, msg.HostDenom)
	}

	maxAllowedSwapLossRate := msg.MaxAllowedSwapLossRate
	if maxAllowedSwapLossRate == "" {
		maxAllowedSwapLossRate = DefaultMaxAllowedSwapLossRate
	}
	maxSwapAmount := msg.MaxSwapAmount
	if maxSwapAmount.IsZero() {
		maxSwapAmount = DefaultMaxSwapAmount
	}

	updatedConfig := types.TradeConfig{
		PoolId: msg.PoolId,

		SwapPrice:            sdk.ZeroDec(),
		PriceUpdateTimestamp: 0,

		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr(maxAllowedSwapLossRate),
		MinSwapAmount:          msg.MinSwapAmount,
		MaxSwapAmount:          maxSwapAmount,
	}

	route.TradeConfig = updatedConfig
	ms.Keeper.SetTradeRoute(ctx, route)

	return &types.MsgUpdateTradeRouteResponse{}, nil
}

func (k msgServer) RestoreInterchainAccount(goCtx context.Context, msg *types.MsgRestoreInterchainAccount) (*types.MsgRestoreInterchainAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get ConnectionEnd (for counterparty connection)
	connectionEnd, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, msg.ConnectionId)
	if !found {
		return nil, errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection %s not found", msg.ConnectionId)
	}
	counterpartyConnection := connectionEnd.Counterparty

	// only allow restoring an account if it already exists
	portID, err := icatypes.NewControllerPortID(msg.AccountOwner)
	if err != nil {
		return nil, err
	}
	_, exists := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, msg.ConnectionId, portID)
	if !exists {
		return nil, errorsmod.Wrapf(types.ErrInvalidInterchainAccountAddress,
			"ICA controller account address not found: %s", msg.AccountOwner)
	}

	appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: msg.ConnectionId,
		HostConnectionId:       counterpartyConnection.ConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))

	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.AccountOwner, appVersion); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register account for owner %s", msg.AccountOwner)
	}

	// If we're restoring a delegation account, we also have to reset record state
	if msg.AccountOwner == types.FormatHostZoneICAOwner(msg.ChainId, types.ICAAccountType_DELEGATION) {
		hostZone, found := k.GetHostZone(ctx, msg.ChainId)
		if !found {
			return nil, types.ErrHostZoneNotFound.Wrapf("delegation ICA supplied, but no associated host zone")
		}

		// Since any ICAs along the original channel will never get relayed,
		// we have to reset the delegation_changes_in_progress field on each validator
		for _, validator := range hostZone.Validators {
			validator.DelegationChangesInProgress = 0
		}
		k.SetHostZone(ctx, hostZone)

		// revert DELEGATION_IN_PROGRESS records for the closed ICA channel (so that they can be staked)
		depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
		for _, depositRecord := range depositRecords {
			// only revert records for the select host zone
			if depositRecord.HostZoneId == hostZone.ChainId && depositRecord.Status == recordtypes.DepositRecord_DELEGATION_IN_PROGRESS {
				depositRecord.Status = recordtypes.DepositRecord_DELEGATION_QUEUE
				k.Logger(ctx).Info(fmt.Sprintf("Setting DepositRecord %d to status DepositRecord_DELEGATION_IN_PROGRESS", depositRecord.Id))
				k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
			}
		}

		// revert epoch unbonding records for the closed ICA channel
		epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)
		epochNumberForPendingUnbondingRecords := []uint64{}
		epochNumberForPendingTransferRecords := []uint64{}
		for _, epochUnbondingRecord := range epochUnbondingRecords {
			// only revert records for the select host zone
			hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
			if !found {
				k.Logger(ctx).Info(fmt.Sprintf("No HostZoneUnbonding found for chainId: %s, epoch: %d", hostZone.ChainId, epochUnbondingRecord.EpochNumber))
				continue
			}

			// Revert UNBONDING_IN_PROGRESS and EXIT_TRANSFER_IN_PROGRESS records
			if hostZoneUnbonding.Status == recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS {
				k.Logger(ctx).Info(fmt.Sprintf("HostZoneUnbonding for %s at EpochNumber %d is stuck in status %s",
					hostZone.ChainId, epochUnbondingRecord.EpochNumber, recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS.String(),
				))
				epochNumberForPendingUnbondingRecords = append(epochNumberForPendingUnbondingRecords, epochUnbondingRecord.EpochNumber)

			} else if hostZoneUnbonding.Status == recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS {
				k.Logger(ctx).Info(fmt.Sprintf("HostZoneUnbonding for %s at EpochNumber %d to in status %s",
					hostZone.ChainId, epochUnbondingRecord.EpochNumber, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS.String(),
				))
				epochNumberForPendingTransferRecords = append(epochNumberForPendingTransferRecords, epochUnbondingRecord.EpochNumber)
			}
		}
		// Revert UNBONDING_IN_PROGRESS records to UNBONDING_QUEUE
		err := k.RecordsKeeper.SetHostZoneUnbondingStatus(ctx, hostZone.ChainId, epochNumberForPendingUnbondingRecords, recordtypes.HostZoneUnbonding_UNBONDING_QUEUE)
		if err != nil {
			errMsg := fmt.Sprintf("unable to update host zone unbonding record status to %s for chainId: %s and epochUnbondingRecordIds: %v, err: %s",
				recordtypes.HostZoneUnbonding_UNBONDING_QUEUE.String(), hostZone.ChainId, epochNumberForPendingUnbondingRecords, err)
			k.Logger(ctx).Error(errMsg)
			return nil, err
		}

		// Revert EXIT_TRANSFER_IN_PROGRESS records to EXIT_TRANSFER_QUEUE
		err = k.RecordsKeeper.SetHostZoneUnbondingStatus(ctx, hostZone.ChainId, epochNumberForPendingTransferRecords, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
		if err != nil {
			errMsg := fmt.Sprintf("unable to update host zone unbonding record status to %s for chainId: %s and epochUnbondingRecordIds: %v, err: %s",
				recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE.String(), hostZone.ChainId, epochNumberForPendingTransferRecords, err)
			k.Logger(ctx).Error(errMsg)
			return nil, err
		}

		// Revert all pending LSM Detokenizations from status DETOKENIZATION_IN_PROGRESS to status DETOKENIZATION_QUEUE
		pendingDeposits := k.RecordsKeeper.GetLSMDepositsForHostZoneWithStatus(ctx, hostZone.ChainId, recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS)
		for _, lsmDeposit := range pendingDeposits {
			k.Logger(ctx).Info(fmt.Sprintf("Setting LSMTokenDeposit %s to status DETOKENIZATION_QUEUE", lsmDeposit.Denom))
			k.RecordsKeeper.UpdateLSMTokenDepositStatus(ctx, lsmDeposit, recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE)
		}
	}

	return &types.MsgRestoreInterchainAccountResponse{}, nil
}

// This kicks off two ICQs, each with a callback, that will update the number of tokens on a validator
// after being slashed. The flow is:
// 1. QueryValidatorSharesToTokensRate (ICQ)
// 2. ValidatorSharesToTokensRate (CALLBACK)
// 3. SubmitDelegationICQ (ICQ)
// 4. DelegatorSharesCallback (CALLBACK)
func (k msgServer) UpdateValidatorSharesExchRate(goCtx context.Context, msg *types.MsgUpdateValidatorSharesExchRate) (*types.MsgUpdateValidatorSharesExchRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.QueryValidatorSharesToTokensRate(ctx, msg.ChainId, msg.Valoper); err != nil {
		return nil, err
	}
	return &types.MsgUpdateValidatorSharesExchRateResponse{}, nil
}

// Submits an ICQ to get the validator's delegated shares
func (k msgServer) CalibrateDelegation(goCtx context.Context, msg *types.MsgCalibrateDelegation) (*types.MsgCalibrateDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound
	}

	if err := k.SubmitCalibrationICQ(ctx, hostZone, msg.Valoper); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for delegation, error : %s", err.Error()))
		return nil, err
	}

	return &types.MsgCalibrateDelegationResponse{}, nil
}

func (k msgServer) UpdateInnerRedemptionRateBounds(goCtx context.Context, msg *types.MsgUpdateInnerRedemptionRateBounds) (*types.MsgUpdateInnerRedemptionRateBoundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Note: we're intentionally not checking the zone is halted
	zone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found: %s", msg.ChainId))
		return nil, types.ErrInvalidHostZone
	}

	// Get the wide bounds
	outerMinSafetyThreshold, outerMaxSafetyThreshold := k.GetOuterSafetyBounds(ctx, zone)

	innerMinSafetyThreshold := msg.MinInnerRedemptionRate
	innerMaxSafetyThreshold := msg.MaxInnerRedemptionRate

	// Confirm the inner bounds are within the outer bounds
	if innerMinSafetyThreshold.LT(outerMinSafetyThreshold) {
		errMsg := fmt.Sprintf("inner min safety threshold (%s) is less than outer min safety threshold (%s)", innerMinSafetyThreshold, outerMinSafetyThreshold)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrInvalidBounds, errMsg)
	}

	if innerMaxSafetyThreshold.GT(outerMaxSafetyThreshold) {
		errMsg := fmt.Sprintf("inner max safety threshold (%s) is greater than outer max safety threshold (%s)", innerMaxSafetyThreshold, outerMaxSafetyThreshold)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrInvalidBounds, errMsg)
	}

	// Set the inner bounds on the host zone
	zone.MinInnerRedemptionRate = innerMinSafetyThreshold
	zone.MaxInnerRedemptionRate = innerMaxSafetyThreshold

	k.SetHostZone(ctx, zone)

	return &types.MsgUpdateInnerRedemptionRateBoundsResponse{}, nil
}

func (k msgServer) ResumeHostZone(goCtx context.Context, msg *types.MsgResumeHostZone) (*types.MsgResumeHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get Host Zone
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		errMsg := fmt.Sprintf("invalid chain id, zone for %s not found", msg.ChainId)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}

	// Check the zone is halted
	if !hostZone.Halted {
		errMsg := fmt.Sprintf("invalid chain id, zone for %s not halted", msg.ChainId)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrHostZoneNotHalted, errMsg)
	}

	// remove from blacklist
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	k.RatelimitKeeper.RemoveDenomFromBlacklist(ctx, stDenom)

	// Resume zone
	hostZone.Halted = false
	k.SetHostZone(ctx, hostZone)

	return &types.MsgResumeHostZoneResponse{}, nil
}
