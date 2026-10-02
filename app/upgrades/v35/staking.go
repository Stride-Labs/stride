package v35

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RaiseMaxUnbondingEntries sets staking MaxEntries to StakingMaxEntries so the mass undelegation
// never hits ErrMaxUnbondingDelegationEntries on a pair that already holds mainnet's 7 entries
// (authority spec §3). Delegation is blocked afterwards, so no pair can approach the new cap.
func RaiseMaxUnbondingEntries(ctx sdk.Context, k stakingkeeper.Keeper) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	params.MaxEntries = StakingMaxEntries
	if err := k.SetParams(ctx, params); err != nil {
		return err
	}
	ctx.Logger().Info(fmt.Sprintf("v35: staking max unbonding entries set to %d", StakingMaxEntries))
	return nil
}

// UndelegateAllDelegations unbonds every x/staking delegation in full with the stock unbonding
// time (authority spec §3). A delegation whose undelegation fails (e.g. missing distribution
// starting info) is logged and skipped; the rest still complete. Failing to list the delegations
// at all fails the upgrade, since leaving everything bonded defeats the step. Only a summary is
// logged because mainnet holds tens of thousands of delegations. Expected side effects: rewards are
// withdrawn by the distribution hook, operators drop below their minimum self-delegation and
// are jailed, and every validator is left at zero tokens for the EndBlocker to unbond.
func UndelegateAllDelegations(ctx sdk.Context, k stakingkeeper.Keeper) error {
	delegations, err := k.GetAllDelegations(ctx)
	if err != nil {
		return fmt.Errorf("v35: unable to list delegations: %w", err)
	}

	// Store iteration order is deterministic, so every node undelegates in the same sequence
	undelegated, skipped := 0, 0
	totalUndelegated := sdkmath.ZeroInt()
	for _, delegation := range delegations {
		amount, err := undelegate(ctx, k, delegation)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: skipping undelegation of %s from %s: %s",
				delegation.DelegatorAddress, delegation.ValidatorAddress, err))
			skipped++
			continue
		}
		undelegated++
		totalUndelegated = totalUndelegated.Add(amount)
	}

	ctx.Logger().Info(fmt.Sprintf("v35: undelegated %d delegations totaling %s ustrd, skipped %d",
		undelegated, totalUndelegated, skipped))
	return nil
}

// undelegate unbonds one delegation in full and returns the tokens moved to unbonding.
func undelegate(ctx sdk.Context, k stakingkeeper.Keeper, delegation stakingtypes.Delegation) (sdkmath.Int, error) {
	delegator, err := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
	if err != nil {
		return sdkmath.Int{}, err
	}
	validator, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
	if err != nil {
		return sdkmath.Int{}, err
	}
	_, amount, err := k.Undelegate(ctx, delegator, validator, delegation.Shares)
	return amount, err
}
