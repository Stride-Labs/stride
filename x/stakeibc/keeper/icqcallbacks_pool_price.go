package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v16/utils"
	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

// TradeRewardBalanceCallback is a callback handler for TradeRewardBalance queries.
// The query response will return the trade ICA account balance for a specific (foreign ibc) denom
// If the balance is non-zero, ICA MsgSends are submitted to initiate a swap on the tradeZone
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func PoolPriceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_PoolPrice,
		"Starting pool spot price callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	chainId := query.ChainId // should be the tradeZoneId, used in logging

	// Unmarshal the query response args, should be a TwapRecord type
	var twapRecord types.OsmosisTwapRecord
	err := twapRecord.Unmarshal(args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to unmarshal the query response")
	}

	price := sdk.ZeroDec()
	fmt.Printf("%+v\n", twapRecord)

	// Unmarshal the callback data containing the tradeRoute we are on
	var tradeRoute types.TradeRoute
	if err := proto.Unmarshal(query.CallbackData, &tradeRoute); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal trade reward balance callback data")
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_PoolPrice,
		"Query response - spot price ratio of %s to %s is %s",
		tradeRoute.RewardDenomOnTradeZone, tradeRoute.TargetDenomOnTradeZone, price))

	// Update the spot price stored on the trade route data in the keeper
	tradeRoute.InputToOutputTwap = price
	k.SetTradeRoute(ctx, tradeRoute)

	return nil
}
