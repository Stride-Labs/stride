package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

	// TODO remove this, only checking the tx landed
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

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

// Example: "query-balance [chain_id] [address] [denom]"
// TODO(TEST-50) Handling the message
func (k msgServer) QueryBalance(goCtx context.Context, msg *types.MsgQueryBalance) (*types.MsgQueryBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// parse input and do some check on that data, throw errors
	// 		parse an input to the IBC packet we'd like to construct (note that input MsgQueryBalance looks like: {ChainId: chain_id, Address: address, Denom: denom, Caller: from_address})
	// 		check (1) host chain is supported (2) address on target chain is valid (3) denom is valid (4) caller addr is valid
	ChainId := msg.ChainId
	// TODO Check ChainId is supported by Stride, try using this approach https://github.com/ingenuity-build/quicksilver/blob/ea71f23c6ef09a57e601f4e544c4be9693f5ba81/x/interchainstaking/keeper/msg_server.go#L37

	// Parse Address addr
	// TODO should this be Address, not Caller? changed temporarily to suppress error
	_, err := sdk.AccAddressFromBech32(msg.Caller)
	if err != nil {
		panic(err)
	}
	//TODO Check Denom is valid denom (can you do this for ICS20s?)
	// Denom := msg.Denom
	// Parse Caller addr
	// _, err := sdk.AccAddressFromBech32(msg.Caller)
	// if err != nil {
	// 	panic(err)
	// }
	ConnectionId := msg.ConnectionId

	// perform some action e.g. send coins. this requires getting attrs and parsing inputs to get addrs, amts, etc. use a keeper to perform the action (e.g. bankKeeper).
	//		(1) construct the ibc transaction (2) submit the ibc tx
	//			target: target chain's bankKeeper module query
	// Construct the packet

	// func (k *Keeper) MakeRequest(
	// 	ctx sdk.Context,
	// 	connection_id string,
	// 	chain_id string,
	// 	query_type string,
	// 	query_params map[string]string,
	// 	period sdk.Int,
	// 	module string,
	// 	callback interface{})

	// TODO do we need to add a callback type for this to work?
	var cb Callback = func(k Keeper, ctx sdk.Context, args []byte, query types.Query) error {
		// panic(err)

		k.Logger(ctx).Info("[TEMP] printing inside the querybalance callback")
		// return k.SetAccountBalance(ctx, zone, query.QueryParameters["address"],
		//  args)

		// address := query.QueryParameters["address"]

		// queryResult := args
		// queryRes := banktypes.QueryAllBalancesResponse{}
		// err := k.cdc.Unmarshal(queryResult, &queryRes)
		// if err != nil {
		// 	k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
		// 	return err
		// }
		// k.Logger(ctx).Info("[TEMP] printing result from query-balances:", queryRes.Balances.String())

		var response stakingtypes.QueryDelegatorDelegationsResponse
		err := k.cdc.Unmarshal(args, &response)
		if err != nil {
			return err
		}

		// TODO get denom dynamically
		delegatorSum := sdk.NewCoin("uatom", sdk.ZeroInt())
		for _, delegation := range response.DelegationResponses {
			delegatorSum = delegatorSum.Add(delegation.Balance)
			if err != nil {
				return err
			}
		}

		// Set Redemption Rate Based On Delegation Balance vs stAsset Supply
		// TODO change local denom
		// get denom with `strided q stakeibc list-host-zone`, currently `stibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9`
		stDenom := "stibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"
		stAssetSupply := k.BankKeeper.GetSupply(ctx, stDenom)
		redemptionRate := delegatorSum.Amount.ToDec().Quo(stAssetSupply.Amount.ToDec())

		// get zone
		hz, found := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
		if found {
			fmt.Errorf("invalid chain id, zone for \"%s\" already registered", ChainId)
		}

		// set the zone
		zone := stakeibctypes.HostZone{
			ChainId:      ChainId,
			ConnectionId: msg.ConnectionId,
			LocalDenom:   hz.LocalDenom,
			BaseDenom:    hz.BaseDenom,
			// Start exchange rate at 1 upon registration
			RedemptionRate:     redemptionRate,
			LastRedemptionRate: hz.RedemptionRate, // previous redemption rate
		}
		// write the zone back to the store
		k.StakeibcKeeper.SetHostZone(ctx, zone)

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("totalDelegations", delegatorSum.String()),
				sdk.NewAttribute("stAssetSupply", stAssetSupply.Amount.String()),
				sdk.NewAttribute("redemptionRate", redemptionRate.String()),
			),
		})

		// set stakeibc:ICAAccount:delegatedBalance
		// oldICAA, found := stakeibckeeper.Keeper.GetICAAccount(ctx)
		// if !found {
		// 	k.Logger(ctx).Error("could not fetch ICAAccount for stakeIbc")
		// }
		// nTokens, err := strconv.ParseInt(strings.Replace(queryRes.Balances.String(), "uatom", "", 1), 10, 32)
		// if err != nil {
		// 	k.Logger(ctx).Error("could not cast QueryBalance result to uint32")
		// }
		// updatedICAAAccount := &stakeibctypes.ICAAccount{Address: oldICAA.Address,
		// 												Balance: oldICAA.Balance,
		// 												DelegatedBalance: int32(nTokens),
		// 												Delegations: oldICAA.Delegations}
		// stakeibckeeper.Keeper.SetICAAccount(ctx, updatedICAAAccount)

		return nil
	}

	// query_type := "cosmos.bank.v1beta1.Query/AllBalances"
	query_type := "cosmos.staking.v1beta1.Query/DelegatorDelegations"

	// balanceQuery := banktypes.QueryAllBalancesRequest{Address: msg.Address}
	// bz, err := k.cdc.Marshal(&balanceQuery)

	delegationQuery := stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: msg.Address}
	bz := k.cdc.MustMarshal(&delegationQuery)

	if err != nil {
		return nil, err
	}

	k.Keeper.MakeRequest(
		ctx,
		ConnectionId,
		ChainId,
		// pass in the target chain module and event/message to query
		// https://buf.build/cosmos/cosmos-sdk/docs/c03d23cee0a9488c835dee787f2deebb:cosmos.bank.v1beta1#cosmos.bank.v1beta1.Query.Balance
		// "cosmos.bank.v1beta1.Query/Balance",
		query_type,
		// pass in arguments to the query here
		// map[string]string{"address": msg.Address},
		bz,
		//TODO set this window to something sensible
		sdk.NewInt(25),
		types.ModuleName,
		cb,
	)
	// TODO how do we display the result here (from the target chain)
	// 		=> for now, just use ctx logging:
	// k.Logger(ctx).Info("ICQ submitted; output = ", ) //, outputFromICQ)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
			sdk.NewAttribute(types.AttributeKeyQueryId, GenerateQueryHash(ConnectionId, ChainId, query_type, bz)),
			sdk.NewAttribute(types.AttributeKeyChainId, ChainId),
			sdk.NewAttribute(types.AttributeKeyConnectionId, ConnectionId),
			sdk.NewAttribute(types.AttributeKeyType, query_type),
			sdk.NewAttribute(types.AttributeKeyHeight, "0"),
			sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(bz)),
		),
	})

	// return; usually a response object or nil
	return &types.MsgQueryBalanceResponse{}, nil
}

func (k msgServer) QueryExchangerate(goCtx context.Context, msg *types.MsgQueryExchangerate) (*types.MsgQueryExchangerateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ChainId := msg.ChainId
	// TODO Check ChainId is supported by Stride, try using this approach https://github.com/ingenuity-build/quicksilver/blob/ea71f23c6ef09a57e601f4e544c4be9693f5ba81/x/interchainstaking/keeper/msg_server.go#L37

	// Parse Address addr
	// TODO should this be Address, not Caller? changed temporarily to suppress error
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	// get zone
	hz, found := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
	if found {
		fmt.Errorf("invalid chain id, zone for \"%s\" already registered", ChainId)
	}
	ConnectionId := hz.ConnectionId

	var cb Callback = func(k Keeper, ctx sdk.Context, args []byte, query types.Query) error {
		var response stakingtypes.QueryDelegatorDelegationsResponse
		err := k.cdc.Unmarshal(args, &response)
		if err != nil {
			return err
		}

		// TODO get denom dynamically -- is it local or base denom?
		delegatorSum := sdk.NewCoin("uatom", sdk.ZeroInt())
		for _, delegation := range response.DelegationResponses {
			delegatorSum = delegatorSum.Add(delegation.Balance)
			if err != nil {
				return err
			}
		}

		// Set Redemption Rate Based On Delegation Balance vs stAsset Supply
		// TODO change local denom
		// get denom with `strided q stakeibc list-host-zone`, currently `stibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9`
		stDenom := "stibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"
		stAssetSupply := k.BankKeeper.GetSupply(ctx, stDenom)
		redemptionRate := delegatorSum.Amount.ToDec().Quo(stAssetSupply.Amount.ToDec())

		// set the zone
		zone := stakeibctypes.HostZone{
			ChainId:      ChainId,
			ConnectionId: hz.ConnectionId,
			LocalDenom:   hz.LocalDenom,
			BaseDenom:    hz.BaseDenom,
			// Start exchange rate at 1 upon registration
			RedemptionRate:     redemptionRate,
			LastRedemptionRate: hz.RedemptionRate, // previous redemption rate
		}
		// write the zone back to the store
		k.StakeibcKeeper.SetHostZone(ctx, zone)

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("totalDelegations", delegatorSum.String()),
				sdk.NewAttribute("stAssetSupply", stAssetSupply.Amount.String()),
				sdk.NewAttribute("redemptionRate", redemptionRate.String()),
			),
		})

		return nil
	}

	query_type := "cosmos.staking.v1beta1.Query/DelegatorDelegations"
	// TODO replace hardcoded addr with host zone's delegation account
	delegationQuery := stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: "cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv"}
	bz := k.cdc.MustMarshal(&delegationQuery)
	if err != nil {
		return nil, err
	}

	// query_type := "cosmos.bank.v1beta1.Query/AllBalances"
	// balanceQuery := banktypes.QueryAllBalancesRequest{Address: msg.Address}
	// bz, err := k.cdc.Marshal(&balanceQuery)

	k.Keeper.MakeRequest(
		ctx,
		ConnectionId,
		ChainId,
		query_type,
		bz,
		sdk.NewInt(25),
		types.ModuleName,
		cb,
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
			sdk.NewAttribute(types.AttributeKeyQueryId, GenerateQueryHash(ConnectionId, ChainId, query_type, bz)),
			sdk.NewAttribute(types.AttributeKeyChainId, ChainId),
			sdk.NewAttribute(types.AttributeKeyConnectionId, ConnectionId),
			sdk.NewAttribute(types.AttributeKeyType, query_type),
			sdk.NewAttribute(types.AttributeKeyHeight, "0"),
			sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(bz)),
		),
	})

	// return; usually a response object or nil
	return &types.MsgQueryExchangerateResponse{}, nil
}
