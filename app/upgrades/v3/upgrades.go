package v3

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimkeeper "github.com/Stride-Labs/stride/x/claim/keeper"
	claimtypes "github.com/Stride-Labs/stride/x/claim/types"
)

// Note: ensure these values are properly set before running upgrade
const (
	UpgradeName        = "v3"
	airdropDistributor = "stride1htlzesdgm6pvunstrgk3p7aj6s66j08f9v3rgl"
	airdropIdentifier  = "osmosis"
	airdropDuration    = time.Hour * 24 * 30 * 12 * 3 // 3 years
)

// CreateUpgradeHandler creates an SDK upgrade handler for v3
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ck claimkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		newVm, err := mm.RunMigrations(ctx, configurator, vm)
		err1 := ck.CreateAirdropAndEpoch(ctx, airdropDistributor, claimtypes.DefaultClaimDenom, uint64(ctx.BlockTime().Unix()), uint64(airdropDuration.Seconds()), airdropIdentifier)
		if err1 != nil {
			panic(err1)
		}
		ck.LoadAllocationData(ctx, allocations)
		return newVm, err
	}
}
