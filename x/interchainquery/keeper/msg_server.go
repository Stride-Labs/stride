package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

func (k msgServer) QueryBalance(goCtx context.Context, msg *types.MsgQueryBalance) (*types.MsgQueryBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO(TEST-84) check:
	//   -  host chain is supported
	//   -  address on target chain is valid
	//   -  denom is valid
	//   -  caller addr is valid
	ChainId := msg.ChainId

	// Parse Address addr; TODO(NOW) should this be Address, not Caller? changed temporarily to suppress error
	_, err := sdk.AccAddressFromBech32(msg.Caller)
	if err != nil {
		panic(err)
	}
	ConnectionId := msg.ConnectionId

	var cb Callback = func(k Keeper, ctx sdk.Context, args []byte, query types.Query) error {

		var response stakingtypes.QueryDelegatorDelegationsResponse
		err := k.cdc.Unmarshal(args, &response)
		if err != nil {
			return err
		}

		// TOD(TEST-85) get denom dynamically
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
			ChainId:            ChainId,
			ConnectionId:       msg.ConnectionId,
			LocalDenom:         hz.LocalDenom,
			BaseDenom:          hz.BaseDenom,
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
	delegationQuery := stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: msg.Address}
	bz := k.cdc.MustMarshal(&delegationQuery)

	if err != nil {
		return nil, err
	}

	k.Keeper.MakeRequest(
		ctx,
		ConnectionId,
		ChainId,
		query_type,
		bz,
		// TODO(TEST-79) understand and use proper period
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
	return &types.MsgQueryBalanceResponse{}, nil
}

// TODO(TEST-78) rename from query-XXX => update-XXX
func (k msgServer) QueryExchangerate(goCtx context.Context, msg *types.MsgQueryExchangerate) (*types.MsgQueryExchangerateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ChainId := msg.ChainId
	// TODO Check ChainId is supported by Stride, try using this approach https://github.com/ingenuity-build/quicksilver/blob/ea71f23c6ef09a57e601f4e544c4be9693f5ba81/x/interchainstaking/keeper/msg_server.go#L37

	// Parse Address addr
	// TODO(NOW) should this be Address, not Caller? changed temporarily to suppress error
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

		// TODO(TEST-85) set denom dynamically
		delegatorSum := sdk.NewCoin("uatom", sdk.ZeroInt())
		for _, delegation := range response.DelegationResponses {
			delegatorSum = delegatorSum.Add(delegation.Balance)
			if err != nil {
				return err
			}
		}

		// Set Redemption Rate Based On Delegation Balance vs stAsset Supply
		// TODO(TEST-85) set denom dynamically
		// get denom with `strided q stakeibc list-host-zone`, currently `stibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9`
		stDenom := "stibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"
		stAssetSupply := k.BankKeeper.GetSupply(ctx, stDenom)
		redemptionRate := delegatorSum.Amount.ToDec().Quo(stAssetSupply.Amount.ToDec())

		// set the zone
		zone := stakeibctypes.HostZone{
			ChainId:            ChainId,
			ConnectionId:       hz.ConnectionId,
			LocalDenom:         hz.LocalDenom,
			BaseDenom:          hz.BaseDenom,
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
	// TODO(TEST-86) get addr dynamically (delegationAddress)
	delegationQuery := stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: "cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv"}
	bz := k.cdc.MustMarshal(&delegationQuery)
	if err != nil {
		return nil, err
	}

	k.Keeper.MakeRequest(
		ctx,
		ConnectionId,
		ChainId,
		query_type,
		bz,
		// TODO(TEST-79) understand and use proper period
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
			// TODO(TEST-79) understand height
			sdk.NewAttribute(types.AttributeKeyHeight, "0"),
			sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(bz)),
		),
	})

	// return; usually a response object or nil
	return &types.MsgQueryExchangerateResponse{}, nil
}

// TODO(TEST-78) rename from query-XXX => update-XXX
func (k msgServer) QueryDelegatedbalance(goCtx context.Context, msg *types.MsgQueryDelegatedbalance) (*types.MsgQueryDelegatedbalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ChainId := msg.ChainId

	// Parse Address addr
	// TODO(NOW) should this be Address, not Caller
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
		queryRes := banktypes.QueryAllBalancesResponse{}
		err := k.cdc.Unmarshal(args, &queryRes)
		if err != nil {
			k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
			return err
		}
		// TODO get denom dynamically
		balance := int32(queryRes.Balances.AmountOf("uatom").Int64())

		// Set delegation account balance to ICQ result
		hz, found := k.StakeibcKeeper.GetHostZone(ctx, ChainId)
		if found {
			fmt.Errorf("invalid chain id, zone for \"%s\" already registered", ChainId)
		}

		da := hz.DelegationAccount
		delegationAccount := stakeibctypes.ICAAccount{Address: da.Address,
			Balance:          balance, // <== updated
			DelegatedBalance: da.DelegatedBalance,
			Delegations:      da.Delegations,
			Target:           da.Target,
		}

		// set the zone
		zone := stakeibctypes.HostZone{
			ChainId:            hz.ChainId,
			ConnectionId:       hz.ConnectionId,
			LocalDenom:         hz.LocalDenom,
			BaseDenom:          hz.BaseDenom,
			DelegationAccount:  &delegationAccount, // <== updated
			RedemptionRate:     hz.RedemptionRate,
			LastRedemptionRate: hz.LastRedemptionRate,
		}
		// write the zone back to the store
		k.StakeibcKeeper.SetHostZone(ctx, zone)

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("totalBalance", string(balance)),
			),
		})

		return nil
	}

	query_type := "cosmos.bank.v1beta1.Query/AllBalances"
	// TODO(NOW) replace hardcoded addr with host zone's delegation account
	balanceQuery := banktypes.QueryAllBalancesRequest{Address: "cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv"}
	bz, err := k.cdc.Marshal(&balanceQuery)
	if err != nil {
		return nil, err
	}

	k.Keeper.MakeRequest(
		ctx,
		ConnectionId,
		ChainId,
		query_type,
		bz,
		// TODO(TEST-79) understand and use proper period
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

	return &types.MsgQueryDelegatedbalanceResponse{}, nil
}
