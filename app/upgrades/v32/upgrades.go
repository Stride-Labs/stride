package v32

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/Stride-Labs/stride/v32/utils"
	stakeibckeeper "github.com/Stride-Labs/stride/v32/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v32/x/stakeibc/types"
)

var (
	UpgradeName = "v32"

	MinDeposit         = sdkmath.NewInt(20_000_000_000)
	ValidatorWeightCap = uint64(20)
)

// CreateUpgradeHandler creates an SDK upgrade handler for v32
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	govKeeper govkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s...", UpgradeName))

		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Updating min deposits...")
		if err := IncreaseMinDeposit(ctx, govKeeper); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Updating max validator weight...")
		IncreaseMaxValidatorWeight(ctx, stakeibcKeeper)

		ctx.Logger().Info("Updating validator weights...")
		if err := UpdateValidatorWeights(ctx, stakeibcKeeper); err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}

// Increase min deposit by 10x to 20k STRD
func IncreaseMinDeposit(ctx context.Context, gk govkeeper.Keeper) error {
	params, err := gk.Params.Get(ctx)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to get params")
	}

	params.MinDeposit = sdk.NewCoins(sdk.NewCoin(utils.BaseStrideDenom, MinDeposit))

	if err := gk.Params.Set(ctx, params); err != nil {
		return errorsmod.Wrap(err, "failed to set params")
	}

	return nil
}

// Increase the max validator weight to
func IncreaseMaxValidatorWeight(ctx sdk.Context, sk stakeibckeeper.Keeper) {
	params := sk.GetParams(ctx)
	params.ValidatorWeightCap = ValidatorWeightCap
	sk.SetParams(ctx, params)
}

// Update validator weights across all host zones
// Phase 1: Add new validators (not yet on-chain) with weight 0
// Phase 2: Set all validator weights to their target values
func UpdateValidatorWeights(ctx sdk.Context, sk stakeibckeeper.Keeper) error {
	for _, chainId := range utils.StringMapKeys(NewValidators) {
		validators := NewValidators[chainId]
		ctx.Logger().Info(fmt.Sprintf("Adding %d new validators to %s...", len(validators), chainId))
		for _, val := range validators {
			validator := stakeibctypes.Validator{
				Name:    val.Name,
				Address: val.Address,
				Weight:  0,
			}
			if err := sk.AddValidatorToHostZone(ctx, chainId, validator, false); err != nil {
				return errorsmod.Wrapf(err, "failed to add validator %s to %s", val.Address, chainId)
			}
		}
	}

	for _, chainId := range utils.StringMapKeys(TargetWeights) {
		weights := TargetWeights[chainId]
		ctx.Logger().Info(fmt.Sprintf("Setting validator weights for %s...", chainId))

		hostZone, found := sk.GetHostZone(ctx, chainId)
		if !found {
			return errorsmod.Wrapf(stakeibctypes.ErrHostZoneNotFound, "host zone %s not found", chainId)
		}

		weightMap := make(map[string]uint64, len(weights))
		for _, w := range weights {
			weightMap[w.Address] = w.Weight
		}

		for _, validator := range hostZone.Validators {
			targetWeight, found := weightMap[validator.Address]
			if !found {
				return errorsmod.Wrapf(stakeibctypes.ErrValidatorNotFound,
					"validator %s on %s not found in target weights", validator.Address, chainId)
			}
			validator.Weight = targetWeight
		}

		sk.SetHostZone(ctx, hostZone)

		var totalWeight uint64
		for _, validator := range hostZone.Validators {
			totalWeight += validator.Weight
		}
		ctx.Logger().Info(fmt.Sprintf("  %s: %d validators, total weight %d",
			chainId, len(hostZone.Validators), totalWeight))
	}

	return nil
}
