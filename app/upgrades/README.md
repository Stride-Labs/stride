# Upgrades

## Increment Version 
```go
// cmd/strided/config/config.go
...
version.Version = "{newVersion}"

// app/app.go
...
Version = "{newVersion}"

// go.mod (will need to update all imports after)
module github.com/Stride-Labs/stride/{newVersion}
```

## Create Upgrade Handler
```go
// app/upgrades/{upgradeVersion}/upgrades.go

package {upgradeVersion}

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	UpgradeName = "{upgradeVersion}"
)

// CreateUpgradeHandler creates an SDK upgrade handler for {upgradeVersion}
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
```

## Register Upgrade Handler
```go
// app/upgrades.go

package app

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (app *StrideApp) setupUpgradeHandlers() {
	// {upgradeVersion} upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		{upgradeVersion}.UpgradeName,
		{upgradeVersion}.CreateUpgradeHandler(app.mm, app.configurator),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("Failed to read upgrade info from disk: %w", err))
	}
    ...

	// If adding a new module, add the new store keys
	switch upgradeInfo.Name {
	...
	case {upgradeVersion}:
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{newmoduletypes.StoreKey},
		}
	}
```

# Migrations (Only required if the state changed)
## Store Old Proto Types
```go
// x/{moduleName}/migrations/{oldVersion}/types/{data_type}.pb.go
```

## Increment the Module's Consensus Version
* The consensus version is different from the chain version - it is specific to each module and is incremented every time state is migrated
```go
// x/{moduleName}/module.go
func (AppModule) ConsensusVersion() uint64 { return 2 }
```

## Define Migration Logic
```go
// x/{moduleName}/migrations/{new-consensus-version}/migrations.go
package {upgradeVersion}

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	{new-consensus-version} "github.com/Stride-Labs/stride/v9/x/records/migrations/{new-consensus-version}"
)

// TODO: Add migration logic to deserialize with old protos and re-serialize with new ones
func MigrateStore(ctx sdk.Context) error {
	store := ctx.KVStore(storeKey)
    ...
}
```

## Specify the Migration in the Upgrade Handler
```go
// app/upgrades/{upgradeVersion}/upgrades.go

import (
	{module}migration "github.com/Stride-Labs/stride/v9/x/{module}/migrations/{new-consensus-version}"
)

// CreateUpgradeHandler creates an SDK upgrade handler for {upgradeVersion}
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	{module}StoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		if err := {module}migration.MigrateStore(ctx, {module}StoreKey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate {module} store")
		}
		vm[{moduleName}] = mm.GetVersionMap()[{moduleName}] 
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
```

## Add Additional Parameters to `CreateUpgradeHandler` Invocation 
```go
// app/upgrades.go
	...
		{upgradeVersion}.CreateUpgradeHandler(app.mm, app.configurator, app.appCodec, app.{module}Keeper),
	...
```