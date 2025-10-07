package keeper

import (
	"context"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v29/x/epochs/types"
	stakingibctypes "github.com/Stride-Labs/stride/v29/x/stakeibc/types"

	"github.com/Stride-Labs/stride/v29/x/claim/types"
)

func (k Keeper) AfterDelegationModified(context context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	ctx := sdk.UnwrapSDKContext(context)
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

func (k Keeper) AfterLiquidStake(context context.Context, addr sdk.AccAddress) {
	ctx := sdk.UnwrapSDKContext(context)
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

func (k Keeper) BeforeEpochStart(ctx context.Context, epochInfo epochstypes.EpochInfo) {
}

func (k Keeper) AfterEpochEnd(context context.Context, epochInfo epochstypes.EpochInfo) {
	// check if epochInfo.Identifier is an airdrop epoch
	// if yes, reset claim status for all users
	// check if epochInfo.Identifier starts with "airdrop"
	ctx := sdk.UnwrapSDKContext(context)
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

var (
	_ stakingtypes.StakingHooks     = Hooks{}
	_ stakingibctypes.StakeIBCHooks = Hooks{}
	_ epochstypes.EpochHooks        = Hooks{}
)

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// ibcstaking hooks
func (h Hooks) AfterLiquidStake(ctx context.Context, addr sdk.AccAddress) {
	h.k.AfterLiquidStake(ctx, addr)
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx context.Context, epochInfo epochstypes.EpochInfo) {
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(ctx context.Context, epochInfo epochstypes.EpochInfo) {
	h.k.AfterEpochEnd(ctx, epochInfo)
}

func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	return nil
}

// staking hooks
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return h.k.AfterDelegationModified(ctx, delAddr, valAddr)
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	return nil
}

func (h Hooks) BeforeSlashingUnbondingDelegation(ctx context.Context, unbondingDelegation stakingtypes.UnbondingDelegation,
	infractionHeight int64, slashFactor sdkmath.LegacyDec,
) error {
	return nil
}

func (h Hooks) BeforeSlashingRedelegation(ctx context.Context, srcValidator stakingtypes.Validator, redelegation stakingtypes.Redelegation,
	infractionHeight int64, slashFactor sdkmath.LegacyDec,
) error {
	return nil
}
