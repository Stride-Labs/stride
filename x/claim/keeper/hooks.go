package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	stakingibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"

	"github.com/Stride-Labs/stride/x/claim/types"
)

func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	_, err := k.ClaimCoinsForAction(ctx, delAddr, types.ActionDelegateStake)
	if err != nil {
		panic(err.Error())
	}
}

func (k Keeper) AfterLiquidStake(ctx sdk.Context, addr sdk.AccAddress) {
	_, err := k.ClaimCoinsForAction(ctx, addr, types.ActionLiquidStake)
	if err != nil {
		panic(err.Error())
	}
}

// ________________________________________________________________________________________

// Hooks wrapper struct for claim keeper
type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}
var _ stakingibctypes.StakeIBCHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// ibcstaking hooks
func (h Hooks) AfterLiquidStake(ctx sdk.Context, addr sdk.AccAddress) {
	h.k.AfterLiquidStake(ctx, addr)
}

// staking hooks
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)   {}
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {}
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.k.AfterDelegationModified(ctx, delAddr, valAddr)
}
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {}
func (h Hooks) BeforeSlashingUnbondingDelegation(ctx sdk.Context, unbondingDelegation stakingtypes.UnbondingDelegation,
	infractionHeight int64, slashFactor sdk.Dec) {
}

func (h Hooks) BeforeSlashingRedelegation(ctx sdk.Context, srcValidator stakingtypes.Validator, redelegation stakingtypes.Redelegation,
	infractionHeight int64, slashFactor sdk.Dec) {
}
