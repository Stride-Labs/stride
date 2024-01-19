package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Helper function to create the singleton HostZone with attributes
func (s *KeeperTestSuite) initializeHostZone() types.HostZone {
	hostZone := types.HostZone{
		ChainId:                "CELESTIA",
		NativeDenom:            "utia",
		TransferChannelId:      "channel-05",
		DelegationAddress:      "tia0384a",
		RewardAddress:          "tia144f42e9",
		DepositAddress:         "stride8abb3e",
		RedemptionAddress:      "stride3400de1",
		ClaimAddress:           "stride00b1a83",
		LastRedemptionRate:     sdk.MustNewDecFromStr("1.0"),
		RedemptionRate:         sdk.MustNewDecFromStr("1.0"),
		MinRedemptionRate:      sdk.MustNewDecFromStr("0.95"),
		MaxRedemptionRate:      sdk.MustNewDecFromStr("1.10"),
		MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.97"),
		MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.07"),
		DelegatedBalance:       sdk.NewInt(1_000_000),
		Halted:                 false,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)
	return hostZone
}

func (s *KeeperTestSuite) TestGetHostZone() {
	savedHostZone := s.initializeHostZone()
	loadedHostZone, err := s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "HostZone failed to load")
	s.Require().Equal(savedHostZone, loadedHostZone)
}

func (s *KeeperTestSuite) TestSetHostZone() {
	hostZone := s.initializeHostZone()

	hostZone.RedemptionRate = hostZone.RedemptionRate.Add(sdk.MustNewDecFromStr("0.1"))
	hostZone.DelegatedBalance = hostZone.DelegatedBalance.Add(sdk.NewInt(100_000))
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	loadedHostZone, err := s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "HostZone failed to load")
	s.Require().Equal(hostZone, loadedHostZone)
}
