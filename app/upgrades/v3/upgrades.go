package v3

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimKeeper "github.com/Stride-Labs/stride/x/claim/keeper"
	claimTypes "github.com/Stride-Labs/stride/x/claim/types"
)

// Note: ensure these values are properly set before running upgrade
const (
	UpgradeName        = "v3"
	airdropDistributor = ""
	airdropIdentifier  = ""
	airdropDuration    = time.Hour * 24 * 30 * 12 * 3 // 3 years
)

// CreateUpgradeHandler creates an SDK upgrade handler for v3
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ck claimKeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		_, err := ck.GetParams(ctx)
		if err != nil {
			gen := claimTypes.DefaultGenesis()
			ck.SetParams(ctx, gen.Params)
			ck.SetClaimRecordsWithWeights(ctx, gen.ClaimRecords)
		}

		ck.CreateAirdropAndEpoch(ctx, airdropDistributor, claimTypes.DefaultClaimDenom, uint64(ctx.BlockTime().Unix()), uint64(airdropDuration.Seconds()), airdropIdentifier)
		ck.LoadAllocationData(ctx, allocations)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
