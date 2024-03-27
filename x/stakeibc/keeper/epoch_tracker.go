package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"

	epochstypes "github.com/Stride-Labs/stride/v21/x/epochs/types"
	"github.com/Stride-Labs/stride/v21/x/stakeibc/types"
)

// SetEpochTracker set a specific epochTracker in the store from its index
func (k Keeper) SetEpochTracker(ctx sdk.Context, epochTracker types.EpochTracker) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochTrackerKeyPrefix))
	b := k.cdc.MustMarshal(&epochTracker)
	store.Set(types.EpochTrackerKey(
		epochTracker.EpochIdentifier,
	), b)
}

// GetEpochTracker returns a epochTracker from its index
func (k Keeper) GetEpochTracker(
	ctx sdk.Context,
	epochIdentifier string,
) (val types.EpochTracker, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochTrackerKeyPrefix))

	b := store.Get(types.EpochTrackerKey(
		epochIdentifier,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveEpochTracker removes a epochTracker from the store
func (k Keeper) RemoveEpochTracker(
	ctx sdk.Context,
	epochIdentifier string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochTrackerKeyPrefix))
	store.Delete(types.EpochTrackerKey(
		epochIdentifier,
	))
}

// GetAllEpochTracker returns all epochTracker
func (k Keeper) GetAllEpochTracker(ctx sdk.Context) (list []types.EpochTracker) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochTrackerKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.EpochTracker
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// Update the epoch information in the stakeibc epoch tracker
func (k Keeper) UpdateEpochTracker(ctx sdk.Context, epochInfo epochstypes.EpochInfo) (epochNumber uint64, err error) {
	epochNumber, err = cast.ToUint64E(epochInfo.CurrentEpoch)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch number to uint64: %v", err))
		return 0, err
	}
	epochDurationNano, err := cast.ToUint64E(epochInfo.Duration.Nanoseconds())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch duration to uint64: %v", err))
		return 0, err
	}
	nextEpochStartTime, err := cast.ToUint64E(epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration).UnixNano())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch duration to uint64: %v", err))
		return 0, err
	}
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochInfo.Identifier,
		EpochNumber:        epochNumber,
		Duration:           epochDurationNano,
		NextEpochStartTime: nextEpochStartTime,
	}
	k.SetEpochTracker(ctx, epochTracker)

	return epochNumber, nil
}

// helper to get what share of the curr epoch we're through
func (k Keeper) GetStrideEpochElapsedShare(ctx sdk.Context) (sdk.Dec, error) {
	// Get the current stride epoch
	epochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		errMsg := fmt.Sprintf("Failed to get epoch tracker for %s", epochstypes.STRIDE_EPOCH)
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(sdkerrors.ErrNotFound, errMsg)
	}

	// Get epoch start time, end time, and duration
	epochDuration, err := cast.ToInt64E(epochTracker.Duration)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert epoch duration to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrIntCast, errMsg)
	}
	epochEndTime, err := cast.ToInt64E(epochTracker.NextEpochStartTime)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert next epoch start time to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrIntCast, errMsg)
	}
	epochStartTime := epochEndTime - epochDuration

	// Confirm the current block time is inside the current epoch's start and end times
	currBlockTime := ctx.BlockTime().UnixNano()
	if currBlockTime < epochStartTime || currBlockTime > epochEndTime {
		errMsg := fmt.Sprintf("current block time %d is not within current epoch (ending at %d)", currBlockTime, epochTracker.NextEpochStartTime)
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrInvalidEpoch, errMsg)
	}

	// Get elapsed share
	elapsedTime := currBlockTime - epochStartTime
	elapsedShare := sdk.NewDec(elapsedTime).Quo(sdk.NewDec(epochDuration))
	if elapsedShare.LT(sdk.ZeroDec()) || elapsedShare.GT(sdk.OneDec()) {
		errMsg := fmt.Sprintf("elapsed share (%s) for epoch is not between 0 and 1", elapsedShare)
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrInvalidEpoch, errMsg)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Epoch elapsed share: %v (Block Time: %d, Epoch End Time: %d)", elapsedShare, currBlockTime, epochEndTime))
	return elapsedShare, nil
}
