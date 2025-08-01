package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/staketia/types"
)

// Helper function to create the singleton HostZone with attributes
func (s *KeeperTestSuite) initializeHostZone() types.HostZone {
	hostZone := types.HostZone{
		ChainId:                   "CELESTIA",
		NativeTokenDenom:          "utia",
		NativeTokenIbcDenom:       "ibc/utia",
		TransferChannelId:         "channel-05",
		DelegationAddress:         "tia0384a",
		RewardAddress:             "tia144f42e9",
		DepositAddress:            "stride8abb3e",
		RedemptionAddress:         "stride3400de1",
		ClaimAddress:              "stride00b1a83",
		RemainingDelegatedBalance: sdk.NewInt(1_000_000),
		Halted:                    false,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)
	return hostZone
}

func (s *KeeperTestSuite) TestGetHostZone() {
	savedHostZone := s.initializeHostZone()
	loadedHostZone := s.MustGetHostZone()
	s.Require().Equal(savedHostZone, loadedHostZone)
}

func (s *KeeperTestSuite) TestRemoveHostZone() {
	s.initializeHostZone()
	s.App.StaketiaKeeper.RemoveHostZone(s.Ctx)
	_, err := s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().ErrorContains(err, "host zone not found")
}

func (s *KeeperTestSuite) TestSetHostZone() {
	hostZone := s.initializeHostZone()

	hostZone.RemainingDelegatedBalance = hostZone.RemainingDelegatedBalance.Add(sdk.NewInt(100_000))
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	loadedHostZone := s.MustGetHostZone()
	s.Require().Equal(hostZone, loadedHostZone)
}

func (s *KeeperTestSuite) TestGetUnhaltedHostZone() {
	initialHostZone := types.HostZone{
		ChainId: "chain-0",
	}

	// Attempt to get a host zone when one has not been created yet - it should error
	_, err := s.App.StaketiaKeeper.GetUnhaltedHostZone(s.Ctx)
	s.Require().ErrorContains(err, "host zone not found")

	// Set a non-halted zone
	initialHostZone.Halted = false
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, initialHostZone)

	// Confirm there's no error when fetching it
	actualHostZone, err := s.App.StaketiaKeeper.GetUnhaltedHostZone(s.Ctx)
	s.Require().NoError(err, "no error expected when host zone is active")
	s.Require().Equal(initialHostZone.ChainId, actualHostZone.ChainId, "chain-id")

	// Set a halted zone
	initialHostZone.Halted = true
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, initialHostZone)

	// Confirm there's a halt error
	_, err = s.App.StaketiaKeeper.GetUnhaltedHostZone(s.Ctx)
	s.Require().ErrorContains(err, "host zone is halted")
}
