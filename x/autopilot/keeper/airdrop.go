package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v16/utils"
	"github.com/Stride-Labs/stride/v16/x/autopilot/types"
	claimtypes "github.com/Stride-Labs/stride/v16/x/claim/types"
	stakeibctypes "github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (k Keeper) TryUpdateAirdropClaim(
	ctx sdk.Context,
	packet channeltypes.Packet,
	fungibleTokenPacketData transfertypes.FungibleTokenPacketData,
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

	// By getting senderAddressConvertedToStrideAddress, we can check if the sender of this packet
	// has a corresponding address on Stride that owns claim records.
	// Importantly, this functionality is gated by chains - i.e. packets that come across the canonical Evmos <> Stride
	// can update claim records associated with _Evmos_ only.
	// This prevents malicious IBC channels from updating claim records (in essence, Stride trusts that the sender address on
	// certain IBC channels triggered the transaction that triggered the IBC transfer).
	// NOTE: An older version of PFM broke this assumption, but this functionality is only turned on for Evmos (has new PFM)
	// and Injective (has no PFM).
	senderAddressConvertedToStrideAddress := utils.ConvertAddressToStrideAddress(fungibleTokenPacketData.Sender)
	if senderAddressConvertedToStrideAddress == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid sender address (%s)", fungibleTokenPacketData.Sender))
	}
	// This is the updated owner of a claim record.
	targetStrideAddressForClaimRecord := fungibleTokenPacketData.Receiver

	// find the airdrop for this host chain ID
	airdrop, found := k.claimKeeper.GetAirdropByChainId(ctx, hostZone.ChainId)
	if !found {
		return errorsmod.Wrapf(claimtypes.ErrAirdropNotFound, "airdrop not found for chain-id %s", hostZone.ChainId)
	}
	if !airdrop.AutopilotEnabled {
		return fmt.Errorf("autopilot claiming is not enabled for host zone %s", hostZone.ChainId)
	}

	airdropId := airdrop.AirdropIdentifier
	k.Logger(ctx).Info(fmt.Sprintf("updating airdrop address %s (senderAddressConvertedToStrideAddress %s) to %s for airdrop %s",
		senderAddressConvertedToStrideAddress, fungibleTokenPacketData.Sender, targetStrideAddressForClaimRecord, airdropId))

	return k.claimKeeper.UpdateAirdropAddress(ctx, senderAddressConvertedToStrideAddress, targetStrideAddressForClaimRecord, airdropId)
}
