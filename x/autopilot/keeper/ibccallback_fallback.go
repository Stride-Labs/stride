package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v16/x/icacallbacks/types"
)

func (k Keeper) TransferFallbackCallback(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ackResponse *icacallbackstypes.AcknowledgementResponse,
	args []byte,
) error {
	return nil
}
