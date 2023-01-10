package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v4/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Marshalls redemption callback arguments
func (k Keeper) MarshalRedemptionCallbackArgs(ctx sdk.Context, redemptionCallback types.RedemptionCallback) ([]byte, error) {
	out, err := proto.Marshal(&redemptionCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalRedemptionCallbackArgs | %s", err.Error()))
		return nil, err
	}
	return out, nil
}

// Unmarshalls redemption callback arguments into a RedemptionCallback struct
func (k Keeper) UnmarshalRedemptionCallbackArgs(ctx sdk.Context, redemptionCallback []byte) (types.RedemptionCallback, error) {
	unmarshalledRedemptionCallback := types.RedemptionCallback{}
	if err := proto.Unmarshal(redemptionCallback, &unmarshalledRedemptionCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalRedemptionCallbackArgs | %s", err.Error()))
		return unmarshalledRedemptionCallback, err
	}
	return unmarshalledRedemptionCallback, nil
}

// ICA Callback after undelegating
//   If successful:
//      * Updates epoch unbonding record status
//   If timeout:
//      * Does nothing
//   If failure:
//		* Reverts epoch unbonding record status
func RedemptionCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	redemptionCallback, err := k.UnmarshalRedemptionCallbackArgs(ctx, args)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshal redemption callback args | %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, errMsg)
	}
	chainId := redemptionCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Redemption,
		"Starting callback for Epoch Unbonding Records: %+v", redemptionCallback.EpochUnbondingRecordIds))

	// No need to reset the unbonding record status since it will get revertted when the channel is restored
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Redemption,
			"TIMEOUT (ack is nil), Packet: %+v", packet))
		return nil
	}

	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Redemption,
			"ICA TX FAILED (ack is empty / ack error), Packet: %+v, Error: ", packet, ackResponse.Error))

		// Reset unbondings record status
		err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, chainId, redemptionCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
		if err != nil {
			return err
		}
		return nil
	}

	k.Logger(ctx).Info(utils.LogCallbackWithHostZone(chainId, ICACallbackID_Redemption, "SUCCESS, Packet: %+v", packet))

	// Confirm host zone exists
	_, found := k.GetHostZone(ctx, chainId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "Host zone not found: %s", chainId)
	}

	// Upon success, update the unbonding record status to CLAIMABLE
	err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, chainId, redemptionCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_CLAIMABLE)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("[REDEMPTION] completed on %s", chainId))
	return nil
}
