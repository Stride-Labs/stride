package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/utils"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var min = utils.Min

func (k msgServer) ClaimUndelegatedTokens(goCtx context.Context, msg *types.MsgClaimUndelegatedTokens) (*types.MsgClaimUndelegatedTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

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

	// grab necessary fields to construct ICA call
	hostZone, found := k.GetHostZone(ctx, msg.HostZoneId)
	if !found {
		errMsg := fmt.Sprintf("Host zone %s not found", msg.HostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
	}
	redemptionAccount, found := k.GetRedemptionAccount(ctx, hostZone)
	if !found {
		errMsg := fmt.Sprintf("Redemption account not found for host zone %s", msg.HostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
	}

	var msgs []sdk.Msg
	msgs = append(msgs, &bankTypes.MsgSend{
		FromAddress: redemptionAccount.Address,
		ToAddress:   userRedemptionRecord.Receiver,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(userRedemptionRecord.Denom, int64(userRedemptionRecord.Amount))),
	})
	
	sequence, err := k.SubmitTxs(ctx, hostZone.GetConnectionId(), msgs, *redemptionAccount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Submit tx error: %s", err.Error()))
		return nil, err
	}

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


	return &types.MsgClaimUndelegatedTokensResponse{}, nil
}
