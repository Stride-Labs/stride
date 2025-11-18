package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v30/x/claim/keeper"
	"github.com/Stride-Labs/stride/v30/x/claim/types"
)

func (s *KeeperTestSuite) TestSetAirdropAllocationsForMultiAirdrops() {
	s.SetupTest()
	msgServer := keeper.NewMsgServerImpl(s.App.ClaimKeeper)

	// Set initial allocations for each airdrop
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	allocations := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           addr2.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(30, 2), // 30%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "juno",
		},
	}

	for _, record := range allocations {
		_, err := msgServer.SetAirdropAllocations(s.Ctx, &types.MsgSetAirdropAllocations{
			Allocator:         distributors[record.AirdropIdentifier].String(),
			AirdropIdentifier: record.AirdropIdentifier,
			Users:             []string{record.Address},
			Weights:           []sdkmath.LegacyDec{record.Weight},
		})
		s.Require().NoError(err)
	}

	// Set second allocations for each airdrop
	addr1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	allocations2 := []types.ClaimRecord{
		{
			Address:           addr1.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(40, 2), // 40%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           addr2.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(20, 2), // 20%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "juno",
		},
	}

	for _, record := range allocations2 {
		_, err := msgServer.SetAirdropAllocations(s.Ctx, &types.MsgSetAirdropAllocations{
			Allocator:         distributors[record.AirdropIdentifier].String(),
			AirdropIdentifier: record.AirdropIdentifier,
			Users:             []string{record.Address},
			Weights:           []sdkmath.LegacyDec{record.Weight},
		})
		s.Require().NoError(err)
	}

	totalWeightStride, err := s.App.ClaimKeeper.GetTotalWeight(s.Ctx, types.DefaultAirdropIdentifier)
	s.Require().NoError(err)
	s.Require().Equal(totalWeightStride, sdkmath.LegacyNewDecWithPrec(90, 2))

	totalWeightJuno, err := s.App.ClaimKeeper.GetTotalWeight(s.Ctx, "juno")
	s.Require().NoError(err)
	s.Require().Equal(totalWeightJuno, sdkmath.LegacyNewDecWithPrec(50, 2))

	claimRecords := s.App.ClaimKeeper.GetClaimRecords(s.Ctx, types.DefaultAirdropIdentifier)
	s.Require().Equal(2, len(claimRecords))

	claimRecords = s.App.ClaimKeeper.GetClaimRecords(s.Ctx, "juno")
	s.Require().Equal(2, len(claimRecords))

	// Multiple airdrop allocations for same user should be ignored
	addr2 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	allocations3 := []types.ClaimRecord{
		{
			Address:           addr1.String(),                      // duplicated airdrop address
			Weight:            sdkmath.LegacyNewDecWithPrec(40, 2), // 40%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: types.DefaultAirdropIdentifier,
		},
		{
			Address:           addr2.String(),
			Weight:            sdkmath.LegacyNewDecWithPrec(20, 2), // 20%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: "juno",
		},
	}
	for _, record := range allocations3 {
		_, err := msgServer.SetAirdropAllocations(s.Ctx, &types.MsgSetAirdropAllocations{
			Allocator:         distributors[record.AirdropIdentifier].String(),
			AirdropIdentifier: record.AirdropIdentifier,
			Users:             []string{record.Address},
			Weights:           []sdkmath.LegacyDec{record.Weight},
		})
		s.Require().NoError(err)
	}

	claimRecords = s.App.ClaimKeeper.GetClaimRecords(s.Ctx, types.DefaultAirdropIdentifier)
	s.Require().Equal(2, len(claimRecords))

	claimRecords = s.App.ClaimKeeper.GetClaimRecords(s.Ctx, "juno")
	s.Require().Equal(3, len(claimRecords))
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

func (s *KeeperTestSuite) TestCreateAirdrop_Successful() {
	s.SetupTest()
	msgServer := keeper.NewMsgServerImpl(s.App.ClaimKeeper)

	// Successfully create a new airdrop
	validMsg := getValidCreateEvmosAirdropMsg(s.Ctx)
	_, err := msgServer.CreateAirdrop(s.Ctx, &validMsg)
	s.Require().NoError(err, "no error expected when adding evmos airdrop")

	// Check that it matches the evmos airdrop
	airdrop := s.App.ClaimKeeper.GetAirdropByIdentifier(s.Ctx, "evmos")
	s.Require().Equal("evmos", airdrop.AirdropIdentifier, "airdrop identifier")
	s.Require().Equal("evmos-1", airdrop.ChainId, "airdrop chain-id")
	s.Require().Equal(true, airdrop.AutopilotEnabled, "airdrop autopilot enabled")
}

func (s *KeeperTestSuite) TestCreateAirdrop_IdentifierAlreadyExists() {
	s.SetupTest()
	msgServer := keeper.NewMsgServerImpl(s.App.ClaimKeeper)

	// Attempt to create an airdrop with an identifier that already exists, it should fail
	validMsg := getValidCreateEvmosAirdropMsg(s.Ctx)
	invalidMsg := validMsg
	invalidMsg.Identifier = types.DefaultAirdropIdentifier

	_, err := msgServer.CreateAirdrop(s.Ctx, &invalidMsg)
	s.Require().ErrorContains(err, "airdrop with same identifier already exists")
}

func (s *KeeperTestSuite) TestCreateAirdrop_ChainIdAlreadyExists() {
	s.SetupTest()
	msgServer := keeper.NewMsgServerImpl(s.App.ClaimKeeper)

	// Attempt to create an airdrop with a chain-id that already exists, it should fail
	validMsg := getValidCreateEvmosAirdropMsg(s.Ctx)
	invalidMsg := validMsg
	invalidMsg.ChainId = "stride-1"

	_, err := msgServer.CreateAirdrop(s.Ctx, &invalidMsg)
	s.Require().ErrorContains(err, "airdrop with same chain-id already exists")
}
