package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v10/x/liquidgov/types"
)

// GetProposal gets a proposal from store by ProposalID.
// Panics if can't unmarshal the proposal.
func (keeper Keeper) GetVote(ctx sdk.Context, creator sdk.AccAddress, host_chain_id string, proposal_id uint64) (types.Vote, bool) {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.VoteKey(creator.String(), host_chain_id, proposal_id))
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
}

// DeleteVote deletes a Vote from store.
// Panics if the Vote doesn't exist.
func (keeper Keeper) DeleteVote(ctx sdk.Context, creator string, chain_id string, proposal_id uint64) {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(types.VoteKey(creator, chain_id, proposal_id))
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
func (keeper Keeper) DepositAvailableNow(ctx sdk.Context, creator string, host_zone_id string) (sdk.Int) {
	deposit, _ := keeper.GetDeposit(ctx, sdk.AccAddress(creator), host_zone_id)
	votes := keeper.GetCreatorVotes(ctx, creator)
	maxVoteAmount := sdk.ZeroInt()
	for _, vote := range votes {
		if vote.HostZoneId == host_zone_id && vote.Amount.GT(maxVoteAmount) {
			maxVoteAmount = vote.Amount
		}
	}
	return deposit.Amount.Sub(maxVoteAmount)
}
