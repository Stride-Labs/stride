package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		HostZoneList: []types.HostZone{
			{
				ChainId:                "A",
				TotalDelegations:       sdkmath.OneInt(),
				RedemptionRate:         sdkmath.LegacyOneDec(),
				LastRedemptionRate:     sdkmath.LegacyOneDec(),
				MinRedemptionRate:      sdkmath.LegacyOneDec(),
				MaxRedemptionRate:      sdkmath.LegacyOneDec(),
				MinInnerRedemptionRate: sdkmath.LegacyOneDec(),
				MaxInnerRedemptionRate: sdkmath.LegacyOneDec(),
				Validators:             []*types.Validator{},
			},
		},
		EpochTrackerList: []types.EpochTracker{
			{EpochIdentifier: "stride_epoch"},
		},
		TradeRoutes: []types.TradeRoute{},
	}

	s.App.StakeibcKeeper.InitGenesis(s.Ctx, genesisState)
	exported := s.App.StakeibcKeeper.ExportGenesis(s.Ctx)

	s.Require().Equal(genesisState, *exported)
}
