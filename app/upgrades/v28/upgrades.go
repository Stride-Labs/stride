package v28

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	consumerkeeper "github.com/cosmos/interchain-security/v6/x/ccv/consumer/keeper"

	"github.com/Stride-Labs/stride/v28/utils"
	icqkeeper "github.com/Stride-Labs/stride/v28/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v28/x/stakeibc/keeper"
)

var (
	UpgradeName = "v28"

	EvmosChainId          = "evmos_9001-2"
	QueryId               = "2c39af4c3d2ecb96d8bbf7f3386468c5909e51fe3364b8d1f9d6fce173dd1f7a"
	QueryValidatorAddress = "evmosvaloper1tdss4m3x7jy9mlepm2dwy8820l7uv6m2vx6z88"
	EvmosDelegationIca    = "evmos1d67tx0zekagfhw6chhgza6qmhyad5qprru0nwazpx5s85ld0wh2sdhhznd"

	// Redemption rate bounds updated to give slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdkmath.LegacyMustNewDecFromStr("0.50")
	RedemptionRateOuterMaxAdjustment = sdkmath.LegacyMustNewDecFromStr("1.00")

	MaxMessagesPerIca = uint64(5)
	BandChainId       = "laozi-mainnet"

	// Action Gov Prop 262
	ReceivingAddress262 = "stride1czvrk3jkvtj8m27kqsqu2yrkhw3h3ykwj3rxh6"
	IncentivesAddress   = "stride1tlxk4as9sgpqkh42cfaxqja0mdj6qculqshy0gg3glazmrnx3y8s8gsvqk"
	IncentivesAmount    = sdkmath.NewInt(8_000_000_000_000)
	SecurityAddress     = "stride14p6dyylhlmr445w9shcp36r5qcypc5d7rftpr2v0dt3pcjvy66zq60qfgg"
	SecurityAmount      = sdkmath.NewInt(1_481_000_000_000)
)

// CreateUpgradeHandler creates an SDK upgrade handler for v27
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	consumerKeeper consumerkeeper.Keeper,
	distrKeeper distrkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	icqKeeper icqkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s...", UpgradeName))

		// Run migrations first
		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		// Initialize consumer ID
		// https://github.com/cosmos/interchain-security/blob/v6.4.1/UPGRADING.md#consumer
		ctx.Logger().Info("Setting consumer ID parameter...")
		InitializeConsumerId(ctx, consumerKeeper)

		// Apply distribution fix
		ctx.Logger().Info("Applying distribution module fix...")
		if err := ApplyDistributionFix(ctx, distrKeeper); err != nil {
			// Log warning but continue with upgrade (non-critical)
			ctx.Logger().Warn("Failed to apply distribution fix, continuing...", "warning", err.Error())
		} else {
			ctx.Logger().Info("Distribution fix successfully applied")
		}

		// Loosen slack from redemption rate bounds
		ctx.Logger().Info("Update redemption rate bounds...")
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		ctx.Logger().Info("Processing stale ICQ...")
		ClearStuckEvmosQuery(ctx, stakeibcKeeper, icqKeeper)

		ctx.Logger().Info("Setting max icas for band...")
		SetMaxIcasBand(ctx, stakeibcKeeper)

		ctx.Logger().Info("Action Gov Prop 262...")
		if err := ActionGovProp262(ctx, stakeibcKeeper, bankKeeper); err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}

// InitializeConsumerId sets the consumer Id parameter in the consumer module,
// to the consumer id for which the consumer is registered on the provider chain.
// The consumer id can be obtained in by querying the provider, e.g. by using the
// QueryConsumerIdFromClientId query.
//
// Steps to retrieve the Stride consumer chain ID from Cosmos Hub provider:
//  1. First, obtain the client ID from Stride using the command:
//     `strided q ccvconsumer provider-info` which returns "07-tendermint-1154"
//  2. Then, use the Provider's QueryConsumerIdFromClientId endpoint to get the corresponding consumer ID:
//     - API endpoint: https://rest.cosmos.directory/cosmoshub/interchain_security/ccv/provider/consumer_id/07-tendermint-1154
//     - This endpoint implements the query defined in the Interchain Security repository at:
//     https://github.com/cosmos/interchain-security/blob/307b1446/proto/interchain_security/ccv/provider/v1/query.proto#L132-L138
func InitializeConsumerId(ctx sdk.Context, consumerKeeper consumerkeeper.Keeper) {
	params := consumerKeeper.GetConsumerParams(ctx)
	params.ConsumerId = "1"
	consumerKeeper.SetParams(ctx, params)
}

// Updates the outer redemption rate bounds
func UpdateRedemptionRateBounds(ctx sdk.Context, k stakeibckeeper.Keeper) {
	ctx.Logger().Info("Updating redemption rate outer bounds...")

	for _, hostZone := range k.GetAllHostZone(ctx) {
		outerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMinAdjustment)
		outerMaxDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMaxAdjustment)

		outerMin := hostZone.RedemptionRate.Sub(outerMinDelta)
		outerMax := hostZone.RedemptionRate.Add(outerMaxDelta)

		hostZone.MinRedemptionRate = outerMin
		hostZone.MaxRedemptionRate = outerMax

		k.SetHostZone(ctx, hostZone)
	}
}

// Cleans up the stale ICQ
func ClearStuckEvmosQuery(ctx sdk.Context, k stakeibckeeper.Keeper, icqKeeper icqkeeper.Keeper) {
	ctx.Logger().Info("Deleting stale ICQ...")
	icqKeeper.DeleteQuery(ctx, QueryId)

	ctx.Logger().Info("Setting validator slash_query_in_progress to false...")
	hostZone, found := k.GetHostZone(ctx, EvmosChainId)
	if !found {
		ctx.Logger().Error("host zone not found")
		return
	}

	// find the right validator and set slash_query_in_progress to false
	for i, validator := range hostZone.Validators {
		if validator.Address == QueryValidatorAddress {
			validator.SlashQueryInProgress = false
			hostZone.Validators[i] = validator
			k.SetHostZone(ctx, hostZone)

			ctx.Logger().Info("Set validator slash_query_in_progress to false")
			return
		}
	}
}

// Add the MaxMessagesPerIcaTx parameter to each host zone
func SetMaxIcasBand(ctx sdk.Context, k stakeibckeeper.Keeper) {
	bandHostZone, found := k.GetHostZone(ctx, BandChainId)
	if !found {
		ctx.Logger().Error("band host zone not found")
		return
	}
	bandHostZone.MaxMessagesPerIcaTx = MaxMessagesPerIca
	k.SetHostZone(ctx, bandHostZone)
}

// Action Gov Prop 262
func ActionGovProp262(ctx sdk.Context, k stakeibckeeper.Keeper, bankKeeper bankkeeper.Keeper) error {
	receivingAddress262, err := sdk.AccAddressFromBech32(ReceivingAddress262)
	if err != nil {
		return errorsmod.Wrap(err, "invalid prop recipient address")
	}

	// Transfer from incentive program address to receiving address
	ctx.Logger().Info("Sending incentives to receiving address...")
	incentivesAddress, err := sdk.AccAddressFromBech32(IncentivesAddress)
	if err != nil {
		return errorsmod.Wrap(err, "invalid prop sender address")
	}

	incentivesCoins := sdk.NewCoins(sdk.NewCoin("ustrd", IncentivesAmount))
	err = utils.SafeSendCoins(false, bankKeeper, ctx, incentivesAddress, receivingAddress262, incentivesCoins)
	if err != nil {
		return errorsmod.Wrap(err, "failed to send tokens from incentives to receiving address")
	}

	// Transfer from security address to receiving address
	ctx.Logger().Info("Sending security to receiving address...")
	securityAddress, err := sdk.AccAddressFromBech32(SecurityAddress)
	if err != nil {
		return errorsmod.Wrap(err, "invalid prop sender address")
	}

	securityCoins := sdk.NewCoins(sdk.NewCoin("ustrd", SecurityAmount))
	err = utils.SafeSendCoins(false, bankKeeper, ctx, securityAddress, receivingAddress262, securityCoins)
	if err != nil {
		return errorsmod.Wrap(err, "failed to send tokens from security to receiving address")
	}

	return nil
}
