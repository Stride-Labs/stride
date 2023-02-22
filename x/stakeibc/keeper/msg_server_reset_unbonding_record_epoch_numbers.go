package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v6/x/stakeibc/types"
)

const ErrorResettingUnbondingRecord = "Reset of epoch unbonding record epoch numbers was not successful"

func (k msgServer) ResetUnbondingRecordEpochNumbers(goCtx context.Context, msg *types.MsgResetUnbondingRecordEpochNumbers) (*types.MsgResetUnbondingRecordEpochNumbersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	initialEpochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)

	// Reset epoch numbers to match the number used in the epoch key
	err := k.RecordsKeeper.ResetEpochUnbondingRecordEpochNumbers(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", ErrorResettingUnbondingRecord, err.Error())
	}

	// Confirm we have the same number of epoch unbonding records after the reset
	finalEpochUnondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)
	if len(initialEpochUnbondingRecords) != len(finalEpochUnondingRecords) {
		return nil, errors.New(fmt.Sprintf("%s: Number of epoch unbonding records changed - Previous: %d, After Reset: %d",
			ErrorResettingUnbondingRecord, len(initialEpochUnbondingRecords), len(finalEpochUnondingRecords)))
	}

	// Validate each unbonding record
	epochNumberMap := make(map[uint64]bool)
	for _, epochUnbondingRecord := range finalEpochUnondingRecords {
		// There should be no duplicate epoch numbers
		if _, epochNumberAlreadyExists := epochNumberMap[epochUnbondingRecord.EpochNumber]; epochNumberAlreadyExists {
			return nil, errors.New(fmt.Sprintf("%s: Duplicate EpochNumber exists", ErrorResettingUnbondingRecord))
		}

		// There should be no 0 epoch numbers
		if epochUnbondingRecord.EpochNumber == 0 {
			return nil, errors.New(fmt.Sprintf("%s: EpochNumber of 0 found in state.", ErrorResettingUnbondingRecord))
		}

		epochNumberMap[epochUnbondingRecord.EpochNumber] = true
	}

	return &types.MsgResetUnbondingRecordEpochNumbersResponse{}, nil
}
