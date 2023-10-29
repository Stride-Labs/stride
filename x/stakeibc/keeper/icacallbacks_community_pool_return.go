package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v14/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// ICACallback registered by hand fires when transfer from Stride to community pool return ICA completes
//
//	If successful: Fire the ICA to call fund-community-pool with the new denom on the chain just sent to
//	If failure or timeout: do nothing, will be retries next epoch
func (k Keeper) CommunityPoolReturnCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	returnCallback := types.CommunityPoolReturnTransferCallback{}
	if err := proto.Unmarshal(args, &returnCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal community pool return transfer callback: %s", err.Error())
	}
	chainId := returnCallback.HostZoneId
	strideDenom := returnCallback.DenomStride
	ibcDenom := returnCallback.IbcDenom
	amount := returnCallback.Amount
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_CommunityPoolReturn, "Community Pool Return Callback, sent %d %s to %s and got %s", amount, strideDenom, chainId, ibcDenom))

	communityPoolHostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for callback chain ID (%s)", chainId)
	}

	// Using the ibc denom as it will appear on the community pool host zone after the transfer
	// create tokens assuming the full amount sent has landed and then fire ICA tx to Fund Community Pool
	token := sdk.NewCoin(ibcDenom, amount)
	k.FundCommunityPool(ctx, communityPoolHostZone, token)

	return nil
}
