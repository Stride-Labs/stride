package keeper

import (
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	pooltypes "cosmossdk.io/x/protocolpool/types" // seems to need sdk v51+ we are on v47

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v14/utils"
	epochstypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// For each hostZone with a valid community pool, trigger the ICQs and ICAs to transfer deposited tokens to holding address
// Since ICQs and ICAs take time to complete, it is almost certain tokens swept in and processed will be swept out in a later epoch
func (k Keeper) SweepInAllDepositedCommunityPoolTokens(ctx sdk.Context) error {
	hostZones := k.GetAllActiveHostZone(ctx)
	for _, hostZone := range hostZones {
		if hostZone.CommunityPoolDepositIcaAddress != "" &&
			hostZone.CommunityPoolHoldingAddress != "" &&
			hostZone.CommunityPoolReturnIcaAddress != "" {
				// ICQ for the host denom of the chain, these are tokens the pool wants staked
				k.ICQCommunityPoolDepositICABalance(ctx, hostZone, hostZone.HostDenom)
				// ICQ for the stToken of the host denom, these are tokens the pool wants redeemed
				//   if stDenom is the denom on stride, ibcStDenom is the ibc denom on hostZone for stDenom
				ibcStDenom := k.GetStakedHostTokenIbcDenomOnHostZone(hostZone)
				k.ICQCommunityPoolDepositICABalance(ctx, hostZone, ibcStDenom)
			}
	}
	return nil
}

// For each hostZone with a valid community pool, trigger the transfers of tokens in the holding address to the return ICA
func (k Keeper) SweepOutAllReturningCommunityPoolTokens(ctx sdk.Context) error {
	hostZones := k.GetAllActiveHostZone(ctx)
	for _, hostZone := range hostZones {
		if hostZone.CommunityPoolDepositIcaAddress != "" &&
			hostZone.CommunityPoolHoldingAddress != "" &&
			hostZone.CommunityPoolReturnIcaAddress != "" {
				k.IBCReturnAllCommunityPoolTokens(ctx, hostZone)
			}
	}	
	return nil
}



// ICQ specific denom for balance in the deposit ICA on the community pool host zone
// The ICQ callback will call IBCTransferCommunityPoolICATokensToStride with found token(s) as input
func (k Keeper) ICQCommunityPoolDepositICABalance(ctx sdk.Context, hostZone types.HostZone, denom string) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for %s in community pool deposit account balance", denom))

	// Get the withdrawal account address from the host zone
	if hostZone.CommunityPoolDepositIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no deposit account found for %s", hostZone.ChainId)
	}

	// Encode the deposit account address for the query request
	// The query request consists of the account address and denom passed in
	_, depositAddressBz, err := bech32.DecodeAndConvert(hostZone.CommunityPoolDepositIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid withdrawal account address, could not decode (%s)", err.Error())
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(depositAddressBz), []byte(denom)...)

	// The response might be a coin, or might just be an int depending on sdk version
	// Since we need the denom later, store the denom as callback data for the query
	callbackData := types.CommunityPoolDepositQueryCallback{
		Denom: denom,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal community pool deposit balance callback data")
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
		CallbackId:      ICQCallbackID_CommunityPoolDepositBalance,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying for pool deposit balance, error: %s", err.Error()))
		return err
	}

	return nil
}


// ibc transfers tokens from the foreign hub community pool deposit ICA address onto Stride hub
// then as an atomic action, liquid stake the tokens with Autopilot commands in the ibc message memos
func (k Keeper) IBCTransferCommunityPoolICATokensToStride(ctx sdk.Context, communityPoolHostZoneId string, token sdk.Coin) error {
	k.Logger(ctx).Info(fmt.Sprintf("Transfering %d %s tokens from community pool deposit ICA to Stride hub holding address", token.Amount.Int64(), token.Denom))

	// TODO: add safety check here on if the amount is greater than a threshold to avoid many, frequent, small transfers
	//       threshold might be a config we tune to avoid ddos type attacks, for now using 0 as hard coded threshold
	if token.Amount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] The amount %v to transfer did not meet the minimum threshold!", token.Amount))
		return errors.New("Transfer Amount below threshold!")
	}

	hostZone, hostZoneFound := k.GetHostZone(ctx, communityPoolHostZoneId)
	if !hostZoneFound || hostZone.Halted {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Host zone not found or halted!"))
		return errors.New("No active host zone found!")
	}

	if hostZone.CommunityPoolDepositIcaAddress == "" || hostZone.CommunityPoolHoldingAddress == "" {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Unknown send or recieve address! DepositICAAddress: %s HoldingAddress: %s", 
			hostZone.CommunityPoolDepositIcaAddress, 
			hostZone.CommunityPoolHoldingAddress))
		return errors.New("Critical addresses missing from hostZone config!")
	}

	// If the denom is the native denom of the community pool chain, assume the pool wants to stake these tokens
	// if the denom is the ibc dennom representing stNativeDenom, assume the pool wants to redeem these staked tokens
	// if the denom is something else, transfer the tokens over but do no staking or redeeming autopilot action
	//   since all tokens in the module account will eventually be sent back to the pool, the tx route can have any denom
	autoPilotAction := NoAction
	if token.Denom == hostZone.HostDenom {
		autoPilotAction = LiquidStake
	}
	ibcStDenom := k.GetStakedHostTokenIbcDenomOnHostZone(hostZone)
	if token.Denom == ibcStDenom {
		autoPilotAction = RedeemStake
	}

	// ibc transfer tokens from foreign hub deposit ICA address to Stride hub "holding" address
	err := k.IBCTransferCommunityPoolTokens(ctx, token, hostZone, autoPilotAction)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Failed to submit transfer to host zone, HostZone: %v, Channel: %v, Coin: %v, SendAddress: %v, RecAddress: %v",
			hostZone.ChainId, hostZone.TransferChannelId, token, hostZone.CommunityPoolDepositIcaAddress, hostZone.CommunityPoolHoldingAddress))
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] err {%s}", err.Error()))
		return err
	}

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "[IBCTransferCommunityPoolICATokensToStride] Transfer community pool tokens to Stride successfully initiated!"))

	return nil
}


// Using tokens in the CommunityPoolReturnIcaAddress, ICA tx to fund community pool
// Note: The denoms of the passed in token have to be ibc denoms which exist on the communityPoolHostZone
func (k Keeper) FundCommunityPool(ctx sdk.Context, communityPoolHostZone types.HostZone, token sdk.Coin) error {
	fundCoins := sdk.NewCoins(token)
	
	var msgs []proto.Message
	msgs = append(msgs, pooltypes.NewMsgFundCommunityPool(
		fundCoins,
		communityPoolHostZone.CommunityPoolReturnIcaAddress,
	))

	// No need to build ICA callback data or input an ICA callback method 
	icaCallbackId := ""
	var icaCallbackData []byte;
	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)
	icaTimeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos

	// Send the transaction through SubmitTx to kick off ICA command -- no ICA callback method name, or callback args needed
	_, err := k.SubmitTxs(ctx, 
		communityPoolHostZone.ConnectionId, 
		msgs, 
		types.ICAAccountType_COMMUNITY_POOL_RETURN, 
		icaTimeoutTimestamp, 
		icaCallbackId, 
		icaCallbackData)
	if err != nil {
		return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs for FundCommunityPool, Messages: %v, err: %s", msgs, err.Error())
	}

	return nil
}


// given a hostZone with native denom, returns the ibc denom on the zone for the staked stDenom 
func (k Keeper) GetStakedHostTokenIbcDenomOnHostZone(hostZone types.HostZone) (ibcStakedDenom string) {
	nativeDenom := hostZone.HostDenom
	stDenomOnStride := types.StAssetDenomFromHostZoneDenom(nativeDenom)
	// we use the hostZone.TransferChannelId because all stDenom originate on stride before being sent out
	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, hostZone.TransferChannelId)
	prefixedDenom := sourcePrefix + stDenomOnStride

	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}
