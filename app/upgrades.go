package app

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	authz "github.com/cosmos/cosmos-sdk/x/authz"

	v2 "github.com/Stride-Labs/stride/v9/app/upgrades/v2"
	v3 "github.com/Stride-Labs/stride/v9/app/upgrades/v3"
	v4 "github.com/Stride-Labs/stride/v9/app/upgrades/v4"
	v5 "github.com/Stride-Labs/stride/v9/app/upgrades/v5"
	v6 "github.com/Stride-Labs/stride/v9/app/upgrades/v6"
	v7 "github.com/Stride-Labs/stride/v9/app/upgrades/v7"
	v8 "github.com/Stride-Labs/stride/v9/app/upgrades/v8"
	v9 "github.com/Stride-Labs/stride/v9/app/upgrades/v9"
	autopilottypes "github.com/Stride-Labs/stride/v9/x/autopilot/types"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	ratelimittypes "github.com/Stride-Labs/stride/v9/x/ratelimit/types"
	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
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

	// v6 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v6.UpgradeName,
		v6.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.appCodec,
			app.ClaimKeeper,
		),
	)

	// v7 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v7.UpgradeName,
		v7.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.appCodec,
			app.AccountKeeper,
			app.BankKeeper,
			app.EpochsKeeper,
			app.ICAHostKeeper,
			app.MintKeeper,
			app.StakeibcKeeper,
		),
	)

	// v8 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v8.UpgradeName,
		v8.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.appCodec,
			app.ClaimKeeper,
			app.AutopilotKeeper,
		),
	)

	// v9 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v9.UpgradeName,
		v9.CreateUpgradeHandler(app.mm, app.configurator, app.ClaimKeeper),
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
	case "v7":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{ratelimittypes.StoreKey, autopilottypes.StoreKey},
		}
	}

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
