package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v9/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Marshal claim callback args
func (k Keeper) MarshalClaimCallbackArgs(ctx sdk.Context, claimCallback types.ClaimCallback) ([]byte, error) {
	out, err := proto.Marshal(&claimCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalClaimCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// Unmarshalls claim callback arguments into a ClaimCallback struct
func (k Keeper) UnmarshalClaimCallbackArgs(ctx sdk.Context, claimCallback []byte) (*types.ClaimCallback, error) {
	unmarshalledDelegateCallback := types.ClaimCallback{}
	if err := proto.Unmarshal(claimCallback, &unmarshalledDelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalClaimCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledDelegateCallback, nil
}

// ICA Callback after claiming unbonded tokens
//   If successful:
//     * Removes the user redemption record
//   If timeout/failure:
//     * Reverts pending flag in the user redemption record so the claim can be re-tried
func ClaimCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	claimCallback, err := k.UnmarshalClaimCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal claim callback args: %s", err.Error()))
	}
	chainId := claimCallback.ChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Claim,
		"Starting claim callback for Redemption Record: %s", claimCallback.UserRedemptionRecordId))

	// Grab the associated user redemption record
	userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, claimCallback.GetUserRedemptionRecordId())
	if !found {
		return errorsmod.Wrapf(types.ErrRecordNotFound, "user redemption record not found %s", claimCallback.GetUserRedemptionRecordId())
	}

	// Check for a timeout
	// If the ICA timed out, update the redemption record so the user can retry the claim
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Claim,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))

		userRedemptionRecord.ClaimIsPending = false
		k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
		return nil
	}

	// Check for a failed transaction (ack error)
	// Upon failure, update the redemption record to allow the user to retry the claim
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Claim,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))

		// after an error, a user should be able to retry the claim
		userRedemptionRecord.ClaimIsPending = false
		k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Claim,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Upon success, remove the record and decrement the unbonded amount on the host zone unbonding record
	k.RecordsKeeper.RemoveUserRedemptionRecord(ctx, claimCallback.GetUserRedemptionRecordId())
	err = k.DecrementHostZoneUnbonding(ctx, userRedemptionRecord, *claimCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("ClaimCallback failed (DecrementHostZoneUnbonding), packet %v, err: %s", packet, err.Error()))
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] success on %s", userRedemptionRecord.GetHostZoneId()))
	return nil
}

// After a user claims their unbonded tokens, the claim amount is decremented from the corresponding host zone unbonding record
func (k Keeper) DecrementHostZoneUnbonding(ctx sdk.Context, userRedemptionRecord recordstypes.UserRedemptionRecord, callbackArgs types.ClaimCallback) error {
	// fetch the hzu associated with the user unbonding record
	hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, callbackArgs.EpochNumber, callbackArgs.ChainId)
	if !found {
		return errorsmod.Wrapf(types.ErrRecordNotFound, "host zone unbonding not found %s", callbackArgs.ChainId)
	}

	// decrement the hzu by the amount claimed
	hostZoneUnbonding.NativeTokenAmount = hostZoneUnbonding.NativeTokenAmount.Sub(userRedemptionRecord.Amount)

	// save the updated hzu on the epoch unbonding record
	epochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, callbackArgs.EpochNumber, callbackArgs.ChainId, hostZoneUnbonding)
	if !success {
		return errorsmod.Wrapf(types.ErrRecordNotFound, "epoch unbonding record not found %s", callbackArgs.ChainId)
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *epochUnbondingRecord)
	return nil
}
