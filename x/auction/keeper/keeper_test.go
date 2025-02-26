package keeper_test

import (
	"bytes"
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v25/app/apptesting"
	"github.com/Stride-Labs/stride/v25/x/auction/keeper"
	"github.com/Stride-Labs/stride/v25/x/auction/types"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	logBuffer bytes.Buffer
}

// Modify SetupTest to include mock setup
func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Create a logger with accessible output
	logger := log.NewTMLogger(&s.logBuffer)
	s.Ctx = s.Ctx.WithLogger(logger)
}

// Dynamically gets the MsgServer for this module's keeper
// this function must be used so that the MsgServer is always created with the most updated App context
//
//	which can change depending on the type of test
//	(e.g. tests with only one Stride chain vs tests with multiple chains and IBC support)
func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.AuctionKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Helper function to get a auction and confirm there's no error
func (s *KeeperTestSuite) MustGetAuction(name string) types.Auction {
	auction, err := s.App.AuctionKeeper.GetAuction(s.Ctx, name)
	s.Require().NoError(err, "no error expected when getting auction")
	return *auction
}
