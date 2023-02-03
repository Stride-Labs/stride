package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	epochstypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	if epochInfo.Identifier == epochstypes.STRIDE_EPOCH {
		// on epoch update proposals
		k.UpdateProposals(ctx)

		// check for outstanding proposals in vote cast window and cast votes
		k.CastVotes(ctx, epochInfo)

		// // for every mature record
		//
		//	// either loop through records here or dequeue from UnlockingRecordQueue
		//	k.CompleteUnlocking(ctx, denom, owner) // returns mature lockups to owners and deletes lockups
	}
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.AfterEpochEnd(ctx, epochInfo)
}

func (k Keeper) UpdateProposals(ctx sdk.Context) {
	hostZone, found := k.stakeibcKeeper.GetHostZone(ctx, "GAIA") // TODO remove hardcode
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("hostzone not found %s...", "GAIA"))
	} else {
		k.MirrorProposals(ctx, hostZone)
	}
}

func (k Keeper) CastVotes(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	// get param for cast window - may be hostzone specific?
	// window := k.GetCastWindowParam(ctx)
	// for every proposal in each host zone check for cast window and cast votes
	proposals := k.GetProposals(ctx)
	for _, proposal := range proposals {
		// if proposal in cast window
		// if epochInfo.CurrentEpochStartTime + window > proposal.votingEndTime {
		// TODO loop host zones
		hostZone, found := k.stakeibcKeeper.GetHostZone(ctx, proposal.HostZoneId)
		if !found {
			k.Logger(ctx).Info(fmt.Sprintf("hostzone not found %s...", proposal.HostZoneId))
		} else {
			// tally votes
			// vote := k.TallyVotes(ctx, proposal)
			// TODO remove hardcode - weighted votes
			vote := govtypesv1beta1.NewVote(proposal.GovProposal.ProposalId, sdk.AccAddress(hostZone.DelegationAccount.Address), govtypesv1beta1.NewNonSplitVoteOption(govtypesv1beta1.OptionYes))
			k.CastVoteOnHost(ctx, hostZone, vote) // submits vote ICA on host
		}
		// cast vote on host
	}
	// }
}

// function may also want to prune old proposals
// that failed to cast votes past end of voting period

// func (k Keeper) TallyVotes(ctx, proposal types.Proposal) vote {
// 	// votes := k.GetVotesForProposal(ctx, proposal)
// 	// yes := sdk.ZeroInt()
// 	// no := sdk.ZeroInt()
// 	// noVeto := sdk.ZeroInt()
// 	// abstain := sdk.ZeroInt()

// 	// for _, vote := range votes {
// 	// 	lockup, err := k.GetLockedTokens(ctx, msg.Signer, msg.denom)
// 	// 	// check for valid lockup
// 	// //	err := k.CheckLockup(ctx, msg.Signer, vote.amount, proposal)
// 	// 	if err
// 	// 		// delete invalid vote
// 	// 		k.DeleteVote(ctx, vote.voter, vote.proposalId, vote.hostZoneChainId)
// 	// 	switch
// 	// 		// add totals for each vote option from lockup amounts
// 	}
// 	// return highest vote percentage
// }
