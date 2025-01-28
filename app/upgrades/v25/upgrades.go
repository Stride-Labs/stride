package v25

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	recordskeeper "github.com/Stride-Labs/stride/v25/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v25/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v25/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v25/x/stakeibc/types"
	staketiakeeper "github.com/Stride-Labs/stride/v25/x/staketia/keeper"
	staketiatypes "github.com/Stride-Labs/stride/v25/x/staketia/types"
)

const (
	UpgradeName = "v25"

	CosmosChainId         = "cosmoshub-4"
	FailedLSMDepositDenom = "cosmosvaloper1yh089p0cre4nhpdqw35uzde5amg3qzexkeggdn/59223"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v25
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bankKeeper bankkeeper.Keeper,
	recordsKeeper recordskeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
	staketiaKeeper staketiakeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v25...")

		// Migrate staketia to stakeibc
		if err := staketiakeeper.InitiateMigration(ctx, staketiaKeeper, bankKeeper, recordsKeeper, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate staketia to stakeibc")
		}

		// Add celestia validators
		if err := AddCelestiaValidators(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to add celestia validators")
		}

		// Reset failed LSM record
		ResetLSMRecord(ctx, recordsKeeper)

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Adds the full celestia validator set, with a 0 delegation for each
func AddCelestiaValidators(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	for _, validatorConfig := range Validators {
		validator := stakeibctypes.Validator{
			Name:    validatorConfig.name,
			Address: validatorConfig.address,
			Weight:  validatorConfig.weight,
		}

		if err := k.AddValidatorToHostZone(ctx, staketiatypes.CelestiaChainId, validator, false); err != nil {
			return err
		}

		// Query and store the validator's sharesToTokens rate
		if err := k.QueryValidatorSharesToTokensRate(ctx, staketiatypes.CelestiaChainId, validatorConfig.address); err != nil {
			return err
		}
	}
	return nil
}

// Reset the failed LSM detokenization record status and decrement the amount by 1
// so that it will succeed on the retry
func ResetLSMRecord(ctx sdk.Context, k recordskeeper.Keeper) {
	ctx.Logger().Info("Resetting failed LSM detokenization record...")

	lsmDeposit, found := k.GetLSMTokenDeposit(ctx, CosmosChainId, FailedLSMDepositDenom)
	if !found {
		// No need to panic in this case since the difference is immaterial
		ctx.Logger().Error("Failed LSM deposit record not found")
		return
	}
	lsmDeposit.Status = recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE
	lsmDeposit.Amount = lsmDeposit.Amount.Sub(sdkmath.OneInt())
	k.SetLSMTokenDeposit(ctx, lsmDeposit)
}
