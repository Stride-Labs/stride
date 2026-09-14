package v34

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icqkeeper "github.com/Stride-Labs/stride/v34/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// ResetStuckSlashQueries clears SlashQueryInProgress on the validators in
// StuckSlashQueryValidators. SubmitDelegationICQ skips any validator with the
// flag set, so without this reset their slash checks never resume.
//
// Missing state is logged and skipped rather than returned as an error, since
// halting the upgrade over this cleanup would be far worse than a no-op.
func ResetStuckSlashQueries(ctx sdk.Context, stakeibcKeeper stakeibckeeper.Keeper) {
	hostZone, found := stakeibcKeeper.GetHostZone(ctx, StuckSlashQueryHostZone)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("v34: host zone %s not found, skipping slash query reset", StuckSlashQueryHostZone))
		return
	}

	validatorsByName := make(map[string]*stakeibctypes.Validator, len(hostZone.Validators))
	for _, validator := range hostZone.Validators {
		validatorsByName[validator.Name] = validator
	}

	// Validators are stored as pointers, so clearing the flag here updates hostZone in place
	for _, name := range StuckSlashQueryValidators {
		validator, found := validatorsByName[name]
		if !found {
			ctx.Logger().Error(fmt.Sprintf("v34: validator %s not found on %s, skipping slash query reset",
				name, StuckSlashQueryHostZone))
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v34: resetting slash query in progress for %s", name))
		validator.SlashQueryInProgress = false
	}
	stakeibcKeeper.SetHostZone(ctx, hostZone)
}

// DeleteStuckQueries removes the ICQs in StuckQueryIds. An ID that no longer
// exists (e.g. the query was answered before the upgrade) is a no-op.
func DeleteStuckQueries(ctx sdk.Context, icqKeeper icqkeeper.Keeper) {
	ctx.Logger().Info(fmt.Sprintf("v34: deleting %d stuck ICQs", len(StuckQueryIds)))
	for _, queryId := range StuckQueryIds {
		icqKeeper.DeleteQuery(ctx, queryId)
	}
}
