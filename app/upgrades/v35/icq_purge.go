package v35

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icqkeeper "github.com/Stride-Labs/stride/v34/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// haqqSlashPathCallbacks are the ICQ callbacks that correct a validator's delegation from
// on-chain shares. A query submitted against the pre-delta state has no reason to exist once
// the delta table is applied, so all of them are thrown out first.
var haqqSlashPathCallbacks = map[string]bool{
	stakeibckeeper.ICQCallbackID_Delegation: true,
	stakeibckeeper.ICQCallbackID_Validator:  true,
	stakeibckeeper.ICQCallbackID_Calibrate:  true,
}

// PurgeHaqqSlashQueries deletes every pending slash-path query for haqq_11235-1 and clears
// SlashQueryInProgress on every haqq validator (spec §5 "Pending ICQs"). Unlike v34's pinned
// query ids this is dynamic, so it cannot go stale between measurement and execution. It
// must run before ReconcileHaqqDelegations.
func PurgeHaqqSlashQueries(ctx sdk.Context, icq icqkeeper.Keeper, sk stakeibckeeper.Keeper) {
	numDeleted := 0
	for _, query := range icq.AllQueries(ctx) {
		// Callback ids are only unique within a module, so the module is part of the match
		if query.CallbackModule != stakeibctypes.ModuleName || query.ChainId != HaqqChainId || !haqqSlashPathCallbacks[query.CallbackId] {
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v35: deleting pending %s ICQ %s for %s", query.CallbackId, query.Id, HaqqChainId))
		icq.DeleteQuery(ctx, query.Id)
		numDeleted++
	}
	ctx.Logger().Info(fmt.Sprintf("v35: %d pending slash-path ICQ(s) deleted for %s", numDeleted, HaqqChainId))

	hostZone, found := sk.GetHostZone(ctx, HaqqChainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v35: host zone %s not found, skipping slash query flag reset", HaqqChainId))
		return
	}
	// Validators are stored as pointers, so clearing the flag updates hostZone in place
	for _, validator := range hostZone.Validators {
		if validator.SlashQueryInProgress {
			ctx.Logger().Info(fmt.Sprintf("v35: clearing SlashQueryInProgress on %s", validator.Address))
			validator.SlashQueryInProgress = false
		}
	}
	sk.SetHostZone(ctx, hostZone)
}

// PurgeWithdrawalBalanceQueries deletes every pending withdrawal-balance query on every
// chain: its callback delegates the withdrawal ICA balance back to validators, and PR 2
// removed the epoch call that submits it (spec §6), so a late answer must not be acted on.
func PurgeWithdrawalBalanceQueries(ctx sdk.Context, icq icqkeeper.Keeper) {
	numDeleted := 0
	for _, query := range icq.AllQueries(ctx) {
		if query.CallbackModule != stakeibctypes.ModuleName || query.CallbackId != stakeibckeeper.ICQCallbackID_WithdrawalHostBalance {
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v35: deleting pending withdrawal-balance ICQ %s for %s", query.Id, query.ChainId))
		icq.DeleteQuery(ctx, query.Id)
		numDeleted++
	}
	ctx.Logger().Info(fmt.Sprintf("v35: %d pending withdrawal-balance ICQ(s) deleted", numDeleted))
}
