package v36_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Stride-Labs/stride/v35/app/apptesting"
	v36 "github.com/Stride-Labs/stride/v35/app/upgrades/v36"
	"github.com/Stride-Labs/stride/v35/utils"
)

// Denoms the test app never mints on its own, so balances in them come only from the test
const (
	treasuryIbcDenom    = "ibc/treasury-test"
	treasuryStCmdxDenom = "stucmdx"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) balances(address string) sdk.Coins {
	return s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(address))
}

// fundTreasuries sets each treasury up like mainnet: a permissionless module account named by its own address,
// holding several denoms (one an 18-decimal amount), and returns what each holds
func (s *UpgradeTestSuite) fundTreasuries() map[string]sdk.Coins {
	bigIbcAmount, ok := sdkmath.NewIntFromString("1197383382313267605")
	s.Require().True(ok)
	funded := map[string]sdk.Coins{
		v36.CommunityPoolGrowthAddress: sdk.NewCoins(
			sdk.NewInt64Coin(utils.BaseStrideDenom, 2_866_671_132_700),
			sdk.NewCoin(treasuryIbcDenom, bigIbcAmount),
			sdk.NewInt64Coin(treasuryStCmdxDenom, 4_568_085_512),
		),
		v36.IncentiveProgramAddress: sdk.NewCoins(
			sdk.NewInt64Coin(utils.BaseStrideDenom, 3_000_000_000_000),
			sdk.NewInt64Coin(treasuryStCmdxDenom, 10_000),
		),
		v36.CommunityPoolSecurityAddress: sdk.NewCoins(sdk.NewInt64Coin(utils.BaseStrideDenom, 174_932_454_147)),
	}
	for _, address := range v36.TreasuryAddresses {
		accAddress := sdk.MustAccAddressFromBech32(address)
		baseAccount := authtypes.NewBaseAccountWithAddress(accAddress)
		moduleAccount := authtypes.NewModuleAccount(baseAccount, address)
		s.App.AccountKeeper.SetAccount(s.Ctx, s.App.AccountKeeper.NewAccount(s.Ctx, moduleAccount))
		for _, coin := range funded[address] {
			s.FundAccount(accAddress, coin)
		}
	}
	return funded
}

func (s *UpgradeTestSuite) TestUpgrade() {
	funded := s.fundTreasuries()
	authorityBefore := s.balances(v36.UpgradeAuthority)

	s.ConfirmUpgradeSucceeded(v36.UpgradeName)

	expected := authorityBefore
	for _, address := range v36.TreasuryAddresses {
		s.Require().True(s.balances(address).IsZero(), "%s is emptied", address)
		expected = expected.Add(funded[address]...)
	}
	s.Require().Equal(expected, s.balances(v36.UpgradeAuthority), "the multisig holds every treasury coin")
}

func (s *UpgradeTestSuite) TestTransferTreasuryToAuthority_EmptyTreasuries() {
	authorityBefore := s.balances(v36.UpgradeAuthority)

	v36.TransferTreasuryToAuthority(s.Ctx, s.App.BankKeeper)

	s.Require().Equal(authorityBefore, s.balances(v36.UpgradeAuthority), "nothing to move, nothing moved")
}
