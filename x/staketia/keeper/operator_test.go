package keeper_test

import (
	"github.com/Stride-Labs/stride/v19/x/staketia/types"
)

// test that the admin address helpers work as expected
func (s *KeeperTestSuite) TestIsAddressHelpers() {

	operatorAddress := s.TestAccs[0].String()
	safeAddress := s.TestAccs[1].String()
	randomAddress := s.TestAccs[2].String()

	// Create a host zone with an operator and safe
	zone := types.HostZone{
		OperatorAddressOnStride: operatorAddress,
		SafeAddressOnStride:     safeAddress,
	}

	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)

	// Confirm that the operator is an OPERATOR admin
	s.Require().NoError(s.App.StaketiaKeeper.CheckIsOperatorAddress(s.Ctx, operatorAddress))
	// Confirm that a random address is not an OPERATOR admin
	s.Require().Error(s.App.StaketiaKeeper.CheckIsOperatorAddress(s.Ctx, randomAddress))

	// Confirm that the safe is a SAFE admin
	s.Require().NoError(s.App.StaketiaKeeper.CheckIsSafeAddress(s.Ctx, safeAddress))
	// Confirm that a random address is not a SAFE admin
	s.Require().Error(s.App.StaketiaKeeper.CheckIsSafeAddress(s.Ctx, randomAddress))

	// Test SafeOrOperator
	s.Require().NoError(s.App.StaketiaKeeper.CheckIsSafeOrOperatorAddress(s.Ctx, operatorAddress))
	s.Require().NoError(s.App.StaketiaKeeper.CheckIsSafeOrOperatorAddress(s.Ctx, safeAddress))
	s.Require().Error(s.App.StaketiaKeeper.CheckIsSafeOrOperatorAddress(s.Ctx, randomAddress))

}
