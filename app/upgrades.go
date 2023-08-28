package app

import (
	"fmt"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	consumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	evmosvestingtypes "github.com/evmos/vesting/x/vesting/types"

	v10 "github.com/Stride-Labs/stride/v14/app/upgrades/v10"
	v11 "github.com/Stride-Labs/stride/v14/app/upgrades/v11"
	v12 "github.com/Stride-Labs/stride/v14/app/upgrades/v12"
	v13 "github.com/Stride-Labs/stride/v14/app/upgrades/v13"
	v14 "github.com/Stride-Labs/stride/v14/app/upgrades/v14"
	v2 "github.com/Stride-Labs/stride/v14/app/upgrades/v2"
	v3 "github.com/Stride-Labs/stride/v14/app/upgrades/v3"
	v4 "github.com/Stride-Labs/stride/v14/app/upgrades/v4"
	v5 "github.com/Stride-Labs/stride/v14/app/upgrades/v5"
	v6 "github.com/Stride-Labs/stride/v14/app/upgrades/v6"
	v7 "github.com/Stride-Labs/stride/v14/app/upgrades/v7"
	v8 "github.com/Stride-Labs/stride/v14/app/upgrades/v8"
	v9 "github.com/Stride-Labs/stride/v14/app/upgrades/v9"
	autopilottypes "github.com/Stride-Labs/stride/v14/x/autopilot/types"
	claimtypes "github.com/Stride-Labs/stride/v14/x/claim/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	icaoracletypes "github.com/Stride-Labs/stride/v14/x/icaoracle/types"
	ratelimittypes "github.com/Stride-Labs/stride/v14/x/ratelimit/types"
	recordtypes "github.com/Stride-Labs/stride/v14/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func (app *StrideApp) setupUpgradeHandlers(appOpts servertypes.AppOptions) {
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
			app.keys[stakeibctypes.StoreKey],
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

	// v10 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v10.UpgradeName,
		v10.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.appCodec,
			app.keys[capabilitytypes.ModuleName],
			app.AccountKeeper,
			app.BankKeeper,
			app.CapabilityKeeper,
			app.IBCKeeper.ChannelKeeper,
			app.ClaimKeeper,
			app.IBCKeeper.ClientKeeper,
			app.ConsensusParamsKeeper,
			app.GovKeeper,
			app.IcacallbacksKeeper,
			app.MintKeeper,
			app.ParamsKeeper,
			app.RatelimitKeeper,
			app.StakeibcKeeper,
		),
	)

	// v11 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v11.UpgradeName,
		v11.CreateUpgradeHandler(
			app.mm,
			app.configurator,
		),
	)

	// v12 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v12.UpgradeName,
		v12.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.appCodec,
			appOpts,
			*app.IBCKeeper,
			&app.ConsumerKeeper,
			app.StakingKeeper,
		),
	)

	// v13 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v13.UpgradeName,
		v13.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.StakeibcKeeper,
		),
	)
	// v14 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v14.UpgradeName,
		v14.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.appCodec,
			app.AccountKeeper,
			app.BankKeeper,
			app.ClaimKeeper,
			&app.ConsumerKeeper,
			app.InterchainqueryKeeper,
			app.StakeibcKeeper,
			app.StakingKeeper,
			app.VestingKeeper,
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
	case "v7":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{ratelimittypes.StoreKey, autopilottypes.StoreKey},
		}
	case "v10":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{crisistypes.StoreKey, consensustypes.StoreKey},
		}
	case "v12":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{consumertypes.ModuleName},
		}
	case "v13":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{icaoracletypes.ModuleName},
		}
	case "v14":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{evmosvestingtypes.ModuleName},
		}
	}

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
