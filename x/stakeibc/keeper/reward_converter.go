package keeper

import (
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v14/utils"
	epochstypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"

	osmosistypes "github.com/osmosis-labs/osmosis/x/gamm/types"
)

// The goal of this code is to allow certain reward token types to be automatically traded into other types
// This happens before the rest of the staking, allocation, distribution etc. would continue as normal

// Reward tokens are any special denoms which are paid out in the withdrawal address
// Most host zones inflate their tokens and the newly minted host denom is what appears in the withdrawal ICA
// This code allows for chains to use foreign denoms as revenue, which can be traded to any other denom first

// 0. Before the normal epochly base denom withdrawal address check in the normal staking flow
//		by doing this check before the normal staking flow, we could trade tokens which are stake-able instead
// 1. Epochly check the reward denom balance in the withdrawal address
//       on callback, send all this reward denom from withdrawl ICA to trade ICA on the trade zone (OSMOSIS)
// 2. Epochly check the reward denom balance in trade ICA
//		on callback, trade all reward denom for target output denom defined by pool and routes in params
// 3. Epochly check the target denom balance in trade ICA
//		on callback, transfer these target denoms from trade ICA to withdrawal ICA on original host zone
// Normal staking flow continues from there. So if the target denom is the original host zone base denom
// as will often be the case, then these tokens will follow the normal staking and distribution flow.

// Normal IBC transfers can't always be used, pfm transfers might be needed to unwind through the reward zone
// even if the converted denom is hostZone host denom, direct routes contain channel info between host and trade zones
// (going through Stride would mess up the IBC denoms further, the host zone ICA can kick off ibc/pfm transfer)

var (
	// Timeout for the IBC transfer of the Tokens across zones, sometimes with pfm for 2 hops
	TransferTimeout = time.Hour * 24 // 1 day
)

// Helper to load all trade-able reward token denoms based on what TransferRoutes from hostZone -> tradeZone are defined
// important: these are the ibc denoms as they exist on the hostZone, not the tradeZone
// these are the denoms for which the balance should be checked in the QueryWithdrawalRewardBalance ICQs
func (k Keeper) GetHostZoneTradeableDenoms(ctx sdk.Context, hostZoneID string) (denoms []string) {
	params := k.GetParams(ctx)
	tradeRoutes := params.TradeRoutes
	for _, route := range tradeRoutes.TransferRoutes {
		if route.StartChainId == hostZoneID {
			denoms = append(denoms, route.StartDenom)
		} 
	}
	return denoms
}

// Helper to load all trade-able reward token denoms based on what TradePools are defined
// Important: these are the ibc denoms as they exist on the tradeZone, not the hostZone
// these are the denoms for which the balance should be checked in the QueryTradeRewardBalance ICQs
func (k Keeper) GetTradeZoneTradeableDenoms(ctx sdk.Context) (denoms []string) {
	params := k.GetParams(ctx)
	tradeRoutes := params.TradeRoutes
	for _, pool := range tradeRoutes.TradePools {
		denoms = append(denoms, pool.InputDenom)
	}
	return denoms
}

// Helper to load all converted token denoms based on what TransferRoutes from tradeZone -> hostZone are defined
// Important: these are the ibc denoms as they exist on the tradeZone, not the hostZone
// these are the denoms for which the balance should be checked in the QueryTradeConvertedBalance ICQs
func (k Keeper) GetTradeZoneConvertedDenoms(ctx sdk.Context, tradeZoneID string) (denoms []string) {
	return k.GetHostZoneTradeableDenoms(ctx, tradeZoneID)
}

// Helper to get which tradeZone this denom on this hostZone can be found to have pools
// Important: these are the ibc denoms as they exist on the hostZone, not the tradeZone
func (k Keeper) GetTradeZoneID(ctx sdk.Context, denomOnHostZone string, hostZoneID string) (tradeZoneID string) {
	params := k.GetParams(ctx)
	tradeRoutes := params.TradeRoutes
	for _, route := range tradeRoutes.TransferRoutes {
		if route.StartDenom == denomOnHostZone && 
			route.StartChainId == hostZoneID {
				return route.FinishChainId
			} 
	}
	return "osmosis-1" // default tradezone, still needs to be defined in the hostZones
}

// Helper to get the tradePool for this input denom on tradeZone originally coming from this hostZone 
// Important: these are the ibc denoms as they exist on the tradeZone, not the hostZone
func (k Keeper) GetTradePool(ctx sdk.Context, rewardDenomOnTradeZone string, hostZoneID string) (types.TradePool, error) {
	params := k.GetParams(ctx)
	tradeRoutes := params.TradeRoutes
	for _, pool := range tradeRoutes.TradePools {
		if pool.InputDenom == rewardDenomOnTradeZone && 
			pool.HostChainId == hostZoneID {
				return pool, nil
			} 
	}
	return types.TradePool{}, errors.New("No trade pool found defined in the params for this source and denom")
}

// Helper to locate the route if it exists in the params based on start denom and the start and finish chains
func (k Keeper) GetTransferRoute(ctx sdk.Context, startDenom string, startChainID string, finishChainID string) (types.TransferRoute, error) {
	params := k.GetParams(ctx)
	tradeRoutes := params.TradeRoutes

	for _, route := range tradeRoutes.TransferRoutes {
		if route.StartDenom == startDenom && 
			route.StartChainId == startChainID && 
				route.FinishChainId == finishChainID {
					return route, nil
		}
	}
	return types.TransferRoute{}, errors.New("No route was found defined in the params for this transfer")
}

// Helper function with internal logic to determine if denoms are foreign ibc denoms and PFM unwinding is needed
// Only use when both startZone and finishZone are not strideZone itself and therefore these addresses are ICAs
func (k Keeper) TransferTokensFromStartZonetoFinishZone(ctx sdk.Context, 
					amount sdk.Int, 
					startZoneDenom string, 
					startZone types.HostZone, 
					startZoneAddress string, 
					finishZone types.HostZone, 
					finishZoneAddress string) error {
	// Get the route if it exists
	route, routeErr := k.GetTransferRoute(ctx, startZoneDenom, startZone.ChainId, finishZone.ChainId)
	if routeErr != nil {
		return routeErr
	}

	// Check if this is a single hop direct ibc transfer or a two-hop ibc unwinding transfer
	if route.MiddleDenom == "" && route.MiddleChainId == "" {
		// found a direct route, use info to kick off standard ibc transfer (from the startZoneAddress ICA)
		sendTokens := sdk.NewCoin(startZoneDenom, amount)
		timeout := uint64(ctx.BlockTime().UnixNano() + (TransferTimeout).Nanoseconds())

		var msgs []proto.Message
		msgs = append(msgs, &transfertypes.MsgTransfer{
			SourcePort:       transfertypes.PortID, //route.StartTransferPort if need custom port
			SourceChannel:    route.StartTransferChannel,
			Token:            sendTokens,
			Sender:           startZoneAddress,
			Receiver:         finishZoneAddress,
			TimeoutTimestamp: timeout,
		})					
		k.Logger(ctx).Info(utils.LogWithHostZone(startZone.ChainId,
			"Preparing MsgTransfer of %v from %s to %s", sendTokens.String(), startZone.ChainId, finishZone.ChainId))
	
		// Send the transaction through SubmitTxsStrideEpoch to kick off transfer from the startZone to the finishZone
		// By making the callbackId = "", no callback data will be needed and callbackArgs will be ignored
		callbackId := ""
		var callbackArgs []byte;
		_, err := k.SubmitTxsStrideEpoch(ctx, startZone.ConnectionId, msgs, types.ICAAccountType_WITHDRAWAL, callbackId, callbackArgs)
		if err != nil {
			return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
		}

		return nil

	} else {
		// found an unwinding route, use info to kick off standard ibc transfer (from the startZoneAddress ICA)
		// and then include the relevant memo for pfm to perform the chained transfer to the finishZoneAddress
		sendTokens := sdk.NewCoin(startZoneDenom, amount)
		timeout := uint64(ctx.BlockTime().UnixNano() + (TransferTimeout).Nanoseconds())

		memoJSON := fmt.Sprintf(`"forward": {"receiver": "%s","port": "%s","channel":"%s","timeout":"10m","retries": 2}`,
			finishZoneAddress, route.FinishTransferPort, route.FinishTransferChannel)

		var msgs []proto.Message
		msgs = append(msgs, &transfertypes.MsgTransfer{
			SourcePort:       transfertypes.PortID, //route.TransferRoute.TransferPort if need custom port
			SourceChannel:    route.StartTransferChannel,
			Token:            sendTokens,
			Sender:           startZoneAddress,
			Receiver:         route.PassthroughIcaAddress, // could be "pfm" or a real address depending on version
			TimeoutTimestamp: timeout,
			Memo:             memoJSON,	
		})					
		k.Logger(ctx).Info(utils.LogWithHostZone(startZone.ChainId,
			"Preparing MsgTransfer of %v from %s to %s", sendTokens.String(), startZone.ChainId, finishZone.ChainId))
	
		// Send the transaction through SubmitTxsStrideEpoch to kick off transfer from the startZone to the finishZone
		// By making the callbackId = "", no callback data will be needed and callbackArgs will be ignored
		callbackId := ""
		var callbackArgs []byte
		_, err := k.SubmitTxsStrideEpoch(ctx, startZone.ConnectionId, msgs, types.ICAAccountType_WITHDRAWAL, callbackId, callbackArgs)
		if err != nil {
			return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
		}

		return nil
	}
}


// Transfer the reward tokens from the hostZone withdrawl ICA to be converted on the tradeZone
func (k Keeper) TransferHostRewardTokensToTrade(ctx sdk.Context, amount sdk.Int, rewardDenom string, hostZone types.HostZone, tradeZone types.HostZone) error {
	return k.TransferTokensFromStartZonetoFinishZone(ctx, 
		amount, 
		rewardDenom,
		hostZone, 
		hostZone.WithdrawalIcaAddress, 
		tradeZone, 
		hostZone.RewardTradeIcaAddress) // this ICA exists on the tradeZone but stored per hostZone
}

// Transfer the converted tokens back from tradeZone to the hostZone withdrawal ICA
func (k Keeper) TransferTradeConvertedTokensToHost(ctx sdk.Context, amount sdk.Int, convertedDenom string, hostZone types.HostZone, tradeZone types.HostZone) error {
	return k.TransferTokensFromStartZonetoFinishZone(ctx, 
		amount, 
		convertedDenom, 
		tradeZone, 
		hostZone.RewardTradeIcaAddress, // this ICA exists on the tradeZone but stored per hostZone
		hostZone, 
		hostZone.WithdrawalIcaAddress)
}

// Trade all the reward tokens in the Trade ICA for the target output token type using ICA remote tx on trade zone
// Params define the inputs, outputs, routes, and pool information on the trade zone for each type of token
func (k Keeper) TradeRewardTokens(ctx sdk.Context, 
					amount sdk.Int,
					rewardDenomOnTradeZone string,
					hostZone types.HostZone, 
					tradeZone types.HostZone) error {
	// Load the tradepool info if it exists
	tradePool, poolErr := k.GetTradePool(ctx, rewardDenomOnTradeZone, hostZone.ChainId)
	if poolErr != nil {
		return poolErr
	}
	tradeTokens := sdk.NewCoin(rewardDenomOnTradeZone, amount)
	// Prepare Osmosis GAMM module MsgSwapExactAmountIn from the trade account to perform the trade
	// If we want to generalize in the future, write swap message generation funcs for each DEX type, 
	// decide which msg generation function to call based on check of which tradeZone was passed in
	var msgs []proto.Message
	if amount.GT(sdk.ZeroInt()) {
		var routes []osmosistypes.SwapAmountInRoute
		routes = append(routes, osmosistypes.SwapAmountInRoute{
			PoolId: tradePool.PoolId,
			TokenOutDenom: tradePool.OutputDenom,
		});
		msgs = append(msgs, &osmosistypes.MsgSwapExactAmountIn{
			Sender: hostZone.RewardTradeIcaAddress, // address exists on trade zone but field stored on hostZone
			Routes: routes,
			TokenIn: tradeTokens,
			TokenOutMinAmount: sdk.ZeroInt(),
		})
		k.Logger(ctx).Info(utils.LogWithHostZone(tradeZone.ChainId,
			"Preparing MsgSwapExactAmountIn of %v from the trade account", tradeTokens.String()))
	}

	// Send the transaction through SubmitTxsStrideEpoch
	// By making the callbackId = "", no callback data will be needed and callbackArgs will be ignored
	callbackId := ""
	var callbackArgs []byte
	_, err := k.SubmitTxsStrideEpoch(ctx, tradeZone.ConnectionId, msgs, types.ICAAccountType_TRADE, callbackId, callbackArgs)
	if err != nil {
		return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	}

	return nil
}


// ICQ calls for remote ICA balances
// There is a single trade zone (hardcoded as Osmosis for now but maybe additional DEXes allowed in the future)
// We have to initialize a single hostZone object for the trade zone once in initialization and then it can be used in all these calls

// Kick off ICQ for the reward denom balance in the withdrawal address
func (k Keeper) WithdrawalRewardBalanceQuery(ctx sdk.Context, hostZone types.HostZone, hostZoneRewardDenom string) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for reward denom in withdrawal account"))

	// Get the withdrawal account address from the host zone
	if hostZone.WithdrawalIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no withdrawal account found for %s", hostZone.ChainId)
	}

	// Encode the withdrawal account address for the query request
	// The query request consists of the withdrawal account address and reward denom
	_, withdrawalAddressBz, err := bech32.DecodeAndConvert(hostZone.WithdrawalIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid withdrawal account address, could not decode (%s)", err.Error())
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(withdrawalAddressBz), []byte(hostZoneRewardDenom)...)

	// Timeout query at end of epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeout := time.Unix(0, int64(strideEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// Marshalling callback data to pass into query
	callbackData := types.WithdrawalRewardBalanceQueryCallback{
		RewardDenomOnHostZone: hostZoneRewardDenom,
		HostZoneId: hostZone.ChainId,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal WithdrawalRewardBalanceQuery callback data")
	}


	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_WithdrawalRewardBalance,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying for withdrawal reward denom balance, error: %s", err.Error()))
		return err
	}

	return nil
}

// Kick off ICQ for how many reward tokens are in the trade ICA associated with this host zone
func (k Keeper) TradeRewardBalanceQuery(ctx sdk.Context, hostZone types.HostZone, tradeZone types.HostZone, tradeZoneRewardDenom string) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for reward denom in trade ICA account"))

	// Get the trade account address from the host zone
	if hostZone.RewardTradeIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no trade account found for %s", hostZone.ChainId)
	}

	// Encode the trade account address for the query request
	// The query request consists of the trade account address and reward denom
	// keep in mind this ICA address actually exists on trade zone but is associated with trades performed for host zone
	_, tradeAddressBz, err := bech32.DecodeAndConvert(hostZone.RewardTradeIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid trade account address, could not decode (%s)", err.Error())
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(tradeAddressBz), []byte(tradeZoneRewardDenom)...)

	// Timeout query at end of epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeout := time.Unix(0, int64(strideEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// Marshalling callback data to pass into query
	callbackData := types.TradeRewardBalanceQueryCallback{
		RewardDenomOnTradeZone: tradeZoneRewardDenom,
		HostZoneId: hostZone.ChainId,
		TradeZoneId: tradeZone.ChainId,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal TradeRewardBalanceQuery callback data")
	}


	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         tradeZone.ChainId,
		ConnectionId:    tradeZone.ConnectionId, // query needs to go to the trade zone, not the host zone
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_TradeRewardBalance,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying trade ICA for reward denom balance, error: %s", err.Error()))
		return err
	}

	return nil
}

// Kick off ICQ for how many converted tokens are in the trade ICA associated with this host zone
func (k Keeper) TradeConvertedBalanceQuery(ctx sdk.Context, hostZone types.HostZone, tradeZone types.HostZone, tradeZoneConvertedDenom string) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for converted denom in trade ICA account"))

	// Get the trade account address from the host zone
	if hostZone.RewardTradeIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no trade account found for %s", hostZone.ChainId)
	}

	// Encode the trade account address for the query request
	// The query request consists of the trade account address and converted denom
	// keep in mind this ICA address actually exists on trade zone but is associated with trades performed for host zone
	_, tradeAddressBz, err := bech32.DecodeAndConvert(hostZone.RewardTradeIcaAddress) 
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid trade account address, could not decode (%s)", err.Error())
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(tradeAddressBz), []byte(tradeZoneConvertedDenom)...)

	// Timeout query at end of epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeout := time.Unix(0, int64(strideEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// Marshalling callback data to pass into query
	callbackData := types.TradeConvertedBalanceQueryCallback{
		ConvertedDenomOnTradeZone: tradeZoneConvertedDenom,
		HostZoneId: hostZone.ChainId,
		TradeZoneId: tradeZone.ChainId,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal WithdrawalRewardBalanceQuery callback data")
	}


	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         tradeZone.ChainId,
		ConnectionId:    tradeZone.ConnectionId, // query needs to go to the trade zone, not the host zone
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_TradeConvertedBalance,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying trade ICA for converted denom balance, error: %s", err.Error()))
		return err
	}

	return nil
}


// Final helper function to be run epochly, kicks off all queries on all token types which match that ICA
func (k Keeper) ConvertAllRewards(ctx sdk.Context) {
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		hostZoneTradableDenoms := k.GetHostZoneTradeableDenoms(ctx, hostZone.ChainId)
		if len(hostZoneTradableDenoms) == 0 {
			continue
		}
		for _, hostZoneRewardDenom := range hostZoneTradableDenoms {
			k.WithdrawalRewardBalanceQuery(ctx, hostZone, hostZoneRewardDenom)
		}

		tradeZoneID := k.GetTradeZoneID(ctx, hostZoneTradableDenoms[0], hostZone.ChainId)
		tradeZone, _ := k.GetHostZone(ctx, tradeZoneID)

		tradeZoneRewardDenoms := k.GetTradeZoneTradeableDenoms(ctx)
		for _, tradeZoneRewardDenom := range tradeZoneRewardDenoms {
			k.TradeRewardBalanceQuery(ctx, hostZone, tradeZone, tradeZoneRewardDenom)
		}

		tradeZoneConvertedDenoms := k.GetTradeZoneConvertedDenoms(ctx, tradeZone.ChainId)
		for _, tradeZoneConvertedDenom := range tradeZoneConvertedDenoms {
			k.TradeConvertedBalanceQuery(ctx, hostZone, tradeZone, tradeZoneConvertedDenom)
		}		
	}
}
