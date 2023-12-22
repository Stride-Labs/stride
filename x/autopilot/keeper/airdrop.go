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
	transferMetadata types.AutopilotTransferMetadata,
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
	senderStrideAddress := utils.ConvertAddressToStrideAddress(transferMetadata.Sender)
	if senderStrideAddress == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid sender address (%s)", transferMetadata.Sender))
	}
	newStrideAddress := transferMetadata.OriginalReceiver

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
		senderStrideAddress, transferMetadata.Sender, newStrideAddress, airdropId))

	if err := k.claimKeeper.UpdateAirdropAddress(ctx, senderStrideAddress, newStrideAddress, airdropId); err != nil {
		return err
	}

	// Finally send token back to the original reciever (since the hashed receiver was used for the transfer)
	fromAddress := sdk.MustAccAddressFromBech32(transferMetadata.HashedReceiver)
	toAddress := sdk.MustAccAddressFromBech32(transferMetadata.OriginalReceiver)
	token := sdk.NewCoin(transferMetadata.Denom, transferMetadata.Amount)

	return k.bankkeeper.SendCoins(ctx, fromAddress, toAddress, sdk.NewCoins(token))
}
