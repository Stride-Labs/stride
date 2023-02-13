package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
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

// Dynamically gets the MsgServer for this module's keeper
// this function must be used so that the MsgServer is always created with the most updated App context
//	which can change depending on the type of test
//	(e.g. tests with only one Stride chain vs tests with multiple chains and IBC support)
func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.StakeibcKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestIsRedemptionRateWithinSafetyBounds() {
	params := s.App.StakeibcKeeper.GetParams(s.Ctx)
	params.MinRedemptionRates["osmosis-1"] = "0.3"
	params.MaxRedemptionRates["osmosis-1"] = "2.0"
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	zone := types.HostZone{
		ChainId:        "osmosis-1",
		RedemptionRate: sdk.NewDec(1),
	}
	rrSafe, err := s.App.StakeibcKeeper.IsRedemptionRateWithinSafetyBounds(s.Ctx, zone)
	s.Require().NoError(err)
	s.Require().True(rrSafe)

	zone.RedemptionRate = sdk.NewDecWithPrec(16, 1) // 1.6
	rrSafe, err = s.App.StakeibcKeeper.IsRedemptionRateWithinSafetyBounds(s.Ctx, zone)
	s.Require().NoError(err)
	s.Require().True(rrSafe)

	zone.RedemptionRate = sdk.NewDecWithPrec(21, 1) // 2.1
	rrSafe, err = s.App.StakeibcKeeper.IsRedemptionRateWithinSafetyBounds(s.Ctx, zone)
	s.Require().Error(err)
	s.Require().False(rrSafe)
}
