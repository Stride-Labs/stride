package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/claim/keeper"
	"github.com/Stride-Labs/stride/x/claim/types"
)

func (suite *KeeperTestSuite) TestSetAirdropAllocationsForMultiAirdrops() {
	suite.SetupTest()
	msgServer := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)

	// Set initial allocations for each airdrop
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	allocations := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           addr2.String(),
			Weight:            sdk.NewDecWithPrec(30, 2), // 30%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "juno",
		},
	}

	for _, record := range allocations {
		msgServer.SetAirdropAllocations(sdk.WrapSDKContext(suite.ctx), &types.MsgSetAirdropAllocations{
			Allocator:         distributors[record.AirdropIdentifier].String(),
			AirdropIdentifier: record.AirdropIdentifier,
			Users:             []string{record.Address},
			Weights:           []sdk.Dec{record.Weight},
		})
	}

	// Set second allocations for each airdrop
	addr1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	allocations2 := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            sdk.NewDecWithPrec(40, 2), // 40%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           addr2.String(),
			Weight:            sdk.NewDecWithPrec(20, 2), // 20%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "juno",
		},
	}

	for _, record := range allocations2 {
		msgServer.SetAirdropAllocations(sdk.WrapSDKContext(suite.ctx), &types.MsgSetAirdropAllocations{
			Allocator:         distributors[record.AirdropIdentifier].String(),
			AirdropIdentifier: record.AirdropIdentifier,
			Users:             []string{record.Address},
			Weights:           []sdk.Dec{record.Weight},
		})
	}

	totalWeightStride, err := suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().NoError(err)
	suite.Require().Equal(totalWeightStride, sdk.NewDecWithPrec(90, 2))

	totalWeightJuno, err := suite.app.ClaimKeeper.GetTotalWeight(suite.ctx, "juno")
	suite.Require().NoError(err)
	suite.Require().Equal(totalWeightJuno, sdk.NewDecWithPrec(50, 2))

	claimRecords := suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().Equal(2, len(claimRecords))

	claimRecords = suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, "juno")
	suite.Require().Equal(2, len(claimRecords))

	// Multiple airdrop allocations for same user should be ignored
	addr2 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	allocations3 := []types.ClaimRecord{
		{
			Address:           addr1.String(),            // duplicated airdrop address
			Weight:            sdk.NewDecWithPrec(40, 2), // 40%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           addr2.String(),
			Weight:            sdk.NewDecWithPrec(20, 2), // 20%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "juno",
		},
	}
	for _, record := range allocations3 {
		msgServer.SetAirdropAllocations(sdk.WrapSDKContext(suite.ctx), &types.MsgSetAirdropAllocations{
			Allocator:         distributors[record.AirdropIdentifier].String(),
			AirdropIdentifier: record.AirdropIdentifier,
			Users:             []string{record.Address},
			Weights:           []sdk.Dec{record.Weight},
		})
	}

	claimRecords = suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().Equal(2, len(claimRecords))

	claimRecords = suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, "juno")
	suite.Require().Equal(3, len(claimRecords))
}
