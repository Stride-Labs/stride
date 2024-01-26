package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v18/app/apptesting"
	"github.com/Stride-Labs/stride/v18/x/autopilot/types"
)

const (
	HostChainId    = "chain-0"
	HostBechPrefix = "cosmos"
	HostAddress    = "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k"
	HostDenom      = "uatom"

	Atom = "uatom"
	Strd = "ustrd"
	Osmo = "uosmo"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	QueryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.QueryClient = types.NewQueryClient(s.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
