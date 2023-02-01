package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v5/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
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
//
//	If successful:
//	   * Updates relevant validator delegations on the host zone struct
//	If timeout/failure:
//	   * Does nothing
func RebalanceCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	rebalanceCallback, err := k.UnmarshalRebalanceCallbackArgs(ctx, args)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal rebalance callback args: %s: %s", err.Error(), types.ErrUnmarshalFailure.Error())
	}
	chainId := rebalanceCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Rebalance, "Starting rebalance callback"))

	// Check for timeout (ack nil)
	// No action is necessary on a timeout
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Rebalance,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// No action is necessary on a failure
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Rebalance,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Rebalance,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Confirm the host zone exists
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return fmt.Errorf("host zone not found %s: %s", chainId, types.ErrInvalidRequest.Error())
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
			return fmt.Errorf("validator not found %s: %s", srcValidator, types.ErrInvalidRequest.Error())
		}

		// Increment the total delegation for the destination validator
		if _, valFound := valAddrMap[dstValidator]; valFound {
			valAddrMap[dstValidator].DelegationAmt = valAddrMap[dstValidator].DelegationAmt.Add(rebalancing.Amt)
		} else {
			return fmt.Errorf("validator not found %s: %s", dstValidator, types.ErrInvalidRequest.Error())
		}
	}
	k.SetHostZone(ctx, hostZone)

	return nil
}
