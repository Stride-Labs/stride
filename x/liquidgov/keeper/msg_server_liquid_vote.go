package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v10/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v10/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	//timestamp "github.com/golang/protobuf/ptypes"
)

func (k msgServer) LiquidVote(goCtx context.Context, msg *types.MsgLiquidVote) (*types.MsgLiquidVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that we have the target proposal defined by host zone + proposal id
	currProp, propFound := k.GetProposal(ctx, msg.ProposalId, msg.HostZoneId)
	if !propFound {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidHostZone, "target proposal is not yet known for hostzone %s and proposal id %d", msg.HostZoneId, msg.ProposalId)		
	}

	// Verify that this proposal still has time left in the voting period (including buffer param)
	now := time.Now() // subtract the buffer so voting isn't possible right up to the hub limit since ICAs take time
	if now.After(currProp.GovProposal.VotingEndTime) {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrTxMsgDataInvalid, "This proposal is no longer in the legal voting period")		
	}

	// Verify that the user has already deposited enough of the right token for this vote amount to be possible
	currDeposit, depositFound := k.GetDeposit(ctx, sdk.AccAddress(msg.Creator), msg.HostZoneId)
	if !depositFound || currDeposit.Amount.LT(msg.Amount) {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "not enough deposit found (%d < %d)", currDeposit.Amount, msg.Amount)
	}

	// See if there are already votes for this chain + proposal from this user, must match the existing option (cancel first to change)
	currVote, voteFound := k.GetVote(ctx, sdk.AccAddress(msg.Creator), msg.HostZoneId, msg.ProposalId)
	if voteFound {
		if currVote.Option.String() != msg.VoteOption.String() {
			return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "current vote does not match previous vote (%v < %v)", currVote.Option, msg.VoteOption)
		}
		if (currVote.Amount.Add(msg.Amount)).LT(currDeposit.Amount) {
			return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "not enough deposit found (%d < %d)", currDeposit.Amount, currVote.Amount.Add(msg.Amount))
		}
		currVote.Amount = currVote.Amount.Add(msg.Amount)
	} else {
		// Holding period should be defined as a param...
		holdingPeriod := time.Hour
		escrowAvailableTime := currProp.GovProposal.VotingEndTime.Add(holdingPeriod)
		escrowAvailable, _ := gogotypes.TimestampProto(escrowAvailableTime)
		currVote, _ = types.NewVote(msg.Creator, msg.HostZoneId, msg.ProposalId, msg.Amount, msg.VoteOption, escrowAvailable)
	}

	k.Keeper.SetVote(ctx, currVote)

	return &types.MsgLiquidVoteResponse{}, nil
}
