package app

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	v2 "github.com/Stride-Labs/stride/v4/app/upgrades/v2"
	v3 "github.com/Stride-Labs/stride/v4/app/upgrades/v3"
	v4 "github.com/Stride-Labs/stride/v4/app/upgrades/v4"
	v5 "github.com/Stride-Labs/stride/v4/app/upgrades/v5"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AuthzHeightAdjustmentUpgradeStoreLoader is used to delete the authz store with the
// wrong height and then re-add authz store with the right height
func AuthzHeightAdjustmentUpgradeStoreLoader(upgradeHeight int64) baseapp.StoreLoader {
	return func(ms sdk.CommitMultiStore) error {
		if upgradeHeight == ms.LastCommitID().Version+1 {
			err := ms.LoadLatestVersionAndUpgrade(&storetypes.StoreUpgrades{
				Deleted: []string{authzkeeper.StoreKey},
			})
			if err != nil {
				panic(err)
			}
		}
		// Otherwise load default store loader
		return baseapp.DefaultStoreLoader(ms)
	}
}

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
		v5.CreateUpgradeHandler(app.mm, app.configurator, app.InterchainqueryKeeper),
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
		app.SetStoreLoader(AuthzHeightAdjustmentUpgradeStoreLoader(upgradeInfo.Height))
	}

	if (storeUpgrades != nil) && (upgradeInfo.Name != "v5") {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
