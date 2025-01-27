package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/strdburner/types"
)

func (s *KeeperTestSuite) TestQueryStrdBurnerAddress() {
	// Query for the strd burner address
	req := &types.QueryStrdBurnerAddressRequest{}
	resp, err := s.App.StrdBurnerKeeper.StrdBurnerAddress(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying strd burner address")
	s.Require().Equal(s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(), resp.Address, "address")
}

func (s *KeeperTestSuite) TestQueryTotalStrdBurned() {
	// Set initial total burned amount
	expectedAmount := sdkmath.NewInt(1000000)
	s.App.StrdBurnerKeeper.SetTotalStrdBurned(s.Ctx, expectedAmount)

	// Query for the total burned amount
	req := &types.QueryTotalStrdBurnedRequest{}
	resp, err := s.App.StrdBurnerKeeper.TotalStrdBurned(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying total strd burned")
	s.Require().Equal(expectedAmount, resp.TotalBurned, "total burned amount")
}
