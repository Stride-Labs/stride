package v27

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// ApplyDistributionFix fixes an issue with a missing slashing event
// that causes some users to be unable to withdraw rewards
// from validator stridevaloper1tlz6ksce084ndhwlq2usghamvh0dut9q4z2gxd.
// The issue affects delegations created before block height 4300034.
func ApplyDistributionFix(ctx sdk.Context, distrKeeper distrkeeper.Keeper) error {
	// Define validator address with the missing slashing event
	valAddr, err := sdk.ValAddressFromBech32("stridevaloper1tlz6ksce084ndhwlq2usghamvh0dut9q4z2gxd")
	if err != nil {
		return err
	}

	// Define periods for slashing event to be inserted
	// These values are derived from mainnet state analysis
	slashingEventPeriod := uint64(3902)
	slashingEventBlock := uint64(4673775)
	slashingEventFraction := sdkmath.LegacyMustNewDecFromStr("0.0001") // 0.01% slash

	// Insert the missing slashing event between blocks 4300034-5047517 (periods 3893-3912)
	// The slashing event represents a 0.01% slash that occurred but wasn't properly recorded
	err = distrKeeper.SetValidatorSlashEvent(
		ctx,
		valAddr,
		slashingEventBlock,
		slashingEventPeriod,
		disttypes.NewValidatorSlashEvent(slashingEventPeriod, slashingEventFraction),
	)
	if err != nil {
		return err
	}

	// Copy historical rewards data from upper bound period (3913) to the slashing event period
	// Note: By using the same historical rewards from the upper bound period, we're effectively
	// not accounting for rewards that accrued on approximately half the blocks between 4300034-5047517.
	// The reward amounts are extremely small as of 2025-03-25:
	// - 0.000000000055967683 INJ   (≈ $0.0000000005910187 USD)
	// - 0.000000006094164748 EVMOS (≈ $0.0000000000284597 USD)
	// At these microscopic values, the simplification has virtually no impact on users.
	// Use hardcoded historical rewards data for period 3913 in case reading from state fails
	// These values are taken from the mainnet state export (https://github.com/Stride-Labs/stride/blob/73b5d9391/app/upgrades/v27/test_dist_genesis.json#L10889-L10945)
	decCoins := sdk.NewDecCoins(
		sdk.NewDecCoinFromDec("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", sdkmath.LegacyMustNewDecFromStr("0.000000000128901120")),
		sdk.NewDecCoinFromDec("ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34", sdkmath.LegacyMustNewDecFromStr("0.000000000070807751")),
		sdk.NewDecCoinFromDec("staevmos", sdkmath.LegacyMustNewDecFromStr("6094164748.469303271892926613")),
		sdk.NewDecCoinFromDec("stinj", sdkmath.LegacyMustNewDecFromStr("55967683.467690438071400748")),
		sdk.NewDecCoinFromDec("stuatom", sdkmath.LegacyMustNewDecFromStr("0.001238741629066745")),
		sdk.NewDecCoinFromDec("stucmdx", sdkmath.LegacyMustNewDecFromStr("0.000668292145029035")),
		sdk.NewDecCoinFromDec("stujuno", sdkmath.LegacyMustNewDecFromStr("0.000637635773775489")),
		sdk.NewDecCoinFromDec("stuluna", sdkmath.LegacyMustNewDecFromStr("0.000366081500196863")),
		sdk.NewDecCoinFromDec("stuosmo", sdkmath.LegacyMustNewDecFromStr("0.002780710046543077")),
		sdk.NewDecCoinFromDec("stustars", sdkmath.LegacyMustNewDecFromStr("0.017148158752737249")),
		sdk.NewDecCoinFromDec("stuumee", sdkmath.LegacyMustNewDecFromStr("0.005597631087158596")),
		sdk.NewDecCoinFromDec("ustrd", sdkmath.LegacyMustNewDecFromStr("0.074031996748679476")),
	)

	// Create historical rewards using the hardcoded values
	historicalRewards := disttypes.NewValidatorHistoricalRewards(decCoins, 1)

	// Set the historical rewards for the slashing event period
	return distrKeeper.SetValidatorHistoricalRewards(ctx, valAddr, slashingEventPeriod, historicalRewards)
}
