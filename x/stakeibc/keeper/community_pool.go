package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v19/utils"
	epochstypes "github.com/Stride-Labs/stride/v19/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v19/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

// For each hostZone with a valid community pool, trigger the ICQs and ICAs to transfer tokens from DepositICA or back to ReturnICA
// Since ICQs and ICAs take time to complete, it is almost certain tokens swept in and processed will be swept out in a later epoch
func (k Keeper) ProcessAllCommunityPoolTokens(ctx sdk.Context) {
	hostZones := k.GetAllActiveHostZone(ctx)
	for _, hostZone := range hostZones {
		if hostZone.CommunityPoolDepositIcaAddress == "" ||
			hostZone.CommunityPoolStakeHoldingAddress == "" ||
			hostZone.CommunityPoolRedeemHoldingAddress == "" ||
			hostZone.CommunityPoolReturnIcaAddress == "" {
			continue
		}

		// stDenom is the ibc denom on hostZone when the hostZone's native denom is staked
		denom := hostZone.HostDenom
		stIbcDenom, err := k.GetStIbcDenomOnHostZone(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId, "Unable to get stToken ibc denom - %s", err.Error()))
			continue
		}

		/****** Stage 1: Query deposit ICA for denom/stDenom, Transfer tokens to stride *******/

		// ICQ for the host denom of the chain, these are tokens the pool wants staked
		err = k.QueryCommunityPoolIcaBalance(ctx, hostZone, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, denom)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId,
				"Failed to submit ICQ for native denom %s in deposit ICA - %s", denom, err.Error()))
		}
		// ICQ for staked tokens of the host denom, these are tokens the pool wants redeemed
		err = k.QueryCommunityPoolIcaBalance(ctx, hostZone, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, stIbcDenom)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId,
				"Failed to submit ICQ for stHostDenom %s in deposit ICA - %s", stIbcDenom, err.Error()))
		}

		/****** Stage 2: LiquidStake denom and transfer to return ICA, or RedeemStake stDenom *******/

		// LiquidStake tokens in the stake holding address and transfer to the return ica
		if err = k.LiquidStakeCommunityPoolTokens(ctx, hostZone); err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId,
				"Failed to liquid staking and transfer community pool tokens in stake holding address - %s", err.Error()))
		}
		// RedeemStake tokens in the redeem holding address, in 30 days they claim to the return ica
		if err = k.RedeemCommunityPoolTokens(ctx, hostZone); err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId,
				"Failed to redeeming stTokens in redeem holding address - %s", err.Error()))
		}

		/****** Stage 3: Query return ICA for denom/stDenom, FundCommunityPool from return ICA *******/

		err = k.QueryCommunityPoolIcaBalance(ctx, hostZone, types.ICAAccountType_COMMUNITY_POOL_RETURN, denom)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId,
				"Failed to submit ICQ for native denom %s in return ICA - %s", denom, err.Error()))
		}
		err = k.QueryCommunityPoolIcaBalance(ctx, hostZone, types.ICAAccountType_COMMUNITY_POOL_RETURN, stIbcDenom)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId,
				"Failed to submit ICQ for stHostDenom %s in return ICA - %s", stIbcDenom, err.Error()))
		}
	}
}

// ICQ specific denom for balance in the deposit ICA or return ICA on the community pool host zone
// Depending on account type and denom, discovered tokens are transferred to Stride or funded to the pool
func (k Keeper) QueryCommunityPoolIcaBalance(
	ctx sdk.Context,
	hostZone types.HostZone,
	icaType types.ICAAccountType,
	denom string,
) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Building ICQ for %s balance in community pool %s address", denom, icaType.String()))

	var icaAddress string
	switch icaType {
	case types.ICAAccountType_COMMUNITY_POOL_DEPOSIT:
		icaAddress = hostZone.CommunityPoolDepositIcaAddress
	case types.ICAAccountType_COMMUNITY_POOL_RETURN:
		icaAddress = hostZone.CommunityPoolReturnIcaAddress
	default:
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "icaType must be either deposit or return!")
	}

	// Verify a valid ica address exists for this host zone
	if icaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no address of type %s found for %s",
			icaType.String(), hostZone.ChainId)
	}

	_, addressBz, err := bech32.DecodeAndConvert(icaAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid %s address, could not decode (%s)",
			icaType.String(), icaAddress)
	}
	queryData := append(banktypes.CreateAccountBalancesPrefix(addressBz), []byte(denom)...)

	// The response might be a coin, or might just be an int depending on sdk version
	// Since we need the denom later, store the denom as callback data for the query
	callbackData := types.CommunityPoolBalanceQueryCallback{
		IcaType: icaType,
		Denom:   denom,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "can't marshal community pool balance callback data %+v", callbackData)
	}

	// Timeout query at end of epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeout := time.Unix(0, int64(strideEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_CommunityPoolIcaBalance,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return errorsmod.Wrapf(err, "Error submitting query for pool ica balance")
	}

	return nil
}

// Liquid stake all native tokens in the stake holding address
func (k Keeper) LiquidStakeCommunityPoolTokens(ctx sdk.Context, hostZone types.HostZone) error {
	// Get the number of native tokens in the stake address
	// The native tokens will be an ibc denom since they've been transferred to stride
	communityPoolStakeAddress, err := sdk.AccAddressFromBech32(hostZone.CommunityPoolStakeHoldingAddress)
	if err != nil {
		return err
	}
	nativeTokens := k.bankKeeper.GetBalance(ctx, communityPoolStakeAddress, hostZone.IbcDenom)

	// If there aren't enough tokens, do nothing
	// (consider specifying a minimum here)
	if nativeTokens.Amount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "No community pool tokens to liquid stake"))
		return nil
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Liquid staking community pool tokens: %+v", nativeTokens))

	// TODO: Move LS function to keeper method instead of message server
	// Liquid stake the balance in the stake holding account
	msgServer := NewMsgServerImpl(k)
	liquidStakeRequest := types.MsgLiquidStake{
		Creator:   hostZone.CommunityPoolStakeHoldingAddress,
		Amount:    nativeTokens.Amount,
		HostDenom: hostZone.HostDenom,
	}
	resp, err := msgServer.LiquidStake(ctx, &liquidStakeRequest)
	if err != nil {
		return types.ErrFailedToLiquidStake.Wrapf(err.Error())
	}

	// If the liquid stake was successful, transfer the stTokens to the return ICA
	return k.TransferHoldingToCommunityPoolReturn(ctx, hostZone, resp.StToken)
}

// Redeem all the stTokens in the redeem holding address
func (k Keeper) RedeemCommunityPoolTokens(ctx sdk.Context, hostZone types.HostZone) error {
	// Get the number of stTokens in the redeem address
	communityPoolRedeemAddress, err := sdk.AccAddressFromBech32(hostZone.CommunityPoolRedeemHoldingAddress)
	if err != nil {
		return err
	}
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stTokens := k.bankKeeper.GetBalance(ctx, communityPoolRedeemAddress, stDenom)

	// If there aren't enough tokens, do nothing
	// (consider a greater than zero minimum threshold to avoid extra transfers)
	if stTokens.Amount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "No community pool tokens to redeem"))
		return nil
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Redeeming community pool tokens: %+v", stTokens))

	// TODO: Move Redeem function to keeper method instead of message server
	// Redeem the stTokens in the redeem holding account
	// The return ICA address will be the recipient of the claim
	msgServer := NewMsgServerImpl(k)
	redeemStakeRequest := types.MsgRedeemStake{
		Creator:  hostZone.CommunityPoolRedeemHoldingAddress,
		Amount:   stTokens.Amount,
		HostZone: hostZone.ChainId,
		Receiver: hostZone.CommunityPoolReturnIcaAddress,
	}
	if _, err := msgServer.RedeemStake(ctx, &redeemStakeRequest); err != nil {
		return types.ErrUnableToRedeemStake.Wrapf(err.Error())
	}

	return nil
}

// Builds a msg to send funds to a community pool
// If the community pool treasury address is specified on the host zone, the tokens are bank sent there
// Otherwise, a MsgFundCommunityPool is used to send tokens to the default community pool address
func (k Keeper) BuildFundCommunityPoolMsg(
	ctx sdk.Context,
	hostZone types.HostZone,
	tokens sdk.Coins,
	senderAccountType types.ICAAccountType,
) (fundMsg []proto.Message, err error) {
	// Get the sender ICA address based on the account type
	var sender string
	switch senderAccountType {
	case types.ICAAccountType_COMMUNITY_POOL_RETURN:
		sender = hostZone.CommunityPoolReturnIcaAddress
	case types.ICAAccountType_WITHDRAWAL:
		sender = hostZone.WithdrawalIcaAddress
	default:
		return nil, errorsmod.Wrapf(types.ErrICATxFailed,
			"fund community pool ICA can only be initiated from either the community pool return or withdrawal ICA account")
	}

	// If the community pool treasury address is specified, bank send there
	if hostZone.CommunityPoolTreasuryAddress != "" {
		fundMsg = []proto.Message{&banktypes.MsgSend{
			FromAddress: sender,
			ToAddress:   hostZone.CommunityPoolTreasuryAddress,
			Amount:      tokens,
		}}
	} else {
		// Otherwise, call MsgFundCommunityPool
		fundMsg = []proto.Message{&disttypes.MsgFundCommunityPool{
			Amount:    tokens,
			Depositor: sender,
		}}
	}

	return fundMsg, nil
}

// Using tokens in the CommunityPoolReturnIcaAddress, trigger ICA tx to fund community pool
// Note: The denom of the passed in token has to be the denom which exists on the hostZone not Stride
func (k Keeper) FundCommunityPool(
	ctx sdk.Context,
	hostZone types.HostZone,
	token sdk.Coin,
	senderAccountType types.ICAAccountType,
) error {
	msgs, err := k.BuildFundCommunityPoolMsg(ctx, hostZone, sdk.NewCoins(token), senderAccountType)
	if err != nil {
		return err
	}

	// Timeout the ICA at the end of the epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeoutTimestamp := uint64(strideEpochTracker.NextEpochStartTime)

	// No need to build ICA callback data or input an ICA callback method
	icaCallbackId := ""
	var icaCallbackData []byte

	// Send the transaction through SubmitTx to kick off ICA command
	_, err = k.SubmitTxs(ctx,
		hostZone.ConnectionId,
		msgs,
		senderAccountType,
		timeoutTimestamp,
		icaCallbackId,
		icaCallbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to SubmitTxs for FundCommunityPool, Messages: %+v", msgs)
	}

	return nil
}
