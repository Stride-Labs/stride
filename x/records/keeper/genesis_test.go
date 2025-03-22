package keeper_test

import (
	"github.com/Stride-Labs/stride/v26/testutil/nullify"
	"github.com/Stride-Labs/stride/v26/x/records/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		UserRedemptionRecordList: []types.UserRedemptionRecord{
			{
				Id: "0",
			},
			{
				Id: "1",
			},
		},
		UserRedemptionRecordCount: 2,
		EpochUnbondingRecordList: []types.EpochUnbondingRecord{
			{
				EpochNumber: 0,
			},
			{
				EpochNumber: 1,
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
		DepositRecordList: []types.DepositRecord{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		DepositRecordCount: 2,
		LsmTokenDepositList: []types.LSMTokenDeposit{
			{
				DepositId: "ID1",
				ChainId:   "chain-1",
				Denom:     "denom1",
			},
			{
				DepositId: "ID2",
				ChainId:   "chain-2",
				Denom:     "denom2",
			},
		},
	}
	s.App.RecordsKeeper.InitGenesis(s.Ctx, genesisState)
	got := s.App.RecordsKeeper.ExportGenesis(s.Ctx)
	s.Require().NotNil(got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	s.Require().Equal(genesisState.PortId, got.PortId)

	s.Require().ElementsMatch(genesisState.DepositRecordList, got.DepositRecordList)
	s.Require().Equal(genesisState.DepositRecordCount, got.DepositRecordCount)
	s.Require().ElementsMatch(genesisState.LsmTokenDepositList, got.LsmTokenDepositList)
}
