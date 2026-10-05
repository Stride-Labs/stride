package distrwrapper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Hooks wraps the standard x/distribution staking hooks so that removing a validator cannot panic.
// Reward withdrawals cap each delegator's rewards at the validator's outstanding rewards, not at
// outstanding minus commission, so rounding can leave accumulated commission a hair above
// outstanding. The stock AfterValidatorRemoved then panics on outstanding.Sub(commission), and
// when the staking EndBlocker removes a matured, empty validator there is no recovery around it,
// so the chain halts. Every other hook passes through unchanged.
type Hooks struct {
	keeper.Hooks
	k keeper.Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

func NewHooks(k keeper.Keeper) Hooks {
	return Hooks{Hooks: k.Hooks(), k: k}
}

// AfterValidatorRemoved caps the validator's commission at its outstanding rewards, then runs the
// standard hook. A validator within bounds is removed exactly as before.
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	if _, err := ClampCommissionToOutstanding(ctx, h.k, valAddr); err != nil {
		return err
	}
	return h.Hooks.AfterValidatorRemoved(ctx, consAddr, valAddr)
}

// ClampCommissionToOutstanding caps a validator's accumulated commission at its outstanding
// rewards, per denom, and reports whether it changed anything. The clamped dust stays in the
// distribution module and reaches the community pool when the validator is removed.
func ClampCommissionToOutstanding(ctx context.Context, k keeper.Keeper, valAddr sdk.ValAddress) (bool, error) {
	commission, err := k.GetValidatorAccumulatedCommission(ctx, valAddr)
	if err != nil {
		return false, err
	}
	outstanding, err := k.GetValidatorOutstandingRewardsCoins(ctx, valAddr)
	if err != nil {
		return false, err
	}

	// SafeSub reports whether any denom of commission exceeds outstanding; Intersect then takes
	// the per-denom minimum and drops denoms outstanding no longer holds
	if _, exceeds := outstanding.SafeSub(commission.Commission); !exceeds {
		return false, nil
	}
	capped := commission.Commission.Intersect(outstanding)
	sdk.UnwrapSDKContext(ctx).Logger().Info(fmt.Sprintf("clamping %s commission from %s to outstanding %s",
		valAddr, commission.Commission, capped))
	commission.Commission = capped
	if err := k.SetValidatorAccumulatedCommission(ctx, valAddr, commission); err != nil {
		return false, err
	}
	return true, nil
}
