package keeper

import (
	"errors"
	"fmt"

	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v14/utils"
	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// Marshalls delegate callback arguments
func (k Keeper) MarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback types.DelegateCallback) ([]byte, error) {
	out, err := proto.Marshal(&delegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalDelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// Unmarshalls delegate callback arguments into a DelegateCallback struct
func (k Keeper) UnmarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback []byte) (*types.DelegateCallback, error) {
	unmarshalledDelegateCallback := types.DelegateCallback{}
	if err := proto.Unmarshal(delegateCallback, &unmarshalledDelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalDelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledDelegateCallback, nil
}

// ICA Callback after delegating deposit records
// * If successful: Updates deposit record status and records delegation changes on the host zone and validators
// * If timeout:    Does nothing
// * If failure:    Reverts deposit record status
func (k Keeper) DelegateCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Deserialize the callback args
	delegateCallback, err := k.UnmarshalDelegateCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal delegate callback args: %s", err.Error()))
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

	// Regardless of failure/success/timeout, indicate that this ICA has completed
	for _, splitDelegation := range delegateCallback.SplitDelegations {
		if err := k.DecrementValidatorDelegationChangesInProgress(&hostZone, splitDelegation.Validator); err != nil {
			// TODO: Revert after v14 upgrade
			if errors.Is(err, types.ErrInvalidValidatorDelegationUpdates) {
				k.Logger(ctx).Error(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Delegate,
					"Invariant failed - delegation changes in progress fell below 0 for %s", splitDelegation.Validator))
				continue
			}
			return err
		}
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

	// Update delegations on the host zone
	for _, splitDelegation := range delegateCallback.SplitDelegations {
		err := k.AddDelegationToValidator(ctx, &hostZone, splitDelegation.Validator, splitDelegation.Amount, ICACallbackID_Delegate)
		if err != nil {
			return errorsmod.Wrapf(err, "Failed to add delegation to validator")
		}
	}
	k.SetHostZone(ctx, hostZone)

	k.RecordsKeeper.RemoveDepositRecord(ctx, cast.ToUint64(recordId))
	k.Logger(ctx).Info(fmt.Sprintf("[DELEGATION] success on %s", chainId))
	return nil
}
