package app

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	v2 "github.com/Stride-Labs/stride/v4/app/upgrades/v2"
	v3 "github.com/Stride-Labs/stride/v4/app/upgrades/v3"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
)

func (app *StrideApp) setupUpgradeHandlers() {
	// v2 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v2.UpgradeName,
		v2.CreateUpgradeHandler(app.mm, app.configurator),
	)

	// v3 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v3.UpgradeName,
		v3.CreateUpgradeHandler(app.mm, app.configurator, app.ClaimKeeper),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("Failed to read upgrade info from disk: %w", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	var storeUpgrades *storetypes.StoreUpgrades

	switch upgradeInfo.Name {
	// no store upgrades
	case "v3":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{claimtypes.StoreKey},
		}
	}

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
