package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v14/testutil/keeper"
	"github.com/Stride-Labs/stride/v14/testutil/nullify"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func createNTradeRoute(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.TradeRoute {
	items := make([]types.TradeRoute, n)
	for i := range items {

		hostChain := strconv.Itoa(i) + "chain"
		rewardChain := strconv.Itoa(i+1) + "chain"
		tradeChain := strconv.Itoa(i+2) + "chain"

		hostICA := types.ICAAccount{
			ChainId: hostChain,
			Type: types.ICAAccountType_WITHDRAWAL,
		}
		rewardICA := types.ICAAccount{
			ChainId: rewardChain,
			Type: types.ICAAccountType_UNWIND,
		}
		tradeICA := types.ICAAccount{
			ChainId: tradeChain,
			Type: types.ICAAccountType_TRADE,
		}

		host_reward_hop := types.TradeHop{
			FromAccount: &hostICA,
			ToAccount: &rewardICA,
		}
		reward_trade_hop := types.TradeHop{
			FromAccount: &rewardICA,
			ToAccount: &tradeICA,
		}
		trade_host_hop := types.TradeHop{
			FromAccount: &tradeICA,
			ToAccount: &hostICA,			
		}

		hostDenom := strconv.Itoa(i) + "denom"
		rewardDenom := strconv.Itoa(i+1) + "denom"

		items[i].RewardDenomOnHostZone = "ibc-" + rewardDenom + "-on-" + hostChain
		items[i].RewardDenomOnRewardZone = rewardDenom
		items[i].RewardDenomOnTradeZone = "ibc-" + rewardDenom + "-on-" + tradeChain
		items[i].TargetDenomOnTradeZone = "ibc-" + hostDenom + "-on-" + tradeChain
		items[i].TargetDenomOnHostZone = hostDenom

		items[i].HostToRewardHop = &host_reward_hop
		items[i].RewardToTradeHop = &reward_trade_hop
		items[i].TradeToHostHop = &trade_host_hop

		items[i].PoolId = uint64(i * 1000)

		keeper.SetTradeRoute(ctx, items[i])
	}
	return items
}

func TestTradeRouteGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNTradeRoute(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetTradeRoute(ctx, item.RewardDenomOnHostZone, item.TargetDenomOnHostZone)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestTradeRouteRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNTradeRoute(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveTradeRoute(ctx, item.RewardDenomOnHostZone, item.TargetDenomOnHostZone)
		_, found := keeper.GetTradeRoute(ctx, item.RewardDenomOnHostZone, item.TargetDenomOnHostZone)
		require.False(t, found)
	}
}

func TestTradeRouteGetAll(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNTradeRoute(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllTradeRoute(ctx)),
	)
}

func TestTradeRouteGetAllForHostZone(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	createNTradeRoute(keeper, ctx, 10)

	chainNum := strconv.Itoa(3)
	chainId :=  chainNum + "chain"
	chainDenom := chainNum + "denom"
	zone := types.HostZone{
		ChainId: chainId,
		HostDenom: chainDenom,
	}
	keeper.SetHostZone(ctx, zone)

	routes := keeper.GetAllTradeRouteForHostZone(ctx, chainId)
	require.Equal(t, 1, len(routes), "Should only be one route for %s", chainId)
	require.Equal(t, chainId, routes[0].HostToRewardHop.FromAccount.ChainId, "Host ICA should be from the requested chain")
	require.Equal(t, chainDenom, routes[0].TargetDenomOnHostZone, "TargetDenomOnHostZone should be the host denom for the zone")
}
