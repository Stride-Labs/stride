package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v8/utils"
	"github.com/Stride-Labs/stride/v8/x/autopilot/types"
	stakeibctypes "github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (k Keeper) TryUpdateAirdropClaim(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	packetMetadata types.ClaimPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.ClaimActive {
		return errors.New("packet forwarding param is not active")
	}

	// verify packet originated on a registered host zone
	if packet.GetDestPort() != transfertypes.PortID {
		return errors.New("packet should ")
	}
	if _, found := k.stakeibcKeeper.GetHostZoneFromTransferChannelID(ctx, packet.GetDestChannel()); !found {
		return errorsmod.Wrapf(stakeibctypes.ErrHostZoneNotFound,
			"host zone not found for transfer channel %s", packet.GetDestChannel())
	}

	// grab relevant addresses
	senderStrideAddress := utils.ConvertAddressToStrideAddress(data.Sender)
	if senderStrideAddress == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid sender address (%s)", data.Sender))
	}
	newStrideAddress := packetMetadata.StrideAddress

	// update the airdrop
	airdropId := packetMetadata.AirdropId
	k.Logger(ctx).Info(fmt.Sprintf("updating airdrop address %s (orig %s) to %s for airdrop %s",
		senderStrideAddress, data.Sender, newStrideAddress, airdropId))

	return k.claimKeeper.UpdateAirdropAddress(ctx, senderStrideAddress, newStrideAddress, airdropId)
}
