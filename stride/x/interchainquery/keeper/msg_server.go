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
					k.Logger(ctx).Error("Error in callback", "error", err, "msg", msg.QueryId, "result", msg.Result, "type", q.QueryType, "params", q.QueryParameters)
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
	// TODO
	var cb Callback = func(k Keeper, ctx sdk.Context, args []byte, query types.Query) error {
		k.Logger(ctx).Info("[TEMP] printing inside the querybalance callback")
		return nil
	}

	// var cb Callback = func(k Keeper, ctx sdk.Context, args []byte, query types.Query) error {
	// 	return k.SetAccountBalance(ctx, zone, query.QueryParameters["address"], args)
	// }

	k.Keeper.MakeRequest(
		ctx,
		ConnectionId,
		ChainId,
		// pass in the target chain module and event/message to query
		// https://buf.build/cosmos/cosmos-sdk/docs/c03d23cee0a9488c835dee787f2deebb:cosmos.bank.v1beta1#cosmos.bank.v1beta1.Query.Balance
		// "cosmos.bank.v1beta1.Query/Balance",
		"cosmos.bank.v1beta1.Query/AllBalances",
		// pass in arguments to the query here
		// map[string]string{"address": msg.Address},
		// TODO revert from hardcode
		map[string]string{"address": "cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv"},
		//TODO set this to something sensible
		sdk.NewInt(25),
		types.ModuleName,
		cb,
	)

	// var packet types.CreatePairPacketData

	// packet.SourceDenom = msg.SourceDenom
	// packet.TargetDenom = msg.TargetDenom

	// // Transmit the packet
	// err := k.TransmitCreatePairPacket(
	// 	ctx,
	// 	packet,
	// 	msg.Port,
	// 	msg.ChannelID,
	// 	clienttypes.ZeroHeight(),
	// 	msg.TimeoutTimestamp,
	// )
	// if err != nil {
	// 	return nil, err
	// }

	// TODO how do we display the result here (from the target chain)
	// 		=> for now, just use ctx logging:
	k.Logger(ctx).Info("ICQ submitted; output = ") //, outputFromICQ)

	// return; usually a response object or nil
	return &types.MsgQueryBalanceResponse{}, nil
}

// func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	// Init variables
// 	// get the sender address
// 	sender, err := sdk.AccAddressFromBech32(msg.Creator)
// 	if err != nil {
// 		k.Logger(ctx).Info("Invalid address")
// 		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "address invalid")
// 	}
// 	// get the coins to send, they need to be in the format {amount}{denom}
// 	// NOTE: int is an int32 or int64 (depending on machine type) so converting from int32 -> int
// 	// is safe. The converse is not true.
// 	coinString := strconv.Itoa(int(msg.Amount)) + msg.Denom
// 	coins, err := sdk.ParseCoinsNormalized(coinString)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "failed to parse %s coins", coins)
// 	}

// 	// Safety checks
// 	// ensure Amount is non-negative, liquid staking 0 or less tokens is invalid
// 	if msg.Amount < 1 {
// 		k.Logger(ctx).Info("amount must be non-negative")
// 		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be non-negative")
// 	}
// 	// check that the token is an IBC token
// 	isIbcToken := types.IsIBCToken(msg.Denom)
// 	if !isIbcToken {
// 		k.Logger(ctx).Info("invalid token denom")
// 		return nil, sdkerrors.Wrapf(types.ErrInvalidToken, "invalid token denom (%s)", msg.Denom)
// 	}

// 	// deposit `amount` of `denom` token to the stakeibc module
// 	// NOTE: Should we add an additional check here? This is a pretty important line of code
// 	// NOTE: If sender doesn't have enough coins, this panics (error is hard to interpret)
// 	sdkerror := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins)
// 	if sdkerror != nil {
// 		k.Logger(ctx).Error("failed to send tokens from Account to Module")
// 		panic(sdkerror)
// 	}

// 	// mint user `amount` of the corresponding stAsset
// 	// NOTE: We should ensure that denoms are unique - we don't want anyone spoofing denoms
// 	err = k.MintStAsset(ctx, sender, msg.Amount, msg.Denom)
// 	if err != nil {
// 		k.Logger(ctx).Info("failed to send tokens from Account to Module")
// 		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "failed to mint stAssets to user")
// 	}

// 	return &types.MsgLiquidStakeResponse{}, nil
// }
