package v33

import (
	"context"
	"fmt"

	ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// CreateUpgradeHandler returns the v33 upgrade handler that migrates Stride
// from ICS consumer to POA. See docs/superpowers/specs/2026-04-27-ics-to-poa-migration-design.md.
//
// poaKeeper is a pointer because POA's keeper methods (including InitGenesis)
// have pointer receivers; passing by value here would not compile.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	consumerKeeper ccvconsumerkeeper.Keeper,
	poaKeeper *poakeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	distrKeeper distrkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s (ICS → POA)...", UpgradeName))

		// 1. Run module migrations. RunMigrations silently skips modules removed
		//    from the manager (ccvconsumer, ccvdistr, slashing, evidence).
		ctx.Logger().Info("v33: running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		// 2. Snapshot current ICS validator set into POA-shaped Validators.
		ctx.Logger().Info("v33: snapshotting ICS validator set...")
		poaValidators, err := SnapshotValidatorsFromICS(ctx, consumerKeeper)
		if err != nil {
			return nil, err
		}

		// 3. Initialize POA state with that set + admin.
		ctx.Logger().Info("v33: initializing POA state...")
		if err := InitializePOA(ctx, cdc, poaKeeper, AdminMultisigAddress, poaValidators); err != nil {
			return nil, err
		}

		// 4. Sweep residual ICS reward module accounts to community pool.
		ctx.Logger().Info("v33: sweeping ICS module accounts to community pool...")
		if err := SweepICSModuleAccounts(ctx, accountKeeper, bankKeeper, distrKeeper); err != nil {
			return nil, err
		}

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete.", UpgradeName))
		return versionMap, nil
	}
}
