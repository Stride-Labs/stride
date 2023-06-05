package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v9/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
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
//     * Updates relevant validator delegations on the host zone struct
//   If timeout/failure:
//     * Does nothing
func RebalanceCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	rebalanceCallback, err := k.UnmarshalRebalanceCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal rebalance callback args: %s", err.Error()))
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
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "host zone not found %s", chainId)
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
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator not found %s", srcValidator)
		}

		// Increment the total delegation for the destination validator
		if _, valFound := valAddrMap[dstValidator]; valFound {
			valAddrMap[dstValidator].DelegationAmt = valAddrMap[dstValidator].DelegationAmt.Add(rebalancing.Amt)
		} else {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator not found %s", dstValidator)
		}
	}
	k.SetHostZone(ctx, hostZone)

	return nil
}
