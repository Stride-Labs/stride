package v17

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icqkeeper "github.com/Stride-Labs/stride/v16/x/interchainquery/keeper"
	ratelimitkeeper "github.com/Stride-Labs/stride/v16/x/ratelimit/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
)

var (
	UpgradeName = "v17"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v15
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	icqKeeper icqkeeper.Keeper,
	ratelimitKeeper ratelimitkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v17...")

		ctx.Logger().Info("Migrating host zones...")

		ctx.Logger().Info("Deleting all pending queries...")

		ctx.Logger().Info("Reseting slash query in progress...")

		ctx.Logger().Info("Updating community pool tax...")

		ctx.Logger().Info("Updating redemption rate bounds...")

		ctx.Logger().Info("Adding rate limits to Osmosis...")

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Migrates the host zones to the new structure which supports community pool liquid staking
// We don't have to perform a true migration here since only new fields were added
// This will also register the relevant community pool ICA addresses
func MigrateHostZones() {

}

// Deletes all currently active queries (to remove ones that were stuck)
func DeleteAllPendingQueries() {

}

// Resets the slash query in progress flag for each validator
func ResetSlashQueryInProgress() {

}

// Increases the community pool tax from 2 to 5%
// This was from prop XXX which passed, but was deleted due to an ICS blacklist
func IncreaseCommunityPoolTax() {

}

// Updates the outer redemption rate bounds
func UpdatingRedemptionRateBounds() {

}

// Rate limits transfers to osmosis and deletes the injective rate limit
func UpdateRateLimits() {

}
