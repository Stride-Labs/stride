package keeper_test

import (
	"math/rand"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v8/testutil/keeper"
	"github.com/Stride-Labs/stride/v8/testutil/nullify"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func createSetNLSMTokenDeposit(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.LSMTokenDeposit {
	items := make([]types.LSMTokenDeposit, n)
	for i := range items {
		validatorAddr := "validatorAddress"
		tokenRecordId := rand.Int()

		items[i].Denom = validatorAddr + strconv.Itoa(tokenRecordId)
		items[i].ValidatorAddress = validatorAddr
		items[i].ChainId = strconv.Itoa(i)
		items[i].Amount = sdkmath.ZeroInt()
		items[i].Status = types.TRANSFER_IN_PROGRESS
		keeper.SetLSMTokenDeposit(ctx, items[i])
	}
	return items
}

func TestLSMTokenDepositGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createSetNLSMTokenDeposit(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetLSMTokenDeposit(ctx, item.ChainId, item.Denom)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestLSMTokenDepositRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createSetNLSMTokenDeposit(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLSMTokenDeposit(ctx, item.ChainId, item.Denom)
		_, found := keeper.GetLSMTokenDeposit(ctx, item.ChainId, item.Denom)
		require.False(t, found)
	}
}

func TestLSMTokenDepositGetAll(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createSetNLSMTokenDeposit(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllLSMTokenDeposit(ctx)),
	)
}
