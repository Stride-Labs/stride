package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v11/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v11/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Marshalls CastVoteOnHost callback arguments
func (k Keeper) MarshalCastVotesCallbackArgs(ctx sdk.Context, castVotesCallback types.CastVotesCallback) ([]byte, error) {
	out, err := proto.Marshal(&castVotesCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalCastVoteOnHostCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// Unmarshalls castvoteonhost callback arguments into a CastVoteOnHostCallback struct
func (k Keeper) UnmarshalCastVotesCallbackArgs(ctx sdk.Context, castVotesCallback []byte) (*types.CastVotesCallback, error) {
	unmarshalledCastVotesCallback := types.CastVotesCallback{}
	if err := proto.Unmarshal(castVotesCallback, &unmarshalledCastVotesCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalReinvestCallbackArgs %s", err.Error()))
		return nil, err
	}
	return &unmarshalledCastVotesCallback, nil
}

// Vote record for retry? and so one vote is sent at a time
func (k Keeper) CastVotesCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	castVotesCallback, err := k.UnmarshalCastVotesCallbackArgs(ctx, args)
	if err != nil {
		return sdkerrors.Wrapf(stakeibctypes.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal delegate callback args: %s", err.Error()))
	}

	chainId := castVotesCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_CastVotes,
		"Starting vote callback for Proposal #%d on host zone %s", castVotesCallback.ProposalId, castVotesCallback.HostZoneId))

	// Confirm chainId and deposit record Id exist
	_, found := k.stakeibcKeeper.GetHostZone(ctx, chainId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "host zone not found %s", chainId)
	}

	// Check for timeout (ack nil)
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_CastVotes,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_CastVotes,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))
		k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_CastVotes,
			"Error with ICA ack: %s", ackResponse.Error))
		return nil
	}


	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_CastVotes,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))
	k.Logger(ctx).Info(fmt.Sprintf("[VOTE] success on %s", chainId))

	return nil
}
