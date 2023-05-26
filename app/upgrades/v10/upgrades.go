package v10

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icacontrollermigrations "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/migrations/v6"
	clientkeeper "github.com/cosmos/ibc-go/v7/modules/core/02-client/keeper"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"

	mintkeeper "github.com/Stride-Labs/stride/v9/x/mint/keeper"
	minttypes "github.com/Stride-Labs/stride/v9/x/mint/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

var (
	UpgradeName     = "v10"
	EpochProvisions = sdk.NewDec(940_469_178)

	StakingProportion                     = "0.1605"
	CommunityPoolGrowthProportion         = "0.2158"
	StrategicReserveProportion            = "0.4879"
	CommunityPoolSecurityBudgetProportion = "0.1358"

	MinInitialDepositRatio = "0.25"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v10
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	legacyParamSubspace paramstypes.Subspace,
	capabilityStoreKey *storetypes.KVStoreKey,
	capabilityKeeper *capabilitykeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	consensusParamsKeeper consensusparamkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	mintKeeper mintkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v10...")

		ctx.Logger().Info("Migrating tendermint consensus params from x/params to x/consensus...")
		baseapp.MigrateParams(ctx, legacyParamSubspace, &consensusParamsKeeper)

		ctx.Logger().Info("Migrating ICA channel capabilities...")
		if err := icacontrollermigrations.MigrateICS27ChannelCapability(
			ctx,
			cdc,
			capabilityStoreKey,
			capabilityKeeper,
			stakeibctypes.ModuleName,
		); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to migrate ICA channel capabilities")
		}

		ctx.Logger().Info("Pruning expired tendermint consensus states...")
		if _, err := ibctmmigrations.PruneExpiredConsensusStates(ctx, cdc, clientKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to prune expired consensus states")
		}

		ctx.Logger().Info("Reducing STRD staking rewards...")
		if err := ReduceSTRDStakingRewards(ctx, mintKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to reduce STRD staking rewards")
		}

		ctx.Logger().Info("Running module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)

		ctx.Logger().Info("Setting MinInitialDepositRatio...")
		SetMinInitialDepositRatio(ctx, govKeeper)

		return vm, err
	}
}

// Cut STRD staking rewards in half from 27.64% to 13.82%
// Reduce epoch provisions by 13.82% from 1,078,767,123 to 940,469,178
func ReduceSTRDStakingRewards(ctx sdk.Context, k mintkeeper.Keeper) error {
	minter := minttypes.NewMinter(EpochProvisions)
	k.SetMinter(ctx, minter)

	stakingProportion := sdk.MustNewDecFromStr(StakingProportion)
	communityPoolGrowthProportion := sdk.MustNewDecFromStr(CommunityPoolGrowthProportion)
	strategicReserveProportion := sdk.MustNewDecFromStr(StrategicReserveProportion)
	communityPoolSecurityBudgetProportion := sdk.MustNewDecFromStr(CommunityPoolSecurityBudgetProportion)

	// Confirm proportions sum to 100
	totalProportions := stakingProportion.
		Add(communityPoolGrowthProportion).
		Add(strategicReserveProportion).
		Add(communityPoolSecurityBudgetProportion)

	if !totalProportions.Equal(sdk.OneDec()) {
		return fmt.Errorf("distribution proprotions do not sum to 1 (%v)", totalProportions)
	}

	distributionProperties := minttypes.DistributionProportions{
		Staking:                     stakingProportion,
		CommunityPoolGrowth:         communityPoolGrowthProportion,
		StrategicReserve:            strategicReserveProportion,
		CommunityPoolSecurityBudget: communityPoolSecurityBudgetProportion,
	}

	params := k.GetParams(ctx)
	params.DistributionProportions = distributionProperties
	k.SetParams(ctx, params)

	return nil
}

// Set the initial deposit ratio to 25%
func SetMinInitialDepositRatio(ctx sdk.Context, k govkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.MinInitialDepositRatio = MinInitialDepositRatio
	k.SetParams(ctx, params)
}
