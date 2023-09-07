package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// Transfers tokens from the community pool deposit ICA account to the host zone holding address for that pool
func (k Keeper) IBCTransferCommunityPoolTokens(ctx sdk.Context, msg *transfertypes.MsgTransfer) error {
	// Submit IBC transfer msg
	msgTransferResponse, err := k.RecordsKeeper.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return err
	}

	// Build callback data
	communityPoolHoldingAddress := msg.Receiver
	transferCallback := types.CommunityPoolTransferCallback{
		CommunityPoolHoldingAddress: communityPoolHoldingAddress,
	}

	k.Logger(ctx).Info(fmt.Sprintf("Marshalling CommunityPoolTransferCallback args: %+v", transferCallback))
	marshalledCallbackArgs, err := proto.Marshal(&transferCallback)
	if err != nil {
		return err
	}

	// Store the callback data
	sequence := msgTransferResponse.Sequence
	callback := icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(msg.SourcePort, msg.SourceChannel, sequence),
		PortId:       msg.SourcePort,
		ChannelId:    msg.SourceChannel,
		Sequence:     sequence,
		CallbackId:   ICACallbackID_CommunityPoolTransfer,
		CallbackArgs: marshalledCallbackArgs,
	}
	k.Logger(ctx).Info(fmt.Sprintf("Storing CommunityPoolTransferCallback data: %+v", callback))
	k.ICACallbacksKeeper.SetCallbackData(ctx, callback)

	return nil
}
