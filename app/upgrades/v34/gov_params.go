package v34

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
)

// UpdateGovParams sets the quorum to GovQuorum and the voting period to
// GovVotingPeriod, leaving every other gov param unchanged.
func UpdateGovParams(ctx context.Context, govKeeper govkeeper.Keeper) error {
	params, err := govKeeper.Params.Get(ctx)
	if err != nil {
		return errorsmod.Wrap(err, "failed to get gov params")
	}

	votingPeriod := GovVotingPeriod
	params.Quorum = GovQuorum
	params.VotingPeriod = &votingPeriod

	// Params.Set doesn't validate, and an invalid combination (e.g. an expedited
	// voting period that's no longer shorter than the voting period) would
	// break every future proposal
	if err := params.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "invalid gov params")
	}
	if err := govKeeper.Params.Set(ctx, params); err != nil {
		return errorsmod.Wrap(err, "failed to set gov params")
	}

	return nil
}
