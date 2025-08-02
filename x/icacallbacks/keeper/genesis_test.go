package keeper_test

import (
	"github.com/Stride-Labs/stride/v28/x/icacallbacks/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		CallbackDataList: []types.CallbackData{
			{
				CallbackKey: "0",
			},
			{
				CallbackKey: "1",
			},
		},
	}

	s.App.IcacallbacksKeeper.InitGenesis(s.Ctx, genesisState)
	got := s.App.IcacallbacksKeeper.ExportGenesis(s.Ctx)
	s.Require().NotNil(got)

	s.Require().Equal(genesisState.PortId, got.PortId)

	s.Require().ElementsMatch(genesisState.CallbackDataList, got.CallbackDataList)
}
