package keeper

import (
	"fmt"

	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// ICA Callback after transfer of tokens in community pool to Stride
//
//	If successful:
//	  * Calls to LiquidStakeCommunityPoolTokens
//	If timeout:
//	  * Does nothing but log error
//	If failure:
//	  * Does nothing but log error
func (k Keeper) CommunityPoolTransferCallback(	ctx sdk.Context, 
												packet channeltypes.Packet, 
												ackResponse *icacallbackstypes.AcknowledgementResponse, 
												args []byte) error {
	// Check for timeout (ack nil)
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(fmt.Sprintf("[CommunityPoolTransferCallback] packet timeout: %v %v",
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(fmt.Sprintf("[CommunityPoolTransferCallback] packet failure: %v %v",
			icacallbackstypes.AckResponseStatus_FAILURE, packet))
		return nil
	}

	k.Logger(ctx).Info(fmt.Sprintf("[CommunityPoolTransferCallback] IBC Transfer success: %v %v",
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Fetch callback args
	var poolTransferCallback types.CommunityPoolTransferCallback
	if err := proto.Unmarshal(args, &poolTransferCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal community pool transfer callback args: %s", err.Error()))
	}
	communityPoolHoldingAddress := poolTransferCallback.CommunityPoolHoldingAddress	

	// Successful transfer of community pool funds from deposit ICA to holding address, trigger liquid staking
	k.LiquidStakeCommunityPoolTokens(ctx, communityPoolHoldingAddress)

	return nil
}
