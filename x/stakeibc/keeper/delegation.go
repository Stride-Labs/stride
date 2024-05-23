package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v22/utils"
	recordstypes "github.com/Stride-Labs/stride/v22/x/records/types"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

// Builds the delegation ICA messags for a given deposit record
// Each validator has a portion of the total amount on the record based on their weight
func (k Keeper) GetDelegationICAMessages(
	ctx sdk.Context,
	hostZone types.HostZone,
	amt sdk.Coin, depositRecord recordstypes.DepositRecord,
) (msgs []proto.Message, delegations []*types.SplitDelegation, err error) {
	// Fetch the relevant ICA
	if hostZone.DelegationIcaAddress == "" {
		return msgs, delegations, errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", hostZone.ChainId)
	}

	// Construct the transaction
	targetDelegatedAmts, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, amt.Amount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target delegation amounts for host zone %s", hostZone.ChainId))
		return msgs, delegations, err
	}

	for _, validator := range hostZone.Validators {
		relativeAmount, ok := targetDelegatedAmts[validator.Address]
		if !ok || !relativeAmount.IsPositive() {
			continue
		}

		msgs = append(msgs, &stakingtypes.MsgDelegate{
			DelegatorAddress: hostZone.DelegationIcaAddress,
			ValidatorAddress: validator.Address,
			Amount:           sdk.NewCoin(amt.Denom, relativeAmount),
		})
		delegations = append(delegations, &types.SplitDelegation{
			Validator: validator.Address,
			Amount:    relativeAmount,
		})
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Preparing MsgDelegates from the delegation account to each validator"))

	if len(msgs) == 0 {
		return msgs, delegations, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Target delegation amount was 0 for each validator")
	}

	return msgs, delegations, nil
}

// Submit undelegate ICA messages in small batches to reduce the gas size per tx
func (k Keeper) BatchSubmitDelegationICAMessages(
	ctx sdk.Context,
	hostZone types.HostZone,
	depositRecord recordstypes.DepositRecord,
	msgs []proto.Message,
	delegations []*types.SplitDelegation,
	batchSize int,
) (numTxsSubmitted uint64, err error) {
	// Iterate the full list of messages and submit in batches
	for start := 0; start < len(msgs); start += batchSize {
		end := start + batchSize
		if end > len(msgs) {
			end = len(msgs)
		}

		msgBatch := msgs[start:end]
		delegationsBatch := delegations[start:end]

		// Store the callback data
		delegateCallback := types.DelegateCallback{
			HostZoneId:       hostZone.ChainId,
			DepositRecordId:  depositRecord.Id,
			SplitDelegations: delegationsBatch,
		}
		marshalledCallbackArgs, err := k.MarshalDelegateCallbackArgs(ctx, delegateCallback)
		if err != nil {
			return 0, err
		}

		// Send the transaction through SubmitTx
		_, err = k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgBatch, types.ICAAccountType_DELEGATION, ICACallbackID_Delegate, marshalledCallbackArgs)
		if err != nil {
			return 0, errorsmod.Wrapf(err, "Failed to SubmitTxs for connectionId %s on %s. Messages: %s", hostZone.ConnectionId, hostZone.ChainId, msgs)
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "ICA MsgDelegates Successfully Sent"))

		// flag the delegation change in progress on each validator
		for _, delegation := range delegationsBatch {
			if err := k.IncrementValidatorDelegationChangesInProgress(&hostZone, delegation.Validator); err != nil {
				return 0, err
			}
		}
		k.SetHostZone(ctx, hostZone)

		numTxsSubmitted += 1
	}

	return numTxsSubmitted, nil
}

// Iterate each deposit record marked DELEGATION_QUEUE and use the delegation ICA to delegate on the host zone
func (k Keeper) StakeExistingDepositsOnHostZones(ctx sdk.Context, epochNumber uint64, depositRecords []recordstypes.DepositRecord) {
	k.Logger(ctx).Info("Staking deposit records...")

	stakeDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		isStakeRecord := record.Status == recordstypes.DepositRecord_DELEGATION_QUEUE
		isBeforeCurrentEpoch := record.DepositEpochNumber < epochNumber
		isNotInProgress := record.DelegationTxsInProgress == 0
		return isStakeRecord && isBeforeCurrentEpoch && isNotInProgress
	})

	if len(stakeDepositRecords) == 0 {
		k.Logger(ctx).Info("No deposit records in state DELEGATION_QUEUE")
		return
	}

	for _, depositRecord := range stakeDepositRecords {
		if depositRecord.Amount.IsZero() {
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId,
			"Processing deposit record %d: %v%s", depositRecord.Id, depositRecord.Amount, depositRecord.Denom))

		hostZone, err := k.GetActiveHostZone(ctx, depositRecord.HostZoneId)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Not processing %d, %s", depositRecord.Id, err.Error()))
			continue
		}

		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Staking %v%s", depositRecord.Amount, hostZone.HostDenom))
		stakeAmount := sdk.NewCoin(hostZone.HostDenom, depositRecord.Amount)

		// Build the list of delegation messages for each validator
		msgs, delegations, err := k.GetDelegationICAMessages(ctx, hostZone, stakeAmount, depositRecord)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s | err: %s", stakeAmount.String(), hostZone.ChainId, err.Error()))
			continue
		}

		// Submit the delegation messages in batchs
		delegateBatchSize := int(hostZone.MaxMessagesPerIcaTx)
		numTxsSubmitted, err := k.BatchSubmitDelegationICAMessages(ctx, hostZone, depositRecord, msgs, delegations, delegateBatchSize)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Faild to submit delegation ICAs: %s", err.Error()))
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Successfully submitted stake"))

		// Increment the number of tx sin progress on the record and update the status
		depositRecord.Status = recordstypes.DepositRecord_DELEGATION_IN_PROGRESS
		depositRecord.DelegationTxsInProgress += numTxsSubmitted
		k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("hostZone", hostZone.ChainId),
				sdk.NewAttribute("newAmountStaked", depositRecord.Amount.String()),
			),
		)
	}
}

// Delegates accrued staking rewards for reinvestment
func (k Keeper) ReinvestRewards(ctx sdk.Context) {
	k.Logger(ctx).Info("Reinvesting tokens...")

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		// only process host zones once withdrawal accounts are registered
		if hostZone.WithdrawalIcaAddress == "" {
			k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Withdrawal account not registered for host zone"))
			continue
		}

		// read clock time on host zone
		blockTime, err := k.GetLightClientTimeSafely(ctx, hostZone.ConnectionId)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not find blockTime for host zone %s, err: %s", hostZone.ConnectionId, err.Error()))
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "BlockTime for host zone: %d", blockTime))

		err = k.SubmitWithdrawalHostBalanceICQ(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Error updating withdrawal balance for host zone %s: %s", hostZone.ConnectionId, err.Error()))
			continue
		}
	}
}
