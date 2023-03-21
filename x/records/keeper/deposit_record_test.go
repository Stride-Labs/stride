package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v7/testutil/keeper"
	"github.com/Stride-Labs/stride/v7/x/records/types"
)

func TestSumDepositRecords(t *testing.T) {
	keeper, _ := keepertest.RecordsKeeper(t)

	depositRecords := []types.DepositRecord{{Amount: math.NewInt(50)}, {Amount: math.NewInt(50)}}
	sum := keeper.SumDepositRecords(depositRecords)
	require.Equal(t, math.NewInt(100), sum)

	// Make sure no records sum equals 0
	depositRecords = []types.DepositRecord{}
	sum = keeper.SumDepositRecords(depositRecords)
	require.Equal(t, math.ZeroInt(), sum)
}

func TestFilterDepositRecords(t *testing.T) {
	keeper, _ := keepertest.RecordsKeeper(t)
	transferRecord := types.DepositRecord{Status: types.DepositRecord_TRANSFER_QUEUE}
	delegationRecord := types.DepositRecord{Status: types.DepositRecord_DELEGATION_QUEUE}
	depositRecords := []types.DepositRecord{transferRecord, delegationRecord}

	filterDelegationRecords := func(record types.DepositRecord) (condition bool) {
		return record.Status == types.DepositRecord_DELEGATION_QUEUE
	}
	delegationRecords := keeper.FilterDepositRecords(depositRecords, filterDelegationRecords)

	require.Len(t, delegationRecords, 1)
	require.Equal(t, delegationRecord, delegationRecords[0])
}
