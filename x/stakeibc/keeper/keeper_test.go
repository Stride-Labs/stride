package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

const (
	Atom         = "uatom"
	StAtom       = "stuatom"
	IbcAtom      = "ibc/uatom"
	GaiaPrefix   = "cosmos"
	HostChainId  = "GAIA"
	Bech32Prefix = "cosmos"

	Osmo        = "uosmo"
	StOsmo      = "stuosmo"
	IbcOsmo     = "ibc/uosmo"
	OsmoPrefix  = "osmo"
	OsmoChainId = "OSMO"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.StakeibcKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
