package v35_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/Stride-Labs/stride/v35/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v35/app/upgrades/v35"
	"github.com/Stride-Labs/stride/v35/utils"
)

// A denom the test app never mints on its own, so balances in it come only from the test
const communityPoolIbcDenom = "ibc/community-pool-test"

// fundCommunityPool deposits the coins into the community pool from a fresh account
func (s *UpgradeTestSuite) fundCommunityPool(coins sdk.Coins) {
	depositor := apptesting.CreateRandomAccounts(1)[0]
	for _, coin := range coins {
		s.FundAccount(depositor, coin)
	}
	s.Require().NoError(s.App.DistrKeeper.FundCommunityPool(s.Ctx, coins, depositor))
}

func (s *UpgradeTestSuite) communityPool() sdk.DecCoins {
	feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)
	return feePool.CommunityPool
}

func (s *UpgradeTestSuite) authorityBalances() sdk.Coins {
	return s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(v35.UpgradeAuthority))
}

func (s *UpgradeTestSuite) TestTransferCommunityPoolToAuthority() {
	funded := sdk.NewCoins(
		sdk.NewInt64Coin(utils.BaseStrideDenom, 5_000_000),
		sdk.NewInt64Coin(communityPoolIbcDenom, 700),
	)
	s.fundCommunityPool(funded)

	// The pool is tracked in decimals; give each denom a sub-unit remainder that cannot be sent
	dust := sdk.NewDecCoins(
		sdk.NewDecCoinFromDec(utils.BaseStrideDenom, sdkmath.LegacyMustNewDecFromStr("0.4")),
		sdk.NewDecCoinFromDec(communityPoolIbcDenom, sdkmath.LegacyMustNewDecFromStr("0.9")),
	)
	feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)
	feePool.CommunityPool = feePool.CommunityPool.Add(dust...)
	s.Require().NoError(s.App.DistrKeeper.FeePool.Set(s.Ctx, feePool))

	distrModuleAddress := s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	distrBalancesBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, distrModuleAddress)
	s.Require().True(s.authorityBalances().IsZero(), "multisig starts empty")

	v35.TransferCommunityPoolToAuthority(s.Ctx, s.App.DistrKeeper)

	s.Require().Equal(funded, s.authorityBalances(), "multisig holds every whole coin")
	s.Require().Equal(dust, s.communityPool(), "only the sub-unit remainder is left")
	s.Require().Equal(distrBalancesBefore.Sub(funded...), s.App.BankKeeper.GetAllBalances(s.Ctx, distrModuleAddress),
		"the coins left the distribution module account")

	// A second run (localnet re-run) finds no whole coins and moves nothing
	v35.TransferCommunityPoolToAuthority(s.Ctx, s.App.DistrKeeper)
	s.Require().Equal(funded, s.authorityBalances(), "idempotent")
	s.Require().Equal(dust, s.communityPool(), "idempotent")
}

func (s *UpgradeTestSuite) TestTransferCommunityPoolToAuthority_EmptyPool() {
	s.Require().True(s.communityPool().IsZero(), "pool starts empty")

	v35.TransferCommunityPoolToAuthority(s.Ctx, s.App.DistrKeeper)

	s.Require().True(s.authorityBalances().IsZero(), "nothing sent")
	s.Require().True(s.communityPool().IsZero(), "pool untouched")
}

// A denom whose pool balance claims more than the distribution module account holds cannot be
// paid out: it is skipped with its pool balance untouched, and the other denoms still transfer
func (s *UpgradeTestSuite) TestTransferCommunityPoolToAuthority_FailedDenomIsSkipped() {
	strd := sdk.NewInt64Coin(utils.BaseStrideDenom, 5_000_000)
	s.fundCommunityPool(sdk.NewCoins(strd, sdk.NewInt64Coin(communityPoolIbcDenom, 700)))

	feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewInt64DecCoin(communityPoolIbcDenom, 1))
	s.Require().NoError(s.App.DistrKeeper.FeePool.Set(s.Ctx, feePool))

	v35.TransferCommunityPoolToAuthority(s.Ctx, s.App.DistrKeeper)

	s.Require().Equal(sdk.NewCoins(strd), s.authorityBalances(), "only the payable denom is sent")
	s.Require().Equal(sdk.NewDecCoins(sdk.NewInt64DecCoin(communityPoolIbcDenom, 701)), s.communityPool(),
		"the unpayable denom stays in the pool in full")
}
