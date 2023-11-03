package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v14/utils"
	epochstypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// For each hostZone with a valid community pool, trigger the ICQs and ICAs to transfer tokens from DepositICA or back to ReturnICA
// Since ICQs and ICAs take time to complete, it is almost certain tokens swept in and processed will be swept out in a later epoch
func (k Keeper) ProcessAllCommunityPoolTokens(ctx sdk.Context) {
	hostZones := k.GetAllActiveHostZone(ctx)
	for _, hostZone := range hostZones {
		if hostZone.CommunityPoolDepositIcaAddress == "" ||
			hostZone.CommunityPoolHoldingAddress == "" ||
			hostZone.CommunityPoolReturnIcaAddress == "" {
			continue
		}

		// stDenom is the ibc denom on hostZone when the hostZone's native denom is staked
		denom := hostZone.HostDenom
		stDenom := k.GetStakedDenomOnHostZone(ctx, hostZone)

		/****** Epoch 1 *******/	
		// ICQ for the host denom of the chain, these are tokens the pool wants staked
		err := k.QueryCommunityPoolBalance(ctx, hostZone, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, denom)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId, "Querying hostDenom %s in deposit- %s", denom, err.Error()))
		}
		// ICQ for staked tokens of the host denom, these are tokens the pool wants redeemed
		err = k.QueryCommunityPoolBalance(ctx, hostZone,types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, stDenom)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId, "Querying stHostDenom %s in deposit - %s", stDenom, err.Error()))
		}

		/****** Epoch 2 *******/
		// LiquidStake tokens of denom, RedeemStake tokens of stDenom before transfer to return ica
		err = k.TransformTokensInHolding(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId, "Transforming tokens in holding address - %s", err.Error()))
		}

		/****** Epoch 3 *******/	
		// ICQ for and denom or stDenom tokens in return ICA and call FundCommunityPool 
		err = k.QueryCommunityPoolBalance(ctx, hostZone, types.ICAAccountType_COMMUNITY_POOL_RETURN, denom)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId, "Querying hostDenom %s in return- %s", denom, err.Error()))
		}
		err = k.QueryCommunityPoolBalance(ctx, hostZone,types.ICAAccountType_COMMUNITY_POOL_RETURN, stDenom)
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId, "Querying stHostDenom %s in return - %s", stDenom, err.Error()))
		}
	}
}

// ICQ specific denom for balance in the deposit ICA or return ICA on the community pool host zone
// Depending on account type and denom, discovered tokens are transferred to Stride or funded to the pool
func (k Keeper) QueryCommunityPoolBalance(ctx sdk.Context, 
											hostZone types.HostZone, 
											icaType types.ICAAccountType, 
											denom string) error {

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, 
		"Building ICQ for %s balance in community pool %s address", denom, icaType.String()))

	icaAddress := ""
	if icaType == types.ICAAccountType_COMMUNITY_POOL_DEPOSIT {
		icaAddress = hostZone.CommunityPoolDepositIcaAddress
	} else if icaType == types.ICAAccountType_COMMUNITY_POOL_RETURN {
		icaAddress = hostZone.CommunityPoolReturnIcaAddress
	} else {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "icaType must be either deposit or return!")		
	}

	// Verify a valid ica address exists for this host zone
	if icaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no address of type %s found for %s",
			icaType.String(), hostZone.ChainId)
	}

	_, addressBz, err := bech32.DecodeAndConvert(icaAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid %s address, could not decode (%s)", 
			icaType.String(), hostZone.CommunityPoolDepositIcaAddress)
	}
	queryData := append(banktypes.CreateAccountBalancesPrefix(addressBz), []byte(denom)...)
	
	// The response might be a coin, or might just be an int depending on sdk version
	// Since we need the denom later, store the denom as callback data for the query
	callbackData := types.CommunityPoolBalanceQueryCallback{
		IcaType: icaType,
		Denom: denom,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "can't marshal community pool balance callback data %+v", callbackData)
	}

	// Timeout query at end of epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	timeout := time.Unix(0, int64(strideEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_CommunityPoolBalance,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return errorsmod.Wrapf(err, "Error submitting query for pool ica balance")
	}

	return nil
}

func (k Keeper) TransformTokensInHolding(ctx sdk.Context, hostZone types.HostZone) error {
	// Use bankKeeper to see all coin types and amounts currently in the stride-side holding module address
	req := banktypes.NewQueryAllBalancesRequest(sdk.AccAddress(hostZone.CommunityPoolHoldingAddress), nil)
	resp, err := k.bankKeeper.AllBalances(ctx, req)
	if err != nil {
		return errorsmod.Wrapf(err, "Couldn't look up balances in holding address")
	}

	// Very important: these are the denoms on stride
	denom := hostZone.IbcDenom
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)

	for _, foundTokens := range resp.Balances {
		if foundTokens.Denom == denom {
			// Liquid Stake these tokens
			// Immediately transfer new stTokens to return ICA
		} else if foundTokens.Denom == stDenom {
			// Redeem Stake these tokens
			// takes 30 days but unstaked claimed to return ICA
		}
	}

	return nil
}

// Using tokens in the CommunityPoolReturnIcaAddress, trigger ICA tx to fund community pool
// Note: The denom of the passed in token has to be the denom which exists on the hostZone not Stride
func (k Keeper) FundCommunityPool(ctx sdk.Context, hostZone types.HostZone, token sdk.Coin) error {
	fundCoins := sdk.NewCoins(token)

	var msgs []proto.Message
	msgs = append(msgs, &disttypes.MsgFundCommunityPool{
		Amount:    fundCoins,
		Depositor: hostZone.CommunityPoolReturnIcaAddress,
	})

	// No need to build ICA callback data or input an ICA callback method
	icaCallbackId := ""
	var icaCallbackData []byte
	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)
	icaTimeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos

	// Send the transaction through SubmitTx to kick off ICA command
	_, err := k.SubmitTxs(ctx,
		hostZone.ConnectionId,
		msgs,
		types.ICAAccountType_COMMUNITY_POOL_RETURN,
		icaTimeoutTimestamp,
		icaCallbackId,
		icaCallbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to SubmitTxs for FundCommunityPool, Messages: %+v", msgs)
	}

	return nil
}
