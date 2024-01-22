package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/keeper"
	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

func (s *KeeperTestSuite) TestAdjustDelegatedBalance() {
	// TODO [sttia]: verify this fails if issues by non-admin
	msgServer := keeper.NewMsgServerImpl(s.App.StaketiaKeeper)

	safeAddress := "SAFEADDR"

	// Create the host zone
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, types.HostZone{
		SafeAddress:      safeAddress,
		DelegatedBalance: sdk.NewInt(0),
	})

	// Call adjust for each test case and confirm the ending delegation
	testCases := []struct {
		address       string
		offset        sdkmath.Int
		endDelegation sdkmath.Int
	}{
		{address: "valA", offset: sdkmath.NewInt(10), endDelegation: sdkmath.NewInt(10)}, // 0 + 10 = 10
		{address: "valB", offset: sdkmath.NewInt(-5), endDelegation: sdkmath.NewInt(5)},  // 10 - 5 = 5
		{address: "valC", offset: sdkmath.NewInt(8), endDelegation: sdkmath.NewInt(13)},  // 5 + 8 = 13
		{address: "valD", offset: sdkmath.NewInt(2), endDelegation: sdkmath.NewInt(15)},  // 13 + 2 = 15
		{address: "valE", offset: sdkmath.NewInt(-6), endDelegation: sdkmath.NewInt(9)},  // 15 - 6 = 9
	}
	for _, tc := range testCases {
		msg := types.MsgAdjustDelegatedBalance{
			Operator:         safeAddress,
			DelegationOffset: tc.offset,
			ValidatorAddress: tc.address,
		}
		_, err := msgServer.AdjustDelegatedBalance(s.Ctx, &msg)
		s.Require().NoError(err, "no error expected when adjusting delegated bal properly for %s", tc.address)

		hostZone := s.MustGetHostZone()
		s.Require().Equal(tc.endDelegation, hostZone.DelegatedBalance, "delegation after change for %s", tc.address)
	}

	// Remove the host zone and try again, it should fail
	s.App.StaketiaKeeper.RemoveHostZone(s.Ctx)
	_, err := msgServer.AdjustDelegatedBalance(s.Ctx, &types.MsgAdjustDelegatedBalance{})
	s.Require().ErrorContains(err, "host zone not found")
}
