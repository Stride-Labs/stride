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
var (
	UpgradeName         = "v3"
	airdropDistributors = []string{
		"stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l",
		"stride104az7rd5yh3p8qn4ary8n3xcwquuwgee4vnvvc",
		"stride17kvetgthwt6caku5qjs2rx2njgh26vmg448r5u",
		"stride1swvv9kpp75e60pvlv5x6mcw5f54qgpph239e5s",
		"stride1ywrhas3ae7z3ljqxmgdzjx8wyaf3djwuh4hdlj",
	}
	airdropIdentifiers = []string{"stride", "gaia", "osmosis", "juno", "stars"}
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
		if err != nil {
			return newVm, err
		}

		// total number of airdrop distributors must be equal to identifiers
		if len(airdropDistributors) == len(airdropIdentifiers) {
			for idx, airdropDistributor := range airdropDistributors {
				err = ck.CreateAirdropAndEpoch(ctx, airdropDistributor, claimtypes.DefaultClaimDenom, uint64(ctx.BlockTime().Unix()), uint64(airdropDuration.Seconds()), airdropIdentifiers[idx])
				if err != nil {
					return newVm, err
				}
			}
		}
		ck.LoadAllocationData(ctx, allocations)
		return newVm, nil
	}
}
