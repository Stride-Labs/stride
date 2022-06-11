package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: &keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) SubmitQueryResponse(goCtx context.Context, msg *types.MsgSubmitQueryResponse) (*types.MsgSubmitQueryResponseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	q, found := k.GetQuery(ctx, msg.QueryId)
	if found {
		for _, module := range k.callbacks {
			if module.Has(msg.QueryId) {
				err := module.Call(ctx, msg.QueryId, msg.Result, q)
				if err != nil {
					k.Logger(ctx).Error("Error in callback", "error", err, "msg", msg.QueryId, "result", msg.Result, "type", q.QueryType, "request", q.Request)
					return nil, err
				}
			}
		}
		//q.LastHeight = sdk.NewInt(ctx.BlockHeight())

		if err := k.SetDatapointForId(ctx, msg.QueryId, msg.Result, sdk.NewInt(msg.Height)); err != nil {
			return nil, err
		}

		if q.Period.IsNegative() {
			k.DeleteQuery(ctx, msg.QueryId)
		} else {
			k.SetQuery(ctx, q)
		}

	} else {
		return &types.MsgSubmitQueryResponseResponse{}, nil // technically this is an error, but will cause the entire tx to fail if we have one 'bad' message, so we can just no-op here.
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &types.MsgSubmitQueryResponseResponse{}, nil
}

// TODO(TEST-80): rm this
// func (k msgServer) QueryBalance(goCtx context.Context, msg *types.MsgQueryBalance) (*types.MsgQueryBalanceResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	// TODO(TEST-84) check host chain is supported
// 	ChainId := msg.ChainId

// 	// Parse caller Address
// 	_, err := sdk.AccAddressFromBech32(msg.Caller)
// 	if err != nil {
// 		panic(err)
// 	}
// 	ConnectionId := msg.ConnectionId

// 	var cb Callback = func(k Keeper, ctx sdk.Context, args []byte, query types.Query) error {

// 		var response stakingtypes.QueryDelegatorDelegationsResponse
// 		err := k.cdc.Unmarshal(args, &response)
// 		if err != nil {
// 			return err
// 		}

// 		// Get denom dynamically
// 		hz, _ := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
// 		delegatorSum := sdk.NewCoin(hz.HostDenom, sdk.ZeroInt())
// 		for _, delegation := range response.DelegationResponses {
// 			delegatorSum = delegatorSum.Add(delegation.Balance)
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		// Set Redemption Rate Based On Delegation Balance vs stAsset Supply
// 		// Get IBC Denom
// 		stAssetSupply := k.BankKeeper.GetSupply(ctx, hz.IBCDenom)
// 		redemptionRate := delegatorSum.Amount.ToDec().Quo(stAssetSupply.Amount.ToDec())

// 		// get zone
// 		hz, found := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
// 		if found {
// 			fmt.Errorf("invalid chain id, zone for \"%s\" already registered", ChainId)
// 		}

// 		// update redemptionRate and LastRedemptionRate on hz
// 		hz.LastRedemptionRate = hz.RedemptionRate
// 		hz.RedemptionRate = redemptionRate
// 		// write the zone back to the store
// 		k.StakeibcKeeper.SetHostZone(ctx, hz)

// 		ctx.EventManager().EmitEvents(sdk.Events{
// 			sdk.NewEvent(
// 				sdk.EventTypeMessage,
// 				sdk.NewAttribute("totalDelegations", delegatorSum.String()),
// 				sdk.NewAttribute("stAssetSupply", stAssetSupply.Amount.String()),
// 				sdk.NewAttribute("redemptionRate", redemptionRate.String()),
// 			),
// 		})

// 		return nil
// 	}

// 	query_type := "cosmos.staking.v1beta1.Query/DelegatorDelegations"
// 	delegationQuery := stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: msg.Address}
// 	bz := k.cdc.MustMarshal(&delegationQuery)

// 	if err != nil {
// 		return nil, err
// 	}

// 	k.Keeper.MakeRequest(
// 		ctx,
// 		ConnectionId,
// 		ChainId,
// 		query_type,
// 		bz,
// 		// TODO(TEST-79) understand and use proper period
// 		sdk.NewInt(25),
// 		types.ModuleName,
// 		cb,
// 	)

// 	ctx.EventManager().EmitEvents(sdk.Events{
// 		sdk.NewEvent(
// 			sdk.EventTypeMessage,
// 			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
// 			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
// 			sdk.NewAttribute(types.AttributeKeyQueryId, GenerateQueryHash(ConnectionId, ChainId, query_type, bz)),
// 			sdk.NewAttribute(types.AttributeKeyChainId, ChainId),
// 			sdk.NewAttribute(types.AttributeKeyConnectionId, ConnectionId),
// 			sdk.NewAttribute(types.AttributeKeyType, query_type),
// 			sdk.NewAttribute(types.AttributeKeyHeight, "0"),
// 			sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(bz)),
// 		),
// 	})

// 	// return; usually a response object or nil
// 	return &types.MsgQueryBalanceResponse{}, nil
// }

// // TODO(TEST-78) rename from query-XXX => update-XXX
// func (k msgServer) QueryExchangerate(goCtx context.Context, msg *types.MsgQueryExchangerate) (*types.MsgQueryExchangerateResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	// TODO(TEST-84) check host chain is supported
// 	ChainId := msg.ChainId

// 	// Parse caller Address
// 	_, err := sdk.AccAddressFromBech32(msg.Creator)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// get zone
// 	hz, found := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
// 	if found {
// 		fmt.Errorf("invalid chain id, zone for \"%s\" already registered", ChainId)
// 	}
// 	ConnectionId := hz.ConnectionId

// 	var cb Callback = func(k Keeper, ctx sdk.Context, args []byte, query types.Query) error {
// 		var response stakingtypes.QueryDelegatorDelegationsResponse
// 		err := k.cdc.Unmarshal(args, &response)
// 		if err != nil {
// 			return err
// 		}

// 		// Get denom dynamically
// 		hz, _ := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
// 		delegatorSum := sdk.NewCoin(hz.HostDenom, sdk.ZeroInt())
// 		for _, delegation := range response.DelegationResponses {
// 			delegatorSum = delegatorSum.Add(delegation.Balance)
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		// Set Redemption Rate Based On Delegation Balance vs stAsset Supply
// 		// Get denom dynamically
// 		stAssetSupply := k.BankKeeper.GetSupply(ctx, hz.IBCDenom)
// 		redemptionRate := delegatorSum.Amount.ToDec().Quo(stAssetSupply.Amount.ToDec())

// 		// update redemptionRate and LastRedemptionRate on hz
// 		hz.LastRedemptionRate = hz.RedemptionRate
// 		hz.RedemptionRate = redemptionRate
// 		// write the zone back to the store
// 		k.StakeibcKeeper.SetHostZone(ctx, hz)

// 		ctx.EventManager().EmitEvents(sdk.Events{
// 			sdk.NewEvent(
// 				sdk.EventTypeMessage,
// 				sdk.NewAttribute("totalDelegations", delegatorSum.String()),
// 				sdk.NewAttribute("stAssetSupply", stAssetSupply.Amount.String()),
// 				sdk.NewAttribute("redemptionRate", redemptionRate.String()),
// 			),
// 		})

// 		return nil
// 	}

// 	query_type := "cosmos.staking.v1beta1.Query/DelegatorDelegations"
// 	// Get delegationAddress dynamically
// 	delegationQuery := stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: hz.DelegationAccount.Address}
// 	bz := k.cdc.MustMarshal(&delegationQuery)
// 	if err != nil {
// 		return nil, err
// 	}

// 	k.Keeper.MakeRequest(
// 		ctx,
// 		ConnectionId,
// 		ChainId,
// 		query_type,
// 		bz,
// 		// TODO(TEST-79) understand and use proper period
// 		sdk.NewInt(25),
// 		types.ModuleName,
// 		cb,
// 	)

// 	ctx.EventManager().EmitEvents(sdk.Events{
// 		sdk.NewEvent(
// 			sdk.EventTypeMessage,
// 			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
// 			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
// 			sdk.NewAttribute(types.AttributeKeyQueryId, GenerateQueryHash(ConnectionId, ChainId, query_type, bz)),
// 			sdk.NewAttribute(types.AttributeKeyChainId, ChainId),
// 			sdk.NewAttribute(types.AttributeKeyConnectionId, ConnectionId),
// 			sdk.NewAttribute(types.AttributeKeyType, query_type),
// 			// TODO(TEST-79) understand height
// 			sdk.NewAttribute(types.AttributeKeyHeight, "0"),
// 			sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(bz)),
// 		),
// 	})

// 	// return; usually a response object or nil
// 	return &types.MsgQueryExchangerateResponse{}, nil
// }

// // TODO(TEST-78) rename from query-XXX => update-XXX
// func (k msgServer) QueryDelegatedbalance(goCtx context.Context, msg *types.MsgQueryDelegatedbalance) (*types.MsgQueryDelegatedbalanceResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	// TODO(TEST-84) check host chain is supported
// 	ChainId := msg.ChainId

// 	// Parse caller Address
// 	_, err := sdk.AccAddressFromBech32(msg.Creator)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// get zone
// 	hz, found := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
// 	if found {
// 		fmt.Errorf("invalid chain id, zone for \"%s\" already registered", ChainId)
// 	}
// 	ConnectionId := hz.ConnectionId

// 	var cb Callback = func(k Keeper, ctx sdk.Context, args []byte, query types.Query) error {
// 		queryRes := banktypes.QueryAllBalancesResponse{}
// 		err := k.cdc.Unmarshal(args, &queryRes)
// 		if err != nil {
// 			k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
// 			return err
// 		}

// 		// Get denom dynamically
// 		hz, _ := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
// 		balance := queryRes.Balances.AmountOf(hz.HostDenom)

// 		// Set delegation account balance to ICQ result
// 		hz, found := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
// 		if found {
// 			fmt.Errorf("invalid chain id, zone for \"%s\" already registered", ChainId)
// 		}

// 		da := hz.DelegationAccount
// 		da.Balance = int32(balance.Int64())
// 		hz.DelegationAccount = da
// 		k.StakeibcKeeper.SetHostZone(ctx, hz)

// 		ctx.EventManager().EmitEvents(sdk.Events{
// 			sdk.NewEvent(
// 				sdk.EventTypeMessage,
// 				sdk.NewAttribute("totalBalance", balance.String()),
// 			),
// 		})

// 		return nil
// 	}

// 	query_type := "cosmos.bank.v1beta1.Query/AllBalances"
// 	balanceQuery := banktypes.QueryAllBalancesRequest{Address: hz.GetDelegationAccount().Address}
// 	bz, err := k.cdc.Marshal(&balanceQuery)
// 	if err != nil {
// 		return nil, err
// 	}

// 	k.Keeper.MakeRequest(
// 		ctx,
// 		ConnectionId,
// 		ChainId,
// 		query_type,
// 		bz,
// 		// TODO(TEST-79) understand and use proper period
// 		sdk.NewInt(25),
// 		types.ModuleName,
// 		cb,
// 	)

// 	ctx.EventManager().EmitEvents(sdk.Events{
// 		sdk.NewEvent(
// 			sdk.EventTypeMessage,
// 			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
// 			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
// 			sdk.NewAttribute(types.AttributeKeyQueryId, GenerateQueryHash(ConnectionId, ChainId, query_type, bz)),
// 			sdk.NewAttribute(types.AttributeKeyChainId, ChainId),
// 			sdk.NewAttribute(types.AttributeKeyConnectionId, ConnectionId),
// 			sdk.NewAttribute(types.AttributeKeyType, query_type),
// 			sdk.NewAttribute(types.AttributeKeyHeight, "0"),
// 			sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(bz)),
// 		),
// 	})

// 	return &types.MsgQueryDelegatedbalanceResponse{}, nil
// }
