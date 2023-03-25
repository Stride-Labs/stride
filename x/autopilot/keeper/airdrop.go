package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v7/utils"
	"github.com/Stride-Labs/stride/v7/x/autopilot/types"
)

func (k Keeper) TryUpdateAirdropClaim(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	packetMetadata *types.ClaimPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.ClaimActive {
		return errors.New("packet forwarding param is not active")
	}

	// grab relevant addresses
	senderStrideAddress := utils.ConvertAddressToStrideAddress(data.Sender)
	if senderStrideAddress == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid sender address (%s)", data.Sender))
	}
	newStrideAddress := packetMetadata.StrideAddress.String()

	// update the airdrop
	airdropId := packetMetadata.AirdropId
	k.Logger(ctx).Info(fmt.Sprintf("updating airdrop address %s (orig %s) to %s for airdrop %s",
		senderStrideAddress, data.Sender, newStrideAddress, airdropId))
	err := k.claimKeeper.UpdateAirdropAddress(ctx, senderStrideAddress, newStrideAddress, airdropId)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to update airdrop address: %v", err))
	}

	return nil
}
