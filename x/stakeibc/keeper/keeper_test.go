package keeper_test

import (
	"testing"

	"github.com/Stride-Labs/stride/app/apptesting"
	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	msgServer types.MsgServer
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.msgServer = keeper.NewMsgServerImpl(suite.App.StakeibcKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
