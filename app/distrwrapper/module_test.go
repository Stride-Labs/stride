package distrwrapper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v32/app/apptesting"
	"github.com/Stride-Labs/stride/v32/app/distrwrapper"
)

type DistrWrapperTestSuite struct {
	apptesting.AppTestHelper
}

func (s *DistrWrapperTestSuite) SetupTest() {
	s.Setup()
	// Tests below assume blocks past 1 so the height-guard doesn't short-circuit.
	s.Ctx = s.Ctx.WithBlockHeight(2)
}

func TestDistrWrapperTestSuite(t *testing.T) {
	suite.Run(t, new(DistrWrapperTestSuite))
}

// fundFeeCollector mints coins and parks them in fee_collector, simulating
// what mint.DistributeMintedCoin would do for the staking-incentives slice.
func (s *DistrWrapperTestSuite) fundFeeCollector(amount sdkmath.Int) sdk.Coin {
	denom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	coin := sdk.NewCoin(denom, amount)
	s.FundModuleAccount(authtypes.FeeCollectorName, coin)
	return coin
}

// callBeginBlock constructs a fresh distrwrapper.AppModule from the test app's
// keepers and invokes BeginBlock on it. This is what the module manager would
// do every block in a real run.
func (s *DistrWrapperTestSuite) callBeginBlock() {
	mod := distrwrapper.NewAppModule(
		s.App.AppCodec(),
		s.App.DistrKeeper,
		s.App.AccountKeeper,
		s.App.BankKeeper,
		s.App.StakingKeeper,
		s.App.GetSubspace("distribution"),
	)
	s.Require().NoError(mod.BeginBlock(s.Ctx))
}

// poaBalance returns the POA module account's balance in the bond denom.
func (s *DistrWrapperTestSuite) poaBalance() sdkmath.Int {
	denom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	addr := s.App.AccountKeeper.GetModuleAddress(poatypes.ModuleName)
	return s.App.BankKeeper.GetBalance(s.Ctx, addr, denom).Amount
}

// feeCollectorBalance returns the fee_collector module account's balance in the bond denom.
func (s *DistrWrapperTestSuite) feeCollectorBalance() sdkmath.Int {
	denom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	addr := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	return s.App.BankKeeper.GetBalance(s.Ctx, addr, denom).Amount
}

func (s *DistrWrapperTestSuite) TestBeginBlock_HeightOne_NoOp() {
	s.Ctx = s.Ctx.WithBlockHeight(1)
	feeAmount := sdkmath.NewInt(1_000_000)
	s.fundFeeCollector(feeAmount)

	s.callBeginBlock()

	s.Require().Equal(feeAmount, s.feeCollectorBalance(),
		"fee_collector should be untouched at height 1")
	s.Require().True(s.poaBalance().IsZero(),
		"POA should not receive anything at height 1")
}

func (s *DistrWrapperTestSuite) TestBeginBlock_EmptyFeeCollector_NoOp() {
	s.callBeginBlock()
	s.Require().True(s.feeCollectorBalance().IsZero())
	s.Require().True(s.poaBalance().IsZero())
}

func (s *DistrWrapperTestSuite) TestBeginBlock_RoutesPOAShare() {
	feeAmount := sdkmath.NewInt(1_000_000)
	s.fundFeeCollector(feeAmount)

	s.callBeginBlock()

	// 15% of 1_000_000 = 150_000
	expectedPOA := sdkmath.NewInt(150_000)
	s.Require().Equal(expectedPOA, s.poaBalance(),
		"POA should receive 15%% of the fee_collector balance")
	s.Require().True(s.feeCollectorBalance().IsZero(),
		"fee_collector should be fully drained")
}

func (s *DistrWrapperTestSuite) TestBeginBlock_AllocatesStakingShareToBondedValidator() {
	// The test app's GenesisStateWithConsumerValSet seeds one Bonded staking
	// validator. After a BeginBlock, that validator should have an outstanding
	// rewards entry for ~85% of the funded amount, minus community tax.
	feeAmount := sdkmath.NewInt(1_000_000)
	s.fundFeeCollector(feeAmount)

	bondedValidators, err := s.App.StakingKeeper.GetBondedValidatorsByPower(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(bondedValidators, "test app should have at least one bonded validator")

	s.callBeginBlock()

	// Sum outstanding rewards across all bonded validators — should equal the
	// staking share (~850_000) minus the community tax slice.
	denom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	totalRewards := sdkmath.LegacyZeroDec()
	for _, v := range bondedValidators {
		valAddr, err := sdk.ValAddressFromBech32(v.OperatorAddress)
		s.Require().NoError(err)
		out, err := s.App.DistrKeeper.GetValidatorOutstandingRewards(s.Ctx, valAddr)
		s.Require().NoError(err)
		totalRewards = totalRewards.Add(out.Rewards.AmountOf(denom))
	}

	communityTax, err := s.App.DistrKeeper.GetCommunityTax(s.Ctx)
	s.Require().NoError(err)
	expectedRewards := sdkmath.LegacyNewDec(850_000).
		Mul(sdkmath.LegacyOneDec().Sub(communityTax))

	// Allow ±1 unit drift for truncation.
	diff := totalRewards.Sub(expectedRewards).Abs()
	s.Require().True(diff.LT(sdkmath.LegacyNewDec(2)),
		"validator outstanding rewards should be ≈%s; got %s (diff %s)",
		expectedRewards, totalRewards, diff)
}

func (s *DistrWrapperTestSuite) TestBeginBlock_CommunityPoolGetsTaxAndRemainder() {
	feeAmount := sdkmath.NewInt(1_000_000)
	s.fundFeeCollector(feeAmount)

	feePoolBefore, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)

	s.callBeginBlock()

	feePoolAfter, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)

	denom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	delta := feePoolAfter.CommunityPool.AmountOf(denom).
		Sub(feePoolBefore.CommunityPool.AmountOf(denom))

	communityTax, err := s.App.DistrKeeper.GetCommunityTax(s.Ctx)
	s.Require().NoError(err)
	expectedFromTax := sdkmath.LegacyNewDec(850_000).Mul(communityTax)

	// Community pool gets at least the tax slice. May also accrue truncation
	// remainders (a few units).
	s.Require().True(delta.GTE(expectedFromTax),
		"community pool delta %s should be ≥ tax slice %s", delta, expectedFromTax)
	s.Require().True(delta.LT(expectedFromTax.Add(sdkmath.LegacyNewDec(10))),
		"community pool delta %s should be ≤ tax slice + small drift", delta)
}

func (s *DistrWrapperTestSuite) TestBeginBlock_NoBondedValidators_StakingShareToCommunityPool() {
	// Wipe the genesis bonded validator(s) so the bonded set is empty.
	validators, err := s.App.StakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)
	for _, v := range validators {
		v.Status = stakingtypes.Unbonded
		s.Require().NoError(s.App.StakingKeeper.SetValidator(s.Ctx, v))
	}

	feeAmount := sdkmath.NewInt(1_000_000)
	s.fundFeeCollector(feeAmount)

	feePoolBefore, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)

	s.callBeginBlock()

	feePoolAfter, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)

	denom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	delta := feePoolAfter.CommunityPool.AmountOf(denom).
		Sub(feePoolBefore.CommunityPool.AmountOf(denom))

	// With no bonded validators, the entire 85% staking share lands in
	// community pool. POA still gets 15%.
	s.Require().Equal(sdkmath.LegacyNewDec(850_000), delta,
		"with no bonded validators, all 85%% should go to community pool")
	s.Require().Equal(sdkmath.NewInt(150_000), s.poaBalance(),
		"POA still gets 15%% even with no bonded validators")
}
