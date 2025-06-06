package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v27/x/autopilot/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v27/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

func (k Keeper) TryRedeemStake(
	ctx sdk.Context,
	packet channeltypes.Packet,
	transferPacketData transfertypes.FungibleTokenPacketData,
	autopilotMetadata types.StakeibcPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.StakeibcActive {
		return fmt.Errorf("packet forwarding param is not active")
	}

	// At this point in the stack, the denom's in the packet data appear as they existed on the sender zone,
	//   but as a denom trace instead of a hash
	// Meaning, for native stTokens, the port and channel on the host zone are part of the denom
	//   (e.g. transfer/{channel-on-hub}/stuatom)
	// Only stride native stTokens can be redeemed, so we confirm that the denom's prefix matches
	//   the packet's "source" channel (i.e. the channel on the host zone)
	if !transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), transferPacketData.Denom) {
		return fmt.Errorf("the ibc token %s is not supported for redeem stake", transferPacketData.Denom)
	}

	voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
	stAssetDenom := transferPacketData.Denom[len(voucherPrefix):]
	if !k.stakeibcKeeper.CheckIsStToken(ctx, stAssetDenom) {
		return fmt.Errorf("not a liquid staking token")
	}

	hostZoneDenom := stakeibctypes.HostZoneDenomFromStAssetDenom(stAssetDenom)

	amount, ok := sdk.NewIntFromString(transferPacketData.Amount)
	if !ok {
		return fmt.Errorf("not a parsable amount field")
	}

	strideAddress := transferPacketData.Receiver
	redemptionReceiver := autopilotMetadata.IbcReceiver

	return k.RunRedeemStake(ctx, strideAddress, redemptionReceiver, hostZoneDenom, amount)
}

func (k Keeper) RunRedeemStake(ctx sdk.Context, strideAddress string, redemptionReceiver string, hostZoneDenom string, amount sdkmath.Int) error {
	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, hostZoneDenom)
	if err != nil {
		return err
	}

	msg := &stakeibctypes.MsgRedeemStake{
		Creator:  strideAddress,
		Amount:   amount,
		HostZone: hostZone.ChainId,
		Receiver: redemptionReceiver,
	}

	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	msgServer := stakeibckeeper.NewMsgServerImpl(k.stakeibcKeeper)
	if _, err = msgServer.RedeemStake(sdk.WrapSDKContext(ctx), msg); err != nil {
		return errorsmod.Wrapf(err, "redeem stake failed")
	}

	return nil
}
