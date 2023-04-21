package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/x/claim/keeper"
	"github.com/Stride-Labs/stride/v9/x/claim/types"
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
		_, err := msgServer.SetAirdropAllocations(sdk.WrapSDKContext(suite.ctx), &types.MsgSetAirdropAllocations{
			Allocator:         distributors[record.AirdropIdentifier].String(),
			AirdropIdentifier: record.AirdropIdentifier,
			Users:             []string{record.Address},
			Weights:           []sdk.Dec{record.Weight},
		})
		suite.Require().NoError(err)
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
		_, err := msgServer.SetAirdropAllocations(sdk.WrapSDKContext(suite.ctx), &types.MsgSetAirdropAllocations{
			Allocator:         distributors[record.AirdropIdentifier].String(),
			AirdropIdentifier: record.AirdropIdentifier,
			Users:             []string{record.Address},
			Weights:           []sdk.Dec{record.Weight},
		})
		suite.Require().NoError(err)
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
		_, err := msgServer.SetAirdropAllocations(sdk.WrapSDKContext(suite.ctx), &types.MsgSetAirdropAllocations{
			Allocator:         distributors[record.AirdropIdentifier].String(),
			AirdropIdentifier: record.AirdropIdentifier,
			Users:             []string{record.Address},
			Weights:           []sdk.Dec{record.Weight},
		})
		suite.Require().NoError(err)
	}

	claimRecords = suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, types.DefaultAirdropIdentifier)
	suite.Require().Equal(2, len(claimRecords))

	claimRecords = suite.app.ClaimKeeper.GetClaimRecords(suite.ctx, "juno")
	suite.Require().Equal(3, len(claimRecords))
}

func getValidCreateEvmosAirdropMsg(ctx sdk.Context) types.MsgCreateAirdrop {
	return types.MsgCreateAirdrop{
		Distributor:      distributors["evmos"].String(),
		Identifier:       "evmos",
		ChainId:          "evmos-1",
		Denom:            sdk.DefaultBondDenom,
		StartTime:        uint64(ctx.BlockTime().Unix()),
		Duration:         uint64(types.DefaultAirdropDuration.Seconds()),
		AutopilotEnabled: true,
	}
}

func (suite *KeeperTestSuite) TestCreateAirdrop_Successful() {
	suite.SetupTest()
	msgServer := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)

	// Successfully create a new airdrop
	validMsg := getValidCreateEvmosAirdropMsg(suite.ctx)
	_, err := msgServer.CreateAirdrop(sdk.WrapSDKContext(suite.ctx), &validMsg)
	suite.Require().NoError(err, "no error expected when adding evmos airdrop")

	// Check that it matches the evmos airdrop
	airdrop := suite.app.ClaimKeeper.GetAirdropByIdentifier(suite.ctx, "evmos")
	suite.Require().Equal("evmos", airdrop.AirdropIdentifier, "airdrop identifier")
	suite.Require().Equal("evmos-1", airdrop.ChainId, "airdrop chain-id")
	suite.Require().Equal(true, airdrop.AutopilotEnabled, "airdrop autopilot enabled")
}

func (suite *KeeperTestSuite) TestCreateAirdrop_IdentifierAlreadyExists() {
	suite.SetupTest()
	msgServer := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)

	// Attempt to create an airdrop with an identifier that already exists, it should fail
	validMsg := getValidCreateEvmosAirdropMsg(suite.ctx)
	invalidMsg := validMsg
	invalidMsg.Identifier = types.DefaultAirdropIdentifier

	_, err := msgServer.CreateAirdrop(sdk.WrapSDKContext(suite.ctx), &invalidMsg)
	suite.Require().ErrorContains(err, "airdrop with same identifier already exists")
}

func (suite *KeeperTestSuite) TestCreateAirdrop_ChainIdAlreadyExists() {
	suite.SetupTest()
	msgServer := keeper.NewMsgServerImpl(suite.app.ClaimKeeper)

	// Attempt to create an airdrop with a chain-id that already exists, it should fail
	validMsg := getValidCreateEvmosAirdropMsg(suite.ctx)
	invalidMsg := validMsg
	invalidMsg.ChainId = "stride-1"

	_, err := msgServer.CreateAirdrop(sdk.WrapSDKContext(suite.ctx), &invalidMsg)
	suite.Require().ErrorContains(err, "airdrop with same chain-id already exists")
}
