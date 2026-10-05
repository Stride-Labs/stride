package distrwrapper_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/Stride-Labs/stride/v35/app/apptesting"
)

// setRemovalRecords gives a validator outstanding rewards and accumulated commission, and funds the
// distribution module with the outstanding amount so the commission payout can be sent.
func (s *DistrWrapperTestSuite) setRemovalRecords(valAddr sdk.ValAddress, outstanding, commission sdk.DecCoins) {
	s.Require().NoError(s.App.DistrKeeper.SetValidatorOutstandingRewards(s.Ctx, valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: outstanding}))
	s.Require().NoError(s.App.DistrKeeper.SetValidatorAccumulatedCommission(s.Ctx, valAddr, distrtypes.ValidatorAccumulatedCommission{Commission: commission}))
	funds, _ := outstanding.TruncateDecimal()
	for _, coin := range funds {
		s.FundModuleAccount(distrtypes.ModuleName, coin)
	}
}

// communityPool returns the community pool's balance in the bond denom.
func (s *DistrWrapperTestSuite) communityPool(denom string) sdkmath.LegacyDec {
	feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
	s.Require().NoError(err)
	return feePool.CommunityPool.AmountOf(denom)
}

// removeValidator runs AfterValidatorRemoved through the app's wired staking hooks, the path the
// staking keeper takes when it deletes a validator.
func (s *DistrWrapperTestSuite) removeValidator(valAddr sdk.ValAddress) error {
	return s.App.StakingKeeper.Hooks().AfterValidatorRemoved(s.Ctx, sdk.ConsAddress(valAddr), valAddr)
}

// Rounding in reward withdrawals can leave a validator's accumulated commission above its
// outstanding rewards. The stock distribution hook panics on outstanding.Sub(commission), which in
// the staking EndBlocker halts the chain; the wrapped hook caps commission at outstanding first,
// so the operator is paid what outstanding holds and a denom outstanding no longer has is dropped.
func (s *DistrWrapperTestSuite) TestAfterValidatorRemoved_CommissionAboveOutstanding() {
	denom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	valAddr := sdk.ValAddress(apptesting.CreateRandomAccounts(1)[0])
	s.setRemovalRecords(valAddr,
		sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 3)),
		sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 5), sdk.NewInt64DecCoin("uosmo", 2)))
	communityPoolBefore := s.communityPool(denom)

	s.Require().NotPanics(func() {
		s.Require().NoError(s.removeValidator(valAddr))
	})

	s.Require().Equal(sdkmath.NewInt(3), s.App.BankKeeper.GetBalance(s.Ctx, sdk.AccAddress(valAddr), denom).Amount,
		"operator paid the clamped commission")
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, sdk.AccAddress(valAddr), "uosmo").IsZero(), "foreign denom dropped")
	s.Require().Equal(communityPoolBefore, s.communityPool(denom), "nothing left over for the community pool")
	commission, err := s.App.DistrKeeper.GetValidatorAccumulatedCommission(s.Ctx, valAddr)
	s.Require().NoError(err)
	s.Require().True(commission.Commission.IsZero(), "commission record removed")
}

// A validator within bounds is removed exactly as the stock hook would: full commission to the
// operator, the remaining outstanding to the community pool.
func (s *DistrWrapperTestSuite) TestAfterValidatorRemoved_CommissionWithinOutstanding() {
	denom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	valAddr := sdk.ValAddress(apptesting.CreateRandomAccounts(1)[0])
	s.setRemovalRecords(valAddr, sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 10)), sdk.NewDecCoins(sdk.NewInt64DecCoin(denom, 4)))
	communityPoolBefore := s.communityPool(denom)

	s.Require().NoError(s.removeValidator(valAddr))

	s.Require().Equal(sdkmath.NewInt(4), s.App.BankKeeper.GetBalance(s.Ctx, sdk.AccAddress(valAddr), denom).Amount,
		"operator paid the full commission")
	s.Require().Equal(communityPoolBefore.Add(sdkmath.LegacyNewDec(6)), s.communityPool(denom),
		"remaining outstanding goes to the community pool")
}
