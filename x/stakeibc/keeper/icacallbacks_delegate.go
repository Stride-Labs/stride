package keeper

import (
	"fmt"

	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/utils"
	"github.com/Stride-Labs/stride/v4/x/icacallbacks"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
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
//   If successful:
//      * Updates deposit record status and records delegation changes on the host zone and validators
//   If timeout:
//      * Does nothing
//   If failure:
//		* Reverts deposit record status
func DelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	// Deserialize the callback args
	delegateCallback, err := k.UnmarshalDelegateCallbackArgs(ctx, args)
	if err != nil {
		return err
	}
	chainId := delegateCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Delegate,
		"Starting callback for Deposit Record: %d", delegateCallback.DepositRecordId))

	// Confirm chainId and deposit record Id exist
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "host zone not found %s", chainId)
	}
	recordId := delegateCallback.DepositRecordId
	depositRecord, found := k.RecordsKeeper.GetDepositRecord(ctx, recordId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "deposit record not found %d", recordId)
	}

	// Check for timeout (ack nil)
	// No need to reset the deposit record status since it will get revertted when the channel is restored
	if ack == nil {
		k.Logger(ctx).Error(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Delegate,
			"TIMEOUT (ack is nil), Packet: %+v", packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// Reset the deposit record status upon failure
	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to fetch txMsgData, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}
	if len(txMsgData.Data) == 0 {
		k.Logger(ctx).Error(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Delegate,
			"ICA TX FAILED (ack is empty / ack error), Packet: %+v", packet))

		// Reset deposit record status
		depositRecord.Status = recordstypes.DepositRecord_DELEGATION_QUEUE
		k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
		return nil
	}

	k.Logger(ctx).Info(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Delegate, "SUCCESS, Packet: %+v", packet))

	// Update delegations on the host zone
	for _, splitDelegation := range delegateCallback.SplitDelegations {
		hostZone.StakedBal = hostZone.StakedBal.Add(splitDelegation.Amount)
		success := k.AddDelegationToValidator(ctx, hostZone, splitDelegation.Validator, splitDelegation.Amount, ICACallbackID_Delegate)
		if !success {
			return sdkerrors.Wrapf(types.ErrValidatorDelegationChg, "Failed to add delegation to validator")
		}
		k.SetHostZone(ctx, hostZone)
	}

	k.RecordsKeeper.RemoveDepositRecord(ctx, cast.ToUint64(recordId))
	k.Logger(ctx).Info(fmt.Sprintf("[DELEGATION] success on %s", chainId))
	return nil
}
