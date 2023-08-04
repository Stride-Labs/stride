package v12

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v3/x/ccv/consumer/keeper"
	consumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	"github.com/spf13/cast"
)

var (
	UpgradeName = "v12"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v12
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	appOpts servertypes.AppOptions,
	ibcKeeper ibckeeper.Keeper,
	consumerKeeper *ccvconsumerkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v12...")
		ibcKeeper.ConnectionKeeper.SetParams(ctx, ibcconnectiontypes.DefaultParams())

		fromVM := make(map[string]uint64)

		// for moduleName, eachModule := range app.MM.Modules {
		// 	fromVM[moduleName] = eachModule.ConsensusVersion()
		// }

		nodeHome := cast.ToString(appOpts.Get(flags.FlagHome))
		consumerUpgradeGenFile := nodeHome + "/config/ccv.json"
		appState, _, err := genutiltypes.GenesisStateFromGenFile(consumerUpgradeGenFile)
		if err != nil {
			return fromVM, fmt.Errorf("failed to unmarshal genesis state: %w", err)
		}

		var consumerGenesis = consumertypes.GenesisState{}
		cdc.MustUnmarshalJSON(appState[consumertypes.ModuleName], &consumerGenesis)

		consumerGenesis.PreCCV = true
		consumerGenesis.Params.SoftOptOutThreshold = "0.05"
		consumerGenesis.Params.RewardDenoms = []string{"ustrd"}
		consumerKeeper.InitGenesis(ctx, &consumerGenesis)
		consumerKeeper.SetDistributionTransmissionChannel(ctx, "channel-0")

		return fromVM, nil
	}
}
