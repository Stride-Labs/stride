package app

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	authz "github.com/cosmos/cosmos-sdk/x/authz"

	v2 "github.com/Stride-Labs/stride/v4/app/upgrades/v2"
	v3 "github.com/Stride-Labs/stride/v4/app/upgrades/v3"
	v4 "github.com/Stride-Labs/stride/v4/app/upgrades/v4"
	v5 "github.com/Stride-Labs/stride/v4/app/upgrades/v5"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
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

	// v4 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v4.UpgradeName,
		v4.CreateUpgradeHandler(app.mm, app.configurator),
	)

	// v5 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v5.UpgradeName,
		v5.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.appCodec,
			app.InterchainqueryKeeper,
			app.StakeibcKeeper,
			app.keys[claimtypes.StoreKey],
			app.keys[icacallbacktypes.StoreKey],
			app.keys[recordtypes.StoreKey],
			app.keys[stakeibctypes.StoreKey],
		),
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
	case "v3":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{claimtypes.StoreKey},
		}
	case "v5":
		storeUpgrades = &storetypes.StoreUpgrades{
			Deleted: []string{authz.ModuleName},
		}
	}

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
