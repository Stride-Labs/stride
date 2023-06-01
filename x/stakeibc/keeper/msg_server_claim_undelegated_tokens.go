package keeper

import (
	"context"
	"fmt"

	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	proto "github.com/cosmos/gogoproto/proto"

	epochstypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type IcaTx struct {
	ConnectionId string
	Msgs         []proto.Message
	Account      types.ICAAccount
	Timeout      uint64
}

func (k msgServer) ClaimUndelegatedTokens(goCtx context.Context, msg *types.MsgClaimUndelegatedTokens) (*types.MsgClaimUndelegatedTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("ClaimUndelegatedTokens %v", msg))
	userRedemptionRecord, err := k.GetClaimableRedemptionRecord(ctx, msg)
	if err != nil {
		errMsg := fmt.Sprintf("unable to find claimable redemption record for msg: %v, error %s", msg, err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrap(types.ErrRecordNotFound, errMsg)
	}

	hostZone, found := k.GetHostZone(ctx, msg.HostZoneId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrHostZoneNotFound, fmt.Sprintf("Host zone %s not found", msg.HostZoneId))
	}

	if hostZone.Halted {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone %s halted", msg.HostZoneId))
		return nil, errorsmod.Wrapf(types.ErrHaltedHostZone, "Host Zone %s halted", msg.HostZoneId)
	}

	icaTx, err := k.GetRedemptionTransferMsg(ctx, userRedemptionRecord, msg.HostZoneId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unable to build redemption transfer message")
	}

	// add callback data
	claimCallback := types.ClaimCallback{
		UserRedemptionRecordId: userRedemptionRecord.Id,
		ChainId:                msg.HostZoneId,
		EpochNumber:            msg.Epoch,
	}
	marshalledCallbackArgs, err := k.MarshalClaimCallbackArgs(ctx, claimCallback)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unable to marshal claim callback args")
	}
	_, err = k.SubmitTxs(ctx, icaTx.ConnectionId, icaTx.Msgs, icaTx.Account, icaTx.Timeout, ICACallbackID_Claim, marshalledCallbackArgs)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Submit tx error: %s", err.Error()))
		return nil, errorsmod.Wrap(err, "unable to submit ICA redemption tx")
	}

	// Set claimIsPending to true, so that the record can't be double claimed
	userRedemptionRecord.ClaimIsPending = true
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
		return nil, errorsmod.Wrap(types.ErrInvalidUserRedemptionRecord, errMsg)
	}

	// check that the record is claimable
	hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, userRedemptionRecord.EpochNumber, msg.HostZoneId)
	if !found {
		errMsg := fmt.Sprintf("Host zone unbonding record %s not found on host zone %s", userRedemptionRecordKey, msg.HostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrInvalidUserRedemptionRecord, errMsg)
	}
	// records associated with host zone unbondings are claimable after the host zone unbonding tokens have been CLAIMABLE to the redemption account
	if hostZoneUnbonding.Status != recordstypes.HostZoneUnbonding_CLAIMABLE {
		errMsg := fmt.Sprintf("User redemption record %s is not claimable, host zone unbonding has status: %s, requires status CLAIMABLE", userRedemptionRecord.Id, hostZoneUnbonding.Status)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrInvalidUserRedemptionRecord, errMsg)
	}
	// records that have claimIsPending set to True have already been claimed (and are pending an ack)
	if userRedemptionRecord.ClaimIsPending {
		errMsg := fmt.Sprintf("User redemption record %s is not claimable, pending ack", userRedemptionRecord.Id)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrInvalidUserRedemptionRecord, errMsg)
	}
	return &userRedemptionRecord, nil
}

func (k Keeper) GetRedemptionTransferMsg(ctx sdk.Context, userRedemptionRecord *recordstypes.UserRedemptionRecord, hostZoneId string) (*IcaTx, error) {
	// grab necessary fields to construct ICA call
	hostZone, found := k.GetHostZone(ctx, hostZoneId)
	if !found {
		errMsg := fmt.Sprintf("Host zone %s not found", hostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrap(types.ErrInvalidHostZone, errMsg)
	}
	redemptionAccount, found := k.GetRedemptionAccount(ctx, hostZone)
	if !found {
		errMsg := fmt.Sprintf("Redemption account not found for host zone %s", hostZoneId)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrap(types.ErrInvalidHostZone, errMsg)
	}

	var msgs []proto.Message
	rrAmt := userRedemptionRecord.Amount
	msgs = append(msgs, &bankTypes.MsgSend{
		FromAddress: redemptionAccount.Address,
		ToAddress:   userRedemptionRecord.Receiver,
		Amount:      sdk.NewCoins(sdk.NewCoin(userRedemptionRecord.Denom, rrAmt)),
	})

	// Give claims a 10 minute timeout
	epochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		errMsg := fmt.Sprintf("Epoch tracker not found for epoch %s", epochstypes.STRIDE_EPOCH)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrap(types.ErrEpochNotFound, errMsg)
	}
	icaTimeOutNanos := k.GetParam(ctx, types.KeyICATimeoutNanos)
	nextEpochStarttime := epochTracker.NextEpochStartTime
	timeout := nextEpochStarttime + icaTimeOutNanos

	icaTx := IcaTx{
		ConnectionId: hostZone.GetConnectionId(),
		Msgs:         msgs,
		Account:      *redemptionAccount,
		Timeout:      timeout,
	}

	return &icaTx, nil
}
