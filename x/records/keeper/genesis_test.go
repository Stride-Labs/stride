package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/x/records/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		UserRedemptionRecordList: []types.UserRedemptionRecord{
			{
				Id:                "0",
				NativeTokenAmount: sdkmath.ZeroInt(),
				StTokenAmount:     sdkmath.ZeroInt(),
			},
			{
				Id:                "1",
				NativeTokenAmount: sdkmath.OneInt(),
				StTokenAmount:     sdkmath.OneInt(),
			},
		},
		UserRedemptionRecordCount: 2,
		EpochUnbondingRecordList: []types.EpochUnbondingRecord{
			{
				EpochNumber:        0,
				HostZoneUnbondings: []*types.HostZoneUnbonding{},
			},
			{
				EpochNumber:        1,
				HostZoneUnbondings: []*types.HostZoneUnbonding{},
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
		DepositRecordList: []types.DepositRecord{
			{
				Id:     0,
				Amount: sdkmath.ZeroInt(),
			},
			{
				Id:     1,
				Amount: sdkmath.OneInt(),
			},
		},
		DepositRecordCount: 2,
		LsmTokenDepositList: []types.LSMTokenDeposit{
			{
				DepositId: "ID1",
				ChainId:   "chain-1",
				Denom:     "denom1",
				Amount:    sdkmath.ZeroInt(),
				StToken:   sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.OneInt()),
			},
			{
				DepositId: "ID2",
				ChainId:   "chain-2",
				Denom:     "denom2",
				Amount:    sdkmath.OneInt(),
				StToken:   sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.OneInt()),
			},
		},
	}
	s.App.RecordsKeeper.InitGenesis(s.Ctx, genesisState)
	got := s.App.RecordsKeeper.ExportGenesis(s.Ctx)
	s.Require().NotNil(got)

	s.Require().Equal(genesisState.PortId, got.PortId)

	s.Require().ElementsMatch(genesisState.DepositRecordList, got.DepositRecordList)
	s.Require().Equal(genesisState.DepositRecordCount, got.DepositRecordCount)
	s.Require().ElementsMatch(genesisState.LsmTokenDepositList, got.LsmTokenDepositList)
}
