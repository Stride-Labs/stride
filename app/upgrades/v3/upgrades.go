package v3

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimkeeper "github.com/Stride-Labs/stride/v9/x/claim/keeper"
	"github.com/Stride-Labs/stride/v9/x/claim/types"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName         = "v3"
	airdropDistributors = []string{
		"stride1cpvl8yf848karqauyhr5jzw6d9n9lnuuu974ev",
		"stride1fmh0ysk5nt9y2cj8hddms5ffj2dhys55xkkjwz",
		"stride1zlu2l3lx5tqvzspvjwsw9u0e907kelhqae3yhk",
		"stride14k9g9zpgaycpey9840nnpa66l4nd6lu7g7t74c",
		"stride12pum4adk5dhp32d90f8g8gfwujm0gwxqnrdlum",
	}
	airdropIdentifiers = []string{"stride", "gaia", "osmosis", "juno", "stars"}
	airdropChainIds    = []string{"stride-1", "cosmoshub-4", "osmosis-1", "juno-1", "stargaze-1"}
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
				err = ck.CreateAirdropAndEpoch(ctx, types.MsgCreateAirdrop{
					Distributor: airdropDistributor,
					Identifier:  airdropIdentifiers[idx],
					ChainId:     airdropChainIds[idx],
					Denom:       claimtypes.DefaultClaimDenom,
					StartTime:   uint64(ctx.BlockTime().Unix()),
					Duration:    uint64(airdropDuration.Seconds()),
				})
				if err != nil {
					return newVm, err
				}
			}
		}
		ck.LoadAllocationData(ctx, allocations)
		return newVm, nil
	}
}
