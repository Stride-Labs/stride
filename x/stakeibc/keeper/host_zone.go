package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v18/utils"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

const (
	MinValidatorsBeforeWeightCapCheck = 10
)

// SetHostZone set a specific hostZone in the store
func (k Keeper) SetHostZone(ctx sdk.Context, hostZone types.HostZone) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	b := k.cdc.MustMarshal(&hostZone)
	store.Set([]byte(hostZone.ChainId), b)
}

// GetHostZone returns a hostZone from its id
func (k Keeper) GetHostZone(ctx sdk.Context, chainId string) (val types.HostZone, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	b := store.Get([]byte(chainId))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetActiveHostZone returns an error if the host zone is not found or if it's found, but is halted
func (k Keeper) GetActiveHostZone(ctx sdk.Context, chainId string) (hostZone types.HostZone, err error) {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return hostZone, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", chainId)
	}
	if hostZone.Halted {
		return hostZone, types.ErrHaltedHostZone.Wrapf("host zone %s is halted", chainId)
	}
	return hostZone, nil
}

// GetHostZoneFromHostDenom returns a HostZone from a HostDenom
func (k Keeper) GetHostZoneFromHostDenom(ctx sdk.Context, denom string) (*types.HostZone, error) {
	var matchZone types.HostZone
	k.IterateHostZones(ctx, func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		if zoneInfo.HostDenom == denom {
			matchZone = zoneInfo
			return nil
		}
		return nil
	})
	if matchZone.ChainId != "" {
		return &matchZone, nil
	}
	return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "No HostZone for %s denom found", denom)
}

// GetHostZoneFromTransferChannelID returns a HostZone from a transfer channel ID
func (k Keeper) GetHostZoneFromTransferChannelID(ctx sdk.Context, channelID string) (hostZone types.HostZone, found bool) {
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		if hostZone.TransferChannelId == channelID {
			return hostZone, true
		}
	}
	return types.HostZone{}, false
}

// RemoveHostZone removes a hostZone from the store
func (k Keeper) RemoveHostZone(ctx sdk.Context, chainId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	store.Delete([]byte(chainId))
}

// GetAllHostZone returns all hostZone
func (k Keeper) GetAllHostZone(ctx sdk.Context) (list []types.HostZone) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.HostZone
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetAllActiveHostZone returns all hostZones that are active (halted = false)
func (k Keeper) GetAllActiveHostZone(ctx sdk.Context) (list []types.HostZone) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.HostZone
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if !val.Halted {
			list = append(list, val)
		}
	}

	return
}

// Increments the validators slash query progress tracker
func (k Keeper) IncrementValidatorSlashQueryProgress(
	ctx sdk.Context,
	chainId string,
	validatorAddress string,
	amount sdkmath.Int,
) error {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return types.ErrHostZoneNotFound
	}

	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return types.ErrValidatorNotFound
	}

	// Increment the progress tracker
	oldProgress := validator.SlashQueryProgressTracker
	newProgress := validator.SlashQueryProgressTracker.Add(amount)
	validator.SlashQueryProgressTracker = newProgress

	// If the checkpoint is zero, it implies the TVL was 0 last time it was set, and we should
	// update it here
	// If the checkpoint is non-zero, only update it if it was just breached
	shouldUpdateCheckpoint := true
	if !validator.SlashQueryCheckpoint.IsZero() {
		oldInterval := oldProgress.Quo(validator.SlashQueryCheckpoint)
		newInterval := newProgress.Quo(validator.SlashQueryCheckpoint)
		shouldUpdateCheckpoint = oldInterval.LT(newInterval)
	}

	// Optionally re-calculate the checkpoint
	// Threshold of 1% means once 1% of TVL has been breached, the query is issued
	if shouldUpdateCheckpoint {
		validator.SlashQueryCheckpoint = k.GetUpdatedSlashQueryCheckpoint(ctx, hostZone.TotalDelegations)
	}

	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	return nil
}

// Increments the number of validator delegation changes in progress by 1
// Note: This modifies the original host zone struct. The calling function must Set this host zone
// for changes to persist
func (k Keeper) IncrementValidatorDelegationChangesInProgress(hostZone *types.HostZone, validatorAddress string) error {
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "validator %s not found", validatorAddress)
	}
	validator.DelegationChangesInProgress += 1
	hostZone.Validators[valIndex] = &validator
	return nil
}

// Decrements the number of validator delegation changes in progress by 1
// Note: This modifies the original host zone struct. The calling function must Set this host zone
// for changes to persist
func (k Keeper) DecrementValidatorDelegationChangesInProgress(hostZone *types.HostZone, validatorAddress string) error {
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "validator %s not found", validatorAddress)
	}
	if validator.DelegationChangesInProgress == 0 {
		return errorsmod.Wrapf(types.ErrInvalidValidatorDelegationUpdates,
			"cannot decrement the number of delegation updates if the validator has 0 updates in progress")
	}
	validator.DelegationChangesInProgress -= 1
	hostZone.Validators[valIndex] = &validator
	return nil
}

// Checks if any validator's portion of the weight is greater than the cap
func (k Keeper) CheckValidatorWeightsBelowCap(ctx sdk.Context, validators []*types.Validator) error {
	// If there's only a few validators, don't enforce this yet
	if len(validators) < MinValidatorsBeforeWeightCapCheck {
		return nil
	}

	// The weight cap in params is an int representing a percentage (e.g. 10 is 10%)
	params := k.GetParams(ctx)
	validatorWeightCap := float64(params.ValidatorWeightCap)

	// Store a map of each validator weight, as well as the total
	totalWeight := float64(0)
	weightsByValidator := map[string]float64{}
	for _, validator := range validators {
		weightsByValidator[validator.Address] = float64(validator.Weight)
		totalWeight += float64(validator.Weight)
	}

	// If the total validator weights are 0, exit prematurely
	if totalWeight == 0 {
		return nil
	}

	// Check if any validator exceeds the cap
	for _, address := range utils.StringMapKeys[float64](weightsByValidator) {
		weightPercentage := weightsByValidator[address] / totalWeight * 100
		if weightPercentage > validatorWeightCap {
			return errorsmod.Wrapf(types.ErrValidatorExceedsWeightCap,
				"validator %s exceeds weight cap, has %v%% of the total weight when the cap is %v%%",
				address, weightPercentage, validatorWeightCap)
		}
	}

	return nil
}

// GetHostZoneFromIBCDenom returns a HostZone from a IBCDenom
func (k Keeper) GetHostZoneFromIBCDenom(ctx sdk.Context, denom string) (*types.HostZone, error) {
	var matchZone types.HostZone
	k.IterateHostZones(ctx, func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		if zoneInfo.IbcDenom == denom {
			matchZone = zoneInfo
			return nil
		}
		return nil
	})
	if matchZone.ChainId != "" {
		return &matchZone, nil
	}
	return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "No HostZone for %s found", denom)
}

// Validate whether a denom is a supported liquid staking token
func (k Keeper) CheckIsStToken(ctx sdk.Context, denom string) bool {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		if types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom) == denom {
			return true
		}
	}
	return false
}

// IterateHostZones iterates zones
func (k Keeper) IterateHostZones(ctx sdk.Context, fn func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))

	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		k.Logger(ctx).Debug(fmt.Sprintf("Iterating HostZone %d", i))
		zone := types.HostZone{}
		k.cdc.MustUnmarshal(iterator.Value(), &zone)

		error := fn(ctx, i, zone)

		if error != nil {
			break
		}
		i++
	}
}
