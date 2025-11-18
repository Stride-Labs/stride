package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v30/x/claim/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	pub2 := secp256k1.GenPrivKey().PubKey()
	addr2 := sdk.AccAddress(pub2.Address())

	pub3 := secp256k1.GenPrivKey().PubKey()
	addr3 := sdk.AccAddress(pub3.Address())

	genesisState := types.GenesisState{
		Params: types.Params{
			Airdrops: []*types.Airdrop{
				{
					AirdropIdentifier:  types.DefaultAirdropIdentifier,
					AirdropStartTime:   s.Ctx.BlockTime(),
					AirdropDuration:    types.DefaultAirdropDuration,
					ClaimDenom:         sdk.DefaultBondDenom,
					DistributorAddress: addr3.String(),
					ClaimedSoFar:       sdkmath.ZeroInt(),
				},
			},
		},
		ClaimRecords: []types.ClaimRecord{
			{
				Address:           addr1.String(),
				Weight:            sdkmath.LegacyNewDecWithPrec(50, 2), // 50%
				ActionCompleted:   []bool{false, false, false},
				AirdropIdentifier: types.DefaultAirdropIdentifier,
			},
			{
				Address:           addr2.String(),
				Weight:            sdkmath.LegacyNewDecWithPrec(50, 2), // 50%
				ActionCompleted:   []bool{false, false, false},
				AirdropIdentifier: "juno",
			},
		},
	}

	s.App.ClaimKeeper.InitGenesis(s.Ctx, genesisState)
	got := s.App.ClaimKeeper.ExportGenesis(s.Ctx)
	s.Require().NotNil(got)

	totalWeightStride, err := s.App.ClaimKeeper.GetTotalWeight(s.Ctx, types.DefaultAirdropIdentifier)
	s.Require().NoError(err)
	s.Require().Equal(totalWeightStride, genesisState.ClaimRecords[0].Weight)

	totalWeightJuno, err := s.App.ClaimKeeper.GetTotalWeight(s.Ctx, types.DefaultAirdropIdentifier)
	s.Require().NoError(err)
	s.Require().Equal(totalWeightJuno, genesisState.ClaimRecords[1].Weight)

	s.Require().Equal(genesisState.Params, got.Params)
	s.Require().ElementsMatch(genesisState.ClaimRecords, got.ClaimRecords)
}
