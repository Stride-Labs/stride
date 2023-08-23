package v14

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/Stride-Labs/stride/v13/x/claim/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v13/x/stakeibc/keeper"
	claimkeeper "github.com/Stride-Labs/stride/v13/x/claim/keeper"

)

var (
	UpgradeName = "v14"

	AirdropDuration  = time.Hour * 24 * 30 * 12 * 3                 // 3 years
	AirdropStartTime = time.Date(2023, 9, 4, 16, 0, 0, 0, time.UTC) // Sept 3, 2023 @ 16:00 UTC (12:00 EST)

	InjectiveAirdropDistributor = ""
	InjectiveAirdropIdentifier  = "injective"
	InjectiveChainId            = "injective-1"

	ComdexAirdropDistributor = ""
	ComdexAirdropIdentifier  = "comdex"
	ComdexChainId            = "comdex-1"

	SommAirdropDistributor = ""
	SommAirdropIdentifier  = "sommelier"
	SommChainId            = "sommelier-3"

	UmeeAirdropDistributor = ""
	UmeeAirdropIdentifier  = "umee"
	UmeeChainId            = "umee-1"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v14
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	claimKeeper claimkeeper.Keeper
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v14...")

		duration := uint64(AirdropDuration.Seconds())
		startTime := uint64(AirdropStartTime.Unix())

		// Add the Injective Airdrop
		ctx.Logger().Info("Adding Injective airdrop...")
		if err := claimKeeper.CreateAirdropAndEpoch(ctx, types.MsgCreateAirdrop{
			Distributor:      InjectiveAirdropDistributor,
			Identifier:       InjectiveAirdropIdentifier,
			ChainId:          InjectiveChainId,
			Denom:            claimtypes.DefaultClaimDenom,
			StartTime:        startTime,
			Duration:         duration,
			AutopilotEnabled: true,
		}); err != nil {
			return vm, err
		}

		// Add the Comdex Airdrop
		ctx.Logger().Info("Adding Comdex airdrop...")
		if err := claimKeeper.CreateAirdropAndEpoch(ctx, types.MsgCreateAirdrop{
			Distributor:      ComdexAirdropDistributor,
			Identifier:       ComdexAirdropIdentifier,
			ChainId:          ComdexChainId,
			Denom:            claimtypes.DefaultClaimDenom,
			StartTime:        startTime,
			Duration:         duration,
			AutopilotEnabled: false,
		}); err != nil {
			return vm, err
		}

		// Add the Somm Airdrop
		ctx.Logger().Info("Adding Somm airdrop...")
		if err := claimKeeper.CreateAirdropAndEpoch(ctx, types.MsgCreateAirdrop{
			Distributor:      SommAirdropDistributor,
			Identifier:       SommAirdropIdentifier,
			ChainId:          SommChainId,
			Denom:            claimtypes.DefaultClaimDenom,
			StartTime:        startTime,
			Duration:         duration,
			AutopilotEnabled: false,
		}); err != nil {
			return vm, err
		}

		// Add the Umee Airdrop
		ctx.Logger().Info("Adding Umee airdrop...")
		if err := claimKeeper.CreateAirdropAndEpoch(ctx, types.MsgCreateAirdrop{
			Distributor:      UmeeAirdropDistributor,
			Identifier:       UmeeAirdropIdentifier,
			ChainId:          UmeeChainId,
			Denom:            claimtypes.DefaultClaimDenom,
			StartTime:        startTime,
			Duration:         duration,
			AutopilotEnabled: false,
		}); err != nil {
			return vm, err
		}

		ctx.Logger().Info("Loading airdrop allocations...")
		claimKeeper.LoadAllocationData(ctx, allocations)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
