package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	keepertest "github.com/Stride-Labs/stride/v27/testutil/keeper"
	"github.com/Stride-Labs/stride/v27/testutil/nullify"
	"github.com/Stride-Labs/stride/v27/x/records/keeper"
	"github.com/Stride-Labs/stride/v27/x/records/types"
)

// Helper function to create a new host zone unbonding record, filling in the sdkmath.Int's
// so that they can be compared
func newHostZoneUnbonding(chainId string, status types.HostZoneUnbonding_Status) types.HostZoneUnbonding {
	return types.HostZoneUnbonding{
		HostZoneId:            chainId,
		Status:                status,
		StTokenAmount:         sdkmath.ZeroInt(),
		NativeTokenAmount:     sdkmath.ZeroInt(),
		NativeTokensToUnbond:  sdkmath.ZeroInt(),
		StTokensToBurn:        sdkmath.ZeroInt(),
		ClaimableNativeTokens: sdkmath.ZeroInt(),
	}
}

func createNEpochUnbondingRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) ([]types.EpochUnbondingRecord, map[string]types.HostZoneUnbonding) {
	hostZoneUnbondingsList := []types.HostZoneUnbonding{
		newHostZoneUnbonding("host-A", types.HostZoneUnbonding_UNBONDING_QUEUE),
		newHostZoneUnbonding("host-B", types.HostZoneUnbonding_UNBONDING_QUEUE),
		newHostZoneUnbonding("host-C", types.HostZoneUnbonding_UNBONDING_QUEUE),
	}
	hostZoneUnbondingsMap := make(map[string]types.HostZoneUnbonding)
	for _, hostZoneUnbonding := range hostZoneUnbondingsList {
		hostZoneUnbondingsMap[hostZoneUnbonding.HostZoneId] = hostZoneUnbonding
	}

	epochUnbondingRecords := make([]types.EpochUnbondingRecord, n)
	for epochNumber, epochUnbondingRecord := range epochUnbondingRecords {
		epochUnbondingRecord.EpochNumber = uint64(epochNumber)

		unbondingsCopy := make([]*types.HostZoneUnbonding, 3)
		for i := range unbondingsCopy {
			hostZoneUnbonding := hostZoneUnbondingsList[i]
			epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, &hostZoneUnbonding)
		}

		epochUnbondingRecords[epochNumber] = epochUnbondingRecord
		keeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	}
	return epochUnbondingRecords, hostZoneUnbondingsMap
}

func TestEpochUnbondingRecordGet(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items, _ := createNEpochUnbondingRecord(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetEpochUnbondingRecord(ctx, item.EpochNumber)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestEpochUnbondingRecordRemove(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items, _ := createNEpochUnbondingRecord(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveEpochUnbondingRecord(ctx, item.EpochNumber)
		_, found := keeper.GetEpochUnbondingRecord(ctx, item.EpochNumber)
		require.False(t, found)
	}
}

func TestEpochUnbondingRecordGetAll(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items, _ := createNEpochUnbondingRecord(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllEpochUnbondingRecord(ctx)),
	)
}

func TestGetAllPreviousEpochUnbondingRecords(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items, _ := createNEpochUnbondingRecord(keeper, ctx, 10)
	currentEpoch := uint64(8)
	fetchedItems := items[:currentEpoch]
	require.ElementsMatch(t,
		nullify.Fill(fetchedItems),
		nullify.Fill(keeper.GetAllPreviousEpochUnbondingRecords(ctx, currentEpoch)),
	)
}

func TestGetHostZoneUnbondingByChainId(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	_, hostZoneUnbondings := createNEpochUnbondingRecord(keeper, ctx, 10)

	expectedHostZoneUnbonding := hostZoneUnbondings["host-B"]
	actualHostZoneUnbonding, found := keeper.GetHostZoneUnbondingByChainId(ctx, 1, "host-B")

	require.True(t, found)
	require.Equal(t,
		*actualHostZoneUnbonding,
		expectedHostZoneUnbonding,
	)
}

func (s *KeeperTestSuite) TestAddHostZoneToEpochUnbondingRecord() {
	epochUnbondingRecords, _ := createNEpochUnbondingRecord(&s.App.RecordsKeeper, s.Ctx, 3)

	epochNumber := uint64(0)
	initialEpochUnbondingRecord := epochUnbondingRecords[int(epochNumber)]

	// Update host zone unbonding for host-C
	updatedHostZoneUnbonding := newHostZoneUnbonding("host-C", types.HostZoneUnbonding_UNBONDING_IN_PROGRESS)

	expectedEpochUnbondingRecord := initialEpochUnbondingRecord
	expectedEpochUnbondingRecord.HostZoneUnbondings[2] = &updatedHostZoneUnbonding

	updatedEpochUnbonding, err := s.App.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(s.Ctx, epochNumber, "host-C", updatedHostZoneUnbonding)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, updatedEpochUnbonding)
	s.Require().NoError(err, "no error expected when updating host-C")
	for i := 0; i < len(expectedEpochUnbondingRecord.HostZoneUnbondings); i++ {
		expectedHostZoneUnbonding := *expectedEpochUnbondingRecord.HostZoneUnbondings[i]
		actualHostZoneUnbonding := *updatedEpochUnbonding.HostZoneUnbondings[i]
		s.Require().Equal(expectedHostZoneUnbonding, actualHostZoneUnbonding, "HZU %d after host-C update", i)
	}

	// Add new host zone to initial epoch unbonding records
	newHostZoneUnbonding := newHostZoneUnbonding("host-D", types.HostZoneUnbonding_UNBONDING_QUEUE)
	expectedEpochUnbondingRecord.HostZoneUnbondings = append(expectedEpochUnbondingRecord.HostZoneUnbondings, &newHostZoneUnbonding)

	updatedEpochUnbonding, err = s.App.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(s.Ctx, epochNumber, "host-D", newHostZoneUnbonding)
	s.Require().NoError(err, "no error expected when adding host-D")
	s.Require().Equal(expectedEpochUnbondingRecord, updatedEpochUnbonding, "EUR after host-D addition")
}

func TestSetHostZoneUnbondingStatus(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)

	initialEpochUnbondingRecords, _ := createNEpochUnbondingRecord(keeper, ctx, 4)

	epochsToUpdate := []uint64{1, 3}
	hostIdToUpdate := "host-B"
	newStatus := types.HostZoneUnbonding_UNBONDING_IN_PROGRESS

	expectedEpochUnbondingRecords := initialEpochUnbondingRecords
	for _, epochUnbondingRecord := range expectedEpochUnbondingRecords {
		for _, epochNumberToUpdate := range epochsToUpdate {
			if epochUnbondingRecord.EpochNumber == epochNumberToUpdate {
				for i, hostUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
					if hostUnbonding.HostZoneId == hostIdToUpdate {
						updatedHostZoneUnbonding := hostUnbonding
						updatedHostZoneUnbonding.Status = newStatus
						epochUnbondingRecord.HostZoneUnbondings[i] = updatedHostZoneUnbonding
					}
				}
			}
		}
	}

	err := keeper.SetHostZoneUnbondingStatus(ctx, hostIdToUpdate, epochsToUpdate, newStatus)
	require.Nil(t, err)

	actualEpochUnbondingRecord := keeper.GetAllEpochUnbondingRecord(ctx)
	require.ElementsMatch(t,
		expectedEpochUnbondingRecords,
		actualEpochUnbondingRecord,
	)
}

func (s *KeeperTestSuite) TestSetHostZoneUnbonding() {
	initialAmount := sdkmath.NewInt(10)
	updatedAmount := sdkmath.NewInt(99)

	// Create two epoch unbonding records, each with two host zone unbondings
	epochUnbondingRecords := []types.EpochUnbondingRecord{
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*types.HostZoneUnbonding{
				{HostZoneId: "chain-0", NativeTokenAmount: initialAmount},
				{HostZoneId: "chain-1", NativeTokenAmount: initialAmount},
			},
		},
		{
			EpochNumber: 2,
			HostZoneUnbondings: []*types.HostZoneUnbonding{
				{HostZoneId: "chain-0", NativeTokenAmount: initialAmount},
				{HostZoneId: "chain-1", NativeTokenAmount: initialAmount},
			},
		},
	}
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	// Update the amount for (epoch-1, chain-0) and (epoch-2, chain-1)
	updatedHostZoneUnbonding1 := types.HostZoneUnbonding{HostZoneId: "chain-0", NativeTokenAmount: updatedAmount}
	err := s.App.RecordsKeeper.SetHostZoneUnbondingRecord(s.Ctx, 1, "chain-0", updatedHostZoneUnbonding1)
	s.Require().NoError(err, "no error expected when setting amount for (epoch-1, chain-0)")

	updatedHostZoneUnbonding2 := types.HostZoneUnbonding{HostZoneId: "chain-1", NativeTokenAmount: updatedAmount}
	err = s.App.RecordsKeeper.SetHostZoneUnbondingRecord(s.Ctx, 2, "chain-1", updatedHostZoneUnbonding2)
	s.Require().NoError(err, "no error expected when setting amount for (epoch-2, chain-1)")

	// Create the mapping of expected native amounts
	expectedAmountMapping := map[uint64]map[string]sdkmath.Int{
		1: {
			"chain-0": updatedAmount,
			"chain-1": initialAmount,
		},
		2: {
			"chain-0": initialAmount,
			"chain-1": updatedAmount,
		},
	}

	// Loop the records and check that the amounts match the updates
	for _, epochUnbondingRecord := range s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx) {
		s.Require().Len(epochUnbondingRecord.HostZoneUnbondings, 2, "there should be two host records per epoch record")

		for _, hostZoneUnbondingRecord := range epochUnbondingRecord.HostZoneUnbondings {
			expectedAmount := expectedAmountMapping[epochUnbondingRecord.EpochNumber][hostZoneUnbondingRecord.HostZoneId]
			s.Require().Equal(expectedAmount.Int64(), hostZoneUnbondingRecord.NativeTokenAmount.Int64(), "updated record amount")
		}
	}

}
