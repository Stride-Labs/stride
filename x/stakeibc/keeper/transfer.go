package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	epochstypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// Transfers tokens from the community pool deposit ICA account to the host zone holding module address for that pool
func (k Keeper) TransferCommunityPoolDepositToHolding(ctx sdk.Context, hostZone types.HostZone, token sdk.Coin) error {
	// Verify that the deposit ica address exists on the host zone and holding address exists on stride
	if hostZone.CommunityPoolDepositIcaAddress == "" || hostZone.CommunityPoolStakeAddress == "" {
		return errors.New("Invalid holding address or deposit address, cannot build valid ICA transfer kickoff command")
	}

	// get the hostZone counterparty transfer channel for sending tokens from hostZone to Stride
	transferChannel, found := k.IBCKeeper.ChannelKeeper.GetChannel(ctx, transfertypes.PortID, hostZone.TransferChannelId)
	if !found {
		return errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "transfer channel %s not found", hostZone.TransferChannelId)
	}
	counterpartyChannelId := transferChannel.Counterparty.ChannelId

	// Timeout both the ICA kick off command and the ibc transfer message at the epoch boundary
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	endEpochTimestamp := uint64(strideEpochTracker.NextEpochStartTime)

	memo := ""
	var msgs []proto.Message
	msgs = append(msgs, transfertypes.NewMsgTransfer(
		transfertypes.PortID,
		counterpartyChannelId, // for transfers of communityPoolHostZone -> Stride
		token,
		hostZone.CommunityPoolDepositIcaAddress, // ICA controlled address on community pool zone
		hostZone.CommunityPoolStakeAddress,      // Stride address, unique to each community pool / hostzone
		clienttypes.Height{},
		endEpochTimestamp,
		memo,
	))

	// No need to build ICA callback data or input an ICA callback method since the callback Stride can see is only
	//  the ICA callback, not the actual transfer callback. The transfer ack returns to the hostZone chain not Stride
	icaCallbackId := ""
	var icaCallbackData []byte

	// Send the transaction through SubmitTx to kick off ICA commands -- no ICA callback method name, or callback args needed
	_, err := k.SubmitTxs(ctx,
		hostZone.ConnectionId,
		msgs,
		types.ICAAccountType_COMMUNITY_POOL_DEPOSIT,
		endEpochTimestamp,
		icaCallbackId,
		icaCallbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	}
	k.Logger(ctx).Info("Successfully sent ICA command to kick off ibc transfer from deposit ICA to holding address")

	return nil
}

// Transfers a given token from the stride-side holding address to the return ICA address on the host zone
func (k Keeper) TransferHoldingToCommunityPoolReturn(ctx sdk.Context, hostZone types.HostZone, coin sdk.Coin) error {
	memo := ""
	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos

	// build and send an IBC message for each coin to transfer all back to the hostZone
	msg := transfertypes.NewMsgTransfer(
		transfertypes.PortID,
		hostZone.TransferChannelId,
		coin,
		hostZone.CommunityPoolStakeAddress,     // from Stride address, unique to each community pool / hostzone
		hostZone.CommunityPoolReturnIcaAddress, // to ICA controlled address on foreign hub
		clienttypes.Height{},
		timeoutTimestamp,
		memo,
	)

	msgTransferResponse, err := k.RecordsKeeper.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return errorsmod.Wrapf(err, "Error submitting ibc transfer for %+v", coin)
	}

	result := fmt.Sprintf("Successfully submitted ibctransfer for %+v with response %+v",
		coin, msgTransferResponse)
	k.Logger(ctx).Info(result)

	return nil
}

// given a hostZone with native denom, returns the ibc denom on the zone for the staked stDenom
func (k Keeper) GetStakedDenomOnHostZone(ctx sdk.Context, hostZone types.HostZone) (ibcStakedDenom string) {
	nativeDenom := hostZone.HostDenom
	stDenomOnStride := types.StAssetDenomFromHostZoneDenom(nativeDenom)

	// use counterparty transfer channel because tokens come through this channel to hostZone
	transferChannel, _ := k.IBCKeeper.ChannelKeeper.GetChannel(ctx, transfertypes.PortID, hostZone.TransferChannelId)
	counterpartyChannelId := transferChannel.Counterparty.ChannelId

	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, counterpartyChannelId)
	prefixedDenom := sourcePrefix + stDenomOnStride

	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}
