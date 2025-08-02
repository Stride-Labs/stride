package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
	"github.com/Stride-Labs/stride/v28/x/icqoracle/keeper"
	"github.com/Stride-Labs/stride/v28/x/icqoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v28/x/interchainquery/types"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	mockICQKeeper types.IcqKeeper
}

// Helper function to setup keeper with mock ICQ keeper
func (s *KeeperTestSuite) SetupMockICQKeeper() {
	mockICQKeeper := MockICQKeeper{
		SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
			return nil
		},
	}
	s.App.ICQOracleKeeper.IcqKeeper = mockICQKeeper
}

// Modify SetupTest to include mock setup
func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.SetupMockICQKeeper()

	// Set the time to test price staleness
	s.Ctx = s.Ctx.WithBlockTime(time.Now().UTC())
}

// Dynamically gets the MsgServer for this module's keeper
// this function must be used so that the MsgServer is always created with the most updated App context
//
//	which can change depending on the type of test
//	(e.g. tests with only one Stride chain vs tests with multiple chains and IBC support)
func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.ICQOracleKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Helper function to get a token price and confirm there's no error
func (s *KeeperTestSuite) MustGetTokenPrice(baseDenom string, quoteDenom string, osmosisPoolId uint64) types.TokenPrice {
	tp, err := s.App.ICQOracleKeeper.GetTokenPrice(s.Ctx, baseDenom, quoteDenom, osmosisPoolId)
	s.Require().NoError(err, "no error expected when getting token price")
	return tp
}
