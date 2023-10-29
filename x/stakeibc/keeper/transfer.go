package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const (
	LiquidStake = "liquidstake"
	RedeemStake = "redeemstake"
	NoAction    = "noaction"
)

// Transfers tokens from the community pool deposit ICA account to the host zone holding module address for that pool
func (k Keeper) IBCTransferCommunityPoolTokens(ctx sdk.Context, token sdk.Coin, communityPoolHostZone types.HostZone, memoAction string) error {

	// The memo may contain autopilot commands to atomically liquid stake tokens when transfer succeeds
	//  both transfer+liquid stake will succeed and stTokens will end in the stride side holding address, 
	//  or neither will and the original base tokens will return to the foreign deposit ICA address
	memoCommands := ""
	if memoAction == LiquidStake {
		autopilotStakeCmd := "{ \"autopilot\": { \"receiver\": \"%s\",  \"stakeibc\": { \"action\": \"LiquidStake\" } } }"
		memoCommands = fmt.Sprintf(autopilotStakeCmd, communityPoolHostZone.CommunityPoolHoldingAddress)
		stakedDenom := types.StAssetDenomFromHostZoneDenom(token.Denom)
		k.Logger(ctx).Info(fmt.Sprintf("[IBCTransferCommunityPoolTokens] Transferring %v %s and then liquid staking to %s", token.Amount, token.Denom, stakedDenom))
	} else if memoAction == RedeemStake {
		autopilotRedeemCmd := "{ \"autopilot\": { \"receiver\": \"%s\",  \"stakeibc\": { \"action\": \"RedeemStake\" } } }"	
		memoCommands = fmt.Sprintf(autopilotRedeemCmd, communityPoolHostZone.CommunityPoolHoldingAddress)
		k.Logger(ctx).Info(fmt.Sprintf("[IBCTransferCommunityPoolTokens] Transferring %v %s and then redeeming stake", token.Amount, token.Denom))
	} else {
		k.Logger(ctx).Info(fmt.Sprintf("[IBCTransferCommunityPoolTokens] Transferring %v %s with no additional action", token.Amount, token.Denom))
	}

	// get community pool chain's transfer channel for sending tokens from hostZone to Stride
	transferChannel, found := k.IBCKeeper.ChannelKeeper.GetChannel(ctx, ibctypes.PortID, communityPoolHostZone.TransferChannelId)
	if !found {
		return errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "transfer channel %s not found", communityPoolHostZone.TransferChannelId)
	}
	incomingTransferChannelId := transferChannel.Counterparty.ChannelId

	// one timeout for ICA command ibc messages to arrive, other timeout for subsequent transfer message itself 
	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)
	icaTimeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos
	transferTimeoutTimestamp := icaTimeoutTimestamp + ibcTransferTimeoutNanos

	var msgs []proto.Message
	msgs = append(msgs, ibctypes.NewMsgTransfer(
		ibctypes.PortID,
		incomingTransferChannelId, // for transfers of communityPoolHostZone -> Stride
		token,
		communityPoolHostZone.CommunityPoolDepositIcaAddress, // ICA controlled address on community pool zone
		communityPoolHostZone.CommunityPoolHoldingAddress, // Stride address, unique to each community pool / hostzone
		clienttypes.Height{},
		transferTimeoutTimestamp,
		memoCommands,
	))

	// No need to build ICA callback data or input an ICA callback method since the callback stride can see would only
	//  be the ICA callback, not the actual transfer callback, because it would happen on the other chain -- 
	//  This is why we use autopilot: so that on transfer complete, the next action (either stake or unstake) happens without callbacks
	icaCallbackId := ""
	var icaCallbackData []byte;

	// Send the transaction through SubmitTx to kick off ICA commands -- no ICA callback method name, or callback args needed
	_, err := k.SubmitTxs(ctx, 
		communityPoolHostZone.ConnectionId, 
		msgs, 
		types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, 
		icaTimeoutTimestamp, 
		icaCallbackId, 
		icaCallbackData)
	if err != nil {
		return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("[IBCTransferCommunityPoolTokens] Successfully sent ICA command to kick off remote ibc transfer from community pool deposit ICA"))

	return nil
}


// Transfers all tokens in the Stride-side holding address over to the communityPoolReturnAddress ICA
func (k Keeper) IBCReturnAllCommunityPoolTokens(ctx sdk.Context, communityPoolHostZone types.HostZone) error {
	// Use bankKeeper to see all coin types and amounts currently in the stride-side holding module address
	req := banktypes.NewQueryAllBalancesRequest(sdk.AccAddress(communityPoolHostZone.CommunityPoolHoldingAddress), nil)
	resp, err := k.bankKeeper.AllBalances(ctx, req)
	if err != nil {
		return err
	}

	memo := ""
	errors := make([]error, 0)
	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)

	for _, foundCoin := range resp.Balances {
		// build and send an IBC message for each coin to transfer all back to the communityPoolHostZone
		timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos
		msg := ibctypes.NewMsgTransfer(
			ibctypes.PortID,
			communityPoolHostZone.TransferChannelId,
			foundCoin,
			communityPoolHostZone.CommunityPoolHoldingAddress, // from Stride address, unique to each community pool / hostzone
			communityPoolHostZone.CommunityPoolReturnIcaAddress, // to ICA controlled address on foreign hub
			clienttypes.Height{},
			timeoutTimestamp,
			memo,
		)

		msgTransferResponse, err := k.RecordsKeeper.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
		if err != nil {
			result := fmt.Sprintf("[IBCReturnAllCommunityPoolTokens] Error submitting ibc transfer for %v %s with message %s", 
				foundCoin.Amount, foundCoin.Denom, err.Error())
			k.Logger(ctx).Error(result)
			errors = append(errors, err) // log and keep track of errors but don't let one failure stop later transfer attempts
		} else {
			result := fmt.Sprintf("[IBCReturnAllCommunityPoolTokens] Successfully submitted ibc transfer for %v %s with response %+v", 
				foundCoin.Amount, foundCoin.Denom, msgTransferResponse)
			k.Logger(ctx).Info(result)

			// If there was no error in sending the transfer msg, store what the transferred denom will be in callback with amount
			callbackArgs := types.CommunityPoolReturnTransferCallback{
				HostZoneId: communityPoolHostZone.ChainId,
				DenomStride: foundCoin.Denom,
				IbcDenom: k.GetIbcDenomOnHostZone(foundCoin.Denom, communityPoolHostZone),
				Amount: foundCoin.Amount,
			}
			callbackArgsBz, err := proto.Marshal(&callbackArgs)
			if err != nil {
				return errorsmod.Wrapf(err, "Unable to marshal community pool return transfer callback data for %+v", callbackArgs)
			}

			// Register a callback by hand when the transfer msg gets the completed ack, different callback for each coin in module
			k.ICACallbacksKeeper.SetCallbackData(ctx, icacallbackstypes.CallbackData{
				CallbackKey:  icacallbackstypes.PacketID(msg.SourcePort, msg.SourceChannel, msgTransferResponse.Sequence),
				PortId:       msg.SourcePort,
				ChannelId:    msg.SourceChannel,
				Sequence:     msgTransferResponse.Sequence,
				CallbackId:   ICACallbackID_CommunityPoolReturn,
				CallbackArgs: callbackArgsBz,
			})			
		}

	}

	// if there were any errors return the first one for now...
	if len(errors) > 0 {
		return errors[0]
	}	
	return nil
}


// helper function to find the ibc denom on the foreign chain of tokens after transfer from stride  
func (k Keeper) GetIbcDenomOnHostZone(strideDenom string, hostZone types.HostZone) (ibcDenom string) {
	// we use the hostZone.TransferChannelId because direction is stride to hostZone
	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, hostZone.TransferChannelId)
	prefixedDenom := sourcePrefix + strideDenom

	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}
