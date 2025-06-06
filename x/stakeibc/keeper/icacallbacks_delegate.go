package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v27/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v27/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v27/x/records/types"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

// ICA Callback after delegating deposit records
// * If successful: Updates deposit record status and records delegation changes on the host zone and validators
// * If timeout:    Does nothing
// * If failure:    Reverts deposit record status
func (k Keeper) DelegateCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Deserialize the callback args
	delegateCallback := types.DelegateCallback{}
	if err := proto.Unmarshal(args, &delegateCallback); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal delegate callback")
	}
	chainId := delegateCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Delegate,
		"Starting delegate callback for Deposit Record: %d", delegateCallback.DepositRecordId))

	// Confirm chainId and deposit record Id exist
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "host zone not found %s", chainId)
	}
	recordId := delegateCallback.DepositRecordId
	depositRecord, found := k.RecordsKeeper.GetDepositRecord(ctx, recordId)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "deposit record not found %d", recordId)
	}

	// Regardless of failure/success/timeout, indicate that this ICA has completed on the deposit record
	if depositRecord.DelegationTxsInProgress == 0 {
		return types.ErrInvalidDelegationsInProgress.Wrapf("delegation changes in progress is already 0 and can't be decremented")
	}
	depositRecord.DelegationTxsInProgress -= 1
	k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)

	// Regardless of failure/success/timeout, indicate that this ICA has completed on each validator
	// Sum up the total delegated in the process
	totalDelegatedInBatch := sdkmath.ZeroInt()
	for _, splitDelegation := range delegateCallback.SplitDelegations {
		if err := k.DecrementValidatorDelegationChangesInProgress(&hostZone, splitDelegation.Validator); err != nil {
			return err
		}
		totalDelegatedInBatch = totalDelegatedInBatch.Add(splitDelegation.Amount)
	}
	k.SetHostZone(ctx, hostZone)

	// Check for timeout (ack nil)
	// No need to reset the deposit record status since it will get reverted when the channel is restored
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Delegate,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// Reset the deposit record status upon failure
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Delegate,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))

		// Reset deposit record status
		depositRecord.Status = recordstypes.DepositRecord_DELEGATION_QUEUE
		k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Delegate,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Decrement the amount on the deposit record
	// If there's nothing left on the deposit record, remove it
	depositRecord.Amount = depositRecord.Amount.Sub(totalDelegatedInBatch)
	if depositRecord.Amount.IsZero() {
		k.RecordsKeeper.RemoveDepositRecord(ctx, recordId)
	} else {
		k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
	}

	// Update delegations on the validators and host zone
	for _, splitDelegation := range delegateCallback.SplitDelegations {
		err := k.AddDelegationToValidator(ctx, &hostZone, splitDelegation.Validator, splitDelegation.Amount, ICACallbackID_Delegate)
		if err != nil {
			return errorsmod.Wrapf(err, "Failed to add delegation to validator")
		}
	}
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(fmt.Sprintf("[DELEGATION] success on %s", chainId))
	return nil
}
