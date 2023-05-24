package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	stakingibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	"github.com/Stride-Labs/stride/v9/x/claim/types"
)

func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	identifiers := k.GetAirdropIdentifiersForUser(ctx, delAddr)
	for _, identifier := range identifiers {
		cacheCtx, write := ctx.CacheContext()
		_, err := k.ClaimCoinsForAction(cacheCtx, delAddr, types.ACTION_DELEGATE_STAKE, identifier)
		if err == nil {
			write()
		} else {
			k.Logger(ctx).Error(fmt.Sprintf("airdrop claim failure for %s on delegation hook: %s", delAddr.String(), err.Error()))
			return err
		}
	}
	return nil
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
	// check if epochInfo.Identifier is an airdrop epoch
	// if yes, reset claim status for all users
	// check if epochInfo.Identifier starts with "airdrop"
	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] checking if epoch %s is an airdrop epoch", epochInfo.Identifier))
	if strings.HasPrefix(epochInfo.Identifier, "airdrop-") {

		airdropIdentifier := strings.TrimPrefix(epochInfo.Identifier, "airdrop-")
		k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] trimmed airdrop identifier: %s", airdropIdentifier))

		airdrop := k.GetAirdropByIdentifier(ctx, airdropIdentifier)
		k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] airdrop found: %v", airdrop))

		if airdrop != nil {
			k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] resetting claims for airdrop %s", epochInfo.Identifier))
			err := k.ResetClaimStatus(ctx, airdropIdentifier)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("[CLAIM] failed to reset claim status for epoch %s: %s", epochInfo.Identifier, err.Error()))
			}
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] airdrop %s not found, skipping reset", airdropIdentifier))
		}
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

func (h Hooks) AfterUnbondingInitiated(ctx sdk.Context, id uint64) error {
	return nil
}

// staking hooks
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return h.k.AfterDelegationModified(ctx, delAddr, valAddr)
}
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	return nil
}
func (h Hooks) BeforeSlashingUnbondingDelegation(ctx sdk.Context, unbondingDelegation stakingtypes.UnbondingDelegation,
	infractionHeight int64, slashFactor sdk.Dec) error {
	return nil
}

func (h Hooks) BeforeSlashingRedelegation(ctx sdk.Context, srcValidator stakingtypes.Validator, redelegation stakingtypes.Redelegation,
	infractionHeight int64, slashFactor sdk.Dec) error {
	return nil
}
