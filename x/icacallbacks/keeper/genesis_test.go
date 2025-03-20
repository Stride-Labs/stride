package keeper_test

import (
	"github.com/Stride-Labs/stride/v26/testutil/nullify"
	"github.com/Stride-Labs/stride/v26/x/icacallbacks/types"
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

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	s.Require().Equal(genesisState.PortId, got.PortId)

	s.Require().ElementsMatch(genesisState.CallbackDataList, got.CallbackDataList)
}
