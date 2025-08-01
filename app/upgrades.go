package app

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/types"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	consumertypes "github.com/cosmos/interchain-security/v4/x/ccv/consumer/types"
	evmosvestingtypes "github.com/evmos/vesting/x/vesting/types"

	v10 "github.com/Stride-Labs/stride/v27/app/upgrades/v10"
	v11 "github.com/Stride-Labs/stride/v27/app/upgrades/v11"
	v12 "github.com/Stride-Labs/stride/v27/app/upgrades/v12"
	v13 "github.com/Stride-Labs/stride/v27/app/upgrades/v13"
	v14 "github.com/Stride-Labs/stride/v27/app/upgrades/v14"
	v15 "github.com/Stride-Labs/stride/v27/app/upgrades/v15"
	v16 "github.com/Stride-Labs/stride/v27/app/upgrades/v16"
	v17 "github.com/Stride-Labs/stride/v27/app/upgrades/v17"
	v18 "github.com/Stride-Labs/stride/v27/app/upgrades/v18"
	v19 "github.com/Stride-Labs/stride/v27/app/upgrades/v19"
	v2 "github.com/Stride-Labs/stride/v27/app/upgrades/v2"
	v20 "github.com/Stride-Labs/stride/v27/app/upgrades/v20"
	v21 "github.com/Stride-Labs/stride/v27/app/upgrades/v21"
	v22 "github.com/Stride-Labs/stride/v27/app/upgrades/v22"
	v23 "github.com/Stride-Labs/stride/v27/app/upgrades/v23"
	v24 "github.com/Stride-Labs/stride/v27/app/upgrades/v24"
	v25 "github.com/Stride-Labs/stride/v27/app/upgrades/v25"
	v26 "github.com/Stride-Labs/stride/v27/app/upgrades/v26"
	v27 "github.com/Stride-Labs/stride/v27/app/upgrades/v27"
	v28 "github.com/Stride-Labs/stride/v27/app/upgrades/v28"
	v3 "github.com/Stride-Labs/stride/v27/app/upgrades/v3"
	v4 "github.com/Stride-Labs/stride/v27/app/upgrades/v4"
	v5 "github.com/Stride-Labs/stride/v27/app/upgrades/v5"
	v6 "github.com/Stride-Labs/stride/v27/app/upgrades/v6"
	v7 "github.com/Stride-Labs/stride/v27/app/upgrades/v7"
	v8 "github.com/Stride-Labs/stride/v27/app/upgrades/v8"
	v9 "github.com/Stride-Labs/stride/v27/app/upgrades/v9"
	airdroptypes "github.com/Stride-Labs/stride/v27/x/airdrop/types"
	auctiontypes "github.com/Stride-Labs/stride/v27/x/auction/types"
	autopilottypes "github.com/Stride-Labs/stride/v27/x/autopilot/types"
	claimtypes "github.com/Stride-Labs/stride/v27/x/claim/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v27/x/icacallbacks/types"
	icaoracletypes "github.com/Stride-Labs/stride/v27/x/icaoracle/types"
	icqoracletypes "github.com/Stride-Labs/stride/v27/x/icqoracle/types"
	recordtypes "github.com/Stride-Labs/stride/v27/x/records/types"
	stakedymtypes "github.com/Stride-Labs/stride/v27/x/stakedym/types"
	stakeibctypes "github.com/Stride-Labs/stride/v27/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v27/x/staketia/types"
	strdburnertypes "github.com/Stride-Labs/stride/v27/x/strdburner/types"
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

	// v15 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v15.UpgradeName,
		v15.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.InterchainqueryKeeper,
			app.StakeibcKeeper,
		),
	)

	// v16 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v16.UpgradeName,
		v16.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.StakeibcKeeper,
			app.RatelimitKeeper,
		),
	)

	// v17 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v17.UpgradeName,
		v17.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.BankKeeper,
			app.DistrKeeper,
			app.InterchainqueryKeeper,
			app.RatelimitKeeper,
			app.StakeibcKeeper,
		),
	)

	// v18 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v18.UpgradeName,
		v18.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.BankKeeper,
			app.GovKeeper,
			app.RecordsKeeper,
			app.StakeibcKeeper,
		),
	)

	// v19 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v19.UpgradeName,
		v19.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.appCodec,
			app.RatelimitKeeper,
			app.WasmKeeper,
		),
	)

	// v20 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v20.UpgradeName,
		v20.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.ConsumerKeeper,
			app.StakeibcKeeper,
		),
	)

	// v21 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v21.UpgradeName,
		v21.CreateUpgradeHandler(
			app.mm,
			app.configurator,
		),
	)

	// v22 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v22.UpgradeName,
		v22.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.StakeibcKeeper,
		),
	)

	// v23 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v23.UpgradeName,
		v23.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.IBCKeeper.ClientKeeper,
			app.RecordsKeeper,
			app.StakeibcKeeper,
		),
	)

	// v24 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v24.UpgradeName,
		v24.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.BankKeeper,
			app.RecordsKeeper,
			app.StakeibcKeeper,
		),
	)

	// v25 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v25.UpgradeName,
		v25.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.BankKeeper,
			app.RecordsKeeper,
			app.StakeibcKeeper,
			app.StaketiaKeeper,
		),
	)

	// v26 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v26.UpgradeName,
		v26.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.ICQOracleKeeper,
		),
	)

	// v27 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v27.UpgradeName,
		v27.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.StakeibcKeeper,
		),
	)

	// v28 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v28.UpgradeName,
		v28.CreateUpgradeHandler(
			app.mm,
			app.configurator,
			app.StakeibcKeeper,
			app.AccountKeeper,
			app.BankKeeper,
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
	case "v17":
		storeUpgrades = &storetypes.StoreUpgrades{
			// Add PFM store key
			Added: []string{packetforwardtypes.ModuleName},
		}
	case "v18":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{staketiatypes.ModuleName},
		}
	case "v19":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{wasmtypes.ModuleName, stakedymtypes.ModuleName},
		}
	case "v22":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{ibchookstypes.StoreKey},
		}
	case "v23":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{ibcwasmtypes.ModuleName, airdroptypes.ModuleName},
		}
	case "v26":
		storeUpgrades = &storetypes.StoreUpgrades{
			Added: []string{icqoracletypes.ModuleName, strdburnertypes.ModuleName, auctiontypes.ModuleName},
		}
	}

	if storeUpgrades != nil {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, storeUpgrades))
	}
}
