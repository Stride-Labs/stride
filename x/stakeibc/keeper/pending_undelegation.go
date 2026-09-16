package keeper

import (
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/store/v2/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// SetPendingUndelegation queues a one-shot undelegation amount for a host zone
// It is submitted through the normal undelegate pipeline at the next day epoch
func (k Keeper) SetPendingUndelegation(ctx sdk.Context, chainId string, amount sdkmath.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingUndelegationKeyPrefix))
	amountBz, err := amount.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte(chainId), amountBz)
}

// GetPendingUndelegation returns the pending undelegation amount for a host zone, if one is queued
func (k Keeper) GetPendingUndelegation(ctx sdk.Context, chainId string) (amount sdkmath.Int, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingUndelegationKeyPrefix))
	amountBz := store.Get([]byte(chainId))
	if amountBz == nil {
		return amount, false
	}
	if err := amount.Unmarshal(amountBz); err != nil {
		panic(err)
	}
	return amount, true
}

// RemovePendingUndelegation deletes the pending undelegation for a host zone
func (k Keeper) RemovePendingUndelegation(ctx sdk.Context, chainId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingUndelegationKeyPrefix))
	store.Delete([]byte(chainId))
}

// ConsumePendingUndelegation subtracts a successfully undelegated batch from the pending amount,
// removing the key once nothing is left
func (k Keeper) ConsumePendingUndelegation(ctx sdk.Context, chainId string, amount sdkmath.Int) {
	pending, found := k.GetPendingUndelegation(ctx, chainId)
	if !found {
		return
	}
	remaining := pending.Sub(amount)
	if !remaining.IsPositive() {
		k.RemovePendingUndelegation(ctx, chainId)
		return
	}
	k.SetPendingUndelegation(ctx, chainId, remaining)
}

// SetPendingUndelegationInFlight records how many undelegate ICA batches of the pending
// undelegation are awaiting an ack; the hook doesn't resubmit while any are outstanding
func (k Keeper) SetPendingUndelegationInFlight(ctx sdk.Context, chainId string, batches uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingUndelegationInFlightKeyPrefix))
	store.Set([]byte(chainId), sdk.Uint64ToBigEndian(batches))
}

// GetPendingUndelegationInFlight returns the number of undelegate ICA batches awaiting an ack
func (k Keeper) GetPendingUndelegationInFlight(ctx sdk.Context, chainId string) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingUndelegationInFlightKeyPrefix))
	batchesBz := store.Get([]byte(chainId))
	if batchesBz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(batchesBz)
}

// RemovePendingUndelegationInFlight clears the in-flight batch count (e.g. when the delegation
// channel is restored and the outstanding batches can never be acked)
func (k Keeper) RemovePendingUndelegationInFlight(ctx sdk.Context, chainId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingUndelegationInFlightKeyPrefix))
	store.Delete([]byte(chainId))
}

// DecrementPendingUndelegationInFlight marks one undelegate ICA batch as acked
func (k Keeper) DecrementPendingUndelegationInFlight(ctx sdk.Context, chainId string) {
	batches := k.GetPendingUndelegationInFlight(ctx, chainId)
	if batches == 0 {
		// Nothing should be acking: the batch was already released (e.g. by a restore)
		k.Logger(ctx).Error(utils.LogWithHostZone(chainId, "Pending undelegation ack received with no batches in flight"))
		return
	}
	if batches == 1 {
		k.RemovePendingUndelegationInFlight(ctx, chainId)
		return
	}
	k.SetPendingUndelegationInFlight(ctx, chainId, batches-1)
}

// IsActiveDelegationChannel reports whether a channel is the host zone's currently active
// delegation ICA channel (a closed, restored channel is not)
func (k Keeper) IsActiveDelegationChannel(ctx sdk.Context, hostZone types.HostZone, channelId string) bool {
	owner := types.FormatHostZoneICAOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	portId, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return false
	}
	activeChannelId, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, hostZone.ConnectionId, portId)
	return found && activeChannelId == channelId
}

// GetAllPendingUndelegations returns every queued pending undelegation
func (k Keeper) GetAllPendingUndelegations(ctx sdk.Context) (list []types.PendingUndelegation) {
	list = []types.PendingUndelegation{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingUndelegationKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var amount sdkmath.Int
		if err := amount.Unmarshal(iterator.Value()); err != nil {
			panic(err)
		}
		list = append(list, types.PendingUndelegation{ChainId: string(iterator.Key()), Amount: amount})
	}

	return list
}

// SubmitPendingUndelegations submits an undelegate ICA for each pending undelegation, using the
// normal validator capacity logic
//
// The ICA is submitted with no epoch unbonding record ids, so the undelegate callback only adjusts
// the validator and host zone delegation balances (nothing is burned). The pending amount stays in
// the store until a batch acks successfully, when that batch's amount is subtracted; a batch that
// fails or times out simply leaves the amount in place to be resubmitted. While any batch is
// awaiting an ack the host zone is skipped so the amount can't be submitted twice. If the channel
// dies with a batch in flight, its ack can never arrive and its timeout may never be processable,
// so RestoreInterchainAccount clears the in-flight count and the next day epoch resubmits
//
// A host zone that unbonds on this day epoch is skipped since InitiateAllHostZoneUnbondings has
// just consumed the same validator capacity, and the delegation balances the capacity is computed
// from are not decremented until that ICA's ack arrives
//
// Any failure (missing channel, insufficient capacity, ICA submit error) is logged and the amount
// is kept so the submission is retried at the next day epoch. This never returns an error or panics
// since it runs from the epoch hook
func (k Keeper) SubmitPendingUndelegations(ctx sdk.Context, epochNumber uint64) {
	for _, pending := range k.GetAllPendingUndelegations(ctx) {
		chainId := pending.ChainId
		amount := pending.Amount

		// A pending amount for a host zone that no longer exists can never be submitted, so drop it
		hostZone, found := k.GetHostZone(ctx, chainId)
		if !found {
			k.Logger(ctx).Error(utils.LogWithHostZone(chainId,
				"Host zone not found for pending undelegation of %v, removing pending key", amount))
			k.RemovePendingUndelegation(ctx, chainId)
			continue
		}

		// A previous batch is still awaiting its ack: resubmitting would undelegate the amount twice
		if inFlight := k.GetPendingUndelegationInFlight(ctx, chainId); inFlight > 0 {
			k.Logger(ctx).Info(utils.LogWithHostZone(chainId,
				"Pending undelegation of %v%s has %d batch(es) in flight, waiting for their acks", amount, hostZone.HostDenom, inFlight))
			continue
		}

		// Defer to the next day epoch if the regular unbonding flow runs for this host zone this epoch,
		// otherwise both would cascade onto the same validators and overshoot the on-chain delegation
		unbondingFrequency := hostZone.GetUnbondingFrequency()
		if epochNumber%unbondingFrequency == 0 {
			k.Logger(ctx).Info(utils.LogWithHostZone(chainId,
				"Host unbonds this epoch, deferring pending undelegation of %v%s to the next day epoch "+
					"(Unbonding Frequency: %d, Epoch: %d)", amount, hostZone.HostDenom, unbondingFrequency, epochNumber))
			continue
		}

		k.Logger(ctx).Info(utils.LogWithHostZone(chainId,
			"Submitting pending undelegation of %v%s", amount, hostZone.HostDenom))

		// Cascade the amount across the validators by unbond capacity and submit it through the same
		// batch submitter as the regular unbonding flow, which also flags the delegation changes in
		// progress on each validator (required so the undelegate callback can decrement them)
		err := utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			msgs, splits, err := k.GetUndelegateMessagesForAmount(ctx, hostZone, amount)
			if err != nil {
				return err
			}
			if len(msgs) == 0 {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no undelegate messages built for pending undelegation")
			}

			batchSize := int(utils.UintToInt(hostZone.MaxMessagesPerIcaTx))
			numTxsSubmitted, err := k.BatchSubmitUndelegateICAMessages(ctx, hostZone, nil, msgs, splits, batchSize)
			if err != nil {
				return err
			}

			// The amount stays pending until each batch acks; the count blocks a duplicate submission
			k.SetPendingUndelegationInFlight(ctx, chainId, numTxsSubmitted)
			EmitUndelegationEvent(ctx, hostZone, amount)
			return nil
		})
		if err != nil {
			k.Logger(ctx).Error(utils.LogWithHostZone(chainId,
				"Unable to submit pending undelegation of %v%s, will retry next day epoch: %s",
				amount, hostZone.HostDenom, err.Error()))
		}
	}
}
