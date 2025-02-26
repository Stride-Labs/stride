package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	"github.com/Stride-Labs/stride/v26/x/stakedym/keeper"
	"github.com/Stride-Labs/stride/v26/x/stakedym/types"
)

const (
	HostChainId     = "chain-0"
	HostNativeDenom = "denom"
	HostIBCDenom    = "ibc/denom"
	StDenom         = "stdenom"

	ValidOperator      = "stride1njt6kn0c2a2w5ax8mlm9k0fmcc8tyjgh7s8hu8"
	ValidTxHashDefault = "BBD978ADDBF580AC2981E351A3EA34AA9D7B57631E9CE21C27C2C63A5B13BDA9"
	ValidTxHashNew     = "FEFC69DCDF00E2BF971A61D34944871F607C84787CA9A69715B360A767FE6862"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

// Dynamically gets the MsgServer for this module's keeper
// this function must be used so that the MsgServer is always created with the most updated App context
//
//	which can change depending on the type of test
//	(e.g. tests with only one Stride chain vs tests with multiple chains and IBC support)
func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.StakedymKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Helper function to get a host zone and confirm there's no error
func (s *KeeperTestSuite) MustGetHostZone() types.HostZone {
	hostZone, err := s.App.StakedymKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "no error expected when getting host zone")
	return hostZone
}
