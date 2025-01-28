package v25

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	recordskeeper "github.com/Stride-Labs/stride/v24/x/records/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v24/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v24/x/stakeibc/types"
	staketiakeeper "github.com/Stride-Labs/stride/v24/x/staketia/keeper"
	staketiatypes "github.com/Stride-Labs/stride/v24/x/staketia/types"
)

const UpgradeName = "v25"

var (
	CommunityPoolGrowthAddress = "stride1lj0m72d70qerts9ksrsphy9nmsd4h0s88ll9gfphmhemh8ewet5qj44jc9"
	BnocsCustodian             = "stride1ff875h5plrnyumhm3cezn85dj4hzjzjqpz99mg"
	BnocsProposalAmount        = sdk.NewInt(17_857_000_000)
	Ustrd                      = "ustrd"
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

		// Implement Bnocs Proposal 256
		if err := ExecuteProp256(ctx, bankKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to implement Bnocs Proposal")
		}

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

// Execute Prop #256 - Signaling proposal to give a grant to Bnocs
// Sends 17,857 STRD from "Community Pool - Growth" to Bnocs recipient address
func ExecuteProp256(ctx sdk.Context, k bankkeeper.Keeper) error {
	communityPoolGrowthAddress := sdk.MustAccAddressFromBech32(CommunityPoolGrowthAddress)
	bnocsCuostidanAddress := sdk.MustAccAddressFromBech32(BnocsCustodian)
	transferCoin := sdk.NewCoin(Ustrd, BnocsProposalAmount)
	return k.SendCoins(ctx, communityPoolGrowthAddress, bnocsCuostidanAddress, sdk.NewCoins(transferCoin))
}
