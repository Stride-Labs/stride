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
	params.SafetyMinRedemptionRateThreshold = 75
	params.SafetyMaxRedemptionRateThreshold = 150
	params.MinRedemptionRates["gaia-1"] = "1.5"
	params.MaxRedemptionRates["gaia-1"] = "2.5"
	params.MinRedemptionRates["osmosis-1"] = "0.3"
	params.MaxRedemptionRates["osmosis-1"] = "2.0"
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	for _, tc := range []struct {
		chainId        string
		redemptionRate sdk.Dec
		expSafe        bool
	}{
		{
			chainId:        "osmosis-1",
			redemptionRate: sdk.NewDecWithPrec(1, 1), // 0.1
			expSafe:        false,
		},
		{
			chainId:        "osmosis-1",
			redemptionRate: sdk.NewDecWithPrec(3, 1), // 0.3
			expSafe:        true,
		},
		{
			chainId:        "osmosis-1",
			redemptionRate: sdk.NewDecWithPrec(15, 1), // 1.5
			expSafe:        true,
		},
		{
			chainId:        "osmosis-1",
			redemptionRate: sdk.NewDecWithPrec(25, 1), // 2.5
			expSafe:        false,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdk.NewDecWithPrec(1, 1), // 0.1
			expSafe:        false,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdk.NewDecWithPrec(3, 1), // 0.3
			expSafe:        false,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdk.NewDecWithPrec(15, 1), // 1.5
			expSafe:        true,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdk.NewDecWithPrec(25, 1), // 2.5
			expSafe:        true,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdk.NewDecWithPrec(1, 1), // 0.1
			expSafe:        false,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdk.NewDecWithPrec(3, 1), // 0.3
			expSafe:        false,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdk.NewDecWithPrec(15, 1), // 1.5
			expSafe:        true,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdk.NewDecWithPrec(25, 1), // 2.5
			expSafe:        false,
		},
	} {
		rrSafe, err := s.App.StakeibcKeeper.IsRedemptionRateWithinSafetyBounds(s.Ctx, types.HostZone{
			ChainId:        tc.chainId,
			RedemptionRate: tc.redemptionRate,
		})
		if tc.expSafe {
			s.Require().NoError(err)
			s.Require().True(rrSafe)
		} else {
			s.Require().Error(err)
			s.Require().False(rrSafe)
		}
	}
}
