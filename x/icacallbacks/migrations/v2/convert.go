package v2

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	icacallbacktypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	oldstakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/migrations/v2/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

const (
	ICACallbackID_Delegate   = "delegate"
	ICACallbackID_Rebalance  = "rebalance"
	ICACallbackID_Undelegate = "undelegate"
)

func convertDelegateCallback(oldDelegateCallback oldstakeibctypes.DelegateCallback) stakeibctypes.DelegateCallback {
	newSplitDelegations := []*stakeibctypes.SplitDelegation{}
	for _, oldSplitDelegation := range oldDelegateCallback.SplitDelegations {
		newSplitDelegation := stakeibctypes.SplitDelegation{
			Validator: oldSplitDelegation.Validator,
			Amount:    sdkmath.NewIntFromUint64(oldSplitDelegation.Amount),
		}
		newSplitDelegations = append(newSplitDelegations, &newSplitDelegation)
	}

	return stakeibctypes.DelegateCallback{
		HostZoneId:       oldDelegateCallback.HostZoneId,
		DepositRecordId:  oldDelegateCallback.DepositRecordId,
		SplitDelegations: newSplitDelegations,
	}
}

func convertUndelegateCallback(oldUndelegateCallback oldstakeibctypes.UndelegateCallback) stakeibctypes.UndelegateCallback {
	newSplitDelegations := []*stakeibctypes.SplitDelegation{}
	for _, oldSplitDelegation := range oldUndelegateCallback.SplitDelegations {
		newSplitDelegation := stakeibctypes.SplitDelegation{
			Validator: oldSplitDelegation.Validator,
			Amount:    sdkmath.NewIntFromUint64(oldSplitDelegation.Amount),
		}
		newSplitDelegations = append(newSplitDelegations, &newSplitDelegation)
	}

	return stakeibctypes.UndelegateCallback{
		HostZoneId:              oldUndelegateCallback.HostZoneId,
		SplitDelegations:        newSplitDelegations,
		EpochUnbondingRecordIds: oldUndelegateCallback.EpochUnbondingRecordIds,
	}
}

func convertRebalanceCallback(oldRebalanceCallback oldstakeibctypes.RebalanceCallback) stakeibctypes.RebalanceCallback {
	newRebalancings := []*stakeibctypes.Rebalancing{}
	for _, oldRebalancing := range oldRebalanceCallback.Rebalancings {
		newRebalancing := stakeibctypes.Rebalancing{
			SrcValidator: oldRebalancing.SrcValidator,
			DstValidator: oldRebalancing.DstValidator,
			Amt:          sdkmath.NewIntFromUint64(oldRebalancing.Amt),
		}
		newRebalancings = append(newRebalancings, &newRebalancing)
	}

	return stakeibctypes.RebalanceCallback{
		HostZoneId:   oldRebalanceCallback.HostZoneId,
		Rebalancings: newRebalancings,
	}
}

func convertCallbackData(oldCallbackData icacallbacktypes.CallbackData) (icacallbacktypes.CallbackData, error) {
	var newCallbackArgs []byte
	switch oldCallbackData.CallbackId {
	case ICACallbackID_Delegate:
		// Deserialize the callback args with the old DelegateCallback type
		oldDelegateCallback := oldstakeibctypes.DelegateCallback{}
		if err := proto.Unmarshal(oldCallbackData.CallbackArgs, &oldDelegateCallback); err != nil {
			return icacallbacktypes.CallbackData{}, errorsmod.Wrapf(stakeibctypes.ErrUnmarshalFailure, err.Error())
		}

		// Convert and serialize with the new DelegateCallback type
		newDelegateCallback := convertDelegateCallback(oldDelegateCallback)
		newDelegateCallbackBz, err := proto.Marshal(&newDelegateCallback)
		if err != nil {
			return icacallbacktypes.CallbackData{}, errorsmod.Wrapf(stakeibctypes.ErrMarshalFailure, err.Error())
		}

		// Update the CallbackData with the new args
		newCallbackArgs = newDelegateCallbackBz

	case ICACallbackID_Undelegate:
		// Deserialize the callback args with the old UndelegateCallback type
		oldUndelegateCallback := oldstakeibctypes.UndelegateCallback{}
		if err := proto.Unmarshal(oldCallbackData.CallbackArgs, &oldUndelegateCallback); err != nil {
			return icacallbacktypes.CallbackData{}, errorsmod.Wrapf(stakeibctypes.ErrUnmarshalFailure, err.Error())
		}

		// Convert and serialize with the new UndelegateCallback type
		newUndelegateCallback := convertUndelegateCallback(oldUndelegateCallback)
		newUndelegateCallbackBz, err := proto.Marshal(&newUndelegateCallback)
		if err != nil {
			return icacallbacktypes.CallbackData{}, errorsmod.Wrapf(stakeibctypes.ErrMarshalFailure, err.Error())
		}

		// Update the CallbackData with the new args
		newCallbackArgs = newUndelegateCallbackBz

	case ICACallbackID_Rebalance:
		// Deserialize the callback args with the old RebalanceCallback type
		oldRebalanceCallback := oldstakeibctypes.RebalanceCallback{}
		if err := proto.Unmarshal(oldCallbackData.CallbackArgs, &oldRebalanceCallback); err != nil {
			return icacallbacktypes.CallbackData{}, errorsmod.Wrapf(stakeibctypes.ErrUnmarshalFailure, err.Error())
		}

		// Convert and serialize with the new RebalanceCallback type
		newRebalanceCallback := convertRebalanceCallback(oldRebalanceCallback)
		newRebalanceCallbackBz, err := proto.Marshal(&newRebalanceCallback)
		if err != nil {
			return icacallbacktypes.CallbackData{}, errorsmod.Wrapf(stakeibctypes.ErrMarshalFailure, err.Error())
		}

		// Update the CallbackData with the new args
		newCallbackArgs = newRebalanceCallbackBz

	default:
		newCallbackArgs = oldCallbackData.CallbackArgs
	}

	newCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  oldCallbackData.CallbackKey,
		PortId:       oldCallbackData.PortId,
		ChannelId:    oldCallbackData.ChannelId,
		Sequence:     oldCallbackData.Sequence,
		CallbackId:   oldCallbackData.CallbackId,
		CallbackArgs: newCallbackArgs,
	}

	return newCallbackData, nil
}
