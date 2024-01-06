package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v16/x/autopilot/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v16/x/icacallbacks/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

// Attempts to do an autopilot liquid stake (and optional forward)
// The liquid stake is only allowed if the inbound packet came along a trusted channel
func (k Keeper) TryLiquidStaking(
	ctx sdk.Context,
	packet channeltypes.Packet,
	transferMetadata transfertypes.FungibleTokenPacketData,
	autopilotMetadata types.StakeibcPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.StakeibcActive {
		return errorsmod.Wrapf(types.ErrPacketForwardingInactive, "autopilot stakeibc routing is inactive")
	}

	// Verify the amount is valid
	amount, ok := sdk.NewIntFromString(transferMetadata.Amount)
	if !ok {
		return errors.New("not a parsable amount field")
	}

	// In this case, we can't process a liquid staking transaction, because we're dealing with native tokens (e.g. STRD, stATOM)
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), transferMetadata.Denom) {
		return fmt.Errorf("native token is not supported for liquid staking (%s)", transferMetadata.Denom)
	}

	// Note: the denom in the packet is the base denom e.g. uatom - not ibc/xxx
	// We need to use the port and channel to build the IBC denom
	prefixedDenom := transfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), transferMetadata.Denom)
	ibcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, transferMetadata.Denom)
	if err != nil {
		return err
	}

	// Verify the IBC denom of the packet matches the host zone, to confirm the packet
	// was sent over a trusted channel
	if hostZone.IbcDenom != ibcDenom {
		return fmt.Errorf("ibc denom %s is not equal to host zone ibc denom %s", ibcDenom, hostZone.IbcDenom)
	}

	return k.RunLiquidStake(ctx, amount, transferMetadata, autopilotMetadata)
}

// Submits a LiquidStake message from the transfer receiver
// If a forwarding recipient is specified, the stTokens are ibc transferred
func (k Keeper) RunLiquidStake(
	ctx sdk.Context,
	amount sdkmath.Int,
	transferMetadata transfertypes.FungibleTokenPacketData,
	autopilotMetadata types.StakeibcPacketMetadata,
) error {
	msg := &stakeibctypes.MsgLiquidStake{
		Creator:   transferMetadata.Receiver,
		Amount:    amount,
		HostDenom: transferMetadata.Denom,
	}

	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	msgServer := stakeibckeeper.NewMsgServerImpl(k.stakeibcKeeper)
	msgResponse, err := msgServer.LiquidStake(
		sdk.WrapSDKContext(ctx),
		msg,
	)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to liquid stake")
	}

	// If the IBCReceiver is empty, there is no forwarding step
	if autopilotMetadata.IbcReceiver == "" {
		return nil
	}

	// Otherwise, if there is forwarding info, submit the IBC transfer
	return k.IBCTransferStToken(ctx, msgResponse.StToken, transferMetadata, autopilotMetadata)
}

// Submits an IBC transfer of the stToken to a non-stride zone (either back to the host zone or to a different zone)
// The sender of the transfer is the hashed receiver of the original autopilot inbound transfer
func (k Keeper) IBCTransferStToken(
	ctx sdk.Context,
	stToken sdk.Coin,
	transferMetadata transfertypes.FungibleTokenPacketData,
	autopilotMetadata types.StakeibcPacketMetadata,
) error {
	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, transferMetadata.Denom)
	if err != nil {
		return err
	}

	// If there's no channelID specified in the packet, default to the channel on the host zone
	channelId := autopilotMetadata.TransferChannel
	if channelId == "" {
		channelId = hostZone.TransferChannelId
	}

	// Generate a hashed address for the sender to the next hop,
	// to prevent impersonation at downstream zones
	// Note: The channel ID here is different than the one used in PFM
	// (we use the outbound channelID, they use the inbound channelID)
	// DOUBLE CHECK ME that it shouldn't matter
	hashedAddress, err := types.GenerateHashedAddress(channelId, transferMetadata.Sender)
	if err != nil {
		return err
	}

	// First we need to bank send to the hashed address
	originalReceiverAddress := transferMetadata.Receiver
	originalReceiver, err := sdk.AccAddressFromBech32(originalReceiverAddress)
	if err != nil {
		return err
	}
	hashedSender, err := sdk.AccAddressFromBech32(hashedAddress)
	if err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoins(ctx, originalReceiver, hashedSender, sdk.NewCoins(stToken)); err != nil {
		return err
	}

	// Use the default transfer timeout of 10 minutes
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + transfertypes.DefaultRelativePacketTimeoutTimestamp

	// Submit the transfer from the hashed address
	transferMsg := &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    channelId,
		Token:            stToken,
		Sender:           hashedAddress,
		Receiver:         autopilotMetadata.IbcReceiver,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             "autopilot-liquid-stake-and-forward",
	}
	transferResponse, err := k.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), transferMsg)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to submit transfer during autopilot liquid stake and forward")
	}
	sequence := transferResponse.Sequence

	// Store the original receiver as the fallback address in case the transfer fails
	transferCallback := types.TransferCallback{
		FallbackAddress: originalReceiverAddress,
	}
	transferCallbackBz, err := proto.Marshal(&transferCallback)
	if err != nil {
		return err
	}

	callbackData := icacallbacktypes.CallbackData{
		CallbackKey:  icacallbacktypes.PacketID(transfertypes.PortID, channelId, sequence),
		PortId:       transfertypes.PortID,
		ChannelId:    channelId,
		Sequence:     sequence,
		CallbackId:   IBCCallbackID_Transfer,
		CallbackArgs: transferCallbackBz,
	}
	k.ibccallbacksKeeper.SetCallbackData(ctx, callbackData)

	return err
}
