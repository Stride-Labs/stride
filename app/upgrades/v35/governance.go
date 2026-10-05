package v35

import (
	"fmt"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// RejectPendingGovProposals closes existing proposals before undelegation can concentrate
// voting power in skipped delegations. Queue removal is required: changing the status alone
// does not prevent the gov EndBlocker from tallying and executing a queued proposal.
func RejectPendingGovProposals(ctx sdk.Context, k govkeeper.Keeper) error {
	iterator, err := k.Proposals.Iterate(ctx, nil)
	if err != nil {
		return err
	}

	// Values closes the iterator before the proposal collection is modified.
	proposals, err := iterator.Values()
	if err != nil {
		return err
	}

	rejected := 0
	for _, proposal := range proposals {
		if proposal.Status != govv1.StatusDepositPeriod && proposal.Status != govv1.StatusVotingPeriod {
			continue
		}
		if err := k.RefundAndDeleteDeposits(ctx, proposal.Id); err != nil {
			return fmt.Errorf("v35: refund proposal %d deposits: %w", proposal.Id, err)
		}
		if err := k.Votes.Clear(ctx, collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposal.Id)); err != nil {
			return err
		}

		// Reuse SDK queue cleanup, then retain the terminal record for proposal history.
		if err := k.DeleteProposal(ctx, proposal.Id); err != nil {
			return err
		}
		proposal.Status = govv1.StatusRejected
		proposal.FailedReason = "Governance closed by v35 wind-down"
		if err := k.SetProposal(ctx, proposal); err != nil {
			return err
		}
		rejected++
	}

	ctx.Logger().Info(fmt.Sprintf("v35: rejected %d pending governance proposals and refunded their deposits", rejected))
	return nil
}
