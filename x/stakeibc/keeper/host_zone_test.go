package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v9/testutil/keeper"
	"github.com/Stride-Labs/stride/v9/testutil/nullify"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func createNHostZone(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.HostZone {
	items := make([]types.HostZone, n)
	for i := range items {
		items[i].ChainId = strconv.Itoa(i)
		items[i].RedemptionRate = sdk.NewDec(1)
		items[i].LastRedemptionRate = sdk.NewDec(1)
		items[i].MinRedemptionRate = sdk.NewDecWithPrec(5, 1)
		items[i].MaxRedemptionRate = sdk.NewDecWithPrec(15, 1)
		items[i].StakedBal = sdkmath.ZeroInt()
		keeper.SetHostZone(ctx, items[i])
	}
	return items
}

func TestHostZoneGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNHostZone(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetHostZone(ctx, item.ChainId)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestHostZoneRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNHostZone(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveHostZone(ctx, item.ChainId)
		_, found := keeper.GetHostZone(ctx, item.ChainId)
		require.False(t, found)
	}
}

func TestHostZoneGetAll(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNHostZone(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllHostZone(ctx)),
	)
}

func TestHostZoneGetAllActiveCase1(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)

	// Case 1: some active some inactive
	numZones := 3
	items := createNHostZone(keeper, ctx, numZones)
	// set the last host zone as halted
	items[numZones-1].Halted = true
	keeper.SetHostZone(ctx, items[numZones-1])

	// only the last host zone is active, so we expect all except that one
	actualActiveHzs := items[:numZones-1]
	getActiveHzResults := keeper.GetAllActiveHostZone(ctx)
	require.ElementsMatch(t,
		nullify.Fill(actualActiveHzs),
		nullify.Fill(getActiveHzResults),
	)
}

func TestHostZoneGetAllActiveCase2(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)

	// Case 2: all active
	numZones := 3
	items := createNHostZone(keeper, ctx, numZones)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllActiveHostZone(ctx)),
	)
}

func TestHostZoneGetAllActiveCase3(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)

	// Case 3: all inactive
	numZones := 3
	items := createNHostZone(keeper, ctx, numZones)
	// set the last host zone as halted
	items[0].Halted = true
	items[1].Halted = true
	items[2].Halted = true
	keeper.SetHostZone(ctx, items[0])
	keeper.SetHostZone(ctx, items[1])
	keeper.SetHostZone(ctx, items[2])
	require.ElementsMatch(t,
		nullify.Fill(types.HostZone{}),
		nullify.Fill(keeper.GetAllActiveHostZone(ctx)),
	)
}

func TestHostZoneGetAllActiveCase4(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)

	// create no zones, check the output is an empty list
	require.ElementsMatch(t,
		nullify.Fill(types.HostZone{}),
		nullify.Fill(keeper.GetAllActiveHostZone(ctx)),
	)
}

func TestGetValidatorFromAddress(t *testing.T) {
	numValidators := 3

	// Create list of validators
	addresses := []string{}
	validators := []*types.Validator{}
	for i := 1; i <= numValidators; i++ {
		address := fmt.Sprintf("val-%d", i)

		addresses = append(addresses, address)
		validators = append(validators, &types.Validator{Address: address})
	}

	// For each validator that was just added, test GetValidatorFromAddress
	for expectedIndex, address := range addresses {
		expectedValidator := *validators[expectedIndex]
		actualValidator, actualIndex, found := keeper.GetValidatorFromAddress(validators, address)

		require.True(t, found)
		require.Equal(t, expectedValidator, actualValidator)
		require.Equal(t, int64(expectedIndex), actualIndex)
	}

	// Test GetValidatorFromAddress for an validator that doesn't exist
	_, _, found := keeper.GetValidatorFromAddress(validators, "fake_validator")
	require.False(t, found)
}

func (s *KeeperTestSuite) TestGetHostZoneFromTransferChannelID() {
	// Store 5 host zones
	expectedHostZones := map[string]types.HostZone{}
	for i := 0; i < 5; i++ {
		chainId := fmt.Sprintf("chain-%d", i)
		channelId := fmt.Sprintf("channel-%d", i)

		hostZone := types.HostZone{
			ChainId:           chainId,
			TransferChannelId: channelId,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
		expectedHostZones[channelId] = hostZone
	}

	// Look up each host zone by the channel ID
	for i := 0; i < 5; i++ {
		channelId := fmt.Sprintf("channel-%d", i)

		expectedHostZone := expectedHostZones[channelId]
		actualHostZone, found := s.App.StakeibcKeeper.GetHostZoneFromTransferChannelID(s.Ctx, channelId)

		s.Require().True(found, "found host zone %d", i)
		s.Require().Equal(expectedHostZone.ChainId, actualHostZone.ChainId, "host zone %d chain-id", i)
	}

	// Lookup a non-existent host zone - should not be found
	_, found := s.App.StakeibcKeeper.GetHostZoneFromTransferChannelID(s.Ctx, "fake_channel")
	s.Require().False(found, "fake channel should not be found")
}
