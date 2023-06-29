package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v10/x/liquidgov/types"
)

// GetProposal gets a proposal from store by ProposalID.
// Panics if can't unmarshal the proposal.
func (keeper Keeper) GetProposal(ctx sdk.Context, proposalID uint64, chain_id string) (types.Proposal, bool) {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.ProposalKey(proposalID, chain_id))
	if bz == nil {
		return types.Proposal{}, false
	}

	var proposal types.Proposal
	if err := keeper.UnmarshalProposal(bz, &proposal); err != nil {
		panic(err)
	}

	return proposal, true
}

// SetProposal sets a proposal to store.
// Panics if can't marshal the proposal.
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal types.Proposal) {
	bz, err := keeper.MarshalProposal(proposal)
	if err != nil {
		panic(err)
	}

	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.ProposalKey(proposal.GovProposal.ProposalId, proposal.HostZoneId), bz)
}

// DeleteProposal deletes a proposal from store.
// Panics if the proposal doesn't exist.
func (keeper Keeper) DeleteProposal(ctx sdk.Context, proposalID uint64, chain_id string) {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(types.ProposalKey(proposalID, chain_id))
}

// GetProposalID gets the highest ICQ'd proposal ID
func (keeper Keeper) GetProposalID(ctx sdk.Context, chain_id string) (proposalID uint64, err error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.ProposalIDKey(chain_id))
	if bz == nil {
		return 0, sdkerrors.Wrap(govtypes.ErrInvalidGenesis, fmt.Sprintf("initial proposal ID for chain_id: %s hasn't been set", chain_id))
	}

	proposalID = govtypes.GetProposalIDFromBytes(bz)
	return proposalID, nil
}

// SetProposalID sets the new proposal ID to the store
func (keeper Keeper) SetProposalID(ctx sdk.Context, proposalID uint64, chain_id string) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.ProposalIDKey(chain_id), govtypes.GetProposalIDBytes(proposalID))
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

func (keeper Keeper) MarshalProposal(proposal types.Proposal) ([]byte, error) {
	// bz, err := keeper.cdc.Marshal(&proposal)
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
