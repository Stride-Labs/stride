package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v4/utils"
	"github.com/Stride-Labs/stride/v4/x/icacallbacks"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Marshalls rebalance callback arguments
func (k Keeper) MarshalRebalanceCallbackArgs(ctx sdk.Context, rebalanceCallback types.RebalanceCallback) ([]byte, error) {
	out, err := proto.Marshal(&rebalanceCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalRebalanceCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// Unmarshalls rebalance callback arguments into a RebalanceCallback struct
func (k Keeper) UnmarshalRebalanceCallbackArgs(ctx sdk.Context, rebalanceCallback []byte) (*types.RebalanceCallback, error) {
	unmarshalledRebalanceCallback := types.RebalanceCallback{}
	if err := proto.Unmarshal(rebalanceCallback, &unmarshalledRebalanceCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalRebalanceCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledRebalanceCallback, nil
}

// ICA Callback after rebalance validators on a host zone
//   If successful:
//      * Updates relevant validator delegations on the host zone struct
//   If timeout/failure:
//      * Does nothing
func RebalanceCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	// Fetch callback args
	rebalanceCallback, err := k.UnmarshalRebalanceCallbackArgs(ctx, args)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal rebalance callback args: %s", err.Error()))
	}
	chainId := rebalanceCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Rebalance, "Starting rebalance callback"))

	// Check for timeout (ack nil)
	// No action is necessary on a timeout
	if ack == nil {
		k.Logger(ctx).Error(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Rebalance,
			"TIMEOUT (ack is nil), Packet: %+v", packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// No action is necessary on a failure
	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("RebalanceCallback failed to fetch txMsgData, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}
	if len(txMsgData.Data) == 0 {
		k.Logger(ctx).Error(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Rebalance,
			"ICA TX FAILED (ack is empty / ack error), Packet: %+v", packet))
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Rebalance, "SUCCESS, Packet: %+v", packet))

	// Confirm the host zone exists
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "host zone not found %s", chainId)
	}

	// Assemble a map from validatorAddress -> validator
	valAddrMap := make(map[string]*types.Validator)
	for _, val := range hostZone.Validators {
		valAddrMap[val.Address] = val
	}

	// For each re-delegation transaction, update the relevant validators on the host zone
	for _, rebalancing := range rebalanceCallback.Rebalancings {
		srcValidator := rebalancing.SrcValidator
		dstValidator := rebalancing.DstValidator

		// Decrement the total delegation from the source validator
		if _, valFound := valAddrMap[srcValidator]; valFound {
			valAddrMap[srcValidator].DelegationAmt = valAddrMap[srcValidator].DelegationAmt.Sub(rebalancing.Amt)
		} else {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator not found %s", srcValidator)
		}

		// Increment the total delegation for the destination validator
		if _, valFound := valAddrMap[dstValidator]; valFound {
			valAddrMap[dstValidator].DelegationAmt = valAddrMap[dstValidator].DelegationAmt.Add(rebalancing.Amt)
		} else {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator not found %s", dstValidator)
		}
	}
	k.SetHostZone(ctx, hostZone)

	return nil
}
