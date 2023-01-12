package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	stakingibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	"github.com/Stride-Labs/stride/v4/x/claim/types"
)

func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	identifiers := k.GetAirdropIdentifiersForUser(ctx, delAddr)
	for _, identifier := range identifiers {
		cacheCtx, write := ctx.CacheContext()
		_, err := k.ClaimCoinsForAction(cacheCtx, delAddr, types.ACTION_DELEGATE_STAKE, identifier)
		if err == nil {
			write()
		} else {
			k.Logger(ctx).Error(fmt.Sprintf("airdrop claim failure for %s on delegation hook: %s", delAddr.String(), err.Error()))
		}
	}
}

func (k Keeper) AfterLiquidStake(ctx sdk.Context, addr sdk.AccAddress) {
	identifiers := k.GetAirdropIdentifiersForUser(ctx, addr)
	for _, identifier := range identifiers {
		cacheCtx, write := ctx.CacheContext()
		_, err := k.ClaimCoinsForAction(cacheCtx, addr, types.ACTION_LIQUID_STAKE, identifier)
		if err == nil {
			write()
		} else {
			k.Logger(ctx).Error(fmt.Sprintf("airdrop claim failure for %s on liquid staking hook: %s", addr.String(), err.Error()))
		}
	}
}

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	err := k.ResetClaimStatus(ctx, epochInfo.Identifier)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to reset claim status for epoch %s: %s", epochInfo.Identifier, err.Error()))
	}
}

// ________________________________________________________________________________________

// Hooks wrapper struct for claim keeper
type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}
var _ stakingibctypes.StakeIBCHooks = Hooks{}
var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// ibcstaking hooks
func (h Hooks) AfterLiquidStake(ctx sdk.Context, addr sdk.AccAddress) {
	h.k.AfterLiquidStake(ctx, addr)
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.AfterEpochEnd(ctx, epochInfo)
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
