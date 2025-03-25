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
	upperBoundPeriod := uint64(3913)
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

	// Copy historical rewards data from upper bound period to the slashing event period
	// Note: By using the same historical rewards from the upper bound period, we're effectively
	// not accounting for rewards that accrued on approximately half the blocks between 4300034-5047517.
	// The reward amounts are extremely small as of 2025-03-25:
	// - 0.000000000055967683 INJ   (≈ $0.0000000005910187 USD)
	// - 0.000000006094164748 EVMOS (≈ $0.0000000000284597 USD)
	// At these microscopic values, the simplification has virtually no impact on users.
	historicalRewards, err := distrKeeper.GetValidatorHistoricalRewards(ctx, valAddr, upperBoundPeriod)
	if err != nil {
		return err
	}

	return distrKeeper.SetValidatorHistoricalRewards(ctx, valAddr, slashingEventPeriod, historicalRewards)
}
