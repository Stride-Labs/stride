package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	epochstypes "github.com/Stride-Labs/stride/v18/x/epochs/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

// Transfers tokens from the community pool deposit ICA account to the host zone stake holding module address for that pool
func (k Keeper) TransferCommunityPoolDepositToHolding(ctx sdk.Context, hostZone types.HostZone, token sdk.Coin) error {
	// Verify that the deposit ica address exists on the host zone and stake holding address exists on stride
	if hostZone.CommunityPoolDepositIcaAddress == "" || hostZone.CommunityPoolStakeHoldingAddress == "" {
		return types.ErrICAAccountNotFound.Wrap(
			"Invalid deposit address or stake holding address, cannot build valid ICA transfer kickoff command")
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

	// Determine the host zone's stToken ibc denom
	nativeDenom := hostZone.HostDenom
	stIbcDenom, err := k.GetStIbcDenomOnHostZone(ctx, hostZone)
	if err != nil {
		return err
	}

	// If the token is the host zone's native token, we send it to the stake holding address to be liquid staked
	// Otherwise, if it's an stToken, we send it to the redeem holding address to be redeemed
	var destinationHoldingAddress string
	switch token.Denom {
	case nativeDenom:
		destinationHoldingAddress = hostZone.CommunityPoolStakeHoldingAddress
	case stIbcDenom:
		destinationHoldingAddress = hostZone.CommunityPoolRedeemHoldingAddress
	default:
		return fmt.Errorf("Invalid community pool transfer denom: %s", token.Denom)
	}

	memo := ""
	var msgs []proto.Message
	msgs = append(msgs, transfertypes.NewMsgTransfer(
		transfertypes.PortID,
		counterpartyChannelId, // for transfers of communityPoolHostZone -> Stride
		token,
		hostZone.CommunityPoolDepositIcaAddress, // ICA controlled address on community pool zone
		destinationHoldingAddress,               // Stride address, unique to each community pool / hostzone
		clienttypes.Height{},
		endEpochTimestamp,
		memo,
	))

	// No need to build ICA callback data or input an ICA callback method since the callback Stride can see is only
	//  the ICA callback, not the actual transfer callback. The transfer ack returns to the hostZone chain not Stride
	icaCallbackId := ""
	var icaCallbackData []byte

	// Send the transaction through SubmitTx to kick off ICA commands -- no ICA callback method name, or callback args needed
	_, err = k.SubmitTxs(ctx,
		hostZone.ConnectionId,
		msgs,
		types.ICAAccountType_COMMUNITY_POOL_DEPOSIT,
		endEpochTimestamp,
		icaCallbackId,
		icaCallbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	}
	k.Logger(ctx).Info("Successfully sent ICA command to kick off ibc transfer from deposit ICA to stake holding address")

	return nil
}

// Transfers a recently minted stToken from the stride-side stake holding address to the return ICA address on the host zone
func (k Keeper) TransferHoldingToCommunityPoolReturn(ctx sdk.Context, hostZone types.HostZone, coin sdk.Coin) error {
	memo := ""
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, epochstypes.STRIDE_EPOCH)
	}
	endEpochTimestamp := uint64(strideEpochTracker.NextEpochStartTime)

	// build and send an IBC message for the coin to transfer it back to the hostZone
	msg := transfertypes.NewMsgTransfer(
		transfertypes.PortID,
		hostZone.TransferChannelId,
		coin,
		hostZone.CommunityPoolStakeHoldingAddress, // from Stride address, unique to each community pool / hostzone
		hostZone.CommunityPoolReturnIcaAddress,    // to ICA controlled address on foreign hub
		clienttypes.Height{},
		endEpochTimestamp,
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
func (k Keeper) GetStIbcDenomOnHostZone(ctx sdk.Context, hostZone types.HostZone) (ibcStakedDenom string, err error) {
	nativeDenom := hostZone.HostDenom
	stDenomOnStride := types.StAssetDenomFromHostZoneDenom(nativeDenom)

	// use counterparty transfer channel because tokens come through this channel to hostZone
	transferChannel, found := k.IBCKeeper.ChannelKeeper.GetChannel(ctx, transfertypes.PortID, hostZone.TransferChannelId)
	if !found {
		return "", channeltypes.ErrChannelNotFound.Wrap(hostZone.TransferChannelId)
	}

	counterpartyChannelId := transferChannel.Counterparty.ChannelId
	if counterpartyChannelId == "" {
		return "", channeltypes.ErrChannelNotFound.Wrapf("counterparty channel not found for %s", hostZone.TransferChannelId)
	}

	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, counterpartyChannelId)
	prefixedDenom := sourcePrefix + stDenomOnStride

	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom(), nil
}
