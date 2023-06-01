package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/autopilot/types"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
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
		return errors.New("airdrop claim packet should be sent along a transfer channel")
	}
	hostZone, found := k.stakeibcKeeper.GetHostZoneFromTransferChannelID(ctx, packet.GetDestChannel())
	if !found {
		return errorsmod.Wrapf(stakeibctypes.ErrHostZoneNotFound,
			"host zone not found for transfer channel %s", packet.GetDestChannel())
	}

	// grab relevant addresses
	senderStrideAddress := utils.ConvertAddressToStrideAddress(data.Sender)
	if senderStrideAddress == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid sender address (%s)", data.Sender))
	}
	newStrideAddress := packetMetadata.StrideAddress

	// find the airdrop for this host chain ID
	airdrop, found := k.claimKeeper.GetAirdropByChainId(ctx, hostZone.ChainId)
	if !found {
		return errorsmod.Wrapf(claimtypes.ErrAirdropNotFound, "airdrop not found for chain-id %s", hostZone.ChainId)
	}
	if !airdrop.AutopilotEnabled {
		return fmt.Errorf("autopilot claiming is not enabled for host zone %s", hostZone.ChainId)
	}

	airdropId := airdrop.AirdropIdentifier
	k.Logger(ctx).Info(fmt.Sprintf("updating airdrop address %s (orig %s) to %s for airdrop %s",
		senderStrideAddress, data.Sender, newStrideAddress, airdropId))

	return k.claimKeeper.UpdateAirdropAddress(ctx, senderStrideAddress, newStrideAddress, airdropId)
}
