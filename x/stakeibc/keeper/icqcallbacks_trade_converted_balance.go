package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"

	icqkeeper "github.com/Stride-Labs/stride/v14/x/interchainquery/keeper"

	"github.com/Stride-Labs/stride/v14/utils"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// TradeConvertedBalanceCallback is a callback handler for TradeConvertedBalance queries.
// The query response will return the trade account balance for a converted (foreign ibc) denom
// If the balance is non-zero, ICA MsgSends are submitted to transfer the discovered balance back to hostZone
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func TradeConvertedBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_TradeConvertedBalance,
		"Starting trade converted balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	tradeZone, found := k.GetHostZone(ctx, chainId)  // this query goes to the trade zone
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args to determine the balance
	tradeConvertedBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}

	// Unmarshal the callback data containing the hostZone, tradeZone and convertedDenom on the trade zone
	var callbackData types.TradeConvertedBalanceQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal trade reward balance callback data")
	}

	hostZone, found := k.GetHostZone(ctx, callbackData.HostZoneId) // this query went to the trade zone, hostZone comes from callback data
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no host zone could be loaded for callback chain ID (%s)", callbackData.HostZoneId)
	}	


	// Confirm the balance is greater than zero
	if tradeConvertedBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(tradeZone.ChainId, ICQCallbackID_TradeConvertedBalance,
			"No balance of reward tokens yet found in address: %s, balance: %v", hostZone.RewardTradeIcaAddress, tradeConvertedBalanceAmount))
		return nil
	}

	// Using ICA commands on the trade address, transfer the found converted tokens from the trade zone to the host zone	
	k.TransferTradeConvertedTokensToHost(ctx, tradeConvertedBalanceAmount, callbackData.ConvertedDenomOnTradeZone, hostZone, tradeZone)
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalRewardBalance, 
		"Sending discovered converted tokens %v %s from tradeZone %s back to hostZone %s", 
		tradeConvertedBalanceAmount, callbackData.ConvertedDenomOnTradeZone, tradeZone.ChainId, hostZone.ChainId))

	return nil
}
