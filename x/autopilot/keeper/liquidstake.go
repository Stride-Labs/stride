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

func (k Keeper) TryLiquidStaking(
	ctx sdk.Context,
	packet channeltypes.Packet,
	newData types.TokenPacketMetadata,
	packetMetadata types.StakeibcPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.StakeibcActive {
		return errorsmod.Wrapf(types.ErrPacketForwardingInactive, "autopilot stakeibc routing is inactive")
	}

	// In this case, we can't process a liquid staking transaction, because we're dealing with STRD tokens
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), newData.Denom) {
		return errors.New("the native token is not supported for liquid staking")
	}

	// Note: denom is base denom e.g. uatom - not ibc/xxx
	var token = sdk.NewCoin(newData.Denom, newData.Amount)

	prefixedDenom := transfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), newData.Denom)
	ibcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, token.Denom)
	if err != nil {
		return fmt.Errorf("host zone not found for denom (%s)", token.Denom)
	}

	if hostZone.IbcDenom != ibcDenom {
		return fmt.Errorf("ibc denom %s is not equal to host zone ibc denom %s", ibcDenom, hostZone.IbcDenom)
	}

	return k.RunLiquidStake(ctx, newData, packetMetadata)
}

func (k Keeper) RunLiquidStake(ctx sdk.Context, transferMetadata types.TokenPacketMetadata, packetMetadata types.StakeibcPacketMetadata) error {
	msg := &stakeibctypes.MsgLiquidStake{
		Creator:   transferMetadata.HashedReceiver,
		Amount:    transferMetadata.Amount,
		HostDenom: transferMetadata.Denom,
	}

	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	msgServer := stakeibckeeper.NewMsgServerImpl(k.stakeibcKeeper)
	result, err := msgServer.LiquidStake(
		sdk.WrapSDKContext(ctx),
		msg,
	)
	if err != nil {
		return errorsmod.Wrapf(err, err.Error())
	}

	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, transferMetadata.Denom)
	if err != nil {
		return err
	}

	// If the IBCReceiver is empty, there is no forwarding step
	// but we still need to transfer the stTokens to the original recipient
	if packetMetadata.IbcReceiver == "" {
		fromAddress := sdk.MustAccAddressFromBech32(transferMetadata.HashedReceiver)
		toAddress := sdk.MustAccAddressFromBech32(transferMetadata.OriginalReceiver)

		return k.bankkeeper.SendCoins(ctx, fromAddress, toAddress, sdk.NewCoins(result.StToken))
	}

	// Otherwise, if there is forwarding info, submit the IBC transfer
	return k.IBCTransferStAsset(ctx, result.StToken, transferMetadata.HashedReceiver, hostZone, packetMetadata)
}

func (k Keeper) IBCTransferStAsset(ctx sdk.Context, stAsset sdk.Coin, sender string, hostZone *stakeibctypes.HostZone, packetMetadata types.StakeibcPacketMetadata) error {
	ibcTransferTimeoutNanos := k.stakeibcKeeper.GetParam(ctx, stakeibctypes.KeyIBCTransferTimeoutNanos)
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos
	channelId := packetMetadata.TransferChannel
	if channelId == "" {
		channelId = hostZone.TransferChannelId
	}
	transferMsg := &transfertypes.MsgTransfer{
		SourcePort:    transfertypes.PortID,
		SourceChannel: channelId,
		Token:         stAsset,
		// TODO: does this reintroduce the bug in PFM where senders can be spoofed?
		// If so, should we instead call PFM directly to forward the packet?
		// Or should we obfuscate the sender, making it a random address?
		Sender:           sender,
		Receiver:         packetMetadata.IbcReceiver,
		TimeoutTimestamp: timeoutTimestamp,
		// TimeoutHeight:    clienttypes.Height{},
		// Memo:             "stTokenIBCTransfer",
	}

	_, err := k.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), transferMsg)
	return err
}
