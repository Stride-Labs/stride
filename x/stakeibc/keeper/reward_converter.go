package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v26/utils"
	epochstypes "github.com/Stride-Labs/stride/v26/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v26/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

const (
	OsmosisSwapTypeUrl       = "/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn"
	LegacyOsmosisSwapTypeUrl = "/osmosis.gamm.v1beta1.MsgSwapExactAmountIn"
)

// JSON Memo for PFM transfers
type PacketForwardMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}
type ForwardMetadata struct {
	Receiver string `json:"receiver"`
	Port     string `json:"port"`
	Channel  string `json:"channel"`
	Timeout  string `json:"timeout"`
	Retries  int64  `json:"retries"`
}

type RewardsSplit struct {
	RebateAmount    sdkmath.Int
	StrideFeeAmount sdkmath.Int
	ReinvestAmount  sdkmath.Int
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// The goal of this code is to allow certain reward token types to be automatically traded into other types
// This happens before the rest of the staking, allocation, distribution etc. would continue as normal
//
// Reward tokens are any special denoms which are paid out in the withdrawal address
// Most host zones inflate their tokens and their native token is what appears in the withdrawal ICA
// The following allows for chains to use foreign denoms as revenue, which can be traded to any other denom first
//
//  1. Epochly check the reward denom balance in the withdrawal address
//     on callback, send all this reward denom from withdrawl ICA to trade ICA on the trade zone (OSMOSIS)
//  2. Off-chain swaps of reward denom to host denom
//  3. Epochly check the host denom balance in trade ICA
//     on callback, transfer these host denom tokens from trade ICA to withdrawal ICA on original host zone
//
// Normal staking flow continues from there. So the host denom tokens will land on the original host zone
// and the normal staking and distribution flow will continue from there.
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Breaks down the split of native rewards into the portions intended for (a) a rebate, (b) stride commission,
// and (c) reinvestment
// For most host zones, the rewards here were generated from normal staking rewards, but in the case of dYdX,
// this is called on the rewards that were converted from USDC to DYDX during the trade route
//
// The rebate percentage is determined by: (% of total TVL contributed by commuity pool) * (rebate percentage)
//
// E.g. Community pool liquid staked 1M, TVL is 10M, rebate is 20%
// Total rewards this epoch are 1000, and the stride fee is 10%
// => Then the rebate is 1000 rewards * 10% stride fee * (1M / 10M) * 20% rebate = 2 tokens
// => Stride fee is 1000 rewards * 10% stride fee - 2 rebate = 98 tokens
// => Reinvestment is 1000 rewards * (100% - 10% stride fee) = 900 tokens
func (k Keeper) CalculateRewardsSplit(
	ctx sdk.Context,
	hostZone types.HostZone,
	rewardsAmount sdkmath.Int,
) (rewardSplit RewardsSplit, err error) {
	// Get the fee rate and total fees from params (e.g. 0.1 for 10% fee)
	strideFeeParam := sdk.NewIntFromUint64(k.GetParams(ctx).StrideCommission)
	totalFeeRate := sdk.NewDecFromInt(strideFeeParam).Quo(sdk.NewDec(100))

	// Get the total fee amount from the fee percentage
	totalFeesAmount := sdk.NewDecFromInt(rewardsAmount).Mul(totalFeeRate).TruncateInt()
	reinvestAmount := rewardsAmount.Sub(totalFeesAmount)

	// Check if the chain has a rebate
	// If there's no rebate, return 0 rebate and send all fees as stride commission
	rebateInfo, chainHasRebate := hostZone.SafelyGetCommunityPoolRebate()
	if !chainHasRebate {
		rewardSplit = RewardsSplit{
			RebateAmount:    sdkmath.ZeroInt(),
			StrideFeeAmount: totalFeesAmount,
			ReinvestAmount:  reinvestAmount,
		}
		return rewardSplit, nil
	}

	// Get supply of stTokens to determine the portion of TVL that the community pool liquid stake makes up
	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stTokenSupply := k.bankKeeper.GetSupply(ctx, stDenom).Amount

	// It shouldn't be possible to have 0 token supply (since there are rewards and there was a community pool stake)
	// This will also prevent a division by 0 error
	if stTokenSupply.IsZero() {
		return rewardSplit, errorsmod.Wrapf(types.ErrDivisionByZero,
			"unable to calculate rebate amount for %s since total delegations are 0", hostZone.ChainId)
	}

	// It also shouldn't be possible for the liquid stake amount to be greater than the full TVL
	if rebateInfo.LiquidStakedStTokenAmount.GT(stTokenSupply) {
		return rewardSplit, errorsmod.Wrapf(types.ErrFeeSplitInvariantFailed,
			"community pool liquid staked amount greater than total delegations")
	}

	// The rebate amount is determined by the contribution of the community pool stake towards the total TVL,
	// multiplied by the rebate fee percentage
	contributionRate := sdk.NewDecFromInt(rebateInfo.LiquidStakedStTokenAmount).Quo(sdk.NewDecFromInt(stTokenSupply))
	rebateAmount := sdk.NewDecFromInt(totalFeesAmount).Mul(contributionRate).Mul(rebateInfo.RebateRate).TruncateInt()
	strideFeeAmount := totalFeesAmount.Sub(rebateAmount)

	rewardSplit = RewardsSplit{
		RebateAmount:    rebateAmount,
		StrideFeeAmount: strideFeeAmount,
		ReinvestAmount:  reinvestAmount,
	}

	return rewardSplit, nil
}

// Builds an authz MsgGrant or MsgRevoke to grant an account trade capabilties on behalf of the trade ICA
func (k Keeper) BuildTradeAuthzMsg(
	ctx sdk.Context,
	tradeRoute types.TradeRoute,
	permissionChange types.AuthzPermissionChange,
	grantee string,
	legacy bool,
) (authzMsg []proto.Message, err error) {
	messageTypeUrl := OsmosisSwapTypeUrl
	if legacy {
		messageTypeUrl = LegacyOsmosisSwapTypeUrl
	}

	switch permissionChange {
	case types.AuthzPermissionChange_GRANT:
		authorization := authz.NewGenericAuthorization(messageTypeUrl)
		expiration := ctx.BlockTime().Add(time.Hour * 24 * 365 * 100) // 100 years

		grant, err := authz.NewGrant(ctx.BlockTime(), authorization, &expiration)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "unable to build grant struct")
		}
		authzMsg = []proto.Message{&authz.MsgGrant{
			Granter: tradeRoute.TradeAccount.Address,
			Grantee: grantee,
			Grant:   grant,
		}}

	case types.AuthzPermissionChange_REVOKE:
		authzMsg = []proto.Message{&authz.MsgRevoke{
			Granter:    tradeRoute.TradeAccount.Address,
			Grantee:    grantee,
			MsgTypeUrl: messageTypeUrl,
		}}

	default:
		return nil, errors.New("invalid permission change")
	}

	return authzMsg, nil
}

// Builds a PFM transfer message to send reward tokens from the host zone,
// through the reward zone (to unwind) and finally to the trade zone
func (k Keeper) BuildHostToTradeTransferMsg(
	ctx sdk.Context,
	amount sdkmath.Int,
	route types.TradeRoute,
) (msg transfertypes.MsgTransfer, err error) {
	// Get the epoch tracker to determine the timeouts
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return msg, errorsmod.Wrap(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}

	// Timeout the first transfer halfway through the epoch, and the second transfer at the end of the epoch
	// The pfm transfer requires a duration instead of a timestamp for the timeout, so we just use half the epoch length
	halfEpochDuration := strideEpochTracker.Duration / 2
	transfer1TimeoutTimestamp := uint64(strideEpochTracker.NextEpochStartTime - halfEpochDuration) // unix nano
	transfer2TimeoutDuration := fmt.Sprintf("%ds", halfEpochDuration/1e9)                          // string in seconds

	startingDenom := route.RewardDenomOnHostZone
	sendTokens := sdk.NewCoin(startingDenom, amount)

	withdrawlIcaAddress := route.HostAccount.Address
	unwindIcaAddress := route.RewardAccount.Address
	tradeIcaAddress := route.TradeAccount.Address

	// Validate ICAs were registered
	if withdrawlIcaAddress == "" {
		return msg, errorsmod.Wrapf(types.ErrICAAccountNotFound, "no host account found for %s", route.Description())
	}
	if unwindIcaAddress == "" {
		return msg, errorsmod.Wrapf(types.ErrICAAccountNotFound, "no reward account found for %s", route.Description())
	}
	if tradeIcaAddress == "" {
		return msg, errorsmod.Wrapf(types.ErrICAAccountNotFound, "no trade account found for %s", route.Description())
	}

	// Build the pfm memo to specify the forwarding logic
	// This transfer channel id is a channel on the reward Zone for transfers to the trade zone
	// (not to be confused with a transfer channel on Stride or the Host Zone)
	memo := PacketForwardMetadata{
		Forward: &ForwardMetadata{
			Receiver: tradeIcaAddress,
			Port:     transfertypes.PortID,
			Channel:  route.RewardToTradeChannelId,
			Timeout:  transfer2TimeoutDuration,
			Retries:  0,
		},
	}
	memoJSON, err := json.Marshal(memo)
	if err != nil {
		return msg, err
	}

	msg = transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    route.HostToRewardChannelId, // channel on hostZone for transfers to rewardZone
		Token:            sendTokens,
		Sender:           withdrawlIcaAddress,
		Receiver:         unwindIcaAddress, // could be "pfm" or a real address depending on version
		TimeoutTimestamp: transfer1TimeoutTimestamp,
		Memo:             string(memoJSON),
	}

	return msg, nil
}

// ICA tx will kick off transfering the reward tokens from the hostZone withdrawl ICA to the tradeZone trade ICA
// This will be two hops to unwind the ibc denom through the rewardZone using pfm in the transfer memo
func (k Keeper) TransferRewardTokensHostToTrade(ctx sdk.Context, amount sdkmath.Int, route types.TradeRoute) error {
	// Confirm the reward amount exceeds the transfer threshold, otherwise exit prematurely
	if route.MinTransferAmount.GT(amount) {
		k.Logger(ctx).Info(fmt.Sprintf("Balance of %v is below transfer minimum of %v, skipping transfer",
			amount, route.MinTransferAmount))
		return nil
	}

	// Build the PFM transfer message from host to trade zone
	msg, err := k.BuildHostToTradeTransferMsg(ctx, amount, route)
	if err != nil {
		return err
	}
	msgs := []proto.Message{&msg}

	hostZoneId := route.HostAccount.ChainId
	rewardZoneId := route.RewardAccount.ChainId
	tradeZoneId := route.TradeAccount.ChainId
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId,
		"Preparing MsgTransfer of %+v from %s to %s to %s", msg.Token, hostZoneId, rewardZoneId, tradeZoneId))

	// Send the ICA tx to kick off transfer from hostZone through rewardZone to the tradeZone (no callbacks)
	hostAccount := route.HostAccount
	withdrawalOwner := types.FormatHostZoneICAOwner(hostAccount.ChainId, hostAccount.Type)
	err = k.SubmitICATxWithoutCallback(ctx, hostAccount.ConnectionId, withdrawalOwner, msgs, msg.TimeoutTimestamp)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to submit ICA tx, Messages: %+v", msgs)
	}

	return nil
}

// ICA tx to kick off transfering the converted tokens back from tradeZone to the hostZone withdrawal ICA
func (k Keeper) TransferConvertedTokensTradeToHost(ctx sdk.Context, amount sdkmath.Int, route types.TradeRoute) error {
	// Timeout for ica tx and the transfer msgs is at end of epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrap(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeout := uint64(strideEpochTracker.NextEpochStartTime)

	convertedDenom := route.HostDenomOnTradeZone
	sendTokens := sdk.NewCoin(convertedDenom, amount)

	// Validate ICAs were registered
	tradeIcaAddress := route.TradeAccount.Address
	withdrawlIcaAddress := route.HostAccount.Address
	if withdrawlIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no host account found for %s", route.Description())
	}
	if tradeIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no trade account found for %s", route.Description())
	}

	var msgs []proto.Message
	msgs = append(msgs, &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    route.TradeToHostChannelId, // channel on tradeZone for transfers to hostZone
		Token:            sendTokens,
		Sender:           tradeIcaAddress,
		Receiver:         withdrawlIcaAddress,
		TimeoutTimestamp: timeout,
		Memo:             "",
	})

	hostZoneId := route.HostAccount.ChainId
	tradeZoneId := route.TradeAccount.ChainId
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId,
		"Preparing MsgTransfer of %+v from %s to %s", sendTokens, tradeZoneId, hostZoneId))

	// Send the ICA tx to kick off transfer from hostZone through rewardZone to the tradeZone (no callbacks)
	tradeAccount := route.TradeAccount
	tradeOwner := types.FormatTradeRouteICAOwnerFromRouteId(tradeAccount.ChainId, route.GetRouteId(), tradeAccount.Type)
	err := k.SubmitICATxWithoutCallback(ctx, tradeAccount.ConnectionId, tradeOwner, msgs, timeout)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to submit ICA tx, Messages: %+v", msgs)
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ICQ calls for remote ICA balances
// There is a single trade zone
// We have to initialize a single hostZone object for the trade zone once in initialization and
// then it can be used in all these calls
///////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Kick off ICQ for the reward denom balance in the withdrawal address
func (k Keeper) WithdrawalRewardBalanceQuery(ctx sdk.Context, route types.TradeRoute) error {
	withdrawalAccount := route.HostAccount
	k.Logger(ctx).Info(utils.LogWithHostZone(withdrawalAccount.ChainId, "Submitting ICQ for reward denom in withdrawal account"))

	// Encode the withdrawal account address for the query request
	// The query request consists of the withdrawal account address and reward denom
	_, withdrawalAddressBz, err := bech32.DecodeAndConvert(withdrawalAccount.Address)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid withdrawal account address (%s), could not decode", withdrawalAccount.Address)
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(withdrawalAddressBz), []byte(route.RewardDenomOnHostZone)...)

	// Timeout the query halfway through the epoch (since that's when the first transfer
	// in the pfm sequence will timeout)
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrap(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeoutDuration := time.Duration(utils.UintToInt(strideEpochTracker.Duration)) / 2

	// We need the trade route keys in the callback to look up the tradeRoute struct
	callbackData := types.TradeRouteCallback{
		RewardDenom: route.RewardDenomOnRewardZone,
		HostDenom:   route.HostDenomOnHostZone,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal TradeRoute callback data")
	}

	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         withdrawalAccount.ChainId,
		ConnectionId:    withdrawalAccount.ConnectionId,
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_WithdrawalRewardBalance,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return err
	}

	return nil
}

// Kick off ICQ for how many converted tokens are in the trade ICA associated with this host zone
func (k Keeper) TradeConvertedBalanceQuery(ctx sdk.Context, route types.TradeRoute) error {
	tradeAccount := route.TradeAccount
	k.Logger(ctx).Info(utils.LogWithHostZone(tradeAccount.ChainId, "Submitting ICQ for converted denom in trade ICA account"))

	// Encode the trade account address for the query request
	// The query request consists of the trade account address and converted denom
	// keep in mind this ICA address actually exists on trade zone but is associated with trades performed for host zone
	_, tradeAddressBz, err := bech32.DecodeAndConvert(tradeAccount.Address)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid trade account address (%s), could not decode", tradeAccount.Address)
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(tradeAddressBz), []byte(route.HostDenomOnTradeZone)...)

	// Timeout query at end of epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrap(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeout := time.Unix(0, utils.UintToInt(strideEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// We need the trade route keys in the callback to look up the tradeRoute struct
	callbackData := types.TradeRouteCallback{
		RewardDenom: route.RewardDenomOnRewardZone,
		HostDenom:   route.HostDenomOnHostZone,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal trade route as callback data")
	}

	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         tradeAccount.ChainId,
		ConnectionId:    tradeAccount.ConnectionId, // query needs to go to the trade zone, not the host zone
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_TradeConvertedBalance,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return err
	}

	return nil
}

// Main epochly trigger for the trade route tokens swap
//
// The current design assumes foreign reward tokens start and end in the hostZone withdrawal address
// Step 1: transfer reward tokens to trade chain
// Step 2: (off-chain) perform the swap in small batches
// Step 3: return the swapped tokens to the withdrawal ICA on hostZone
func (k Keeper) TransferAllRewardTokens(ctx sdk.Context) {
	for _, route := range k.GetAllTradeRoutes(ctx) {
		// Step 1: ICQ reward balance on hostZone, transfer funds with unwinding to trade chain
		if err := k.WithdrawalRewardBalanceQuery(ctx, route); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to submit query for reward balance in withdrawal ICA: %s", err))
		}
		// Step 3: ICQ converted tokens in trade ICA, transfer funds back to hostZone withdrawal ICA
		if err := k.TradeConvertedBalanceQuery(ctx, route); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to submit query for converted balance in trade ICA: %s", err))
		}
	}
}
