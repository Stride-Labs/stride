package v35_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v35/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v35/app/upgrades/v35"
	"github.com/Stride-Labs/stride/v35/utils"
)

func (s *UpgradeTestSuite) TestRaiseMaxUnbondingEntries() {
	before, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	before.MaxEntries = 7
	before.UnbondingTime = 14 * 24 * time.Hour
	s.Require().NoError(s.App.StakingKeeper.SetParams(s.Ctx, before))

	s.Require().NoError(v35.RaiseMaxUnbondingEntries(s.Ctx, s.App.StakingKeeper))

	after, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint32(v35.StakingMaxEntries), after.MaxEntries)
	expected := before
	expected.MaxEntries = v35.StakingMaxEntries
	s.Require().Equal(expected, after, "only max entries changes")
}

// seedValidator builds a validator record the way v33's setupGovenatorState does, then runs the
// AfterValidatorCreated hook so distribution tracks it and Delegate can write starting info.
// Returns the operator address, which doubles as the self-delegation account.
func (s *UpgradeTestSuite) seedValidator(seed byte, status stakingtypes.BondStatus, minSelfDelegation int64) (sdk.ValAddress, sdk.AccAddress) {
	pubKey := ed25519.GenPrivKeyFromSecret([]byte{seed}).PubKey()
	valAddr := sdk.ValAddress(pubKey.Address())
	operator := sdk.AccAddress(pubKey.Address())

	validator, err := stakingtypes.NewValidator(valAddr.String(), pubKey, stakingtypes.Description{Moniker: fmt.Sprintf("validator-%d", seed)})
	s.Require().NoError(err)
	validator.Status = status
	validator.MinSelfDelegation = sdkmath.NewInt(minSelfDelegation)
	s.Require().NoError(s.App.StakingKeeper.SetValidator(s.Ctx, validator))
	s.Require().NoError(s.App.StakingKeeper.SetValidatorByConsAddr(s.Ctx, validator))
	s.Require().NoError(s.App.StakingKeeper.SetValidatorByPowerIndex(s.Ctx, validator))
	s.Require().NoError(s.App.StakingKeeper.Hooks().AfterValidatorCreated(s.Ctx, valAddr))
	return valAddr, operator
}

// delegate funds the delegator and delegates through the keeper, so the pools are funded and the
// distribution starting info exists (a bare SetDelegation would make the undelegation fail).
func (s *UpgradeTestSuite) delegate(delegator sdk.AccAddress, valAddr sdk.ValAddress, amount int64) sdkmath.Int {
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	tokens := sdkmath.NewInt(amount)
	s.FundAccount(delegator, sdk.NewCoin(bondDenom, tokens))

	validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)
	_, err = s.App.StakingKeeper.Delegate(s.Ctx, delegator, tokens, stakingtypes.Unbonded, validator, true)
	s.Require().NoError(err)
	return tokens
}

func (s *UpgradeTestSuite) mustGetValidator(valAddr sdk.ValAddress) stakingtypes.Validator {
	validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)
	return validator
}

func (s *UpgradeTestSuite) unbondingEntries(delegator sdk.AccAddress, valAddr sdk.ValAddress) []stakingtypes.UnbondingDelegationEntry {
	unbonding, err := s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, delegator, valAddr)
	s.Require().NoError(err, "unbonding delegation %s -> %s", delegator, valAddr)
	return unbonding.Entries
}

func (s *UpgradeTestSuite) hasDelegation(delegator sdk.AccAddress, valAddr sdk.ValAddress) bool {
	_, err := s.App.StakingKeeper.GetDelegation(s.Ctx, delegator, valAddr)
	return err == nil
}

func (s *UpgradeTestSuite) TestUndelegateAllDelegations() {
	// ----- arrange -----
	// Three validators (authority spec §5): bonded with a self-delegation that will drop below the
	// minimum, unbonded, and bonded with a pre-existing unbonding entry and a broken delegation
	jailedVal, operator := s.seedValidator(1, stakingtypes.Bonded, 1_000)
	unbondedVal, unbondedOperator := s.seedValidator(2, stakingtypes.Unbonded, 1)
	bondedVal, _ := s.seedValidator(3, stakingtypes.Bonded, 1)
	delegators := apptesting.CreateRandomAccounts(4)

	selfTokens := s.delegate(operator, jailedVal, 5_000)
	jailedValTokens := s.delegate(delegators[0], jailedVal, 7_000)
	unbondedValTokens := s.delegate(delegators[1], unbondedVal, 3_000)
	bondedValTokens := s.delegate(delegators[2], bondedVal, 8_000)
	skippedTokens := s.delegate(delegators[3], bondedVal, 2_000)

	// The unbonded validator earns 10% commission on 1,000 ustrd of rewards, so removing it in the
	// handler runs the commission payout in AfterValidatorRemoved (operator 100, delegator 900)
	s.setCommissionRate(unbondedVal, "0.1")
	s.FundModuleAccount(distrtypes.ModuleName, sdk.NewInt64Coin(utils.BaseStrideDenom, 1_000))
	s.Require().NoError(s.App.DistrKeeper.AllocateTokensToValidator(s.Ctx, s.mustGetValidator(unbondedVal),
		sdk.NewDecCoins(sdk.NewInt64DecCoin(utils.BaseStrideDenom, 1_000))))
	operatorBalanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, unbondedOperator, utils.BaseStrideDenom)
	delegatorBalanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, delegators[1], utils.BaseStrideDenom)

	// A pre-existing unbonding entry from an earlier block, which the handler must leave alone
	earlierTime := s.Ctx.BlockTime().Add(-time.Hour)
	earlierCtx := s.Ctx.WithBlockTime(earlierTime)
	priorCompletion, priorAmount, err := s.App.StakingKeeper.Undelegate(earlierCtx, delegators[2], bondedVal, sdkmath.LegacyNewDec(1_000))
	s.Require().NoError(err)
	bondedValTokens = bondedValTokens.Sub(priorAmount)

	// The handler runs at a later block, so its entries are distinct from the pre-existing one
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1).WithBlockTime(s.Ctx.BlockTime().Add(time.Minute))

	// The skipped delegation: its distribution starting info is gone, so the reward withdrawal fails
	s.Require().NoError(s.App.DistrKeeper.DeleteDelegatorStartingInfo(s.Ctx, bondedVal, delegators[3]))

	unbondingTime, err := s.App.StakingKeeper.UnbondingTime(s.Ctx)
	s.Require().NoError(err)
	expectedCompletion := s.Ctx.BlockTime().Add(unbondingTime)

	// A fresh event manager so only the handler's own events are inspected
	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())

	// ----- act -----
	s.Require().NoError(v35.UndelegateAllDelegations(s.Ctx, s.App.StakingKeeper))

	// ----- assert: no events reach the block result (the loop runs on a throwaway manager) -----
	for _, event := range s.Ctx.EventManager().Events() {
		s.Require().NotEqual(distrtypes.EventTypeWithdrawRewards, event.Type, "hook events must not be emitted")
	}

	// ----- assert: only the broken delegation remains, untouched -----
	remaining, err := s.App.StakingKeeper.GetAllDelegations(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(remaining, 1, "every delegation but the skipped one is gone")
	s.Require().Equal(delegators[3].String(), remaining[0].DelegatorAddress)
	s.Require().Equal(sdkmath.LegacyNewDecFromInt(skippedTokens), remaining[0].Shares, "the skipped delegation keeps its shares")
	_, err = s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, delegators[3], bondedVal)
	s.Require().ErrorIs(err, stakingtypes.ErrNoUnbondingDelegation, "the skipped delegation has no unbonding entry")

	// ----- assert: one new entry per completed pair with the stock unbonding time -----
	for _, pair := range []struct {
		delegator sdk.AccAddress
		validator sdk.ValAddress
		tokens    sdkmath.Int
	}{
		{operator, jailedVal, selfTokens},
		{delegators[0], jailedVal, jailedValTokens},
		{delegators[1], unbondedVal, unbondedValTokens},
	} {
		entries := s.unbondingEntries(pair.delegator, pair.validator)
		s.Require().Len(entries, 1, "%s -> %s", pair.delegator, pair.validator)
		s.Require().Equal(expectedCompletion, entries[0].CompletionTime)
		s.Require().Equal(pair.tokens, entries[0].Balance)
		s.Require().Equal(s.Ctx.BlockHeight(), entries[0].CreationHeight)
	}

	// The pre-existing entry keeps its completion time and balance; the handler's entry is appended
	entries := s.unbondingEntries(delegators[2], bondedVal)
	s.Require().Len(entries, 2)
	s.Require().Equal(priorCompletion, entries[0].CompletionTime, "pre-existing entry untouched")
	s.Require().Equal(priorAmount, entries[0].Balance, "pre-existing entry untouched")
	s.Require().Equal(expectedCompletion, entries[1].CompletionTime)
	s.Require().Equal(bondedValTokens, entries[1].Balance)

	// ----- assert: validators -----
	// Bonded validators stay at zero tokens for the EndBlocker to unbond; an already-unbonded
	// validator with no shares left is removed by Unbond on the spot
	s.Require().True(s.mustGetValidator(jailedVal).Jailed, "operator dropped below min self-delegation")
	s.Require().True(s.mustGetValidator(jailedVal).Tokens.IsZero())
	_, err = s.App.StakingKeeper.GetValidator(s.Ctx, unbondedVal)
	s.Require().ErrorIs(err, stakingtypes.ErrNoValidatorFound, "unbonded validator removed once empty")

	// Removal paid out the commission to the operator and the rest went to the delegator; the
	// distribution records for the removed validator are cleared
	operatorBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, unbondedOperator, utils.BaseStrideDenom)
	delegatorBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, delegators[1], utils.BaseStrideDenom)
	s.Require().Equal(sdkmath.NewInt(100), operatorBalanceAfter.Amount.Sub(operatorBalanceBefore.Amount), "commission paid to operator")
	s.Require().Equal(sdkmath.NewInt(900), delegatorBalanceAfter.Amount.Sub(delegatorBalanceBefore.Amount), "remaining rewards paid to delegator")
	outstanding, err := s.App.DistrKeeper.GetValidatorOutstandingRewards(s.Ctx, unbondedVal)
	s.Require().NoError(err)
	s.Require().True(outstanding.Rewards.IsZero(), "outstanding rewards cleared")
	commission, err := s.App.DistrKeeper.GetValidatorAccumulatedCommission(s.Ctx, unbondedVal)
	s.Require().NoError(err)
	s.Require().True(commission.Commission.IsZero(), "accumulated commission cleared")

	s.Require().False(s.mustGetValidator(bondedVal).Jailed)
	s.Require().Equal(skippedTokens, s.mustGetValidator(bondedVal).Tokens, "only the skipped delegation's tokens stay on the validator")
}

// A pair already at mainnet's 7 entries is skipped at MaxEntries 7 and succeeds once step 3 has
// raised the cap: the handler runs the raise before the undelegation.
func (s *UpgradeTestSuite) TestUndelegateAllDelegations_AfterRaiseMaxUnbondingEntries() {
	params, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	params.MaxEntries = 7
	s.Require().NoError(s.App.StakingKeeper.SetParams(s.Ctx, params))

	valAddr, _ := s.seedValidator(4, stakingtypes.Bonded, 1)
	delegator := apptesting.CreateRandomAccounts(1)[0]
	tokens := s.delegate(delegator, valAddr, 9_000)

	// Seven entries at distinct creation heights (same-height entries merge into one)
	for height := int64(1); height <= 7; height++ {
		_, err := s.App.StakingKeeper.SetUnbondingDelegationEntry(s.Ctx, delegator, valAddr, height, s.Ctx.BlockTime(), sdkmath.NewInt(1))
		s.Require().NoError(err)
	}
	s.Ctx = s.Ctx.WithBlockHeight(10)

	// At the mainnet cap the pair is skipped and the delegation stays
	s.Require().NoError(v35.UndelegateAllDelegations(s.Ctx, s.App.StakingKeeper))
	s.Require().True(s.hasDelegation(delegator, valAddr), "skipped at max entries")
	s.Require().Len(s.unbondingEntries(delegator, valAddr), 7)

	// After the raise the same delegation unbonds as the eighth entry
	s.Require().NoError(v35.RaiseMaxUnbondingEntries(s.Ctx, s.App.StakingKeeper))
	s.Require().NoError(v35.UndelegateAllDelegations(s.Ctx, s.App.StakingKeeper))
	s.Require().False(s.hasDelegation(delegator, valAddr), "undelegated after the raise")
	entries := s.unbondingEntries(delegator, valAddr)
	s.Require().Len(entries, 8)
	s.Require().Equal(tokens, entries[7].Balance)
}

// setCommissionRate gives a seeded validator a non-zero commission rate
func (s *UpgradeTestSuite) setCommissionRate(valAddr sdk.ValAddress, rate string) {
	validator := s.mustGetValidator(valAddr)
	validator.Commission = stakingtypes.NewCommission(sdkmath.LegacyMustNewDecFromStr(rate), sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec())
	s.Require().NoError(s.App.StakingKeeper.SetValidator(s.Ctx, validator))
}

// If a delegation's undelegation fails after Unbond has already written (here the bonded pool is
// empty, so moving tokens to the not-bonded pool fails), the cache context discards the partial
// write: the delegation is skipped intact, not left deleted without an unbonding entry.
func (s *UpgradeTestSuite) TestUndelegateAllDelegations_FailureAfterUnbondIsRolledBack() {
	valAddr, _ := s.seedValidator(5, stakingtypes.Bonded, 1)
	delegator := apptesting.CreateRandomAccounts(1)[0]
	tokens := s.delegate(delegator, valAddr, 6_000)

	// Drain the bonded pool so bondedTokensToNotBonded fails for the bonded validator
	bondedPool := authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
	s.Require().NoError(s.App.BankKeeper.SendCoinsFromModuleToAccount(
		s.Ctx, stakingtypes.BondedPoolName, apptesting.CreateRandomAccounts(1)[0], s.App.BankKeeper.GetAllBalances(s.Ctx, bondedPool)))

	s.Require().NoError(v35.UndelegateAllDelegations(s.Ctx, s.App.StakingKeeper))

	delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, delegator, valAddr)
	s.Require().NoError(err, "delegation kept")
	s.Require().Equal(sdkmath.LegacyNewDecFromInt(tokens), delegation.Shares, "original shares")
	s.Require().Equal(tokens, s.mustGetValidator(valAddr).Tokens, "validator keeps its tokens")
	_, err = s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, delegator, valAddr)
	s.Require().ErrorIs(err, stakingtypes.ErrNoUnbondingDelegation, "no unbonding entry")
}

// Removing the empty unbonded validator runs distribution's AfterValidatorRemoved, which panics
// when accumulated commission exceeds outstanding rewards. The panic must become a skip for that
// delegation only, not halt the upgrade.
func (s *UpgradeTestSuite) TestUndelegateAllDelegations_PanicIsSkipped() {
	unbondedVal, _ := s.seedValidator(6, stakingtypes.Unbonded, 1)
	bondedVal, _ := s.seedValidator(7, stakingtypes.Bonded, 1)
	delegators := apptesting.CreateRandomAccounts(2)
	panicTokens := s.delegate(delegators[0], unbondedVal, 3_000)
	s.delegate(delegators[1], bondedVal, 4_000)

	// 1 ustrd of commission against zero outstanding rewards underflows outstanding.Sub(commission)
	s.Require().NoError(s.App.DistrKeeper.SetValidatorAccumulatedCommission(s.Ctx, unbondedVal, distrtypes.ValidatorAccumulatedCommission{
		Commission: sdk.NewDecCoins(sdk.NewInt64DecCoin(utils.BaseStrideDenom, 1)),
	}))

	s.Require().NotPanics(func() {
		s.Require().NoError(v35.UndelegateAllDelegations(s.Ctx, s.App.StakingKeeper))
	})

	delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, delegators[0], unbondedVal)
	s.Require().NoError(err, "panicking delegation skipped, still present")
	s.Require().Equal(sdkmath.LegacyNewDecFromInt(panicTokens), delegation.Shares)
	s.Require().Equal(panicTokens, s.mustGetValidator(unbondedVal).Tokens)
	s.Require().False(s.hasDelegation(delegators[1], bondedVal), "other delegations still unbond")
	s.Require().Len(s.unbondingEntries(delegators[1], bondedVal), 1)
}
