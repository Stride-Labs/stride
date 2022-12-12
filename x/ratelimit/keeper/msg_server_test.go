package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/keeper"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func (suite *KeeperTestSuite) TestAdd_Quota() {
	suite.SetupTest()
	msgServer := keeper.NewMsgServerImpl(suite.App.RatelimitKeeper)

	validAddr, _ := apptesting.GenerateTestAddrs()
	_, err := msgServer.AddQuota(sdk.WrapSDKContext(suite.Ctx), &types.MsgAddQuota{
		Creator:         validAddr,
		Name:            "quota",
		MaxPercentRecv:  10,
		MaxPercentSend:  20,
		DurationMinutes: 30,
	})
	suite.Require().NoError(err)

	_, found := suite.App.RatelimitKeeper.GetQuota(suite.Ctx, "quota")
	suite.Require().True(found)

	// check quota duplication
	_, err = msgServer.AddQuota(sdk.WrapSDKContext(suite.Ctx), &types.MsgAddQuota{
		Creator:         validAddr,
		Name:            "quota",
		MaxPercentRecv:  10,
		MaxPercentSend:  20,
		DurationMinutes: 30,
	})
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestRemove_Quota() {
	suite.SetupTest()
	msgServer := keeper.NewMsgServerImpl(suite.App.RatelimitKeeper)

	validAddr, _ := apptesting.GenerateTestAddrs()
	_, err := msgServer.RemoveQuota(sdk.WrapSDKContext(suite.Ctx), &types.MsgRemoveQuota{
		Creator: validAddr,
		Name:    "quota",
	})
	suite.Require().NoError(err)

	_, found := suite.App.RatelimitKeeper.GetQuota(suite.Ctx, "quota")
	suite.Require().False(found)
}
