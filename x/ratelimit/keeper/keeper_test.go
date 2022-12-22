package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/keeper"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	QueryClient types.QueryClient
	MsgServer   types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.QueryClient = types.NewQueryClient(s.QueryHelper)
	s.MsgServer = keeper.NewMsgServerImpl(s.App.RatelimitKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
