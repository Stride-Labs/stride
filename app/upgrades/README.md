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
```go
// x/{moduleName}/module.go
func (AppModule) ConsensusVersion() uint64 { return 2 }
```

## Specify the Migration in the Upgrade Handler
```go
// app/upgrades/{upgradeVersion}/upgrades.go

// CreateUpgradeHandler creates an SDK upgrade handler for {upgradeVersion}
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm[{moduleName}] = 2 // <- ADD THIS
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
```

## Add Migration Handler
```go
// x/{moduleName}/keeper/migrations.go

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

    {upgradeVersion} "github.com/Stride-Labs/stride/v3/x/records/migrations/{upgradeVersion}"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return {upgradeVersion}.MigrateStore(ctx)
}
```

## Define Migration Logic
```go
// x/{moduleName}/migrations/{upgradeVersion}/migrations.go
package {upgradeVersion}

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	{oldVersion} "github.com/Stride-Labs/stride/v3/x/records/migrations/{oldVersion}"
)

// TODO: Add migration logic to deserialize with old protos and re-serialize with new ones
func MigrateStore(ctx sdk.Context) error {
	store := ctx.KVStore(storeKey)
    ...
}
```

## Register Migration Handler
```go
// x/{moduleName}/module.go
...
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
	migrator := keeper.NewMigrator(am.keeper)

	if err := cfg.RegisterMigration(types.ModuleName, 1, migrator.Migrate1to2); err != nil {
		panic(fmt.Errorf("failed to migrate %s to {upgradeVersion}: %w", types.ModuleName, err))
	}
}
```