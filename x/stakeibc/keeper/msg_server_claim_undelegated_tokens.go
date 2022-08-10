package keeper

import (
	"context"
	"fmt"

	"github.com/spf13/cast"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	epochstypes "github.com/Stride-Labs/stride/x/epochs/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

type IcaTx struct {
	ConnectionId string
	Msgs         []sdk.Msg
	Account      types.ICAAccount
	Timeout      uint64
}

func (k msgServer) ClaimUndelegatedTokens(goCtx context.Context, msg *types.MsgClaimUndelegatedTokens) (*types.MsgClaimUndelegatedTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	userRedemptionRecord, err := k.GetClaimableRedemptionRecord(ctx, msg)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "unable to find claimable redemption record")
	}

	icaTx, err := k.GetRedemptionTransferMsg(ctx, userRedemptionRecord, msg.HostZoneId)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "unable to build redemption transfer message")
	}

	// add callback data
	redemptionCallback := types.RedemptionCallback{
		UserRedemptionRecordId: userRedemptionRecord.Id,
	}
	marshalledCallbackArgs := k.MarshalRedemptionCallbackArgs(ctx, redemptionCallback)
	_, err = k.SubmitTxs(ctx, icaTx.ConnectionId, icaTx.Msgs, icaTx.Account, icaTx.Timeout, "redemption", marshalledCallbackArgs)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Submit tx error: %s", err.Error()))
		return nil, sdkerrors.Wrapf(err, "unable to submit ICA redemption tx")
	}

	// Set isClaimable to false, so that the record can't be claimed again
	userRedemptionRecord.IsClaimable = false
	k.RecordsKeeper.SetUserRedemptionRecord(ctx, *userRedemptionRecord)

	return &types.MsgClaimUndelegatedTokensResponse{}, nil
}

func (k Keeper) GetClaimableRedemptionRecord(ctx sdk.Context, msg *types.MsgClaimUndelegatedTokens) (*recordstypes.UserRedemptionRecord, error) {
	// grab the UserRedemptionRecord from the store
	userRedemptionRecordKey := recordstypes.UserRedemptionRecordKeyFormatter(msg.HostZoneId, msg.Epoch, msg.Sender)
	userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, userRedemptionRecordKey)
	if !found {
		errMsg := fmt.Sprintf("User redemption record %s not found on host zone %s", userRedemptionRecordKey, msg.HostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrInvalidUserRedemptionRecord, "could not get user redemption record: %s", userRedemptionRecordKey)
	}

	// check that the record is claimable
	if !userRedemptionRecord.GetIsClaimable() {
		errMsg := fmt.Sprintf("User redemption record %s is not claimable", userRedemptionRecord.Id)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrInvalidUserRedemptionRecord, "user redemption record is not claimable: %s", userRedemptionRecordKey)
	}
	return &userRedemptionRecord, nil
}

func (k Keeper) GetRedemptionTransferMsg(ctx sdk.Context, userRedemptionRecord *recordstypes.UserRedemptionRecord, hostZoneId string) (*IcaTx, error) {
	// grab necessary fields to construct ICA call
	hostZone, found := k.GetHostZone(ctx, hostZoneId)
	if !found {
		errMsg := fmt.Sprintf("Host zone %s not found", hostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
	}
	redemptionAccount, found := k.GetRedemptionAccount(ctx, hostZone)
	if !found {
		errMsg := fmt.Sprintf("Redemption account not found for host zone %s", hostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
	}

	var msgs []sdk.Msg
	msgs = append(msgs, &bankTypes.MsgSend{
		FromAddress: redemptionAccount.Address,
		ToAddress:   userRedemptionRecord.Receiver,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(userRedemptionRecord.Denom, cast.ToInt64(userRedemptionRecord.Amount))),
	})

	// Give claims a 10 minute timeout
	epoch_tracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		errMsg := fmt.Sprintf("Epoch tracker not found for epoch %s", epochstypes.STRIDE_EPOCH)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrap(types.ErrEpochNotFound, errMsg)
	}
	timeout := cast.ToUint64(epoch_tracker.NextEpochStartTime) + cast.ToUint64(k.GetParam(ctx, types.KeyICATimeoutNanos))

	icaTx := IcaTx{
		ConnectionId: hostZone.GetConnectionId(),
		Msgs:         msgs,
		Account:      *redemptionAccount,
		Timeout:      timeout,
	}

	return &icaTx, nil
}
