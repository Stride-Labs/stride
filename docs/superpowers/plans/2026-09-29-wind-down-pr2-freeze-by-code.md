# Wind-Down PR 2: Freeze by Code Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Freeze every stakeibc host zone's redemption rate and stop every epoch flow that compounds or moves stake, by deleting call sites rather than halting zones, while the flows that pay open redemptions keep running.

**Architecture:** Nine calls are removed from `BeforeEpochStart` in `x/stakeibc/keeper/hooks.go` (rate update, reinvest, delegate, rebalance, reward-token transfer, withdrawal-address set, deposit-record creation, epoch-unbonding-record creation, reward-collector auction); the keeper functions behind them stay defined. The delegator-shares slash callback stops rewriting the rate, the 5,000 base-unit calibration cap is lifted, and the two ICQ messages that reach the slash path (`MsgUpdateValidatorSharesExchRate`, `MsgCalibrateDelegation`) become admin-only in `ValidateBasic`. Nothing is configurable and nothing can be toggled back.

**Tech Stack:** Go 1.25, Cosmos SDK v0.54.3, ibc-go v11.2.0, testify suites (`KeeperTestSuite` in `x/stakeibc/keeper`, plain `testing` in `x/stakeibc/types`).

**Spec:** `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` §5 "Messages gated", §6, §11 "Freeze by code", §12 item 2.

> **Branching:** PR 1 branches off `wind-down-design-consolidation`. Each later PR branches
> off the previous PR's branch (PR 2 off PR 1, PR 3 off PR 2, and so on) and the PRs are
> implemented and merged strictly in order: 1, 2, 3, 4, 5. Branch names:
> `wind-down-pr1-remove-handlers`, `wind-down-pr2-freeze-by-code`,
> `wind-down-pr3-upgrade-handler`, `wind-down-pr4-admin-txs`, `wind-down-pr5-sweep-tx`.
> The Go module path stays `github.com/Stride-Labs/stride/v34` in every PR; the bump to
> `/v35` is a manual step after all five land and is out of scope for every plan.

This plan is executed on `wind-down-pr2-freeze-by-code`, created from `wind-down-pr1-remove-handlers` (`git checkout wind-down-pr1-remove-handlers && git checkout -b wind-down-pr2-freeze-by-code`). PR 1's removals are assumed present: the `LiquidStake`/`RedeemStake` msg handlers are gone and their keeper paths remain (`Keeper.LiquidStake`, `Keeper.RedeemStake`); `reward_allocation.go` calls `k.LiquidStake` directly.

## Global Constraints

- No host zone is halted by this PR; `HostZone.Halted` is never written.
- Deleted from `BeforeEpochStart`, exactly these nine calls: `UpdateRedemptionRates`, `ReinvestRewards`, `StakeExistingDepositsOnHostZones`, `RebalanceAllHostZones`, `TransferAllRewardTokens`, `SetWithdrawalAddress`, `CreateDepositRecordsForEpoch`, `CreateEpochUnbondingRecord`, `AuctionOffRewardCollectorBalance`.
- Kept in `BeforeEpochStart`, exactly these seven: `UpdateEpochTracker`, `InitiateAllHostZoneUnbondings`, `SubmitPendingUndelegations`, `CleanupEpochUnbondingRecords`, `ClaimAccruedStakingRewards`, `TransferExistingDepositsToHostZones` (still gated on `DepositInterval`), `SweepUnbondedTokensAllHostZones`.
- The keeper functions behind the deleted calls are NOT deleted and their existing tests stay green (`records_test.go`, `delegation_test.go`, `redemption_rate_test.go`, `reward_allocation_test.go`, `rebalance_test.go`, `reward_converter_test.go` all keep calling them directly).
- `SlashValidatorOnHostZone` keeps correcting `Validator.Delegation`, `Validator.Weight` and `HostZone.TotalDelegations`; it no longer calls `UpdateRedemptionRateForHostZone`.
- `CalibrationThreshold` and its check are removed from `CalibrateDelegationCallback`; `types.ErrCalibrationThresholdExceeded` (code 1552) stays registered (its only reference is `errors.go`; an unused registered error is harmless and removing a code from the registry is not worth a line in this diff).
- `MsgUpdateValidatorSharesExchRate.ValidateBasic` and `MsgCalibrateDelegation.ValidateBasic` call `utils.ValidateAdminAddress(msg.Creator)` after the bech32 check, returning its error unchanged (it wraps `sdkerrors.ErrInvalidAddress` with `"address (%s) is not an admin"`).
- No proto, CLI, or module-path change. Module path stays `github.com/Stride-Labs/stride/v34`.
- Every commit message ends with `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`.
- Known pre-existing failure to ignore: `utils` `TestCreateModuleAccount` fails on main too.

---

## File Structure

| File | Change |
|---|---|
| `x/stakeibc/keeper/hooks.go` | Rewrite `BeforeEpochStart`: nine calls and three unused interval locals removed; nothing else in the file changes. |
| `x/stakeibc/keeper/hooks_test.go` | New. Behavioural tests per epoch type plus a source-level guard on `BeforeEpochStart`. |
| `x/stakeibc/keeper/icqcallbacks_delegator_shares.go` | Remove the two-line rate refresh at the end of `SlashValidatorOnHostZone`. |
| `x/stakeibc/keeper/icqcallbacks_delegator_shares_test.go` | Add `TestDelegatorSharesCallback_RedemptionRateFrozen`. |
| `x/stakeibc/keeper/icqcallbacks_callibrate_delegation.go` | Remove `CalibrationThreshold`, its check and the now-unused `sdkmath` import. |
| `x/stakeibc/keeper/icqcallbacks_callibrate_delegation_test.go` | The two "exceeds threshold" cases now apply; case names updated. |
| `x/stakeibc/types/message_update_delegation.go` | Admin gate in `ValidateBasic`. |
| `x/stakeibc/types/message_calibrate_delegation.go` | Admin gate in `ValidateBasic`. |
| `x/stakeibc/types/message_update_delegation_test.go` | New. |
| `x/stakeibc/types/message_calibrate_delegation_test.go` | New. |

Verified before writing this plan: no test, CLI test, or script constructs either ICQ message with a non-admin creator (`grep -rn "MsgUpdateValidatorSharesExchRate{\|MsgCalibrateDelegation{\|NewMsgUpdateValidatorSharesExchRate(\|NewMsgCalibrateDelegation(" x/ app/` finds only the type files, `codec.go` and `client/cli/tx.go`; `scripts/local-to-mainnet/commands.sh` already signs `update-delegation` with `--from admin`). `TestDelegatorSharesCallback_Successful` does not assert on `RedemptionRate` (the test file never mentions it), so it needs no change. There is no existing `hooks_test.go`.

---

### Task 1: Delete the compounding calls from `BeforeEpochStart` and prove the freeze

**Files:**
- Modify: `x/stakeibc/keeper/hooks.go:19-99`
- Create: `x/stakeibc/keeper/hooks_test.go`

**Interfaces:**
- Consumes: `KeeperTestSuite` helpers already in package `keeper_test`: `SetupUpdateRedemptionRates` (`redemption_rate_test.go`), `SetupInitiateAllHostZoneUnbondings` (`unbonding_test.go`), `SetupTestRewardAllocation`, `checkModuleAccountBalance`, `getTotalPoAValidatorStTokenBalance` (`reward_allocation_test.go`).
- Produces: `func (s *KeeperTestSuite) epochInfoForHook(identifier string, epochNumber int64) epochstypes.EpochInfo` (test helper; Tasks 2-4 do not use it).
- Review: yes (this is the change that freezes the rate on mainnet).

- [ ] **Step 1: Write the failing hook tests**

Create `x/stakeibc/keeper/hooks_test.go`:

```go
package keeper_test

import (
	"os"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/Stride-Labs/stride/v34/x/epochs/types"
	minttypes "github.com/Stride-Labs/stride/v34/x/mint/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// epochInfoForHook builds the EpochInfo the epochs module hands to BeforeEpochStart. The
// current epoch starts at the block time so the tracker's next start time is in the future,
// which the ICA timeouts of the kept day-epoch flows need.
func (s *KeeperTestSuite) epochInfoForHook(identifier string, epochNumber int64) epochstypes.EpochInfo {
	return epochstypes.EpochInfo{
		Identifier:            identifier,
		Duration:              time.Hour,
		CurrentEpoch:          epochNumber,
		CurrentEpochStartTime: s.Ctx.BlockTime(),
		EpochCountingStarted:  true,
	}
}

// The stride epoch used to refresh the redemption rate, create a deposit record for the epoch
// and submit the withdrawal-balance ICQ that starts the reinvest pipeline. After the freeze it
// does none of that, while the epoch tracker (a kept call) still advances.
func (s *KeeperTestSuite) TestBeforeEpochStart_StrideEpoch_RateFrozen() {
	// Same fixture as TestUpdateRedemptionRatesSuccessful: the old hook would have moved the
	// rate from 1.0 to (2 + 3 + 4 + 5) / 10 = 1.4
	s.SetupUpdateRedemptionRates(UpdateRedemptionRateTestCase{
		totalDelegation:       sdkmath.NewInt(2),
		undelegatedBal:        sdkmath.NewInt(3),
		justDepositedNative:   sdkmath.NewInt(4),
		justDepositedLSM:      sdkmath.NewInt(5),
		stSupply:              sdkmath.NewInt(10),
		initialRedemptionRate: sdkmath.LegacyNewDec(1),
	})
	initialDepositRecords := len(s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx))
	s.Require().Equal(2, initialDepositRecords, "fixture seeds two deposit records")

	s.App.StakeibcKeeper.BeforeEpochStart(s.Ctx, s.epochInfoForHook(epochstypes.STRIDE_EPOCH, 7))

	// Frozen: the rate is exactly what it was
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(sdkmath.LegacyNewDec(1), hostZone.RedemptionRate, "redemption rate must not move on a stride epoch")

	// No deposit record was created for the epoch
	s.Require().Len(s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx), initialDepositRecords, "no deposit record created for the epoch")

	// No ICQ was submitted (the withdrawal-balance query is the root of the reinvest pipeline)
	s.Require().Empty(s.App.InterchainqueryKeeper.AllQueries(s.Ctx), "no interchain query submitted")

	// Kept: the epoch tracker advanced
	tracker, found := s.App.StakeibcKeeper.GetEpochTracker(s.Ctx, epochstypes.STRIDE_EPOCH)
	s.Require().True(found, "stride epoch tracker set")
	s.Require().Equal(uint64(7), tracker.EpochNumber, "epoch tracker updated by the hook")
}

// The day epoch still submits queued undelegations (kept) but no longer creates an epoch
// unbonding record for the new epoch (deleted).
func (s *KeeperTestSuite) TestBeforeEpochStart_DayEpoch_UnbondsWithoutNewRecord() {
	// Seeds GAIA and OSMO host zones, a queued unbonding record at epoch 5, and a GAIA.DELEGATION
	// ICA channel; both zones unbond on epoch 12 (see TestInitiateAllHostZoneUnbondings_Successful)
	s.SetupInitiateAllHostZoneUnbondings()
	dayEpoch := int64(12)
	_, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, uint64(dayEpoch))
	s.Require().False(found, "no epoch unbonding record for the new epoch before the hook")

	s.App.StakeibcKeeper.BeforeEpochStart(s.Ctx, s.epochInfoForHook(epochstypes.DAY_EPOCH, dayEpoch))

	// Kept: the queued records were submitted as undelegate ICAs
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)

	// Deleted: nothing appends to an epoch unbonding record any more, so none is created
	_, found = s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, uint64(dayEpoch))
	s.Require().False(found, "no epoch unbonding record created for the new epoch")
}

// The mint epoch used to liquid stake 15% of the reward collector's host fees for the POA
// validators and send the rest to the auction module. Both stop: the balance stays put and no
// stTokens are minted.
func (s *KeeperTestSuite) TestBeforeEpochStart_MintEpoch_RewardCollectorUntouched() {
	s.SetupTestRewardAllocation()
	rewardAmount := sdkmath.NewInt(1000)
	s.FundModuleAccount(types.RewardCollectorName, sdk.NewCoin(IbcAtom, rewardAmount))

	s.App.StakeibcKeeper.BeforeEpochStart(s.Ctx, s.epochInfoForHook(epochstypes.MINT_EPOCH, 3))

	s.checkModuleAccountBalance(types.RewardCollectorName, IbcAtom, rewardAmount)
	s.Require().Equal(sdkmath.ZeroInt().Int64(), s.getTotalPoAValidatorStTokenBalance(StAtom).Int64(),
		"no stTokens minted to the POA validators")
}

// Source-level guard: the freeze is a set of deleted call sites, and the cheapest unambiguous
// way to keep them deleted is to assert on the text of BeforeEpochStart. A behavioural test
// can only show one consequence per call; this one names every call by identifier, so a
// re-added call (or a merge that resurrects one) fails the build with the offending name in
// the message. It reads hooks.go relative to the package directory, which is where `go test`
// runs.
func (s *KeeperTestSuite) TestBeforeEpochStart_SourceHasNoCompoundingCalls() {
	source, err := os.ReadFile("hooks.go")
	s.Require().NoError(err, "hooks.go readable from the package directory")

	start := strings.Index(string(source), "func (k Keeper) BeforeEpochStart(")
	end := strings.Index(string(source), "func (k Keeper) AfterEpochEnd(")
	s.Require().True(start >= 0 && end > start, "BeforeEpochStart and AfterEpochEnd both defined, in that order")
	body := string(source[start:end])

	deletedCalls := []string{
		"k.UpdateRedemptionRates(",
		"k.ReinvestRewards(",
		"k.StakeExistingDepositsOnHostZones(",
		"k.RebalanceAllHostZones(",
		"k.TransferAllRewardTokens(",
		"k.SetWithdrawalAddress(",
		"k.CreateDepositRecordsForEpoch(",
		"k.CreateEpochUnbondingRecord(",
		"k.AuctionOffRewardCollectorBalance(",
	}
	for _, call := range deletedCalls {
		s.Require().NotContains(body, call, "%s must not be called from BeforeEpochStart (spec §6)", call)
	}

	keptCalls := []string{
		"k.UpdateEpochTracker(",
		"k.InitiateAllHostZoneUnbondings(",
		"k.SubmitPendingUndelegations(",
		"k.CleanupEpochUnbondingRecords(",
		"k.ClaimAccruedStakingRewards(",
		"k.TransferExistingDepositsToHostZones(",
		"k.SweepUnbondedTokensAllHostZones(",
	}
	for _, call := range keptCalls {
		s.Require().Contains(body, call, "%s must still be called from BeforeEpochStart (spec §6)", call)
	}

	// The mint-epoch branch had exactly one call and it was deleted, so the branch is gone too
	s.Require().NotContains(body, "MINT_EPOCH", "no mint-epoch branch remains")
}

// Compile-time pin that the keeper functions behind the deleted calls still exist: they are
// dead from the hook but kept until the follow-up cleanup (spec §6, §12), and other tests
// call them directly.
var (
	_ = keeper.Keeper.UpdateRedemptionRates
	_ = keeper.Keeper.ReinvestRewards
	_ = keeper.Keeper.StakeExistingDepositsOnHostZones
	_ = keeper.Keeper.RebalanceAllHostZones
	_ = keeper.Keeper.TransferAllRewardTokens
	_ = keeper.Keeper.SetWithdrawalAddress
	_ = keeper.Keeper.CreateDepositRecordsForEpoch
	_ = keeper.Keeper.CreateEpochUnbondingRecord
	_ = keeper.Keeper.AuctionOffRewardCollectorBalance
	_ = minttypes.ModuleName
)
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```bash
go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestBeforeEpochStart' -v 2>&1 | tail -40
```

Expected: `TestBeforeEpochStart_StrideEpoch_RateFrozen` FAILS on `redemption rate must not move on a stride epoch` (actual `1.400000000000000000`), `TestBeforeEpochStart_DayEpoch_UnbondsWithoutNewRecord` FAILS on `no epoch unbonding record created for the new epoch`, `TestBeforeEpochStart_MintEpoch_RewardCollectorUntouched` FAILS on the reward collector balance (actual 0), `TestBeforeEpochStart_SourceHasNoCompoundingCalls` FAILS on `k.UpdateRedemptionRates( must not be called`.

- [ ] **Step 3: Rewrite `BeforeEpochStart`**

In `x/stakeibc/keeper/hooks.go`, replace the whole function at lines 19-99 (from `func (k Keeper) BeforeEpochStart(` through its closing `}` before `func (k Keeper) AfterEpochEnd(`) with:

```go
// BeforeEpochStart drives the flows that pay the redemptions still open after the v35 upgrade
// and nothing else. The protocol is winding down (docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md,
// §6): the calls that compounded or moved stake (rate refresh, reinvest, delegate, rebalance,
// reward-token transfer, withdrawal-address set, deposit and epoch-unbonding record creation,
// the reward-collector auction) were deleted rather than gated, so the redemption rate of every
// host zone is frozen at its last pre-upgrade value and cannot be toggled back. Their keeper
// functions remain defined until the follow-up cleanup.
func (k Keeper) BeforeEpochStart(context context.Context, epochInfo epochstypes.EpochInfo) {
	ctx := sdk.UnwrapSDKContext(context)

	// Update the stakeibc epoch tracker: the day tracker feeds the ICA timeout of every
	// undelegate batch and the stride tracker feeds the redemption sweep ICA
	epochNumber, err := k.UpdateEpochTracker(ctx, epochInfo)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Unable to update epoch tracker, err: %s", err.Error()))
		return
	}

	// Day Epoch - submit and clean up unbondings
	if epochInfo.Identifier == epochstypes.DAY_EPOCH {
		// Submit the queued redemption records of any host zone that unbonds this epoch
		k.InitiateAllHostZoneUnbondings(ctx, epochNumber)
		// Submit any one-shot undelegations queued by an upgrade handler (e.g. the v34 Injective
		// reconciliation). Store-driven, so this is a no-op when nothing is pending. A host zone
		// that unbonds this epoch is deferred so the two flows don't compete for validator capacity
		k.SubmitPendingUndelegations(ctx, epochNumber)
		// Delete epoch unbonding records once every host zone's unbonding on them is claimed
		k.CleanupEpochUnbondingRecords(ctx, epochNumber)
	}

	// Stride Epoch - move what is already in flight toward the redemption accounts
	if epochInfo.Identifier == epochstypes.STRIDE_EPOCH {
		depositInterval := k.GetParam(ctx, types.KeyDepositInterval)

		// Withdraw accrued staking rewards to the withdrawal ICA, where the wind-down sweeps them
		k.ClaimAccruedStakingRewards(ctx)

		// Carry any native tokens still sitting in a deposit address to the delegation ICA so they
		// leave with the ICA balance instead of being stranded on Stride
		if epochNumber%depositInterval == 0 {
			depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
			k.TransferExistingDepositsToHostZones(ctx, epochNumber, depositRecords)
		}

		// Check previous epochs to see if unbondings finished, and send the relevant tokens
		// to the redemption account
		k.SweepUnbondedTokensAllHostZones(ctx)
	}
}
```

The `MINT_EPOCH` branch is deleted entirely. The imports of `hooks.go` are unchanged: `time`, `proto`, `utils` and `types` are still used by `DisableHubTokenization`, `SetWithdrawalAddress` and `ClaimAccruedStakingRewards` further down the file, which stay.

- [ ] **Step 4: Build and run the hook tests to verify they pass**

Run:

```bash
go build ./... && go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestBeforeEpochStart' -v 2>&1 | tail -20
```

Expected: build clean; all four `TestBeforeEpochStart_*` PASS.

- [ ] **Step 5: Run the whole stakeibc keeper suite to confirm the kept keeper functions' tests are still green**

Run:

```bash
go test ./x/stakeibc/... 2>&1 | tail -5
```

Expected: `ok  	github.com/Stride-Labs/stride/v34/x/stakeibc/keeper` and `ok` for the other stakeibc packages. In particular `TestUpdateRedemptionRatesSuccessful`, `TestLiquidStakeRewardCollectorBalance_Success`, `TestCreateDepositRecordsForEpoch` (records_test.go) and `TestStakeExistingDepositsOnHostZones` (delegation_test.go) still pass because they call the keeper functions directly.

- [ ] **Step 6: Commit**

```bash
git add x/stakeibc/keeper/hooks.go x/stakeibc/keeper/hooks_test.go
git commit -m "feat(stakeibc): freeze the redemption rate by deleting the compounding epoch calls

Removes UpdateRedemptionRates, ReinvestRewards, StakeExistingDepositsOnHostZones,
RebalanceAllHostZones, TransferAllRewardTokens, SetWithdrawalAddress,
CreateDepositRecordsForEpoch, CreateEpochUnbondingRecord and
AuctionOffRewardCollectorBalance from BeforeEpochStart (wind-down spec §6). The
flows that pay open redemptions stay. Keeper functions remain defined.

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Parallel-safe tasks

Tasks 2, 3 and 4 touch disjoint files and none consumes an interface produced by another or by Task 1. Each depends only on the branch base.

### Task 2: The slash callback corrects the delegation but no longer rewrites the rate

**Files:**
- Modify: `x/stakeibc/keeper/icqcallbacks_delegator_shares.go:250-256`
- Modify: `x/stakeibc/keeper/icqcallbacks_delegator_shares_test.go` (append one test)

**Interfaces:**
- Consumes: `SetupDelegatorSharesICQCallback()` and `DelegatorSharesICQCallbackTestCase` from the same test file (host zone `GAIA`, `TotalDelegations` 10,000, queried validator at index 1 with 1,000 delegated, 5% slash expected; no `RedemptionRate` set).
- Produces: nothing.
- Depends on: none.
- Review: yes (accounting path).

- [ ] **Step 1: Write the failing test**

Append to `x/stakeibc/keeper/icqcallbacks_delegator_shares_test.go` (add `sdk "github.com/cosmos/cosmos-sdk/types"` and `minttypes "github.com/Stride-Labs/stride/v34/x/mint/types"` to the imports):

```go
// A slash found by the query still lowers the validator's and the zone's delegation, but the
// redemption rate is frozen (wind-down spec §6): the callback no longer refreshes it. With a
// stToken supply of 10,000 against 9,950 delegated after the slash, the old refresh would have
// set the rate to 0.995; it must stay at the seeded 1.5.
func (s *KeeperTestSuite) TestDelegatorSharesCallback_RedemptionRateFrozen() {
	tc := s.SetupDelegatorSharesICQCallback()

	frozenRate := sdkmath.LegacyMustNewDecFromStr("1.5")
	hostZone := tc.hostZone
	hostZone.RedemptionRate = frozenRate
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	stSupply := sdk.NewCoins(sdk.NewCoin(StAtom, sdkmath.NewInt(10_000)))
	s.Require().NoError(s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, stSupply), "mint stToken supply")

	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "delegator shares callback error")

	updatedHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")

	// The delegation accounting reflects the slash
	s.Require().Equal(tc.expectedSlashAmount.Int64(), tc.hostZone.TotalDelegations.Sub(updatedHostZone.TotalDelegations).Int64(), "total delegations slashed")
	s.Require().Equal(tc.expectedDelegationAmount.Int64(), updatedHostZone.Validators[tc.valIndexQueried].Delegation.Int64(), "validator delegation slashed")

	// The rate did not move
	s.Require().Equal(frozenRate, updatedHostZone.RedemptionRate, "redemption rate must not be rewritten by the slash callback")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestDelegatorSharesCallback_RedemptionRateFrozen' -v 2>&1 | tail -15
```

Expected: FAIL on `redemption rate must not be rewritten by the slash callback` (actual `0.995000000000000000`).

- [ ] **Step 3: Remove the rate refresh from `SlashValidatorOnHostZone`**

In `x/stakeibc/keeper/icqcallbacks_delegator_shares.go`, delete these lines at the end of `SlashValidatorOnHostZone` (currently 250-253):

```go
	// Update the redemption rate
	depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
	k.UpdateRedemptionRateForHostZone(ctx, hostZone, depositRecords)

```

so the function ends:

```go
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Delegation,
		"Delegation updated to: %v, Weight updated to: %v", validator.Delegation, validator.Weight))

	// The redemption rate is frozen for the wind-down (spec §6): a slash lowers the backing but
	// not the rate, and the coverage check on Osmosis is where the difference shows up
	return nil
}
```

No import changes (`k.RecordsKeeper` is a keeper field, not an import).

- [ ] **Step 4: Run the delegator-shares tests to verify they pass**

Run:

```bash
go build ./... && go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestDelegatorSharesCallback' -v 2>&1 | tail -30
```

Expected: build clean; every `TestDelegatorSharesCallback_*` PASS including `_Successful` (it never asserted on the rate) and the new `_RedemptionRateFrozen`.

- [ ] **Step 5: Commit**

```bash
git add x/stakeibc/keeper/icqcallbacks_delegator_shares.go x/stakeibc/keeper/icqcallbacks_delegator_shares_test.go
git commit -m "feat(stakeibc): slash callback no longer rewrites the redemption rate

The delegator-shares callback keeps correcting the validator and host zone
delegation on a detected slash but the rate stays frozen (wind-down spec §6).

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 3: Lift the calibration cap

**Files:**
- Modify: `x/stakeibc/keeper/icqcallbacks_callibrate_delegation.go:4-16, 66-77`
- Modify: `x/stakeibc/keeper/icqcallbacks_callibrate_delegation_test.go:55-95`

**Interfaces:**
- Consumes: nothing new.
- Produces: nothing (the exported `CalibrationThreshold` var is removed; verified no other file references it).
- Depends on: none.
- Review: yes (accounting path; the cap bounded what a permissionless caller could move and Task 4 is what replaces that bound).

- [ ] **Step 1: Update the success table so the failing cases assert the new behaviour**

In `x/stakeibc/keeper/icqcallbacks_callibrate_delegation_test.go`, inside `TestCalibrateDelegation_Success`, replace the last four test cases (from `name: "negative delegation change at threshold boundary"` through the closing of `"positive delegation change exceeds threshold"`) with:

```go
		{
			// Current delegation: 12,500 tokens
			// Query response:     10,000 shares * 0.75 sharesToTokens = 7,500 tokens (-5,000)
			name:                  "large negative delegation change",
			currentDelegation:     sdkmath.NewInt(12_500),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("10000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(7_500),
		},
		{
			// Current delegation: 10,000 tokens
			// Query response:     20,000 shares * 0.75 sharesToTokens = 15,000 tokens (+5,000)
			name:                  "large positive delegation change",
			currentDelegation:     sdkmath.NewInt(10_000),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("20000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(15_000),
		},
		{
			// Current delegation: 12,501 tokens
			// Query response:     10,000 shares * 0.75 sharesToTokens = 7,500 tokens (-5,001)
			// There is no longer a cap on the change (wind-down spec §5), so this applies too
			name:                  "negative delegation change above the former cap",
			currentDelegation:     sdkmath.NewInt(12_501),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("10000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(7_500),
		},
		{
			// Current delegation: 9,999 tokens
			// Query response:     20,000 shares * 0.75 sharesToTokens = 15,000 tokens (+5,001)
			name:                  "positive delegation change above the former cap",
			currentDelegation:     sdkmath.NewInt(9_999),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("20000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(15_000),
		},
		{
			// Current delegation: 1,000,000,000 tokens (the whole zone)
			// Query response:     1,000,000,000 shares * 0.5 sharesToTokens = 500,000,000 tokens (-500,000,000)
			name:                  "delegation change of half the zone",
			currentDelegation:     sdkmath.NewInt(1_000_000_000),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("1000000000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.5"),
			expectedEndDelegation: sdkmath.NewInt(500_000_000),
		},
```

`initialTotalDelegations` at the top of the test is `1_000_000`; the last case moves `TotalDelegations` by -500,000,000, which the existing assertion computes as `initialTotalDelegations.Add(expectedDelegationChange)` and compares as `int64`, so a negative total is accepted by the test (the callback does not guard it; `TotalDelegations` is an `sdkmath.Int`). Keep it: the point of the case is that no size bound exists any more.

`TestCalibrateDelegation_Failure` has no threshold case and needs no change.

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestCalibrateDelegation_Success' -v 2>&1 | tail -15
```

Expected: FAIL on `negative delegation change above the former cap - validator delegation` (actual 12501, expected 7500).

- [ ] **Step 3: Remove the cap**

In `x/stakeibc/keeper/icqcallbacks_callibrate_delegation.go`:

Delete lines 15-16:

```go
// CalibrationThreshold is the max amount of tokens by which a calibration can alter internal record keeping of delegations
var CalibrationThreshold = sdkmath.NewInt(5000)
```

Delete the `sdkmath "cosmossdk.io/math"` import (nothing else in the file uses it after the removal; `delegatedTokens` and `delegationChange` are `sdkmath.Int` values obtained through methods, which needs no import).

Replace the block

```go
	// if the delegation change is more than the calibration threshold constant,
	// return nil so the query submission succeeds
	// Note: There should be no stateful changes above this line
	delegationChange := validator.Delegation.Sub(delegatedTokens)
	if delegationChange.Abs().GT(CalibrationThreshold) {
		k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
			"Delegation change is GT CalibrationThreshold, failing calibration callback"))
		return nil
	}
	validator.Delegation = validator.Delegation.Sub(delegationChange)
```

with

```go
	// Apply the whole difference. The 5,000 base-unit cap that used to bound this existed to
	// limit what a permissionless caller could move; MsgCalibrateDelegation is admin-only now
	// (wind-down spec §5) and the day-0 refresh needs to true up drifts of any size
	// Note: There should be no stateful changes above this line
	delegationChange := validator.Delegation.Sub(delegatedTokens)
	validator.Delegation = validator.Delegation.Sub(delegationChange)
```

- [ ] **Step 4: Run the calibration tests to verify they pass**

Run:

```bash
go build ./... && go vet ./x/stakeibc/keeper/ && go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestCalibrateDelegation' -v 2>&1 | tail -15
```

Expected: build and vet clean (vet catches an unused import if the `sdkmath` removal was missed); `TestCalibrateDelegation_Success` and `TestCalibrateDelegation_Failure` PASS.

- [ ] **Step 5: Confirm nothing else referenced the cap**

Run:

```bash
grep -rn "CalibrationThreshold" --include='*.go' x/ app/
```

Expected: exactly one line, `x/stakeibc/types/errors.go:59: ErrCalibrationThresholdExceeded = ...` (kept registered, see Global Constraints).

- [ ] **Step 6: Commit**

```bash
git add x/stakeibc/keeper/icqcallbacks_callibrate_delegation.go x/stakeibc/keeper/icqcallbacks_callibrate_delegation_test.go
git commit -m "feat(stakeibc): lift the 5,000 base-unit calibration cap

The cap bounded what a permissionless MsgCalibrateDelegation could move; the
message is admin-only for the wind-down and the day-0 refresh must true up
drifts of any size (spec §5).

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 4: Admin-gate the two ICQ messages

**Files:**
- Modify: `x/stakeibc/types/message_update_delegation.go:40-58`
- Modify: `x/stakeibc/types/message_calibrate_delegation.go:39-56`
- Create: `x/stakeibc/types/message_update_delegation_test.go`
- Create: `x/stakeibc/types/message_calibrate_delegation_test.go`

**Interfaces:**
- Consumes: `utils.ValidateAdminAddress(address string) error` (`utils/utils.go:41`), `apptesting.GetAdminAddress() (string, bool)`, `apptesting.GenerateTestAddrs() (validNonAdmin, invalid string)`.
- Produces: nothing.
- Depends on: none.
- Review: yes (auth).

- [ ] **Step 1: Write the failing tests**

Create `x/stakeibc/types/message_update_delegation_test.go`:

```go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgUpdateValidatorSharesExchRate_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChainId := "chain-0"
	validValoper := "cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p"

	tests := []struct {
		name string
		msg  types.MsgUpdateValidatorSharesExchRate
		err  string
	}{
		{
			name: "valid admin message",
			msg: types.MsgUpdateValidatorSharesExchRate{
				Creator: adminAddress,
				ChainId: validChainId,
				Valoper: validValoper,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgUpdateValidatorSharesExchRate{
				Creator: invalidAddress,
				ChainId: validChainId,
				Valoper: validValoper,
			},
			err: "invalid creator address",
		},
		{
			name: "non-admin creator",
			msg: types.MsgUpdateValidatorSharesExchRate{
				Creator: validNonAdminAddress,
				ChainId: validChainId,
				Valoper: validValoper,
			},
			err: "is not an admin",
		},
		{
			name: "missing chain id",
			msg: types.MsgUpdateValidatorSharesExchRate{
				Creator: adminAddress,
				ChainId: "",
				Valoper: validValoper,
			},
			err: "chainid is required",
		},
		{
			name: "missing valoper",
			msg: types.MsgUpdateValidatorSharesExchRate{
				Creator: adminAddress,
				ChainId: validChainId,
				Valoper: "",
			},
			err: "valoper is required",
		},
		{
			name: "valoper without the valoper prefix",
			msg: types.MsgUpdateValidatorSharesExchRate{
				Creator: adminAddress,
				ChainId: validChainId,
				Valoper: "cosmos1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgl2scj",
			},
			err: "must contrain 'valoper'",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
```

Create `x/stakeibc/types/message_calibrate_delegation_test.go`:

```go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgCalibrateDelegation_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChainId := "chain-0"
	validValoper := "cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p"

	tests := []struct {
		name string
		msg  types.MsgCalibrateDelegation
		err  string
	}{
		{
			name: "valid admin message",
			msg: types.MsgCalibrateDelegation{
				Creator: adminAddress,
				ChainId: validChainId,
				Valoper: validValoper,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgCalibrateDelegation{
				Creator: invalidAddress,
				ChainId: validChainId,
				Valoper: validValoper,
			},
			err: "invalid creator address",
		},
		{
			name: "non-admin creator",
			msg: types.MsgCalibrateDelegation{
				Creator: validNonAdminAddress,
				ChainId: validChainId,
				Valoper: validValoper,
			},
			err: "is not an admin",
		},
		{
			name: "missing chain id",
			msg: types.MsgCalibrateDelegation{
				Creator: adminAddress,
				ChainId: "",
				Valoper: validValoper,
			},
			err: "chainid is required",
		},
		{
			name: "missing valoper",
			msg: types.MsgCalibrateDelegation{
				Creator: adminAddress,
				ChainId: validChainId,
				Valoper: "",
			},
			err: "valoper is required",
		},
		{
			name: "valoper without the valoper prefix",
			msg: types.MsgCalibrateDelegation{
				Creator: adminAddress,
				ChainId: validChainId,
				Valoper: "cosmos1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgl2scj",
			},
			err: "must contrain 'valoper'",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
```

(`apptesting.SetupConfig()` sets the `stride` bech32 prefix so the admin address parses; other `message_*_test.go` files in this package rely on a `TestMain`/init that already does this. If `go test` reports `invalid creator address` for the admin case, the config was not set: keep the explicit call.)

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```bash
go test ./x/stakeibc/types/... -run 'TestMsgUpdateValidatorSharesExchRate_ValidateBasic|TestMsgCalibrateDelegation_ValidateBasic' -v 2>&1 | tail -20
```

Expected: both FAIL only on the `non-admin creator` subtest (`An error is expected but got nil`); every other subtest passes already.

- [ ] **Step 3: Add the gate to both messages**

In `x/stakeibc/types/message_update_delegation.go`, add `"github.com/Stride-Labs/stride/v34/utils"` to the imports and change `ValidateBasic` so it begins:

```go
func (msg *MsgUpdateValidatorSharesExchRate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// Admin-only for the wind-down: this message reaches the slash path, which corrects
	// delegations of any size now that the calibration cap is gone (spec §5)
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}

	// basic checks on host denom
	if len(msg.ChainId) == 0 {
```

The rest of the function is unchanged.

In `x/stakeibc/types/message_calibrate_delegation.go`, add the same import and the same block after the bech32 check:

```go
func (msg *MsgCalibrateDelegation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// Admin-only for the wind-down: this message reaches the slash path, which corrects
	// delegations of any size now that the calibration cap is gone (spec §5)
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}

	if len(msg.ChainId) == 0 {
```

The rest of the function is unchanged. The import grouping in this repo puts `github.com/Stride-Labs/stride/v34/...` in its own block after the SDK imports (see `message_add_validators.go` for the exact layout); `make lint` enforces it.

- [ ] **Step 4: Run the tests to verify they pass**

Run:

```bash
go build ./... && go test ./x/stakeibc/types/... 2>&1 | tail -5
```

Expected: build clean; `ok  	github.com/Stride-Labs/stride/v34/x/stakeibc/types`.

- [ ] **Step 5: Confirm no other caller sends these messages from a non-admin**

Run:

```bash
grep -rn "MsgUpdateValidatorSharesExchRate{\|MsgCalibrateDelegation{\|NewMsgUpdateValidatorSharesExchRate(\|NewMsgCalibrateDelegation(" --include='*.go' x/ app/ | grep -v "pb.go\|/types/message_\|/types/codec.go"
```

Expected: only `x/stakeibc/client/cli/tx.go` (lines ~601 and ~632), which take the signer from `--from` and need no change. `scripts/local-to-mainnet/commands.sh` and its template already use `--from admin`.

- [ ] **Step 6: Commit**

```bash
git add x/stakeibc/types/message_update_delegation.go x/stakeibc/types/message_calibrate_delegation.go x/stakeibc/types/message_update_delegation_test.go x/stakeibc/types/message_calibrate_delegation_test.go
git commit -m "feat(stakeibc): admin-gate MsgUpdateValidatorSharesExchRate and MsgCalibrateDelegation

Both reach the slash path, which now corrects delegations of any size; only the
protocol admin may trigger it (wind-down spec §5).

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Merge gate

After Tasks 2-4 are merged onto Task 1's commit on `wind-down-pr2-freeze-by-code`:

```bash
go build ./... && make lint && go test ./... 2>&1 | grep -v "^ok\|no test files" | tail -20
```

Expected: lint clean; the only failing package is `utils` (`TestCreateModuleAccount`, pre-existing on main). `go test ./x/stakeibc/... ./x/autopilot/... ./app/...` all `ok`.

Manual dry run (not automated, ops step recorded for the release): on localstride, upgrade with a deposit and a redemption in flight and watch three stride epochs and one day epoch: the redemption rate in `strided q stakeibc list-host-zone` must not change, no new deposit record appears in `strided q records list-deposit-record`, and the queued redemption is still submitted at the day epoch.

## Self-review

1. **Spec coverage.** §6 deleted table: all nine calls, Task 1. §6 kept table: seven calls, Task 1 (guard test asserts each is still present). §6 "Also deleted" (slash callback rate rewrite): Task 2. §5 "Messages gated": Task 4 (gates) and Task 3 (cap). §11 freeze bullet: "each deleted call absent / each kept call runs on a non-halted zone" → Task 1 guard test plus the three behavioural tests on non-halted zones; "RedemptionRate unchanged across a stride epoch with deposit records and rewards present" → `TestBeforeEpochStart_StrideEpoch_RateFrozen` (deposit records present; the reinvest pipeline that would have handled rewards is what the ICQ assertion covers); "slash-callback test that the delegation is corrected and the rate is not" → Task 2. §12 item 2 lists exactly this set.
2. **Placeholders.** None; every code step shows the code and every test the assertions.
3. **Type consistency.** `epochInfoForHook` returns `epochstypes.EpochInfo` with `Duration time.Duration`, `CurrentEpoch int64`, `CurrentEpochStartTime time.Time` (matches `x/epochs/types/genesis.pb.go`). `GetEpochTracker(ctx, identifier) (types.EpochTracker, bool)`. `GetEpochUnbondingRecord(ctx, epochNumber uint64) (EpochUnbondingRecord, bool)`. `checkModuleAccountBalance(moduleName, denom string, expected sdkmath.Int)` and `getTotalPoAValidatorStTokenBalance(denom string) sdkmath.Int` exist in `reward_allocation_test.go`. `k.GetParam(ctx, key []byte) uint64` has a pointer receiver on `*Keeper`; calling it on the value receiver `k Keeper` inside `BeforeEpochStart` is what the current code already does, so it compiles.
4. **Review tags.** All four tasks are `Review: yes`: the hook edit changes mainnet accounting behaviour, the callback and calibration edits are accounting paths, the gates are auth.
