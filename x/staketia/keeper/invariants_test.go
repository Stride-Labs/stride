package keeper_test

import (
	"github.com/Stride-Labs/stride/v24/x/staketia/types"
)

func (s *KeeperTestSuite) TestHaltZone() {
	// Set a non-halted host zone
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, types.HostZone{
		NativeTokenDenom: HostNativeDenom,
		Halted:           false,
	})

	// Halt the zone
	s.App.StaketiaKeeper.HaltZone(s.Ctx)

	// Confirm it's halted
	hostZone := s.MustGetHostZone()
	s.Require().True(hostZone.Halted, "host zone should be halted")

	// Confirm denom is blacklisted
	isBlacklisted := s.App.RatelimitKeeper.IsDenomBlacklisted(s.Ctx, StDenom)
	s.Require().True(isBlacklisted, "halt zone should blacklist the stAsset denom")
}
