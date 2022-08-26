package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/app/apptesting"
	"github.com/Stride-Labs/stride/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/x/interchainquery/types"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	msgServer types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.msgServer = keeper.NewMsgServerImpl(s.App.InterchainqueryKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
