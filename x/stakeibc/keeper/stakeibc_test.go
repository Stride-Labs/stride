package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/testutil/keeper"
	"github.com/Stride-Labs/stride/testutil/nullify"
	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	strideapp "github.com/Stride-Labs/stride/app"
	_ "github.com/stretchr/testify/suite"
)

//========================= LIQUID STAKING TESTS =========================== 

func (suite *KeeperTestSuite) TestLiquidStake() {
	suite.SetupTest()
	
	isCheckTx := false
	app := strideapp.Setup(isCheckTx)
	msgServer := keeper.NewMsgServerImpl(app.StakeibcKeeper)
	msg := types.NewMsgLiquidStake("stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7", 1, "uatom")
	_, err := msgServer.LiquidStake(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Require().NoError(err)
}


//========================= DELEGATION TESTS =========================== 

func createTestDelegation(keeper *keeper.Keeper, ctx sdk.Context) types.Delegation {
	item := types.Delegation{}
	keeper.SetDelegation(ctx, item)
	return item
}

func TestDelegationGet(t *testing.T) {
	k, ctx := keepertest.StakeibcKeeper(t)
	expected := createTestDelegation(k, ctx)
	actual, found := k.GetDelegation(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&expected),
		nullify.Fill(&actual),
	)
}

func TestDelegationRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	createTestDelegation(keeper, ctx)
	keeper.RemoveDelegation(ctx)
	_, found := keeper.GetDelegation(ctx)
	require.False(t, found)
}
