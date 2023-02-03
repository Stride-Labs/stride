package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v5/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Marshalls CastVoteOnHost callback arguments
func (k Keeper) MarshalCastVoteOnHostCallbackArgs(ctx sdk.Context, castVoteOnHostCallback types.CastVoteOnHostCallback) ([]byte, error) {
	out, err := proto.Marshal(&castVoteOnHostCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalCastVoteOnHostCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// Unmarshalls castvoteonhost callback arguments into a CastVoteOnHostCallback struct
func (k Keeper) UnmarshalCastVoteOnHostCallbackArgs(ctx sdk.Context, castVoteOnHostCallback []byte) (*types.CastVoteOnHostCallback, error) {
	unmarshalledCastVoteOnHostCallback := types.CastVoteOnHostCallback{}
	if err := proto.Unmarshal(castVoteOnHostCallback, &unmarshalledCastVoteOnHostCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalReinvestCallbackArgs %s", err.Error()))
		return nil, err
	}
	return &unmarshalledCastVoteOnHostCallback, nil
}

// Vote record for retry? and so one vote is sent at a time
func CastVoteOnHostCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	castVoteOnHostCallback, err := k.UnmarshalCastVoteOnHostCallbackArgs(ctx, args)
	if err != nil {
		return sdkerrors.Wrapf(stakeibctypes.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal delegate callback args: %s", err.Error()))
	}

	chainId := castVoteOnHostCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_CastVoteOnHost,
		"Starting vote callback for Proposal #%d on host zone %s", castVoteOnHostCallback.GovVote.ProposalId, castVoteOnHostCallback.HostZoneId))

	// Confirm chainId and deposit record Id exist
	_, found := k.stakeibcKeeper.GetHostZone(ctx, chainId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "host zone not found %s", chainId)
	}

	// Check for timeout (ack nil)
	// No need to reset the vote record status since it will get reverted when the channel is restored
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_CastVoteOnHost,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// Reset the vote record status upon failure
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_CastVoteOnHost,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))

		// Reset deposit record status
		// depositRecord.Status = recordstypes.DepositRecord_DELEGATION_QUEUE
		// k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_CastVoteOnHost,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))
	// Check for timeout (ack nil)

	// Check for a failed transaction (ack error)

	// on success delete proposal
	k.DeleteProposal(ctx, castVoteOnHostCallback.GovVote.ProposalId, castVoteOnHostCallback.HostZoneId)
	k.Logger(ctx).Info(fmt.Sprintf("[VOTE] success on %s", chainId))
	return nil
}
