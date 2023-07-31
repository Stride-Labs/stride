package keeper

import (
	//"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
)

// GetProposal gets a proposal from store by ProposalID.
// Panics if can't unmarshal the proposal.
func (keeper Keeper) GetProposal(ctx sdk.Context, chainId string, proposalID uint64) (*types.Proposal, bool) {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.ProposalKey(chainId, proposalID))
	if bz == nil {
		return &types.Proposal{}, false
	}

	var proposal types.Proposal
	if err := keeper.UnmarshalProposal(bz, &proposal); err != nil {
		panic(err)
	}

	return &proposal, true
}

// SetProposal sets a proposal to store.
// Panics if can't marshal the proposal.
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal types.Proposal) {
	//keeper.Logger(ctx).Info(fmt.Sprintf("About to marshall Proposal %+v", proposal))	
	bz, err := keeper.MarshalProposal(&proposal)
	if err != nil {
		//keeper.Logger(ctx).Error(fmt.Sprintf("Error marshalling Proposal %+v", err))
		panic(err)
	}

	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.ProposalKey(proposal.HostZoneId, proposal.GovProposal.ProposalId), bz)
}

// DeleteProposal deletes a proposal from store.
// Panics if the proposal doesn't exist.
func (keeper Keeper) DeleteProposal(ctx sdk.Context, chainId string, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(types.ProposalKey(chainId, proposalID))
}

// IterateProposals iterates over the all the proposals and performs a callback function.
// Panics when the iterator encounters a proposal which can't be unmarshaled.
func (keeper Keeper) IterateProposals(ctx sdk.Context, cb func(proposal types.Proposal) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.ProposalsKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var proposal types.Proposal
		err := keeper.UnmarshalProposal(iterator.Value(), &proposal)
		if err != nil {
			panic(err)
		}

		if cb(proposal) {
			break
		}
	}
}

// GetProposals returns all the proposals from store
func (keeper Keeper) GetProposals(ctx sdk.Context) (proposals types.Proposals) {
	keeper.IterateProposals(ctx, func(proposal types.Proposal) bool {
		proposals = append(proposals, proposal)
		return false
	})
	return
}

func (keeper Keeper) MarshalProposal(proposal *types.Proposal) ([]byte, error) {
	bz, err := proposal.Marshal()
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func (keeper Keeper) UnmarshalProposal(bz []byte, proposal *types.Proposal) error {
	// err := keeper.cdc.Unmarshal(bz, proposal)
	err := proposal.Unmarshal(bz)
	if err != nil {
		return err
	}
	return nil
}
