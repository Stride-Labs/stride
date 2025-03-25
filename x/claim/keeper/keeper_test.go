package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	"github.com/Stride-Labs/stride/v26/x/claim/types"
	minttypes "github.com/Stride-Labs/stride/v26/x/mint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	apptesting.AppTestHelper
}

var distributors map[string]sdk.AccAddress

func (s *KeeperTestSuite) SetupTest() {
	distributors = make(map[string]sdk.AccAddress)

	// Initiate a distributor account for stride user airdrop
	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	s.SetNewAccount(addr1)
	distributors[types.DefaultAirdropIdentifier] = addr1

	// Initiate a distributor account for juno user airdrop
	pub2 := secp256k1.GenPrivKey().PubKey()
	addr2 := sdk.AccAddress(pub2.Address())
	s.SetNewAccount(addr2)
	distributors["juno"] = addr2

	// Initiate a distributor account for juno user airdrop
	pub3 := secp256k1.GenPrivKey().PubKey()
	addr3 := sdk.AccAddress(pub3.Address())
	s.SetNewAccount(addr3)
	distributors["osmosis"] = addr3

	// Initiate a distributor account for evmos user airdrop
	pub4 := secp256k1.GenPrivKey().PubKey()
	addr4 := sdk.AccAddress(pub4.Address())
	s.SetNewAccount(addr4)
	distributors["evmos"] = addr4

	// Mint coins to airdrop module
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(300000000))))
	if err != nil {
		panic(err)
	}
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, addr1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000))))
	if err != nil {
		panic(err)
	}
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, addr2, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000))))
	if err != nil {
		panic(err)
	}
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, addr3, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000))))
	if err != nil {
		panic(err)
	}

	// Stride airdrop
	airdropStartTime := time.Now()
	err = s.App.ClaimKeeper.CreateAirdropAndEpoch(s.Ctx, types.MsgCreateAirdrop{
		Distributor:      addr1.String(),
		Identifier:       types.DefaultAirdropIdentifier,
		ChainId:          "stride-1",
		Denom:            sdk.DefaultBondDenom,
		StartTime:        uint64(airdropStartTime.Unix()),
		Duration:         uint64(types.DefaultAirdropDuration.Seconds()),
		AutopilotEnabled: false,
	})
	if err != nil {
		panic(err)
	}

	// Juno airdrop
	err = s.App.ClaimKeeper.CreateAirdropAndEpoch(s.Ctx, types.MsgCreateAirdrop{
		Distributor: addr2.String(),
		Identifier:  "juno",
		ChainId:     "juno-1",
		Denom:       sdk.DefaultBondDenom,
		StartTime:   uint64(airdropStartTime.Add(time.Hour).Unix()),
		Duration:    uint64(types.DefaultAirdropDuration.Seconds()),
	})
	if err != nil {
		panic(err)
	}

	// Osmosis airdrop
	err = s.App.ClaimKeeper.CreateAirdropAndEpoch(s.Ctx, types.MsgCreateAirdrop{
		Distributor: addr3.String(),
		Identifier:  "osmosis",
		ChainId:     "osmosis-1",
		Denom:       sdk.DefaultBondDenom,
		StartTime:   uint64(airdropStartTime.Unix()),
		Duration:    uint64(types.DefaultAirdropDuration.Seconds()),
	})
	if err != nil {
		panic(err)
	}

	s.Ctx = s.Ctx.WithBlockTime(airdropStartTime)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
