package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

func (k msgServer) ResetUnbondingRecordEpochNumbers(goCtx context.Context, msg *types.MsgResetUnbondingRecordEpochNumbers) (*types.MsgResetUnbondingRecordEpochNumbersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	initialEpochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)

	// Reset epoch numbers to match the number used in the epoch key
	err := k.RecordsKeeper.ResetEpochUnbondingRecordEpochNumbers(ctx)
	if err != nil {
		return nil, fmt.Errorf("Reset of epoch unbonding record epoch numbers was not successful: %s", err.Error())
	}

	// Confirm epoch numbers were set correctly
	epochNumberMap := make(map[uint64]bool)
	for _, epochUnbondingRecord := range initialEpochUnbondingRecords {
		// There should be no duplicate epoch numbers
		if _, epochNumberAlreadyExists := epochNumberMap[epochUnbondingRecord.EpochNumber]; epochNumberAlreadyExists {
			return nil, errors.New("Reset of epoch unbonding record epoch numbers was not successful. Duplicate EpochNumber exist")
		}

		// There should be no 0 epoch numbers
		if epochUnbondingRecord.EpochNumber == 0 {
			return nil, errors.New("Reset of epoch unbonding record epoch numbers was not successful. EpochNumber of 0 found in state.")
		}

		epochNumberMap[epochUnbondingRecord.EpochNumber] = true
	}

	return &types.MsgResetUnbondingRecordEpochNumbersResponse{}, nil
}
