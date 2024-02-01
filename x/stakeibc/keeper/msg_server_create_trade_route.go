package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
)

var (
	DefaultMaxAllowedSwapLossRate = "0.05"
	DefaultMaxSwapAmount          = sdkmath.NewIntWithDecimal(10, 24) // 10e24
)

// Gov tx to register a trade route that swaps reward tokens for a different denom
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
//				 "max_allowed_swap_loss_rate": "0.05"
//				 "min_swap_amount": "10000000",
//				 "max_swap_amount": "1000000000"
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
	_, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenomOnReward, msg.HostDenomOnHost)
	if found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteAlreadyExists,
			"trade route already exists for rewardDenom %s, hostDenom %s", msg.RewardDenomOnReward, msg.HostDenomOnHost)
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
	tradeRouteId := types.GetTradeRouteId(msg.RewardDenomOnReward, msg.HostDenomOnHost)
	hostICA := types.ICAAccount{
		ChainId:      msg.HostChainId,
		Type:         types.ICAAccountType_WITHDRAWAL,
		ConnectionId: hostZone.ConnectionId,
		Address:      hostZone.WithdrawalIcaAddress,
	}

	unwindConnectionId := msg.StrideToRewardConnectionId
	unwindICAType := types.ICAAccountType_CONVERTER_UNWIND
	unwindICA, err := ms.Keeper.RegisterTradeRouteICAAccount(ctx, tradeRouteId, unwindConnectionId, unwindICAType)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register the unwind ICA account")
	}

	tradeConnectionId := msg.StrideToTradeConnectionId
	tradeICAType := types.ICAAccountType_CONVERTER_TRADE
	tradeICA, err := ms.Keeper.RegisterTradeRouteICAAccount(ctx, tradeRouteId, tradeConnectionId, tradeICAType)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register the trade ICA account")
	}

	// If a max allowed swap loss is not provided, use the default
	maxAllowedSwapLossRate := msg.MaxAllowedSwapLossRate
	if maxAllowedSwapLossRate == "" {
		maxAllowedSwapLossRate = DefaultMaxAllowedSwapLossRate
	}
	maxSwapAmount := msg.MaxSwapAmount
	if maxSwapAmount.IsZero() {
		maxSwapAmount = DefaultMaxSwapAmount
	}

	// Create the trade config to specify parameters needed for the swap
	tradeConfig := types.TradeConfig{
		PoolId:               msg.PoolId,
		SwapPrice:            sdk.ZeroDec(), // this should only ever be set by ICQ so initialize to blank
		PriceUpdateTimestamp: 0,

		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr(maxAllowedSwapLossRate),
		MinSwapAmount:          msg.MinSwapAmount,
		MaxSwapAmount:          maxSwapAmount,
	}

	// Finally build and store the main trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnHostZone:   msg.RewardDenomOnHost,
		RewardDenomOnRewardZone: msg.RewardDenomOnReward,
		RewardDenomOnTradeZone:  msg.RewardDenomOnTrade,
		HostDenomOnTradeZone:    msg.HostDenomOnTrade,
		HostDenomOnHostZone:     msg.HostDenomOnHost,

		HostAccount:   hostICA,
		RewardAccount: unwindICA,
		TradeAccount:  tradeICA,

		HostToRewardChannelId:  msg.HostToRewardTransferChannelId,
		RewardToTradeChannelId: msg.RewardToTradeTransferChannelId,
		TradeToHostChannelId:   msg.TradeToHostTransferChannelId,

		TradeConfig: tradeConfig,
	}

	ms.Keeper.SetTradeRoute(ctx, tradeRoute)

	return &types.MsgCreateTradeRouteResponse{}, nil
}

// Registers a new TradeRoute ICAAccount, given the type
// Stores down the connection and chainId now, and the address upon callback
func (k Keeper) RegisterTradeRouteICAAccount(
	ctx sdk.Context,
	tradeRouteId string,
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
	owner := types.FormatTradeRouteICAOwnerFromRouteId(chainId, tradeRouteId, icaAccountType)
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
