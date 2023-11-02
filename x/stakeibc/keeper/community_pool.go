package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v14/utils"
	epochstypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// For each hostZone with a valid community pool, trigger the ICQs and ICAs to transfer tokens from DepositICA or back to ReturnICA
// Since ICQs and ICAs take time to complete, it is almost certain tokens swept in and processed will be swept out in a later epoch
func (k Keeper) SweepAllCommunityPoolTokens(ctx sdk.Context) {
	hostZones := k.GetAllActiveHostZone(ctx)
	for _, hostZone := range hostZones {
		if hostZone.CommunityPoolDepositIcaAddress != "" &&
			hostZone.CommunityPoolHoldingAddress != "" &&
			hostZone.CommunityPoolReturnIcaAddress != "" {
				// ICQ for the host denom of the chain, these are tokens the pool wants staked
				err:= k.QueryCommunityPoolDepositBalance(ctx, hostZone, hostZone.HostDenom)
				if err != nil {
					k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId, "Querying hostDenom %s - %s", hostZone.HostDenom, err.Error()))
				}
				// ICQ for the stToken of the host denom, these are tokens the pool wants redeemed
				//   if stDenom is the denom on stride, ibcStDenom is the ibc denom on hostZone for stDenom
				ibcStDenom := k.GetStakedHostTokenDenomOnHostZone(hostZone)
				err = k.QueryCommunityPoolDepositBalance(ctx, hostZone, ibcStDenom)
				if err != nil {
					k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId, "Querying stHostDenom %s - %s", ibcStDenom, err.Error()))
				}				
				// Transfer out all all tokens in the holding address back to the Return ICA
				err = k.ReturnAllCommunityPoolTokens(ctx, hostZone)
				if err != nil {
					k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId, "Returning from holding address - %s", err.Error()))
				}				
			}
	}
}


// ICQ specific denom for balance in the deposit ICA on the community pool host zone
// The ICQ callback will call IBCTransferCommunityPoolICATokensToStride with found token(s) as input
func (k Keeper) QueryCommunityPoolDepositBalance(ctx sdk.Context, hostZone types.HostZone, denom string) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for %s in community pool deposit account balance", denom))

	// Get the withdrawal account address from the host zone
	if hostZone.CommunityPoolDepositIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no deposit account found for %s", hostZone.ChainId)
	}

	// Encode the deposit account address for the query request
	// The query request consists of the account address and denom passed in
	_, depositAddressBz, err := bech32.DecodeAndConvert(hostZone.CommunityPoolDepositIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid withdrawal account address, could not decode (%s)", hostZone.CommunityPoolDepositIcaAddress)
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(depositAddressBz), []byte(denom)...)

	// The response might be a coin, or might just be an int depending on sdk version
	// Since we need the denom later, store the denom as callback data for the query
	callbackData := types.CommunityPoolDepositQueryCallback{
		Denom: denom,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal community pool deposit balance callback data %v", callbackData)
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
		CallbackId:      ICQCallbackID_CommunityPoolDepositBalance,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return errorsmod.Wrapf(err, "Error querying for pool deposit balance")
	}

	return nil
}

// Using tokens in the CommunityPoolReturnIcaAddress, ICA tx to fund community pool
// Note: The denoms of the passed in token have to be ibc denoms which exist on the communityPoolHostZone
func (k Keeper) FundCommunityPool(ctx sdk.Context, communityPoolHostZone types.HostZone, token sdk.Coin) error {
	fundCoins := sdk.NewCoins(token)
	
	var msgs []proto.Message
	msgs = append(msgs, &disttypes.MsgFundCommunityPool{
		Amount: fundCoins,
		Depositor: communityPoolHostZone.CommunityPoolReturnIcaAddress,
	})

	// No need to build ICA callback data or input an ICA callback method 
	icaCallbackId := ""
	var icaCallbackData []byte;
	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)
	icaTimeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos

	// Send the transaction through SubmitTx to kick off ICA command -- no ICA callback method name, or callback args needed
	_, err := k.SubmitTxs(ctx, 
		communityPoolHostZone.ConnectionId, 
		msgs, 
		types.ICAAccountType_COMMUNITY_POOL_RETURN, 
		icaTimeoutTimestamp, 
		icaCallbackId, 
		icaCallbackData)
	if err != nil {
		return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs for FundCommunityPool, Messages: %v, err: %s", msgs, err.Error())
	}

	return nil
}
