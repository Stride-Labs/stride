package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
)

// Gov tx to to register a trade route that swaps reward tokens for a different denom
//
// Example proposal:
//
//		{
//		   "title": "Create a new trade route for host chain X",
//		   "metadata": "Create a new trade route for host chain X",
//		   "summary": "Create a new trade route for host chain X",
//		   "messages":[
//		      {
//		         "@type": "/stride.stakeibc.MsgCreateTradeRoute",
//		         "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
//
//		         "stride_to_host_connection_id": "connection-0",
//				 "stride_to_reward_connection_id": "connection-1",
//			     "stride_to_trade_connection_id": "connection-2",
//
//				 "host_to_reward_transfer_channel_id": "channel-0",
//				 "reward_to_trade_transfer_channel_id": "channel-1",
//			     "trade_to_host_transfer_channel_id": "channel-2",
//
//				 "reward_denom_on_host": "ibc/rewardTokenXXX",
//				 "reward_denom_on_reward": "rewardToken",
//				 "reward_denom_on_trade": "ibc/rewardTokenYYY",
//				 "host_denom_on_trade": "ibc/hostTokenZZZ",
//				 "host_denom_on_host": "hostToken",
//
//				 "pool_id": 1,
//				 "min_swap_amount": 10000000,
//				 "max_swap_amount": 1000000000
//			  }
//		   ],
//		   "deposit": "2000000000ustrd"
//	   }
//
// >>> strided tx gov submit-proposal {proposal_file.json} --from wallet
func (ms msgServer) CreateTradeRoute(goCtx context.Context, msg *types.MsgCreateTradeRoute) (*types.MsgCreateTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	// Validate trade route does not already exist for this denom
	_, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenomOnHost, msg.HostDenomOnHost)
	if found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteAlreadyExists,
			"startDenom: %s, endDenom: %s", msg.RewardDenomOnHost, msg.HostDenomOnHost)
	}

	// Confirm the host chain exists and the withdrawal address has been initialized
	hostZone, err := ms.Keeper.GetActiveHostZone(ctx, msg.HostChainId)
	if err != nil {
		return nil, err
	}
	if hostZone.WithdrawalIcaAddress == "" {
		return nil, errorsmod.Wrapf(types.ErrICAAccountNotFound, "withdrawal account not initialized on host zone")
	}

	// Register the new ICA accounts
	hostICA := types.ICAAccount{
		ChainId:      msg.HostChainId,
		Type:         types.ICAAccountType_WITHDRAWAL,
		ConnectionId: hostZone.ConnectionId,
		Address:      hostZone.WithdrawalIcaAddress,
	}
	rewardICA, err := ms.Keeper.RegisterTradeRouteICAAccount(ctx, msg.StrideToRewardConnectionId, types.ICAAccountType_UNWIND)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register the unwind ICA account")
	}
	tradeICA, err := ms.Keeper.RegisterTradeRouteICAAccount(ctx, msg.StrideToTradeConnectionId, types.ICAAccountType_TRADE)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register the trade ICA account")
	}

	// Create the hops between each zone
	hostToRewardHop := types.TradeHop{
		TransferChannelId: msg.HostToRewardTransferChannelId,
		FromAccount:       hostICA,
		ToAccount:         rewardICA,
	}

	rewardToTradeHop := types.TradeHop{
		TransferChannelId: msg.RewardToTradeTransferChannelId,
		FromAccount:       rewardICA,
		ToAccount:         tradeICA,
	}

	tradeToHostHop := types.TradeHop{
		TransferChannelId: msg.TradeToHostTransferChannelId,
		FromAccount:       tradeICA,
		ToAccount:         hostICA,
	}

	// Finally build and store the main trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnHostZone:   msg.RewardDenomOnHost,
		RewardDenomOnRewardZone: msg.RewardDenomOnReward,
		RewardDenomOnTradeZone:  msg.RewardDenomOnTrade,
		TargetDenomOnTradeZone:  msg.HostDenomOnTrade,
		TargetDenomOnHostZone:   msg.HostDenomOnHost,
		HostToRewardHop:         hostToRewardHop,
		RewardToTradeHop:        rewardToTradeHop,
		TradeToHostHop:          tradeToHostHop,
		PoolId:                  msg.PoolId,
		SpotPrice:               "", // this should only ever be set by ICQ so initialize to blank
		MinSwapAmount:           sdkmath.NewIntFromUint64(msg.MinSwapAmount),
		MaxSwapAmount:           sdkmath.NewIntFromUint64(msg.MaxSwapAmount),
	}

	ms.Keeper.SetTradeRoute(ctx, tradeRoute)

	return &types.MsgCreateTradeRouteResponse{}, nil
}

func (k Keeper) RegisterTradeRouteICAAccount(
	ctx sdk.Context,
	connectionId string,
	icaAccountType types.ICAAccountType,
) (account types.ICAAccount, err error) {
	// Get the chain ID and counterparty connection-id from the connection ID on Stride
	chainId, err := k.GetChainIdFromConnectionId(ctx, connectionId)
	if err != nil {
		return account, err
	}
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionId)
	if !found {
		return account, errorsmod.Wrap(connectiontypes.ErrConnectionNotFound, connectionId)
	}
	counterpartyConnectionId := connection.Counterparty.ConnectionId

	// Build the appVersion, owner, and portId needed for registration
	appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: connectionId,
		HostConnectionId:       counterpartyConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
	owner := types.FormatICAAccountOwner(chainId, icaAccountType)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return account, err
	}

	// Create the associate ICAAccount object
	account = types.ICAAccount{
		ChainId:      chainId,
		Type:         icaAccountType,
		ConnectionId: connectionId,
	}

	// Check if an ICA account has already been created
	// (in the event that this trade route was removed and then added back)
	// If so, there's no need to register a new ICA
	_, channelFound := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, connectionId, portID)
	icaAddress, icaFound := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, connectionId, portID)
	if channelFound && icaFound {
		account = types.ICAAccount{
			ChainId:      chainId,
			Type:         icaAccountType,
			ConnectionId: connectionId,
			Address:      icaAddress,
		}
		return account, nil
	}

	// Otherwise, if there's no account already, register a new one
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, connectionId, owner, appVersion); err != nil {
		return account, err
	}

	return account, nil
}
