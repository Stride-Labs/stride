package keeper

import (
	"sort"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"

	proto "github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v11/utils"
	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
)

// GetProposal gets a proposal from store by ProposalID.
// Panics if can't unmarshal the proposal.
func (keeper Keeper) GetVote(ctx sdk.Context, creator string, hostZoneId string, proposalId uint64) (types.Vote, bool) {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.VoteKey(creator, hostZoneId, proposalId))
	if bz == nil {
		return types.Vote{}, false
	}

	var vote types.Vote
	if err := keeper.UnmarshalVote(bz, &vote); err != nil {
		panic(err)
	}

	return vote, true
}

// SetVote sets a proposal to store.
// Panics if can't marshal the Vote.
func (keeper Keeper) SetVote(ctx sdk.Context, vote types.Vote) {
	bz, err := keeper.MarshalVote(vote)
	if err != nil {
		panic(err)
	}

	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.VoteKey(vote.Creator, vote.HostZoneId, vote.ProposalId), bz)

	// Kick off ICA to update the foreign hub about the new votes
	err = keeper.CastVotes(ctx, vote.HostZoneId, vote.ProposalId)
	if err != nil {
		keeper.Logger(ctx).Error(err.Error())
	}
}

// DeleteVote deletes a Vote from store.
// Panics if the Vote doesn't exist.
func (keeper Keeper) DeleteVote(ctx sdk.Context, creator string, hostZoneId string, proposalId uint64) {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(types.VoteKey(creator, hostZoneId, proposalId))

	// Kick off ICA to update the foreign hub about the changed votes
	err := keeper.CastVotes(ctx, hostZoneId, proposalId)	
	if err != nil {
		keeper.Logger(ctx).Error(err.Error())
	}
}

// IterateVotes iterates over the all the Votes and performs a callback function.
// Panics when the iterator encounters a Vote which can't be unmarshaled.
func (keeper Keeper) IterateVotes(ctx sdk.Context, cb func(vote types.Vote) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.VotesKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var vote types.Vote
		err := keeper.UnmarshalVote(iterator.Value(), &vote)
		if err != nil {
			panic(err)
		}

		if cb(vote) {
			break
		}
	}
}

// GetVotes returns all the Votes from store
func (keeper Keeper) GetVotes(ctx sdk.Context) (votes []types.Vote) {
	keeper.IterateVotes(ctx, func(vote types.Vote) bool {
		votes = append(votes, vote)
		return false
	})
	return
}

// GetCreatorVotes returns all the Votes in the store from a specific creator
func (keeper Keeper) GetCreatorVotes(ctx sdk.Context, creator string) (votes []types.Vote) {
	keeper.IterateVotes(ctx, func(vote types.Vote) bool {
		if vote.Creator == creator {
			votes = append(votes, vote)
		}
		return false
	})
	return
}

// GetProposalVotes returns all the Votes in the store on a specific proposal
func (keeper Keeper) GetProposalVotes(ctx sdk.Context, hostZoneId string, proposalId uint64) (votes []types.Vote) {
	keeper.IterateVotes(ctx, func(vote types.Vote) bool {
		if vote.HostZoneId == hostZoneId && vote.ProposalId == proposalId {
			votes = append(votes, vote)
		}
		return false
	})
	return
}


func (keeper Keeper) MarshalVote(vote types.Vote) ([]byte, error) {
	bz, err := vote.Marshal()
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func (keeper Keeper) UnmarshalVote(bz []byte, vote *types.Vote) error {
	err := vote.Unmarshal(bz)
	if err != nil {
		return err
	}
	return nil
}


// Helper function to iterate votes and see how much of deposit is available now
func (keeper Keeper) DepositAvailableNow(ctx sdk.Context, creator string, hostZoneId string) (sdk.Int) {
	deposit, _ := keeper.GetDeposit(ctx, creator, hostZoneId)
	votes := keeper.GetCreatorVotes(ctx, creator)
	maxVoteAmount := sdk.ZeroInt()
	for _, vote := range votes {
		if vote.HostZoneId == hostZoneId && vote.Amount.GT(maxVoteAmount) {
			maxVoteAmount = vote.Amount
		}
	}
	return deposit.Amount.Sub(maxVoteAmount)
}

// Helper function to tally current votes on a given proposal
func (keeper Keeper) TallyVotes(ctx sdk.Context, hostZoneId string, proposalId uint64) (map[govtypes.VoteOption]sdk.Int) {
	votes := keeper.GetProposalVotes(ctx, hostZoneId, proposalId)
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "All votes loaded: %+v", votes))	

	var tally = make(map[govtypes.VoteOption]sdk.Int)
	for _, vote := range votes {
		keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Vote found: %+v", vote))		
		currCount, found := tally[vote.Option]
		if !found {
			currCount = sdk.ZeroInt()
		}
		keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "CurrCount is: %d", currCount.Int64()))
		tally[vote.Option] = currCount.Add(vote.Amount)
		keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Tally is: %+v", tally))		
	}
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Tally complete: %+v", tally))
	return tally
}

// Helper function to create a single weighted vote type representing all combined individual votes
func (keeper Keeper) ScaleVotes(ctx sdk.Context, hostZoneId string, proposalId uint64) (weightedVote govtypes.WeightedVoteOptions) {
	hostZone, _ := keeper.stakeibcKeeper.GetHostZone(ctx, hostZoneId)
	stTotal := (sdk.NewDecFromInt(hostZone.StakedBal).Quo(hostZone.RedemptionRate)).TruncateInt()
	remains := sdk.NewInt(stTotal.Int64())

	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "stakedBal %+v RR %+v", hostZone.StakedBal, hostZone.RedemptionRate))	
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Total stTokens: %d", stTotal.Int64()))	

	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "About to tally"))
	tally := keeper.TallyVotes(ctx, hostZoneId, proposalId)
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Tally is: %+v", tally))	
	for option, count := range tally {
		//voteOption, _ := govtypes.VoteOptionFromString(govtypes.VoteOption_name[int32(optionIdx)])
		if option == govtypes.OptionAbstain {
			continue
		}
		scaledOption := govtypes.WeightedVoteOption{Option: option, Weight: sdk.NewDecFromInt(count).Quo(sdk.NewDecFromInt(stTotal))}
		weightedVote = append(weightedVote, scaledOption)
		remains = remains.Sub(count)
	}
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "After loop weighted vote is: %+v", weightedVote))

	// Default is to abstain for all the stTokens which exist but were not used to liquid vote
	abstainOption := govtypes.WeightedVoteOption{Option: govtypes.OptionAbstain, Weight: sdk.NewDecFromInt(remains).Quo(sdk.NewDecFromInt(stTotal))}
	weightedVote = append(weightedVote, abstainOption)

	sort.Slice(weightedVote, func(i, j int) bool {
		return weightedVote[i].Option < weightedVote[j].Option;
	})
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Finally weighted vote is: %+v", weightedVote))

	return weightedVote
}

// Initiate the ICA which casts the current weighted vote on the hub for given proposal
func (keeper Keeper) CastVotes(ctx sdk.Context, hostZoneId string, proposalId uint64) error {
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "About to cast votes for proposal %d:", proposalId))

	hostZone, found := keeper.stakeibcKeeper.GetHostZone(ctx, hostZoneId)
	if !found {
		return errorsmod.Wrapf(stakeibctypes.ErrHostZoneNotFound, "no registered zone for queried hostZoneId (%s)", hostZoneId)
	}
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Found host zone %+v:", hostZone))	

	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		return errorsmod.Wrapf(stakeibctypes.ErrICAAccountNotFound, "no delegation account found for %s", hostZoneId)
	}
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Delegation address is %+v:", delegationAccount.Address))

	// compute the scaled votes for each option type and build a weighted vote message
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "About to tally and scale the votes"))	
	weightedVotes := keeper.ScaleVotes(ctx, hostZoneId, proposalId)
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Weighted votes %+v:", weightedVotes))

	var msgs []proto.Message
	msgs = append(msgs, &govtypes.MsgVoteWeighted{
		ProposalId: proposalId,
		Voter: delegationAccount.Address,
		Options: weightedVotes,
	})

	castVotesCallback := types.CastVotesCallback{
		HostZoneId: hostZoneId,
		ProposalId: proposalId,
		Options: weightedVotes,
	}
	keeper.Logger(ctx).Info(utils.LogWithHostZone(hostZoneId, "Marshalling CastVotesCallback args: %+v", castVotesCallback))
	marshalledCallbackArgs, err := keeper.MarshalCastVotesCallbackArgs(ctx, castVotesCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx on the Stride Epoch
	_, err = keeper.stakeibcKeeper.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, *delegationAccount, ICACallbackID_CastVotes, marshalledCallbackArgs)
	if err != nil {
		return errorsmod.Wrapf(stakeibctypes.ErrICATxFailed, "Failed to SubmitTxs, Messages: %v, err: %s", msgs, err.Error())
	}

	return nil
}
