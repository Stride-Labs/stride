package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v27/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v27/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v27/x/records/types"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
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
		return unmarshalledRedemptionCallback, errorsmod.Wrap(err, "unable to unmarshal redemption callback args")
	}
	return unmarshalledRedemptionCallback, nil
}

// ICA Callback after undelegating
// * If successful: Updates epoch unbonding record status
// * If timeout:    Does nothing
// * If failure:    Reverts epoch unbonding record status
func (k Keeper) RedemptionCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	redemptionCallback, err := k.UnmarshalRedemptionCallbackArgs(ctx, args)
	if err != nil {
		return err
	}
	chainId := redemptionCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Redemption,
		"Starting redemption callback for Epoch Unbonding Records: %+v", redemptionCallback.EpochUnbondingRecordIds))

	// Check for timeout (ack nil)
	// No need to reset the unbonding record status since it will get reverted when the channel is restored
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Redemption,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// Reset the unbonding record status upon failure
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Redemption,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))

		// Reset unbondings record status
		err = k.RecordsKeeper.SetHostZoneUnbondingStatus(ctx, chainId, redemptionCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
		if err != nil {
			return err
		}
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Redemption,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Confirm host zone exists
	_, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "Host zone not found: %s", chainId)
	}

	// Upon success, update the unbonding record status to CLAIMABLE and set the number of
	// claimable tokens for each epoch unbonding record
	for _, epochNumber := range redemptionCallback.EpochUnbondingRecordIds {
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochNumber, chainId)
		if !found {
			return recordstypes.ErrHostUnbondingRecordNotFound.Wrapf("unbonding record not found for epoch %d and chain %s",
				epochNumber, chainId)
		}

		hostZoneUnbonding.ClaimableNativeTokens = hostZoneUnbonding.NativeTokenAmount
		hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_CLAIMABLE
		if err := k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochNumber, chainId, *hostZoneUnbonding); err != nil {
			return err
		}
	}

	k.Logger(ctx).Info(fmt.Sprintf("[REDEMPTION] completed on %s", chainId))
	return nil
}
