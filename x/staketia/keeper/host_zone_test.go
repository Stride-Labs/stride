package keeper_test

import (
	sdk_types "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Helper function to create the singleton HostZone with attributes
func (s *KeeperTestSuite) initializeHostZone() types.HostZone {
	hostZone := types.HostZone{
		ChainId: "CELESTIA",
		NativeDenom: "utia",
		TransferChannelId: "channel-05",
		DelegationAddress: "tia0384a",
		RewardAddress: "tia144f42e9",
		DepositAddress: "stride8abb3e",
		RedemptionAddress: "stride3400de1",
		ClaimAddress: "stride00b1a83",
		LastRedemptionRate: sdk_types.MustNewDecFromStr("1.0"),
		RedemptionRate: sdk_types.MustNewDecFromStr("1.0"),
		MinRedemptionRate: sdk_types.MustNewDecFromStr("0.95"),
		MaxRedemptionRate: sdk_types.MustNewDecFromStr("1.10"),
		MinInnerRedemptionRate: sdk_types.MustNewDecFromStr("0.97"),
		MaxInnerRedemptionRate: sdk_types.MustNewDecFromStr("1.07"),
		DelegatedBalance: sdk_types.NewInt(1_000_000),
		Halted: false,
	}
	s.App.StakeTiaKeeper.SetHostZone(s.Ctx, hostZone)
	return hostZone
}

func (s *KeeperTestSuite) TestGetHostZone() {
	savedHostZone := s.initializeHostZone()
	loadedHostZone, err := s.App.StakeTiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "HostZone failed to load")
	s.Require().Equal(savedHostZone, loadedHostZone)
}

