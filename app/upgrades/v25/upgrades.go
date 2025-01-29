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

var (
	UpgradeName = "v25"

	// Redemption rate bounds updated to give ~3 months of slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("0.10")

	// Osmosis will have a slighly larger buffer with the redemption rate
	// since their yield is less predictable
	OsmosisChainId              = "osmosis-1"
	OsmosisRedemptionRateBuffer = sdk.MustNewDecFromStr("0.02")

	// Inner redemption rate adjustment variables
	RedemptionRateInnerAdjustment = sdk.MustNewDecFromStr("0.001")

	// Info for failed LSM record
	CosmosChainId         = "cosmoshub-4"
	FailedLSMDepositDenom = "cosmosvaloper1yh089p0cre4nhpdqw35uzde5amg3qzexkeggdn/59223"
)

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

		if err := staketiaKeeper.ArchiveFailedTransferRecord(ctx, 427); err != nil {
			return vm, errorsmod.Wrapf(err, "Unable to archive transfer record")
		}
		staketiaKeeper.RemoveTransferInProgressRecordId(ctx, "channel-38", 26)

		// Migrate staketia to stakeibc
		if err := staketiakeeper.InitiateMigration(ctx, staketiaKeeper, bankKeeper, recordsKeeper, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate staketia to stakeibc")
		}

		// Add celestia validators
		if err := AddCelestiaValidators(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to add celestia validators")
		}

		// Implement Bnocs Proposal 256
		// if err := ExecuteProp256(ctx, bankKeeper); err != nil {
		// 	return vm, errorsmod.Wrapf(err, "unable to implement Bnocs Proposal")
		// }

		// Update redemption rate bounds
		// UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		// Update Celestia inner bounds
		// UpdateCelestiaInnerBounds(ctx, stakeibcKeeper)

		// Reset failed LSM record
		// ResetLSMRecord(ctx, recordsKeeper)

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
// See more here: https://www.mintscan.io/stride/proposals/256
// And here: https://common.xyz/stride/discussion/25922-proposal-for-grant-bnocscom-dashboard-for-the-stride-ecosystem
func ExecuteProp256(ctx sdk.Context, k bankkeeper.Keeper) error {
	communityPoolGrowthAddress := sdk.MustAccAddressFromBech32(CommunityPoolGrowthAddress)
	bnocsCuostidanAddress := sdk.MustAccAddressFromBech32(BnocsCustodian)
	transferCoin := sdk.NewCoin(Ustrd, BnocsProposalAmount)
	return k.SendCoins(ctx, communityPoolGrowthAddress, bnocsCuostidanAddress, sdk.NewCoins(transferCoin))
}

// Updates the outer redemption rate bounds
func UpdateRedemptionRateBounds(ctx sdk.Context, k stakeibckeeper.Keeper) {
	ctx.Logger().Info("Updating redemption rate outer bounds...")

	for _, hostZone := range k.GetAllHostZone(ctx) {
		// Give osmosis a bit more slack since OSMO stakers collect real yield
		outerAdjustment := RedemptionRateOuterMaxAdjustment
		if hostZone.ChainId == OsmosisChainId {
			outerAdjustment = outerAdjustment.Add(OsmosisRedemptionRateBuffer)
		}

		outerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMinAdjustment)
		outerMaxDelta := hostZone.RedemptionRate.Mul(outerAdjustment)

		outerMin := hostZone.RedemptionRate.Sub(outerMinDelta)
		outerMax := hostZone.RedemptionRate.Add(outerMaxDelta)

		hostZone.MinRedemptionRate = outerMin
		hostZone.MaxRedemptionRate = outerMax

		k.SetHostZone(ctx, hostZone)
	}
}

// Tighten Celestia's inner bounds as a safety measure
func UpdateCelestiaInnerBounds(ctx sdk.Context, k stakeibckeeper.Keeper) {
	ctx.Logger().Info("Tightening Celestia inner bounds...")

	hostZone, found := k.GetHostZone(ctx, staketiatypes.CelestiaChainId)
	if !found {
		ctx.Logger().Error("Celestia host zone not found, could not update inner bounds")
		return
	}

	innerRedemptionRateDelta := hostZone.RedemptionRate.Mul(RedemptionRateInnerAdjustment)

	minInnerRedemptionRate := hostZone.RedemptionRate.Sub(innerRedemptionRateDelta)
	maxInnerRedemptionRate := hostZone.RedemptionRate.Add(innerRedemptionRateDelta)

	hostZone.MinInnerRedemptionRate = minInnerRedemptionRate
	hostZone.MaxInnerRedemptionRate = maxInnerRedemptionRate

	k.SetHostZone(ctx, hostZone)
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
