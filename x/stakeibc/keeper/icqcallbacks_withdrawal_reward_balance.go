package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"

	icqkeeper "github.com/Stride-Labs/stride/v20/x/interchainquery/keeper"

	"github.com/Stride-Labs/stride/v20/utils"
	icqtypes "github.com/Stride-Labs/stride/v20/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v20/x/stakeibc/types"
)

// WithdrawalRewardBalanceCallback is a callback handler for WithdrawalRewardBalance queries.
// The query response will return the withdrawal account balance for the reward denom in the case
// of a host zone with a trade route (e.g. USDC in the case of the dYdX trade route)
// If the balance is non-zero, ICA MsgSends are submitted to transfer the discovered balance to the tradeZone
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func WithdrawalRewardBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_WithdrawalRewardBalance,
		"Starting withdrawal reward balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	chainId := query.ChainId

	// Unmarshal the query response args to determine the balance
	withdrawalRewardBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}

	// Confirm the balance is greater than zero, or else exit early without further action
	if withdrawalRewardBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalRewardBalance,
			"Not enough reward tokens yet found in withdrawalICA, balance: %v", withdrawalRewardBalanceAmount))
		return nil
	}

	// Unmarshal the callback data containing the tradeRoute we are on
	var tradeRouteCallback types.TradeRouteCallback
	if err := proto.Unmarshal(query.CallbackData, &tradeRouteCallback); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal trade reward balance callback data")
	}

	// Lookup the trade route from the keys in the callback
	tradeRoute, found := k.GetTradeRoute(ctx, tradeRouteCallback.RewardDenom, tradeRouteCallback.HostDenom)
	if !found {
		return types.ErrTradeRouteNotFound.Wrapf("trade route from %s to %s not found",
			tradeRouteCallback.RewardDenom, tradeRouteCallback.HostDenom)
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalRewardBalance,
		"Query response - Withdrawal Reward Balance: %v %s", withdrawalRewardBalanceAmount, tradeRoute.RewardDenomOnHostZone))

	// Transfer the reward amount to the trade zone so it can be swapped for the native token
	if err := k.TransferRewardTokensHostToTrade(ctx, withdrawalRewardBalanceAmount, tradeRoute); err != nil {
		return errorsmod.Wrapf(err, "initiating transfer of reward tokens to trade ICA failed")
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalRewardBalance,
		"Sending discovered reward tokens %v %s from hostZone to tradeZone",
		withdrawalRewardBalanceAmount, tradeRoute.RewardDenomOnRewardZone))

	return nil
}
