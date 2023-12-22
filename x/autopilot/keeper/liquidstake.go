package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v16/x/autopilot/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

// Attempts to do an autopilot liquid stake (and optional forward)
// The liquid stake is only allowed if the inbound packet came along a trusted channel
func (k Keeper) TryLiquidStaking(
	ctx sdk.Context,
	packet channeltypes.Packet,
	transferMetadata types.AutopilotTransferMetadata,
	actionMetadata types.StakeibcPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.StakeibcActive {
		return errorsmod.Wrapf(types.ErrPacketForwardingInactive, "autopilot stakeibc routing is inactive")
	}

	// In this case, we can't process a liquid staking transaction, because we're dealing with STRD tokens
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), transferMetadata.Denom) {
		return errors.New("the native token is not supported for liquid staking")
	}

	// Note: the denom in the packet is the base denom e.g. uatom - not ibc/xxx
	// We need to use the port and channel to build the IBC denom
	prefixedDenom := transfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), transferMetadata.Denom)
	ibcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, transferMetadata.Denom)
	if err != nil {
		return fmt.Errorf("host zone not found for denom (%s)", transferMetadata.Denom)
	}

	// Verify the IBC denom of the packet matches the host zone, to confirm the packet
	// was sent over a trusted channel
	if hostZone.IbcDenom != ibcDenom {
		return fmt.Errorf("ibc denom %s is not equal to host zone ibc denom %s", ibcDenom, hostZone.IbcDenom)
	}

	return k.RunLiquidStake(ctx, transferMetadata, actionMetadata)
}

// Submits a LiquidStake message from the hashed receiver
// If a forwarding recipient is specified, the stTokens are ibc transferred
// If there is no forwarding recipient, they are bank sent to the original receiver
func (k Keeper) RunLiquidStake(
	ctx sdk.Context,
	transferMetadata types.AutopilotTransferMetadata,
	actionMetadata types.StakeibcPacketMetadata,
) error {
	msg := &stakeibctypes.MsgLiquidStake{
		Creator:   transferMetadata.HashedReceiver,
		Amount:    transferMetadata.Amount,
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
		return errorsmod.Wrapf(err, err.Error())
	}

	// If the IBCReceiver is empty, there is no forwarding step
	// but we still need to transfer the stTokens to the original recipient
	if actionMetadata.IbcReceiver == "" {
		fromAddress := sdk.MustAccAddressFromBech32(transferMetadata.HashedReceiver)
		toAddress := sdk.MustAccAddressFromBech32(transferMetadata.OriginalReceiver)

		return k.bankkeeper.SendCoins(ctx, fromAddress, toAddress, sdk.NewCoins(msgResponse.StToken))
	}

	// Otherwise, if there is forwarding info, submit the IBC transfer
	return k.IBCTransferStToken(ctx, msgResponse.StToken, transferMetadata, actionMetadata)
}

// Submits an IBC transfer of the stToken to a non-stride zone (either back to the host zone or to a different zone)
// The sender of the transfer is the hashed receiver of the original autopilot inbound transfer
func (k Keeper) IBCTransferStToken(
	ctx sdk.Context,
	stToken sdk.Coin,
	transferMetadata types.AutopilotTransferMetadata,
	actionMetadata types.StakeibcPacketMetadata,
) error {
	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, transferMetadata.Denom)
	if err != nil {
		return err
	}

	// Use the default transfer timeout of 10 minutes
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + transfertypes.DefaultRelativePacketTimeoutTimestamp

	// If there's no channelID specified in the packet, default to the channel on the host zone
	channelId := actionMetadata.TransferChannel
	if channelId == "" {
		channelId = hostZone.TransferChannelId
	}

	// The transfer message is sent from the hashed receiver to prevent impersonation
	transferMsg := &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    channelId,
		Token:            stToken,
		Sender:           transferMetadata.HashedReceiver,
		Receiver:         actionMetadata.IbcReceiver,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             "autopilot-liquid-stake-and-forward",
	}

	_, err = k.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), transferMsg)
	return err
}
