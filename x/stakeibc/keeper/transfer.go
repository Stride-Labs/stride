package keeper

import (
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	autopilottypes "github.com/Stride-Labs/stride/v14/x/autopilot/types"
	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const (
	LiquidStake = "liquidstake"
	RedeemStake = "redeemstake"
	NoAction    = "noaction"
)

// Transfers tokens from the community pool deposit ICA account to the host zone holding module address for that pool
func (k Keeper) TransferCommunityPoolTokens(ctx sdk.Context, token sdk.Coin, hostZone types.HostZone, autoPilotAction string) error {

	// The memo may contain autopilot commands to atomically liquid stake/redeem tokens when transfer succeeds
	//  both transfer+liquid stake will succeed and tokens will end in the stride side holding address, 
	//  or neither will and the original base tokens will return to the foreign deposit ICA address
	memoCommands := ""
	autopilotMetadata := autopilottypes.RawPacketMetadata{}
	autopilotMetadata.Autopilot.Receiver = hostZone.CommunityPoolHoldingAddress
	if autoPilotAction == LiquidStake {
		autopilotMetadata.Autopilot.Stakeibc.Action = "LiquidStake"		
		autopilotCmdBz, err := json.Marshal(autopilotMetadata)
		if err != nil {
			errorsmod.Wrapf(err, "Couldn't build autopilot json for %v", autopilotMetadata)
		}

		memoCommands = string(autopilotCmdBz)
		k.Logger(ctx).Info(fmt.Sprintf("[TransferCommunityPoolTokens] Transferring %v %s and then liquid staking with memo %s", 
			token.Amount, token.Denom, memoCommands))
	} else if autoPilotAction == RedeemStake {
		autopilotMetadata.Autopilot.Stakeibc.Action = "RedeemStake"		
		autopilotCmdBz, err := json.Marshal(autopilotMetadata)
		if err != nil {
			errorsmod.Wrapf(err, "Couldn't build autopilot json for %v", autopilotMetadata)
		}

		memoCommands = string(autopilotCmdBz)
		k.Logger(ctx).Info(fmt.Sprintf("[TransferCommunityPoolTokens] Transferring %v %s and then redeeming stake with memo %s", 
			token.Amount, token.Denom, memoCommands))
	} else {
		k.Logger(ctx).Info(fmt.Sprintf("[TransferCommunityPoolTokens] Transferring %v %s with no additional action", 
			token.Amount, token.Denom))
	}

	// get community pool chain's transfer channel for sending tokens from hostZone to Stride
	transferChannel, found := k.IBCKeeper.ChannelKeeper.GetChannel(ctx, ibctypes.PortID, hostZone.TransferChannelId)
	if !found {
		return errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "transfer channel %s not found", hostZone.TransferChannelId)
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
		hostZone.CommunityPoolDepositIcaAddress, // ICA controlled address on community pool zone
		hostZone.CommunityPoolHoldingAddress, // Stride address, unique to each community pool / hostzone
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
		hostZone.ConnectionId, 
		msgs, 
		types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, 
		icaTimeoutTimestamp, 
		icaCallbackId, 
		icaCallbackData)
	if err != nil {
		return errorsmod.Wrapf(types.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("[TransferCommunityPoolTokens] Successfully sent ICA command to kick off remote ibc transfer from community pool deposit ICA"))

	return nil
}


// Transfers all tokens in the Stride-side holding address over to the communityPoolReturnAddress ICA
func (k Keeper) ReturnAllCommunityPoolTokens(ctx sdk.Context, hostZone types.HostZone) error {
	// Use bankKeeper to see all coin types and amounts currently in the stride-side holding module address
	req := banktypes.NewQueryAllBalancesRequest(sdk.AccAddress(hostZone.CommunityPoolHoldingAddress), nil)
	resp, err := k.bankKeeper.AllBalances(ctx, req)
	if err != nil {
		return errorsmod.Wrapf(err, "Couldn't look up balances in holding address")
	}

	var transferErr error;
	for _, foundCoin := range resp.Balances {
		if transferErr = k.TransferCoinToReturn(ctx, hostZone, foundCoin); transferErr != nil {
			k.Logger(ctx).Error(errorsmod.Wrapf(transferErr, "error in token transfer %+v", foundCoin).Error())
			continue
		}
	}

	return transferErr // will be nil if no errors in loop
}


func (k Keeper) TransferCoinToReturn(ctx sdk.Context, hostZone types.HostZone, coin sdk.Coin) error {
	memo := ""
	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos

	// build and send an IBC message for each coin to transfer all back to the hostZone
	msg := ibctypes.NewMsgTransfer(
		ibctypes.PortID,
		hostZone.TransferChannelId,
		coin,
		hostZone.CommunityPoolHoldingAddress, // from Stride address, unique to each community pool / hostzone
		hostZone.CommunityPoolReturnIcaAddress, // to ICA controlled address on foreign hub
		clienttypes.Height{},
		timeoutTimestamp,
		memo,
	)

	msgTransferResponse, err := k.RecordsKeeper.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
			return errorsmod.Wrapf(err, "Error submitting ibc transfer for %+v", coin)
	} else {
		result := fmt.Sprintf("Successfully submitted ibctransfer for %+v with response %+v", 
				coin, msgTransferResponse)
		k.Logger(ctx).Info(result)

		// If there was no error in sending the transfer msg, store what the transferred denom will be in callback with amount
		callbackArgs := types.CommunityPoolReturnTransferCallback{
			HostZoneId: hostZone.ChainId,
			DenomStride: coin.Denom,
			IbcDenom: k.GetDenomOnHostZone(coin.Denom, hostZone),
			Amount: coin.Amount,
		}
		callbackArgsBz, err := proto.Marshal(&callbackArgs)
		if err != nil {
			return errorsmod.Wrapf(err, "Unable to marshal pool return transfer callback %+v", callbackArgs)
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
	return nil
}


// helper function to find the ibc denom on the foreign chain of tokens after transfer from stride  
func (k Keeper) GetDenomOnHostZone(strideDenom string, hostZone types.HostZone) (ibcDenom string) {
	// we use the hostZone.TransferChannelId because direction is stride to hostZone
	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, hostZone.TransferChannelId)
	prefixedDenom := sourcePrefix + strideDenom

	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}

// given a hostZone with native denom, returns the ibc denom on the zone for the staked stDenom 
func (k Keeper) GetStakedHostTokenDenomOnHostZone(hostZone types.HostZone) (ibcStakedDenom string) {
	nativeDenom := hostZone.HostDenom
	stDenomOnStride := types.StAssetDenomFromHostZoneDenom(nativeDenom)
	return k.GetDenomOnHostZone(stDenomOnStride, hostZone)
}
