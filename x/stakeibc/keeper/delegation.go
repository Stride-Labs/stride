package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cast"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/Stride-Labs/stride/v18/utils"
	recordstypes "github.com/Stride-Labs/stride/v18/x/records/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

func (k Keeper) DelegateOnHost(ctx sdk.Context, hostZone types.HostZone, amt sdk.Coin, depositRecord recordstypes.DepositRecord) error {
	// TODO: Remove this block and use connection-id from host zone
	// the relevant ICA is the delegate account
	owner := types.FormatHostZoneICAOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "%s has no associated portId", owner)
	}
	connectionId, found := k.GetConnectionIdFromICAPortId(ctx, portID)
	if !found {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "unable to find ICA connection Id for port %s", portID)
	}

	// Fetch the relevant ICA
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", hostZone.ChainId)
	}

	// Construct the transaction
	targetDelegatedAmts, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, amt.Amount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target delegation amounts for host zone %s", hostZone.ChainId))
		return err
	}

	var splitDelegations []*types.SplitDelegation
	var msgs []proto.Message
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
		splitDelegations = append(splitDelegations, &types.SplitDelegation{
			Validator: validator.Address,
			Amount:    relativeAmount,
		})
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Preparing MsgDelegates from the delegation account to each validator"))

	if len(msgs) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Target delegation amount was 0 for each validator")
	}

	// add callback data
	delegateCallback := types.DelegateCallback{
		HostZoneId:       hostZone.ChainId,
		DepositRecordId:  depositRecord.Id,
		SplitDelegations: splitDelegations,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Marshalling DelegateCallback args: %+v", delegateCallback))
	marshalledCallbackArgs, err := k.MarshalDelegateCallbackArgs(ctx, delegateCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, types.ICAAccountType_DELEGATION, ICACallbackID_Delegate, marshalledCallbackArgs)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to SubmitTxs for connectionId %s on %s. Messages: %s", connectionId, hostZone.ChainId, msgs)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "ICA MsgDelegates Successfully Sent"))

	// flag the delegation change in progress on each validator
	for _, splitDelegation := range splitDelegations {
		if err := k.IncrementValidatorDelegationChangesInProgress(&hostZone, splitDelegation.Validator); err != nil {
			return err
		}
	}
	k.SetHostZone(ctx, hostZone)

	// update the record state to DELEGATION_IN_PROGRESS
	depositRecord.Status = recordstypes.DepositRecord_DELEGATION_IN_PROGRESS
	k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)

	return nil
}

// Iterate each deposit record marked DELEGATION_QUEUE and use the delegation ICA to delegate on the host zone
func (k Keeper) StakeExistingDepositsOnHostZones(ctx sdk.Context, epochNumber uint64, depositRecords []recordstypes.DepositRecord) {
	k.Logger(ctx).Info("Staking deposit records...")

	stakeDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		isStakeRecord := record.Status == recordstypes.DepositRecord_DELEGATION_QUEUE
		isBeforeCurrentEpoch := record.DepositEpochNumber < epochNumber
		return isStakeRecord && isBeforeCurrentEpoch
	})

	if len(stakeDepositRecords) == 0 {
		k.Logger(ctx).Info("No deposit records in state DELEGATION_QUEUE")
		return
	}

	// limit the number of staking deposits to process per epoch
	maxDepositRecordsToStake := utils.Min(len(stakeDepositRecords), cast.ToInt(k.GetParam(ctx, types.KeyMaxStakeICACallsPerEpoch)))
	if maxDepositRecordsToStake < len(stakeDepositRecords) {
		k.Logger(ctx).Info(fmt.Sprintf("  MaxStakeICACallsPerEpoch limit reached - Only staking %d out of %d deposit records", maxDepositRecordsToStake, len(stakeDepositRecords)))
	}

	for _, depositRecord := range stakeDepositRecords[:maxDepositRecordsToStake] {
		if depositRecord.Amount.IsZero() {
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId,
			"Processing deposit record %d: %v%s", depositRecord.Id, depositRecord.Amount, depositRecord.Denom))

		hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
		if !hostZoneFound {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Host zone not found for deposit record {%d}", depositRecord.Id))
			continue
		}

		if hostZone.Halted {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Host zone halted for deposit record {%d}", depositRecord.Id))
			continue
		}

		if hostZone.DelegationIcaAddress == "" {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Zone %s is missing a delegation address!", hostZone.ChainId))
			continue
		}

		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Staking %v%s", depositRecord.Amount, hostZone.HostDenom))
		stakeAmount := sdk.NewCoin(hostZone.HostDenom, depositRecord.Amount)

		err := k.DelegateOnHost(ctx, hostZone, stakeAmount, depositRecord)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s | err: %s", stakeAmount.String(), hostZone.ChainId, err.Error()))
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Successfully submitted stake"))

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

		err = k.UpdateWithdrawalBalance(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Error updating withdrawal balance for host zone %s: %s", hostZone.ConnectionId, err.Error()))
			continue
		}
	}
}
