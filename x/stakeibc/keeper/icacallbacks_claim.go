package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v4/x/icacallbacks"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

func (k Keeper) MarshalClaimCallbackArgs(ctx sdk.Context, claimCallback types.ClaimCallback) ([]byte, error) {
	out, err := proto.Marshal(&claimCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalClaimCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

func (k Keeper) UnmarshalClaimCallbackArgs(ctx sdk.Context, claimCallback []byte) (*types.ClaimCallback, error) {
	unmarshalledDelegateCallback := types.ClaimCallback{}
	if err := proto.Unmarshal(claimCallback, &unmarshalledDelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalClaimCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledDelegateCallback, nil
}

func ClaimCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	// deserialize the args
	claimCallback, err := k.UnmarshalClaimCallbackArgs(ctx, args)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("ClaimCallback %v", claimCallback))
	userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, claimCallback.GetUserRedemptionRecordId())
	if !found {
		return sdkerrors.Wrapf(types.ErrRecordNotFound, "user redemption record not found %s", claimCallback.GetUserRedemptionRecordId())
	}

	// handle timeout
	if ack == nil {
		k.Logger(ctx).Error(fmt.Sprintf("ClaimCallback timeout, ack is nil, packet %v", packet))
		// after a timeout, a user should be able to retry the claim
		userRedemptionRecord.ClaimIsPending = false
		k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
		return nil
	}

	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to unmarshal txMsgData, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}

	k.Logger(ctx).Info("ClaimCallback executing", "packet", packet, "txMsgData", txMsgData, "args", args)
	// handle failed tx on host chain
	if len(txMsgData.Data) == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("ClaimCallback failed, packet %v", packet))
		// after an error, a user should be able to retry the claim
		userRedemptionRecord.ClaimIsPending = false
		k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
		return nil
	}

	// claim successfully processed
	// remove the record and decrement the hzu
	k.RecordsKeeper.RemoveUserRedemptionRecord(ctx, claimCallback.GetUserRedemptionRecordId())
	err = k.DecrementHostZoneUnbonding(ctx, userRedemptionRecord, *claimCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("ClaimCallback failed (DecrementHostZoneUnbonding), packet %v, err: %s", packet, err.Error()))
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] success on %s", userRedemptionRecord.GetHostZoneId()))
	return nil
}

func (k Keeper) DecrementHostZoneUnbonding(ctx sdk.Context, userRedemptionRecord recordstypes.UserRedemptionRecord, callbackArgs types.ClaimCallback) error {
	// fetch the hzu associated with the user unbonding record
	hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, callbackArgs.EpochNumber, callbackArgs.ChainId)
	if !found {
		return sdkerrors.Wrapf(types.ErrRecordNotFound, "host zone unbonding not found %s", callbackArgs.ChainId)
	}
	// decrement the hzu by the amount claimed
	hostZoneUnbonding.NativeTokenAmount = hostZoneUnbonding.NativeTokenAmount.Sub(userRedemptionRecord.Amount)
	// save the updated hzu on the epoch unbonding record
	epochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, callbackArgs.EpochNumber, callbackArgs.ChainId, hostZoneUnbonding)
	if !success {
		return sdkerrors.Wrapf(types.ErrRecordNotFound, "epoch unbonding record not found %s", callbackArgs.ChainId)
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *epochUnbondingRecord)
	return nil
}
