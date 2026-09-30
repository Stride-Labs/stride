// x/stakeibc/keeper/wind_down_undelegate.go
package keeper

import (
	"github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v34/utils"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// UndelegateFromValidators is the wind-down drain (spec §7): it submits one MsgUndelegate per
// listed validator (or per funded validator when the list is empty) for that validator's
// recorded delegation minus its offset, through the same record-less batch path the v34
// pending-undelegation pipeline uses. The callback decrements the recorded delegations on
// ack; nothing is changed here.
//
// It refuses, before anything is submitted: a deprecated zone, a zone that still has an
// unbonding record queued or retrying (the record-driven path could never submit it on a
// drained zone, STRIDE-07), a zone with a stored pending undelegation, and a target validator
// with a delegation change or a slash query in progress.
func (k Keeper) UndelegateFromValidators(ctx sdk.Context, msg *types.MsgUndelegateFromValidators) (numBatches uint64, err error) {
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return 0, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", msg.ChainId)
	}
	if hostZone.Deprecated {
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "host zone %s is deprecated", msg.ChainId)
	}
	if err := k.checkNoQueuedUnbondings(ctx, msg.ChainId); err != nil {
		return 0, err
	}

	msgs, splits, err := k.BuildUndelegateFromValidatorsMsgs(hostZone, msg.Validators)
	if err != nil {
		return 0, err
	}

	// Same batch size as the epochly unbonding; a zone that never set it gets the default
	batchSize := int(utils.UintToInt(hostZone.MaxMessagesPerIcaTx))
	if batchSize == 0 {
		batchSize = int(DefaultMaxMessagesPerIcaTx)
	}
	numBatches, err = k.BatchSubmitUndelegateICAMessages(ctx, hostZone, nil, msgs, splits, batchSize)
	if err != nil {
		return 0, err
	}

	// Register the batches in flight so the record-less callback's decrement finds a counter
	inFlight := k.GetPendingUndelegationInFlight(ctx, msg.ChainId)
	k.SetPendingUndelegationInFlight(ctx, msg.ChainId, inFlight+numBatches)

	totalUnbondAmount := k.CalculateTotalUnbondedInBatch(splits)
	k.Logger(ctx).Info(utils.LogWithHostZone(msg.ChainId,
		"Wind-down undelegation of %v%s across %d validator(s) in %d batch(es)",
		totalUnbondAmount, hostZone.HostDenom, len(msgs), numBatches))
	EmitUndelegationEvent(ctx, hostZone, totalUnbondAmount)

	return numBatches, nil
}

// checkNoQueuedUnbondings rejects the drain while any unbonding record for the zone is still
// waiting for the day epoch (queued or retrying) with a real amount, has an undelegate ICA
// in flight, or while a one-shot
// pending undelegation (the v34 mechanism) is stored for the zone: the record-less success
// callback calls ConsumePendingUndelegation, so a drain ack would silently eat that amount
func (k Keeper) checkNoQueuedUnbondings(ctx sdk.Context, chainId string) error {
	if pending, found := k.GetPendingUndelegation(ctx, chainId); found {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"host zone %s has a pending undelegation of %v queued by an upgrade; let the day epoch submit it before draining", chainId, pending)
	}
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, chainId)
		if !found {
			continue
		}
		// A record-driven undelegate ICA that has not acked yet: UndelegationTxsInProgress is
		// incremented on submission and decremented on every ack, success or failure, so the
		// drain waits (spec section 6). An UNBONDING_IN_PROGRESS record with the counter at zero
		// is just waiting out the unbonding period and does not block
		if hostZoneUnbonding.UndelegationTxsInProgress > 0 {
			return types.ErrHostZoneUnbondingPending.Wrapf(
				"epoch %d record for %s has %d undelegate ICA(s) awaiting an ack; wait for them before draining",
				epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding.UndelegationTxsInProgress)
		}
		queued := hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_QUEUE ||
			hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE
		hasAmount := !hostZoneUnbonding.NativeTokenAmount.IsNil() && hostZoneUnbonding.NativeTokenAmount.IsPositive()
		if queued && hasAmount {
			return types.ErrHostZoneUnbondingPending.Wrapf(
				"epoch %d record for %s is %s with %v; wait for the day epoch to submit it before draining",
				epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding.Status, hostZoneUnbonding.NativeTokenAmount)
		}
	}
	return nil
}

// BuildUndelegateFromValidatorsMsgs builds one MsgUndelegate and one SplitUndelegation per
// target validator. An empty request means every validator with a positive recorded
// delegation. Each amount is the recorded delegation minus the offset, passed through the same
// rounding safety the epochly unbonding applies to a full drain of a slashed validator.
func (k Keeper) BuildUndelegateFromValidatorsMsgs(
	hostZone types.HostZone,
	requested []types.ValidatorUndelegation,
) (msgs []proto.Message, splits []*types.SplitUndelegation, err error) {
	targets := requested
	if len(targets) == 0 {
		for _, validator := range hostZone.Validators {
			if !validator.Delegation.IsNil() && validator.Delegation.IsPositive() {
				targets = append(targets, types.ValidatorUndelegation{Address: validator.Address, Offset: sdkmath.ZeroInt()})
			}
		}
	}
	if len(targets) == 0 {
		return nil, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "no validator on %s has a delegation to undelegate", hostZone.ChainId)
	}

	for _, target := range targets {
		validator, _, found := GetValidatorFromAddress(hostZone.Validators, target.Address)
		if !found {
			return nil, nil, types.ErrValidatorNotFound.Wrapf("validator %s not found on %s", target.Address, hostZone.ChainId)
		}
		if validator.DelegationChangesInProgress > 0 {
			return nil, nil, types.ErrInvalidDelegationsInProgress.Wrapf(
				"validator %s has %d delegation change(s) in progress", target.Address, validator.DelegationChangesInProgress)
		}
		// The record-driven path also excludes these; a validator mid-slash-query has a recorded
		// delegation that may be about to move, so ops wait for the day-0 refresh callbacks
		if validator.SlashQueryInProgress {
			return nil, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
				"validator %s has a slash query in progress; wait for its callback before draining", target.Address)
		}

		delegation := validator.Delegation
		if delegation.IsNil() {
			delegation = sdkmath.ZeroInt()
		}
		offset := target.Offset
		if offset.IsNil() {
			offset = sdkmath.ZeroInt()
		}
		amount := delegation.Sub(offset)
		if !amount.IsPositive() {
			return nil, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
				"validator %s: delegation %v minus offset %v is not positive", target.Address, delegation, offset)
		}

		capacity := ValidatorUnbondCapacity{ValidatorAddress: validator.Address, CurrentDelegation: delegation}
		amount = k.applySharesRoundingSafety(hostZone, capacity, amount)

		msgs = append(msgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: hostZone.DelegationIcaAddress,
			ValidatorAddress: validator.Address,
			Amount:           sdk.NewCoin(hostZone.HostDenom, amount),
		})
		splits = append(splits, &types.SplitUndelegation{
			Validator:         validator.Address,
			NativeTokenAmount: amount,
		})
	}
	return msgs, splits, nil
}
