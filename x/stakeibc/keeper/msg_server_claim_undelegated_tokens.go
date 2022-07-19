package keeper

import (
	"context"
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"

	epochstypes "github.com/Stride-Labs/stride/x/epochs/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (k msgServer) ClaimUndelegatedTokens(goCtx context.Context, msg *types.MsgClaimUndelegatedTokens) (*types.MsgClaimUndelegatedTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	userRedemptionRecord, err := k.GetClaimableRedemptionRecord(ctx, msg)
	if err != nil {
		return nil, err
	}

	redemptionAccount, connectionId, err := k.GetRedemptionAccountFromHostZoneId(ctx, msg.HostZoneId)
	if err != nil {
		return nil, err
	}

	timeout, err := k.GetIcaTimeout(ctx)
	if err != nil {
		return nil, err
	}

	msgs := k.GetRedemptionTransferMessage(*userRedemptionRecord, redemptionAccount.Address)
	sequence, err := k.SubmitTxs(ctx, connectionId, msgs, *redemptionAccount, timeout)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Submit tx error: %s", err.Error()))
		return nil, err
	}

	k.FlagRedemptionRecordsAsClaimed(ctx, *userRedemptionRecord, sequence)

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

func (k Keeper) GetRedemptionAccountFromHostZoneId(ctx sdk.Context, hostZoneId string) (*types.ICAAccount, string, error) {
	// grab necessary fields to construct ICA call
	hostZone, found := k.GetHostZone(ctx, hostZoneId)
	if !found {
		errMsg := fmt.Sprintf("Host zone %s not found", hostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, "", sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
	}

	redemptionAccount, found := k.GetRedemptionAccount(ctx, hostZone)
	if !found {
		errMsg := fmt.Sprintf("Redemption account not found for host zone %s", hostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, "", sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
	}
	return redemptionAccount, hostZone.GetConnectionId(), nil
}

func (k Keeper) GetRedemptionTransferMessage(userRedemptionRecord recordstypes.UserRedemptionRecord, redemptionAccountAddress string) []sdk.Msg {
	var msgs []sdk.Msg
	msgs = append(msgs, &bankTypes.MsgSend{
		FromAddress: redemptionAccountAddress,
		ToAddress:   userRedemptionRecord.Receiver,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(userRedemptionRecord.Denom, int64(userRedemptionRecord.Amount))),
	})
	return msgs
}

func (k Keeper) GetIcaTimeout(ctx sdk.Context) (uint64, error) {
	// Give claims a 10 minute timeout
	epoch_tracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		errMsg := fmt.Sprintf("Epoch tracker not found for epoch %s", epochstypes.STRIDE_EPOCH)
		k.Logger(ctx).Error(errMsg)
		return 0, sdkerrors.Wrap(types.ErrEpochNotFound, errMsg)
	}
	timeout := uint64(epoch_tracker.NextEpochStartTime) + uint64(k.GetParam(ctx, types.KeyICATimeoutNanos))
	return timeout, nil
}

func (k Keeper) FlagRedemptionRecordsAsClaimed(ctx sdk.Context, userRedemptionRecord recordstypes.UserRedemptionRecord, sequence uint64) {
	// Set isClaimable to false, so that the record can't be claimed again
	userRedemptionRecord.IsClaimable = false
	k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)

	// Store the sequence number to record id mapping
	pendingClaims := types.PendingClaims{
		Sequence: fmt.Sprint(sequence),
		// NOTE: we could extend this to process multiple claims in the future, given this field is repeated
		UserRedemptionRecordIds: []string{userRedemptionRecord.Id},
	}
	k.SetPendingClaims(ctx, pendingClaims)
}
