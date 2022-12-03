package keeper

import (
	"fmt"

	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/x/icacallbacks"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

func (k Keeper) MarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback types.DelegateCallback) ([]byte, error) {
	out, err := proto.Marshal(&delegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalDelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

func (k Keeper) UnmarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback []byte) (*types.DelegateCallback, error) {
	unmarshalledDelegateCallback := types.DelegateCallback{}
	if err := proto.Unmarshal(delegateCallback, &unmarshalledDelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalDelegateCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledDelegateCallback, nil
}

func DelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	k.Logger(ctx).Info("DelegateCallback executing", "packet", packet)
	// deserialize the args
	delegateCallback, err := k.UnmarshalDelegateCallbackArgs(ctx, args)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("DelegateCallback %v", delegateCallback))
	hostZone := delegateCallback.GetHostZoneId()
	zone, found := k.GetHostZone(ctx, hostZone)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "host zone not found %s", hostZone)
	}
	recordId := delegateCallback.GetDepositRecordId()
	depositRecord, found := k.RecordsKeeper.GetDepositRecord(ctx, recordId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "deposit record not found %d", recordId)
	}

	if ack == nil {
		// timeout
		k.Logger(ctx).Error(fmt.Sprintf("DelegateCallback timeout, ack is nil, packet %v", packet))
		return nil
	}

	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to fetch txMsgData, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}

	if len(txMsgData.Data) == 0 {
		// failed transaction
		depositRecord.Status = recordstypes.DepositRecord_DELEGATION_QUEUE
		k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
		k.Logger(ctx).Error(fmt.Sprintf("DelegateCallback tx failed, ack is empty (ack error), packet %v", packet))
		return nil
	}

	for _, splitDelegation := range delegateCallback.SplitDelegations {
		amount, err := cast.ToInt64E(splitDelegation.Amount)
		if err != nil {
			return err
		}
		validator := splitDelegation.Validator
		k.Logger(ctx).Info(fmt.Sprintf("incrementing stakedBal %d on %s", amount, validator))

		zone.StakedBal += splitDelegation.Amount
		success := k.AddDelegationToValidator(ctx, zone, validator, amount)
		if !success {
			return sdkerrors.Wrapf(types.ErrValidatorDelegationChg, "Failed to add delegation to validator")
		}
		k.SetHostZone(ctx, zone)
	}

	k.RecordsKeeper.RemoveDepositRecord(ctx, cast.ToUint64(recordId))
	k.Logger(ctx).Info(fmt.Sprintf("[DELEGATION] success on %s", hostZone))
	return nil
}
