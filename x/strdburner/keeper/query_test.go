package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
)

func (s *KeeperTestSuite) TestQueryStrdBurnerAddress() {
	// Query for the strd burner address
	req := &types.QueryStrdBurnerAddressRequest{}
	resp, err := s.App.StrdBurnerKeeper.StrdBurnerAddress(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying strd burner address")
	s.Require().Equal(s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(), resp.Address, "address")
}

func (s *KeeperTestSuite) TestQueryTotalStrdBurned() {
	// Set initial total burned amount
	expectedProtocol := sdkmath.NewInt(1000000)
	expectedUser := sdkmath.NewInt(2000000)
	expectedTotal := sdkmath.NewInt(3000000)
	s.App.StrdBurnerKeeper.SetProtocolStrdBurned(s.Ctx, expectedProtocol)
	s.App.StrdBurnerKeeper.SetTotalUserStrdBurned(s.Ctx, expectedUser)

	// Query for the total burned amount
	req := &types.QueryTotalStrdBurnedRequest{}
	resp, err := s.App.StrdBurnerKeeper.TotalStrdBurned(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying total strd burned")
	s.Require().Equal(expectedProtocol, resp.ProtocolBurned, "protocol burned amount")
	s.Require().Equal(expectedUser, resp.TotalUserBurned, "total user burned amount")
	s.Require().Equal(expectedTotal, resp.TotalBurned, "total burned amount")
}

func (s *KeeperTestSuite) TestQueryBurnedByAddress() {
	acc := s.TestAccs[0]

	// Set initial total burned amount
	expectedAmount := sdkmath.NewInt(1000000)
	s.App.StrdBurnerKeeper.SetStrdBurnedByAddress(s.Ctx, acc, expectedAmount)

	// Query for the total burned amount
	req := &types.QueryStrdBurnedByAddressRequest{
		Address: acc.String(),
	}
	resp, err := s.App.StrdBurnerKeeper.StrdBurnedByAddress(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying total strd burned")
	s.Require().Equal(expectedAmount, resp.BurnedAmount, "total burned amount")
}
