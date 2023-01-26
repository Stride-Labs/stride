package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

func (k msgServer) ResetUnbondingRecordEpochNumbers(goCtx context.Context, msg *types.MsgResetUnbondingRecordEpochNumbers) (*types.MsgResetUnbondingRecordEpochNumbersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Reset epoch numbers to match the number used in the epoch key
	k.RecordsKeeper.ResetEpochUnbondingRecordEpochNumbers(ctx)

	// Confirm epoch numbers were set correctly (there should be no 0 epoch numbers)
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		if epochUnbondingRecord.EpochNumber == 0 {
			return nil, errors.New("Reset of epoch unbonding record epoch numbers was not successful. EpochNumber of 0 found in state.")
		}
	}

	return &types.MsgResetUnbondingRecordEpochNumbersResponse{}, nil
}
