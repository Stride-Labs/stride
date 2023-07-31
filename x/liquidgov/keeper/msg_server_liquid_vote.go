package keeper

import (
	"context"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	//timestamp "github.com/golang/protobuf/ptypes"

	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"
)

func (k msgServer) LiquidVote(goCtx context.Context, msg *types.MsgLiquidVote) (*types.MsgLiquidVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.Keeper.Logger(ctx).Info(fmt.Sprintf("Liquid Voting with %d escrowed tokens", msg.Amount))

	// Verify that we have the target proposal defined by host zone + proposal id
	currProp, propFound := k.GetProposal(ctx, msg.HostZoneId, msg.ProposalId)
	k.Keeper.Logger(ctx).Info(fmt.Sprintf("Voting ID %d found this local proposal %+v", msg.ProposalId, currProp))
	if !propFound {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidHostZone, "target proposal is not yet known for hostzone %s and proposal id %d", msg.HostZoneId, msg.ProposalId)		
	}

	// Verify that this proposal still has time left in the voting period (including buffer param)
	now := time.Now() // subtract the buffer so voting isn't possible right up to the hub limit since ICAs take time
	if now.After(currProp.GovProposal.VotingEndTime) {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrTxMsgDataInvalid, "This proposal is no longer in the legal voting period")		
	}

	// Verify that the user has already deposited enough of the right token for this vote amount to be possible
	currDeposit, depositFound := k.GetDeposit(ctx, msg.Creator, msg.HostZoneId)
	if !depositFound || currDeposit.Amount.LT(msg.Amount) {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "not enough deposit found (%d < %d)", currDeposit.Amount, msg.Amount)
	}

	k.Keeper.Logger(ctx).Info(fmt.Sprintf("Vote has time left and enough deposit found %+v", currDeposit))

	// See if there are already votes for this chain + proposal from this user, must match the existing option (cancel first to change)
	currVote, voteFound := k.GetVote(ctx, msg.Creator, msg.HostZoneId, msg.ProposalId)
	if voteFound {
		if currVote.Option.String() != msg.VoteOption.String() {
			return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "current vote does not match previous vote (%v < %v)", currVote.Option, msg.VoteOption)
		}
		if (currVote.Amount.Add(msg.Amount)).LT(currDeposit.Amount) {
			return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "not enough deposit found (%d < %d)", currDeposit.Amount, currVote.Amount.Add(msg.Amount))
		}
		currVote.Amount = currVote.Amount.Add(msg.Amount)
	} else {
		// Holding period should be defined as a param related to the economic security of liquid voting...
		holdingPeriod := time.Hour
		escrowAvailableTime := currProp.GovProposal.VotingEndTime.Add(holdingPeriod)
		escrowAvailable, _ := gogotypes.TimestampProto(escrowAvailableTime)
		currVote, _ = types.NewVote(msg.Creator, msg.HostZoneId, msg.ProposalId, msg.Amount, msg.VoteOption, escrowAvailable)
	}

	k.Keeper.SetVote(ctx, currVote)

	return &types.MsgLiquidVoteResponse{}, nil
}
