package v10

import (
	"fmt"
	stdlog "log"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	consumertypes "github.com/cosmos/interchain-security/x/ccv/consumer/types"

	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/x/ccv/consumer/keeper"
)

var (
	UpgradeName             = "v07-Theta"
	EvmosAirdropDistributor = "stride10dy5pmc2fq7fnmufjfschkfrxaqnpykl6ezy5j"
	EvmosAirdropIdentifier  = "evmos"
	AirdropDuration         = time.Hour * 24 * 30 * 12 * 3 // 3 years
	ResetAirdropIdentifiers = []string{"stride", "gaia", "osmosis", "juno", "stars"}
	AirdropStartTime        = time.Date(2023, 4, 3, 16, 0, 0, 0, time.UTC) // April 3, 2023 @ 16:00 UTC (12:00 EST)
)

// CreateUpgradeHandler creates an SDK upgrade handler for v10
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	ibcKeeper ibckeeper.Keeper,
	consumerKeeper *ccvconsumerkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v9...")
		ibcKeeper.ConnectionKeeper.SetParams(ctx, ibcconnectiontypes.DefaultParams())

		fromVM := make(map[string]uint64)

		// for moduleName, eachModule := range app.MM.Modules {
		// 	fromVM[moduleName] = eachModule.ConsensusVersion()
		// }

		// TODO: should have a way to read from current node home
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			stdlog.Println("Failed to get home dir %2", err)
		}
		nodeHome := userHomeDir + "/.sovereign/config/genesis.json"
		appState, _, err := genutiltypes.GenesisStateFromGenFile(nodeHome)
		if err != nil {
			return fromVM, fmt.Errorf("failed to unmarshal genesis state: %w", err)
		}

		var consumerGenesis = consumertypes.GenesisState{}
		cdc.MustUnmarshalJSON(appState[consumertypes.ModuleName], &consumerGenesis)

		consumerGenesis.PreCCV = true
		// Temp fix
		consumerGenesis.Params.SoftOptOutThreshold = "0.05"
		consumerKeeper.InitGenesis(ctx, &consumerGenesis)

		ctx.Logger().Info("start to run module migrations...")

		// TODO: temporary fix
		return fromVM, nil
		// return app.MM.RunMigrations(ctx, app.configurator, fromVM)
	}
}
