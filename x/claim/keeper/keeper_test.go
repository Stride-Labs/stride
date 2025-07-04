package keeper_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v27/app"
	"github.com/Stride-Labs/stride/v27/x/claim/types"
	minttypes "github.com/Stride-Labs/stride/v27/x/mint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	// querier sdk.Querier
	app *app.StrideApp
}

var distributors map[string]sdk.AccAddress

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.InitStrideTestApp(true)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})
	distributors = make(map[string]sdk.AccAddress)

	// Initiate a distributor account for stride user airdrop
	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
	distributors[types.DefaultAirdropIdentifier] = addr1

	// Initiate a distributor account for juno user airdrop
	pub2 := secp256k1.GenPrivKey().PubKey()
	addr2 := sdk.AccAddress(pub2.Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr2, nil, 0, 0))
	distributors["juno"] = addr2

	// Initiate a distributor account for juno user airdrop
	pub3 := secp256k1.GenPrivKey().PubKey()
	addr3 := sdk.AccAddress(pub3.Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr3, nil, 0, 0))
	distributors["osmosis"] = addr3

	// Initiate a distributor account for evmos user airdrop
	pub4 := secp256k1.GenPrivKey().PubKey()
	addr4 := sdk.AccAddress(pub4.Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr4, nil, 0, 0))
	distributors["evmos"] = addr4

	// Mint coins to airdrop module
	err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(300000000))))
	if err != nil {
		panic(err)
	}
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000))))
	if err != nil {
		panic(err)
	}
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr2, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000))))
	if err != nil {
		panic(err)
	}
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr3, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000))))
	if err != nil {
		panic(err)
	}

	// Stride airdrop
	airdropStartTime := time.Now()
	err = suite.app.ClaimKeeper.CreateAirdropAndEpoch(suite.ctx, types.MsgCreateAirdrop{
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
	err = suite.app.ClaimKeeper.CreateAirdropAndEpoch(suite.ctx, types.MsgCreateAirdrop{
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
	err = suite.app.ClaimKeeper.CreateAirdropAndEpoch(suite.ctx, types.MsgCreateAirdrop{
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

	suite.ctx = suite.ctx.WithBlockTime(airdropStartTime)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
