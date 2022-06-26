package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func (k msgServer) ClaimUndelegatedTokens(goCtx context.Context, msg *types.MsgClaimUndelegatedTokens) (*types.MsgClaimUndelegatedTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// grab our host zone
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		errMsg := fmt.Sprintf("Host zone %s not found", msg.HostZone)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
	}
	// grab necessary fields to construct ICA call
	var msgs []sdk.Msg
	redemptionAccount, found := k.GetRedemptionAccount(ctx, hostZone)
	if !found {
		errMsg := fmt.Sprintf("Redemption account not found for host zone %s", msg.HostZone)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
	}
	// go through the desired number of records and claim them
	numRecordsToClaim := min(int(msg.MaxClaims), len(hostZone.ClaimableRecordIds))
	for i := 0; i < numRecordsToClaim; i++ {
		record, found := k.recordsKeeper.GetUserRedemptionRecord(ctx, hostZone.ClaimableRecordIds[i])
		if !found {
			errMsg := fmt.Sprintf("User redemption record %s not found on host zone %s", hostZone.ClaimableRecordIds[i], hostZone.ChainId)
			k.Logger(ctx).Error(errMsg)
			return nil, sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
		}
		msgs = append(msgs, &bankTypes.MsgSend{
			FromAddress: redemptionAccount.Address,
			ToAddress:   record.Receiver,
		})
	}
	// TODO we should do some error handling here, in case this call fails
	err := k.SubmitTxs(ctx, hostZone.GetConnectionId(), msgs, *redemptionAccount)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}
	// now go through and delete these records
	for i := 0; i < numRecordsToClaim; i++ {
		k.recordsKeeper.RemoveUserRedemptionRecord(ctx, hostZone.ClaimableRecordIds[i])
	}
	// finally clean up these records from claimable records
	hostZone.ClaimableRecordIds = hostZone.ClaimableRecordIds[numRecordsToClaim:]
	return &types.MsgClaimUndelegatedTokensResponse{}, nil
}
