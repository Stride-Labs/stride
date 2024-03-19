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

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v19/utils"
	epochstypes "github.com/Stride-Labs/stride/v19/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v19/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
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

type FeeInfo struct {
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
//  2. Epochly check the reward denom balance in trade ICA
//     on callback, trade all reward denom for host denom defined by pool and routes in params
//  3. Epochly check the host denom balance in trade ICA
//     on callback, transfer these host denom tokens from trade ICA to withdrawal ICA on original host zone
//
// Normal staking flow continues from there. So the host denom tokens will land on the original host zone
// and the normal staking and distribution flow will continue from there.
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO [rebate]: Update when switching to stDYDX
// Calculates the portion of fees that were contributed to from the community pool liquid stake
// This is determined by dividing community pool's liquid stake size by the total TVL
func (k Keeper) GetRebateEligibleFeeRate(
	ctx sdk.Context,
	hostZone types.HostZone,
	rebateInfo types.CommunityPoolRebate,
) (contributionRate sdk.Dec, err error) {
	// It shouldn't be possible to have 0 delegations (since there are rewards and there was a community pool stake)
	// This will also prevent a division by 0 error
	if hostZone.TotalDelegations.IsZero() {
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrDivisionByZero,
			"unable to calculate rebate amount for %s since total delegations are 0", hostZone.ChainId)
	}

	// It also shouldn't be possible for the liquid stake amount to be greater than the full TVL
	if rebateInfo.LiquidStakeAmount.GT(hostZone.TotalDelegations) {
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrFeeSplitInvariantFailed,
			"community pool liquid staked amount greater than total delegations")
	}

	contributionRate = sdk.NewDecFromInt(rebateInfo.LiquidStakeAmount).Quo(sdk.NewDecFromInt(hostZone.TotalDelegations))

	return contributionRate, nil
}

// Breaks down the split of non-native-denom rewards into the portions intended for a rebate vs the remainder
// that's used for fees and reinvestment
// For instance, in the case of dYdX, this is used on the USDC that has not been pushed through the trade route
// yet, and has not yet been converted to DYDX
//
// The rebate percentage is determined by: (% of total TVL contributed by commuity pool) * (rebate percentage)
//
// E.g. Community pool liquid staked 1M, TVL is 10M, rebate is 20%
// Total rewards this epoch are 1000, and the stride fee is 10%
// => Then the rebate is 1000 rewards * 10% stride fee * (1M / 10M) * 20% rebate = 2
func (k Keeper) CalculateRewardsSplitBeforeRebate(
	ctx sdk.Context,
	hostZone types.HostZone,
	rewardAmount sdkmath.Int,
) (rebateAmount sdkmath.Int, remainingAmount sdkmath.Int, err error) {
	// Get the rebate info from the host zone if applicable
	// If there's no rebate, return 0 rebate and the full reward as the remainder
	rebateInfo, chainHasRebate := hostZone.SafelyGetCommunityPoolRebate()
	if !chainHasRebate {
		return sdkmath.ZeroInt(), rewardAmount, nil
	}

	// Get the fee rate from params (e.g. 0.1 for a 10% fee)
	strideFeeRate := k.GetParams(ctx).GetStrideCommissionRate()

	// Calculate the portion of fees that are rebate-eligible
	// (equal to the % of TVL attributable to the community pool's liquid stake)
	contributionRate, err := k.GetRebateEligibleFeeRate(ctx, hostZone, rebateInfo)
	if err != nil {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
	}

	// The rebate amount is determined by the contribution of the community pool stake towards the total TVL,
	// multiplied by the rebate fee percentage
	totalFeesAmount := sdk.NewDecFromInt(rewardAmount).Mul(strideFeeRate).TruncateInt()
	rebateAmount = sdk.NewDecFromInt(totalFeesAmount).Mul(contributionRate).Mul(rebateInfo.RebatePercentage).TruncateInt()
	remainingAmount = rewardAmount.Sub(rebateAmount)

	return rebateAmount, remainingAmount, nil
}

// Given the native-denom reward balance from an ICQ, calculates the relevant portions earmarked as a
// stride fee vs for reinvestment
// This is called *after* a rebate has been issued (if there is one to begin with)
// The reward amount is denominated in the host zone's native denom - this is either directly from
// staking rewards, or, in the case of a host zone with a trade route, this is called with the converted tokens
// For instance, with dYdX, this is applied to the DYDX tokens that were converted from USDC
//
// If the chain doesn't have a rebate in place, the split is decided entirely from the stride commission percent
// However, if the chain does have a rebate, we need to factor that into the calculation, by scaling
// up the rewards to find the amount before the rebate
//
// For instance, if 1000 rewards were collected and 2 were sent as a rebate, then the stride fee should be based
// on the original 1000 rewards instead of the remaining 998 in the query response:
//
//	Community pool liquid staked 1M, TVL is 10M, rebate is 20%, stride fee is 10%
//	If 998 native tokens were queried, we have to scale that up to 1000 original reward tokens
//
//	Effective Rebate Pct = 10% fees * (1M LS / 10M TVL) * 20% rebate = 0.20% (aka 0.002)
//	Effective Stride Fee Pct = 10% fees - 0.20% effective rebate = 9.8%
//	Original Reward Amount = 998 Queried Rewards / (1 - 0.002 effective rebate rate) = 1000 original rewards
//	Then stride fees are 9.8% of that 1000 original rewards = 98
func (k Keeper) CalculateRewardsSplitAfterRebate(
	ctx sdk.Context,
	hostZone types.HostZone,
	rewardsAmount sdkmath.Int,
) (strideFeeAmount sdkmath.Int, reinvestAmount sdkmath.Int, err error) {
	// Get the fee rate and total fees from params (e.g. 0.1 for 10% fee)
	totalFeeRate := k.GetParams(ctx).GetStrideCommissionRate()

	// Check if the chain has a rebate
	rebateInfo, chainHasRebate := hostZone.SafelyGetCommunityPoolRebate()

	// If there's no rebate, the fee split just uses the commission
	// Otherwise, the rebate must be considered in the fee split
	if !chainHasRebate {
		strideFeeAmount = sdk.NewDecFromInt(rewardsAmount).Mul(totalFeeRate).TruncateInt()
	} else {
		// Calculate the portion of fees that are rebate-eligible
		// (equal to the % of TVL attributable to the community pool's liquid stake)
		contributionRate, err := k.GetRebateEligibleFeeRate(ctx, hostZone, rebateInfo)
		if err != nil {
			return sdkmath.ZeroInt(), sdkmath.ZeroInt(), err
		}

		// The effective rebate rate is the portion of TVL contributed to by the liquid stake * the rebate percentage
		// The effective stride fee poriton is the remaining percentage
		effectiveRebateRate := totalFeeRate.Mul(contributionRate).Mul(rebateInfo.RebatePercentage)
		effectiveStrideFeeRate := totalFeeRate.Sub(effectiveRebateRate)

		// Before calculating the fee, we have to scale up the rewards amount to the amount before the rebate
		originalRewardScalingFactor := sdk.OneDec().Sub(effectiveRebateRate)
		originalRewardsAmount := sdk.NewDecFromInt(rewardsAmount).Quo(originalRewardScalingFactor).TruncateDec()
		strideFeeAmount = originalRewardsAmount.Mul(effectiveStrideFeeRate).Ceil().TruncateInt() // ceiling since rebate was truncated
	}

	// Using the strideFeeAmount, back into the reinvest amount
	reinvestAmount = rewardsAmount.Sub(strideFeeAmount)

	return strideFeeAmount, reinvestAmount, nil
}

// Builds an authz MsgGrant or MsgRevoke to grant an account trade capabilties on behalf of the trade ICA
func (k Keeper) BuildTradeAuthzMsg(
	ctx sdk.Context,
	tradeRoute types.TradeRoute,
	permissionChange types.AuthzPermissionChange,
	grantee string,
) (authzMsg []proto.Message, err error) {
	swapMsgTypeUrl := "/" + proto.MessageName(&types.MsgSwapExactAmountIn{})

	switch permissionChange {
	case types.AuthzPermissionChange_GRANT:
		authorization := authz.NewGenericAuthorization(swapMsgTypeUrl)
		grant, err := authz.NewGrant(ctx.BlockTime(), authorization, nil)
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
			MsgTypeUrl: swapMsgTypeUrl,
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
		return msg, errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
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
	// If the min swap amount was not set it would be ZeroInt, if positive we need to compare to the amount given
	//  then if the min swap amount is greater than the current amount, do nothing this epoch to avoid small transfers
	//  Particularly important for the PFM hop if the reward chain has frictional transfer fees (like noble chain)
	if route.TradeConfig.MinSwapAmount.GT(amount) {
		return nil
	}

	// Similarly, if there's no price on the trade route yet, don't initiate the transfer because
	// we know the swap will not be submitted
	if route.TradeConfig.SwapPrice.IsZero() {
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
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
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

// Builds the Osmosis swap message to trade reward tokens for host tokens
// Depending on min and max swap amounts set in the route, it is possible not the full amount given will swap
// The minimum amount of tokens that can come out of the trade is calculated using a price from the pool
func (k Keeper) BuildSwapMsg(rewardAmount sdkmath.Int, route types.TradeRoute) (msg types.MsgSwapExactAmountIn, err error) {
	// Validate the trade ICA was registered
	tradeIcaAddress := route.TradeAccount.Address
	if tradeIcaAddress == "" {
		return msg, errorsmod.Wrapf(types.ErrICAAccountNotFound, "no trade account found for %s", route.Description())
	}

	// If the max swap amount was not set it would be ZeroInt, if positive we need to compare to the amount given
	//  then if max swap amount is LTE to amount full swap is possible so amount is fine, otherwise set amount to max
	tradeConfig := route.TradeConfig
	if tradeConfig.MaxSwapAmount.IsPositive() && rewardAmount.GT(tradeConfig.MaxSwapAmount) {
		rewardAmount = tradeConfig.MaxSwapAmount
	}

	// See if pool swap price has been set to a valid ratio
	// The only time this should not be set is right after the pool is added,
	// before an ICQ has been submitted for the price
	if tradeConfig.SwapPrice.IsZero() {
		return msg, fmt.Errorf("Price not found for pool %d", tradeConfig.PoolId)
	}

	// If there is a valid price, use it to set a floor for the acceptable minimum output tokens
	// minOut is the minimum number of HostDenom tokens we must receive or the swap will fail
	//
	// To calculate minOut, we first convert the rewardAmount into units of HostDenom,
	//   and then we multiply by (1 - MaxAllowedSwapLossRate)
	//
	// The price on the trade route represents the ratio of host denom to reward denom
	// So, to convert from units of RewardTokens to units of HostTokens,
	// we multiply the reward amount by the price:
	//   AmountInHost = AmountInReward * SwapPrice
	rewardAmountConverted := sdk.NewDecFromInt(rewardAmount).Mul(tradeConfig.SwapPrice)
	minOutPercentage := sdk.OneDec().Sub(tradeConfig.MaxAllowedSwapLossRate)
	minOut := rewardAmountConverted.Mul(minOutPercentage).TruncateInt()

	tradeTokens := sdk.NewCoin(route.RewardDenomOnTradeZone, rewardAmount)

	// Prepare Osmosis GAMM module MsgSwapExactAmountIn from the trade account to perform the trade
	// If we want to generalize in the future, write swap message generation funcs for each DEX type,
	// decide which msg generation function to call based on check of which tradeZone was passed in
	routes := []types.SwapAmountInRoute{{
		PoolId:        tradeConfig.PoolId,
		TokenOutDenom: route.HostDenomOnTradeZone,
	}}
	msg = types.MsgSwapExactAmountIn{
		Sender:            tradeIcaAddress,
		Routes:            routes,
		TokenIn:           tradeTokens,
		TokenOutMinAmount: minOut,
	}

	return msg, nil
}

// DEPRECATED: The on-chain swap has been deprecated in favor of an authz controller
// Trade reward tokens in the Trade ICA for the host denom tokens using ICA remote tx on trade zone
// The amount represents the total amount of the reward token in the trade ICA found by the calling ICQ
func (k Keeper) SwapRewardTokens(ctx sdk.Context, rewardAmount sdkmath.Int, route types.TradeRoute) error {
	// If the min swap amount was not set it would be ZeroInt, if positive we need to compare to the amount given
	//  then if the min swap amount is greater than the current amount, do nothing this epoch to avoid small swaps
	tradeConfig := route.TradeConfig
	if tradeConfig.MinSwapAmount.IsPositive() && tradeConfig.MinSwapAmount.GT(rewardAmount) {
		return nil
	}

	// Build the Osmosis swap message to convert reward tokens to host tokens
	msg, err := k.BuildSwapMsg(rewardAmount, route)
	if err != nil {
		return err
	}
	msgs := []proto.Message{&msg}

	tradeAccount := route.TradeAccount
	k.Logger(ctx).Info(utils.LogWithHostZone(tradeAccount.ChainId,
		"Preparing MsgSwapExactAmountIn of %+v from the trade account", msg.TokenIn))

	// Timeout the swap at the end of the epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.HOUR_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.HOUR_EPOCH)
	}
	timeout := uint64(strideEpochTracker.NextEpochStartTime)

	// Send the ICA tx to perform the swap on the tradeZone
	tradeOwner := types.FormatTradeRouteICAOwnerFromRouteId(tradeAccount.ChainId, route.GetRouteId(), tradeAccount.Type)
	err = k.SubmitICATxWithoutCallback(ctx, tradeAccount.ConnectionId, tradeOwner, msgs, timeout)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to submit ICA tx for the swap, Messages: %v", msgs)
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ICQ calls for remote ICA balances
// There is a single trade zone (hardcoded as Osmosis for now but maybe additional DEXes allowed in the future)
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
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeoutDuration := time.Duration(strideEpochTracker.Duration) / 2

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

// Kick off ICQ for how many reward tokens are in the trade ICA associated with this host zone
func (k Keeper) TradeRewardBalanceQuery(ctx sdk.Context, route types.TradeRoute) error {
	tradeAccount := route.TradeAccount
	k.Logger(ctx).Info(utils.LogWithHostZone(tradeAccount.ChainId, "Submitting ICQ for reward denom in trade ICA account"))

	// Encode the trade account address for the query request
	// The query request consists of the trade account address and reward denom
	// keep in mind this ICA address actually exists on trade zone but is associated with trades performed for host zone
	_, tradeAddressBz, err := bech32.DecodeAndConvert(tradeAccount.Address)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid trade account address (%s), could not decode", tradeAccount.Address)
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(tradeAddressBz), []byte(route.RewardDenomOnTradeZone)...)

	// Timeout query at end of epoch
	hourEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.HOUR_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.HOUR_EPOCH)
	}
	timeout := time.Unix(0, int64(hourEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// We need the trade route keys in the callback to look up the tradeRoute struct
	callbackData := types.TradeRouteCallback{
		RewardDenom: route.RewardDenomOnRewardZone,
		HostDenom:   route.HostDenomOnHostZone,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal TradeRewardBalanceQuery callback data")
	}

	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         tradeAccount.ChainId,
		ConnectionId:    tradeAccount.ConnectionId, // query needs to go to the trade zone, not the host zone
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_TradeRewardBalance,
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
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeout := time.Unix(0, int64(strideEpochTracker.NextEpochStartTime))
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

// DEPRECATED: The on-chain swap has been deprecated in favor of an authz controller. Price is no longer needed
// Kick off ICQ for the spot price on the pool given the input and output denoms implied by the given TradeRoute
// the callback for this query is responsible for updating the returned spot price on the keeper data
func (k Keeper) PoolPriceQuery(ctx sdk.Context, route types.TradeRoute) error {
	tradeAccount := route.TradeAccount
	k.Logger(ctx).Info(utils.LogWithHostZone(tradeAccount.ChainId, "Submitting ICQ for spot price in this pool"))

	// Build query request data which consists of the TWAP store key built from each denom
	queryData := icqtypes.FormatOsmosisMostRecentTWAPKey(
		route.TradeConfig.PoolId,
		route.RewardDenomOnTradeZone,
		route.HostDenomOnTradeZone,
	)

	// Timeout query at end of epoch
	hourEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.HOUR_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.HOUR_EPOCH)
	}
	timeout := time.Unix(0, int64(hourEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// We need the trade route keys in the callback to look up the tradeRoute struct
	callbackData := types.TradeRouteCallback{
		RewardDenom: route.RewardDenomOnRewardZone,
		HostDenom:   route.HostDenomOnHostZone,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal TradeRewardBalanceQuery callback data")
	}

	// Submit the ICQ for the trade pool spot price query
	query := icqtypes.Query{
		ChainId:         tradeAccount.ChainId,
		ConnectionId:    tradeAccount.ConnectionId, // query needs to go to the trade zone, not the host zone
		QueryType:       icqtypes.TWAP_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_PoolPrice,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying pool spot price, error: %s", err.Error()))
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// The current design assumes foreign reward tokens start and end in the hostZone withdrawal address
// Step 1: transfer reward tokens to trade chain
// Step 2: perform the swap with as many reward tokens as possible
// Step 3: return the swapped tokens to the withdrawal ICA on hostZone
// Independently there is an ICQ to get the swap price and update it in the keeper state
//
// Because the swaps have limits on how many tokens can be used to avoid slippage,
// the swaps and price checks happen on a faster (hourly) cadence than the transfers (stride epochly)
////////////////////////////////////////////////////////////////////////////////////////////////////

// Helper function to be run stride epochly, kicks off queries on specific denoms on route
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

// Helper function to be run hourly, kicks off query which will kick off actual swaps to happen
func (k Keeper) SwapAllRewardTokens(ctx sdk.Context) {
	for _, route := range k.GetAllTradeRoutes(ctx) {
		// Step 2: ICQ reward balance in trade ICA, swap tokens according to limiting rules
		if err := k.TradeRewardBalanceQuery(ctx, route); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to submit query for reward balance in trade ICA: %s", err))
		}
	}
}

// Helper function to be run hourly, kicks off query to get and update the swap price in keeper data
func (k Keeper) UpdateAllSwapPrices(ctx sdk.Context) {
	for _, route := range k.GetAllTradeRoutes(ctx) {
		// ICQ swap price for the specific pair on this route and update keeper on callback
		if err := k.PoolPriceQuery(ctx, route); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to submit query for pool spot price: %s", err))
		}
	}
}
