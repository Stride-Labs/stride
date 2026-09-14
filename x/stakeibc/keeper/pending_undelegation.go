package keeper

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/store/v2/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
// normal validator capacity logic, and clears the pending key once the ICA has been submitted
//
// The ICA is submitted with no epoch unbonding record ids, so the undelegate callback only adjusts
// the validator and host zone delegation balances (nothing is burned). If the ICA fails or times
// out, the callback re-queues the amount so it's resubmitted at a later day epoch
//
// A host zone that unbonds on this day epoch is skipped (key kept) since InitiateAllHostZoneUnbondings
// has just consumed the same validator capacity, and the delegation balances the capacity is computed
// from are not decremented until that ICA's ack arrives
//
// Any failure (missing channel, insufficient capacity, ICA submit error) is logged and the key is
// kept so the submission is retried at the next day epoch. This never returns an error or panics
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

			batchSize := int(utils.UintToInt(hostZone.MaxMessagesPerIcaTx))
			if _, err := k.BatchSubmitUndelegateICAMessages(ctx, hostZone, nil, msgs, splits, batchSize); err != nil {
				return err
			}

			k.RemovePendingUndelegation(ctx, chainId)
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
