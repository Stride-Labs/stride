package v2

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"github.com/stretchr/testify/require"

	icacallbacktypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"
	oldstakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/migrations/v2/types"
)

func TestConvertDelegateCallback(t *testing.T) {
	hostZoneId := "hz"
	depositRecordId := uint64(1)
	val1 := "val1"
	val2 := "val2"

	// Define old callback type and convert to new type
	oldDelegateCallback := oldstakeibctypes.DelegateCallback{
		HostZoneId:      hostZoneId,
		DepositRecordId: depositRecordId,
		SplitDelegations: []*oldstakeibctypes.SplitDelegation{
			{
				Validator: val1,
				Amount:    uint64(1),
			},
			{
				Validator: val2,
				Amount:    uint64(2),
			},
		},
	}
	newDelegateCallback := convertDelegateCallback(oldDelegateCallback)

	// Check unchanged fields
	require.Equal(t, hostZoneId, newDelegateCallback.HostZoneId, "host zone id")
	require.Equal(t, depositRecordId, newDelegateCallback.DepositRecordId, "deposit record id")
	require.Equal(t, val1, newDelegateCallback.SplitDelegations[0].Validator, "validator 1 address")
	require.Equal(t, val2, newDelegateCallback.SplitDelegations[1].Validator, "validator 2 address")

	// Check update fields
	require.Equal(t, sdkmath.NewInt(1), newDelegateCallback.SplitDelegations[0].Amount, "validator 1 amount")
	require.Equal(t, sdkmath.NewInt(2), newDelegateCallback.SplitDelegations[1].Amount, "validator 2 amount")
}

func TestConvertUndelegateCallback(t *testing.T) {
	hostZoneId := "hz"
	epochUnbondingIds := []uint64{1}
	val1 := "val1"
	val2 := "val2"

	// Define old callback type and convert to new type
	oldUndelegateCallback := oldstakeibctypes.UndelegateCallback{
		HostZoneId: hostZoneId,
		SplitDelegations: []*oldstakeibctypes.SplitDelegation{
			{
				Validator: val1,
				Amount:    uint64(1),
			},
			{
				Validator: val2,
				Amount:    uint64(2),
			},
		},
		EpochUnbondingRecordIds: epochUnbondingIds,
	}
	newUndelegateCallback := convertUndelegateCallback(oldUndelegateCallback)

	// Check unchanged fields
	require.Equal(t, hostZoneId, newUndelegateCallback.HostZoneId, "host zone id")
	require.Equal(t, epochUnbondingIds[0], newUndelegateCallback.EpochUnbondingRecordIds[0], "epoch unbonding record id")
	require.Equal(t, val1, newUndelegateCallback.SplitDelegations[0].Validator, "validator 1 address")
	require.Equal(t, val2, newUndelegateCallback.SplitDelegations[1].Validator, "validator 2 address")

	// Check update fields
	require.Equal(t, sdkmath.NewInt(1), newUndelegateCallback.SplitDelegations[0].Amount, "validator 1 amount")
	require.Equal(t, sdkmath.NewInt(2), newUndelegateCallback.SplitDelegations[1].Amount, "validator 2 amount")
}

func TestConvertRebalanceCallback(t *testing.T) {
	hostZoneId := "hz1"
	srcVal1 := "src_val1"
	srcVal2 := "src_val2"
	dstVal1 := "dst_val1"
	dstVal2 := "dst_val2"

	// Define old callback type and convert to new type
	oldRebalanceCallback := oldstakeibctypes.RebalanceCallback{
		HostZoneId: hostZoneId,
		Rebalancings: []*oldstakeibctypes.Rebalancing{
			{
				SrcValidator: srcVal1,
				DstValidator: dstVal1,
				Amt:          uint64(1),
			},
			{
				SrcValidator: srcVal2,
				DstValidator: dstVal2,
				Amt:          uint64(2),
			},
		},
	}
	newRebalanceCallback := convertRebalanceCallback(oldRebalanceCallback)

	// Check unchanged fields
	require.Equal(t, hostZoneId, newRebalanceCallback.HostZoneId, "host zone id")

	require.Equal(t, srcVal1, newRebalanceCallback.Rebalancings[0].SrcValidator, "srcValidator 1 address")
	require.Equal(t, dstVal1, newRebalanceCallback.Rebalancings[0].DstValidator, "dstValidator 1 address")

	require.Equal(t, srcVal2, newRebalanceCallback.Rebalancings[1].SrcValidator, "srcValidator 2 address")
	require.Equal(t, dstVal2, newRebalanceCallback.Rebalancings[1].DstValidator, "dstValidator 2 address")

	// Check updated fields
	require.Equal(t, sdkmath.NewInt(1), newRebalanceCallback.Rebalancings[0].Amt)
	require.Equal(t, sdkmath.NewInt(2), newRebalanceCallback.Rebalancings[1].Amt)
}

func TestConvertCallbackData_Delegate_Success(t *testing.T) {
	// Marshal delegate callback for callback args in CallbackData struct
	delegateCallbackBz, err := proto.Marshal(&oldstakeibctypes.DelegateCallback{
		SplitDelegations: []*oldstakeibctypes.SplitDelegation{{Amount: uint64(1)}},
	})
	require.NoError(t, err, "no error expected when marshalling delegate callback")

	// Define old delegate callback data type and convert to new type
	oldCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  "key",
		PortId:       "port",
		ChannelId:    "channel",
		Sequence:     uint64(1),
		CallbackId:   ICACallbackID_Delegate,
		CallbackArgs: delegateCallbackBz,
	}
	newCallbackData, err := convertCallbackData(oldCallbackData)
	require.NoError(t, err)

	// If the callback was was type delegate, the callback args SHOULD have changed
	require.NotEqual(t, oldCallbackData.CallbackArgs, newCallbackData.CallbackArgs, "callback args should have changed")

	// All other fields SHOULD NOT have changed
	require.Equal(t, oldCallbackData.CallbackKey, newCallbackData.CallbackKey, "callback key")
	require.Equal(t, oldCallbackData.PortId, newCallbackData.PortId, "port ID")
	require.Equal(t, oldCallbackData.ChannelId, newCallbackData.ChannelId, "channel ID")
	require.Equal(t, oldCallbackData.Sequence, newCallbackData.Sequence, "sequence")
	require.Equal(t, oldCallbackData.CallbackId, newCallbackData.CallbackId, "callback ID")
}

func TestConvertCallbackData_Delegate_Error(t *testing.T) {
	// Define old delegate callback data type with invalid callback args
	oldCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  "key",
		PortId:       "port",
		ChannelId:    "channel",
		Sequence:     uint64(1),
		CallbackId:   ICACallbackID_Delegate,
		CallbackArgs: []byte{1, 2, 3},
	}

	// The convert function should fail since it cannot unmarshal the callback args into a DelegateCallback
	_, err := convertCallbackData(oldCallbackData)
	require.ErrorContains(t, err, "unable to unmarshal data structure")
}

func TestConvertCallbackData_Rebalance_Success(t *testing.T) {
	// Marshal rebalance callback for callback args in CallbackData struct
	rebalanceCallbackBz, err := proto.Marshal(&oldstakeibctypes.RebalanceCallback{
		Rebalancings: []*oldstakeibctypes.Rebalancing{{Amt: uint64(1)}},
	})
	require.NoError(t, err, "no error expected when marshalling rebalance callback")

	// Define old rebalance callback data type and convert to new type
	oldCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  "key",
		PortId:       "port",
		ChannelId:    "channel",
		Sequence:     uint64(1),
		CallbackId:   ICACallbackID_Rebalance,
		CallbackArgs: rebalanceCallbackBz,
	}
	newCallbackData, err := convertCallbackData(oldCallbackData)
	require.NoError(t, err)

	// If the callback was was type delegate, the callback args SHOULD have changed
	require.NotEqual(t, oldCallbackData.CallbackArgs, newCallbackData.CallbackArgs, "callback args should have changed")

	// All other fields SHOULD NOT have changed
	require.Equal(t, oldCallbackData.CallbackKey, newCallbackData.CallbackKey, "callback key")
	require.Equal(t, oldCallbackData.PortId, newCallbackData.PortId, "port ID")
	require.Equal(t, oldCallbackData.ChannelId, newCallbackData.ChannelId, "channel ID")
	require.Equal(t, oldCallbackData.Sequence, newCallbackData.Sequence, "sequence")
	require.Equal(t, oldCallbackData.CallbackId, newCallbackData.CallbackId, "callback ID")
}

func TestConvertCallbackData_Rebalance_Error(t *testing.T) {
	// Define old rebalance callback data type and convert to new type
	oldCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  "key",
		PortId:       "port",
		ChannelId:    "channel",
		Sequence:     uint64(1),
		CallbackId:   ICACallbackID_Rebalance,
		CallbackArgs: []byte{1, 2, 3},
	}

	// The convert function should fail since it cannot unmarshal the callback args into a RebalanceCallback
	_, err := convertCallbackData(oldCallbackData)
	require.ErrorContains(t, err, "unable to unmarshal data structure")
}

func TestConvertCallbackData_Other(t *testing.T) {
	oldCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  "key",
		PortId:       "port",
		ChannelId:    "channel",
		Sequence:     uint64(1),
		CallbackId:   "different_callback",
		CallbackArgs: []byte{1, 2, 3},
	}

	// If the callback is not delegate or rebalance, it should not have been changed
	newCallbackData, err := convertCallbackData(oldCallbackData)
	require.NoError(t, err)
	require.Equal(t, oldCallbackData, newCallbackData)
}
