package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v29/utils"
	"github.com/Stride-Labs/stride/v29/x/staketia/types"
)

// IBC transfers all TIA in the deposit account and sends it to the delegation account
func (k Keeper) PrepareDelegation(ctx sdk.Context, epochNumber uint64, epochDuration time.Duration) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(types.CelestiaChainId, "Preparing delegation for epoch %d", epochNumber))

	// Only send the transfer if the host zone isn't halted
	hostZone, err := k.GetUnhaltedHostZone(ctx)
	if err != nil {
		return err
	}

	// safety check: if any delegation records are in progress, do not allow another transfer
	delegationRecords := k.GetAllActiveDelegationRecords(ctx)
	for _, record := range delegationRecords {
		if record.Status == types.TRANSFER_IN_PROGRESS {
			return errorsmod.Wrapf(types.ErrInvariantBroken,
				"cannot prepare delegation while a transfer is in progress, record ID %d", record.Id)
		}
	}

	// Transfer the full deposit balance which will include new liquid stakes, as well as reinvestment
	depositAddress := sdk.MustAccAddressFromBech32(hostZone.DepositAddress)
	nativeTokens := k.bankKeeper.GetBalance(ctx, depositAddress, hostZone.NativeTokenIbcDenom)

	// If there's nothing to delegate, exit early - no need to create a new record
	if nativeTokens.Amount.IsZero() {
		k.Logger(ctx).Info(utils.LogWithHostZone(types.CelestiaChainId, "No new liquid stakes for epoch %d", epochNumber))
		return nil
	}

	// Create a new delgation record with status TRANSFER IN PROGRESS
	delegationRecord := types.DelegationRecord{
		Id:           epochNumber,
		NativeAmount: nativeTokens.Amount,
		Status:       types.TRANSFER_IN_PROGRESS,
	}
	err = k.SafelySetDelegationRecord(ctx, delegationRecord)
	if err != nil {
		return err
	}

	// Timeout the transfer at the end of the epoch
	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(epochDuration).UnixNano())

	// Transfer the native tokens to the host chain
	transferMsgDepositToDelegation := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostZone.TransferChannelId,
		Token:            nativeTokens,
		Sender:           hostZone.DepositAddress,
		Receiver:         hostZone.DelegationAddress,
		TimeoutTimestamp: timeoutTimestamp,
	}
	msgResponse, err := k.transferKeeper.Transfer(ctx, &transferMsgDepositToDelegation)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to submit transfer from deposit to delegation acct in PrepareDelegation")
	}

	// Store the record ID so that we can access it during the packet callback to update the record status
	k.SetTransferInProgressRecordId(ctx, hostZone.TransferChannelId, msgResponse.Sequence, delegationRecord.Id)

	return nil
}

// Confirms a delegation has completed on the host zone, increments the internal delegated balance,
// and archives the record
func (k Keeper) ConfirmDelegation(ctx sdk.Context, recordId uint64, txHash string, sender string) (err error) {
	// grab unbonding record, verify record is ready to be delegated, and a hash hasn't already been posted
	delegationRecord, found := k.GetDelegationRecord(ctx, recordId)
	if !found {
		return types.ErrDelegationRecordNotFound.Wrapf("delegation record not found for %v", recordId)
	}
	if delegationRecord.Status != types.DELEGATION_QUEUE {
		return types.ErrDelegationRecordInvalidState.Wrapf("delegation record %v is not in the correct state", recordId)
	}
	if delegationRecord.TxHash != "" {
		return types.ErrDelegationRecordInvalidState.Wrapf("delegation record %v already has a txHash", recordId)
	}

	// note: we're intentionally not checking that the host zone is halted, because we still want to process this tx in that case
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}
	stakeibcHostZone, err := k.stakeibcKeeper.GetActiveHostZone(ctx, types.CelestiaChainId)
	if err != nil {
		return err
	}

	// verify delegation record is nonzero
	if !delegationRecord.NativeAmount.IsPositive() {
		return types.ErrDelegationRecordInvalidState.Wrapf("delegation record %v has non positive delegation", recordId)
	}

	// update delegation record to archive it
	delegationRecord.TxHash = txHash
	delegationRecord.Status = types.DELEGATION_COMPLETE
	k.ArchiveDelegationRecord(ctx, delegationRecord)

	// increment delegation on Host Zone
	hostZone.RemainingDelegatedBalance = hostZone.RemainingDelegatedBalance.Add(delegationRecord.NativeAmount)
	stakeibcHostZone.TotalDelegations = stakeibcHostZone.TotalDelegations.Add(delegationRecord.NativeAmount)
	k.SetHostZone(ctx, hostZone)
	k.stakeibcKeeper.SetHostZone(ctx, stakeibcHostZone)

	EmitSuccessfulConfirmDelegationEvent(ctx, recordId, delegationRecord.NativeAmount, txHash, sender)
	return nil
}

// Runs prepare delegations with a cache context wrapper so revert any partial state changes
func (k Keeper) SafelyPrepareDelegation(ctx sdk.Context, epochNumber uint64, epochDuration time.Duration) error {
	return utils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		return k.PrepareDelegation(ctx, epochNumber, epochDuration)
	})
}
