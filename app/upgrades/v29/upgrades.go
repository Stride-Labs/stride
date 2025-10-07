package v29

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consumerkeeper "github.com/cosmos/interchain-security/v6/x/ccv/consumer/keeper"

	strdburnerkeeper "github.com/Stride-Labs/stride/v29/x/strdburner/keeper"
	"github.com/Stride-Labs/stride/v29/x/strdburner/types"
)

var (
	UpgradeName = "v29"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v29
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	consumerKeeper consumerkeeper.Keeper,
	strdburnerKeeper strdburnerkeeper.Keeper,
	strdburnerStoreKey storetypes.StoreKey,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s...", UpgradeName))

		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Updating CCV params...")
		DisableCcvRewards(ctx, consumerKeeper)

		ctx.Logger().Info("Migrating burner totals...")
		MigrateBurnerTotals(ctx, strdburnerStoreKey, strdburnerKeeper)

		return versionMap, nil
	}
}

// Set consumer redistribution fraction to 100% stop sending CCV rewards
func DisableCcvRewards(ctx sdk.Context, ck consumerkeeper.Keeper) {
	params := ck.GetConsumerParams(ctx)
	params.ConsumerRedistributionFraction = "1.0"
	ck.SetParams(ctx, params)
}

// Split the burner total into protocol + user
func MigrateBurnerTotals(ctx sdk.Context, storeKey storetypes.StoreKey, bk strdburnerkeeper.Keeper) {
	// Read the total protocol burn from the legacy store
	store := ctx.KVStore(storeKey)
	protocolBurnedAmountBz := store.Get([]byte(types.TotalStrdBurnedKey))

	protocolBurnedAmount := sdkmath.ZeroInt()
	if protocolBurnedAmountBz != nil {
		protocolBurnedAmount = sdkmath.NewIntFromUint64(sdk.BigEndianToUint64(protocolBurnedAmountBz))
	}

	// Set the total amount as the "protocol" amount, and set the total user amount to 0
	bk.SetProtocolStrdBurned(ctx, protocolBurnedAmount)
	bk.SetTotalUserStrdBurned(ctx, sdkmath.ZeroInt())
}
