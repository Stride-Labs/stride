package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v27/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v27/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
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
		return nil, errorsmod.Wrap(err, "unable to unmarshal rebalance callback args")
	}
	return &unmarshalledRebalanceCallback, nil
}

// ICA Callback after rebalance validators on a host zone
// * If successful:      Updates relevant validator delegations on the host zone struct
// * If timeout/failure: Does nothing
func (k Keeper) RebalanceCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	rebalanceCallback, err := k.UnmarshalRebalanceCallbackArgs(ctx, args)
	if err != nil {
		return err
	}
	chainId := rebalanceCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Rebalance, "Starting rebalance callback"))

	// Regardless of failure/success/timeout, indicate that this ICA has completed
	hostZone, found := k.GetHostZone(ctx, rebalanceCallback.HostZoneId)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "Host zone not found: %s", rebalanceCallback.HostZoneId)
	}
	for _, rebalancing := range rebalanceCallback.Rebalancings {
		if err := k.DecrementValidatorDelegationChangesInProgress(&hostZone, rebalancing.SrcValidator); err != nil {
			return err
		}
		if err := k.DecrementValidatorDelegationChangesInProgress(&hostZone, rebalancing.DstValidator); err != nil {
			return err
		}
	}
	k.SetHostZone(ctx, hostZone)

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

	// Assemble a map from validatorAddress -> validator
	valAddrMap := make(map[string]*types.Validator)
	for _, val := range hostZone.Validators {
		valAddrMap[val.Address] = val
	}

	// For each re-delegation transaction, update the relevant validators on the host zone
	for _, rebalancing := range rebalanceCallback.Rebalancings {
		srcValidator := rebalancing.SrcValidator
		dstValidator := rebalancing.DstValidator

		if _, valFound := valAddrMap[srcValidator]; !valFound {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "source validator not found %s", srcValidator)
		}
		if _, valFound := valAddrMap[dstValidator]; !valFound {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "destination validator not found %s", dstValidator)
		}

		// Decrement the delegation from the source validator and increment the delegation
		// for the destination validator
		valAddrMap[srcValidator].Delegation = valAddrMap[srcValidator].Delegation.Sub(rebalancing.Amt)
		valAddrMap[dstValidator].Delegation = valAddrMap[dstValidator].Delegation.Add(rebalancing.Amt)

		k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Rebalance,
			"  Decrementing delegation on %s by %v", srcValidator, rebalancing.Amt))
		k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Rebalance,
			"  Incrementing delegation on %s by %v", dstValidator, rebalancing.Amt))
	}

	k.SetHostZone(ctx, hostZone)

	return nil
}
