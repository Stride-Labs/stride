# v35: upgrade authority to the multisig, undelegate all STRD

Date: 2026-10-02. Extends the v35 wind-down upgrade
(`2026-09-18-protocol-wind-down-design.md`). Verified against mainnet on 2026-10-02 via
`stride-api.polkachu.com` and against the code at `ecf073afd` (cosmos-sdk v0.54.3).

## §1. Goal

All STRD is going to be transferred off the chain, so every delegation has to be unbonded.
Once stake is gone, x/gov cannot pass anything (the tally fails on zero bonded tokens), so
the chain needs an upgrade path that does not go through governance. v35 does both:

1. Make the 2-of-3 team multisig `stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh` the
   chain-wide authority, so it can submit `MsgSoftwareUpgrade` (and any other
   authority-gated message) directly as a signed transaction.
2. Undelegate every x/staking delegation, with the stock 14-day unbonding time.
3. Close gov submission and block any new delegation so the undelegation sticks.

## §2. Facts the design rests on

- SDK v0.54 added `AuthorityParams{authority}` to consensus params (CometBFT
  `ConsensusParams.Authority`). `sdk.ValidateAuthority` (`types/authority.go:11`) prefers
  `ctx.ConsensusParams().Authority.Authority` over the keeper's own authority whenever it is
  set. The x/upgrade msg server (`x/upgrade/keeper/msg_server.go:30,45`), x/gov
  `MsgUpdateParams`, x/consensus `MsgUpdateParams` and every other SDK module use it.
  Stride's own modules (stakeibc, icaoracle, icqoracle) compare the string directly and stay
  gov-gated; nothing in the wind-down needs them.
- `MsgSoftwareUpgrade` is signed by its `authority` field. A tx from anyone else fails
  signature verification (authority = multisig) or the authority check (authority = own
  address). x/gov refuses proposals whose inner messages are not signed by the gov account
  (`x/gov/keeper/proposal.go:62`), so the gov door is closed from both sides.
- Verified on a local node (2026-10-02 spike): direct multisig `MsgSoftwareUpgrade` stores
  the plan and the node halts at the height; cancel works; the multisig can lower gov params
  back; single-key and wrong-authority submissions are rejected with the expected errors.
- `strided q consensus params` omits the authority field. Verify on mainnet via REST
  `/cosmos/consensus/v1/params` or gRPC.
- A new consensus authority takes effect the block after it is written
  (`x/consensus/keeper/keeper.go:65-68`). Irrelevant in an upgrade handler.
- Staking on mainnet: `unbonding_time` 14d, `max_entries` 7, 100 bonded / 1 unbonding /
  125 unbonded validators, ~7.03M STRD bonded, ~3.40M STRD not bonded, ~40k delegations on
  the top 40 validators alone. The POA module owns the consensus validator set; x/staking is
  still wrapped by `ccvstaking`, whose EndBlock runs `BlockValidatorUpdates` (so matured
  unbondings complete and validators with no tokens are removed) and discards the
  validator updates.
- Staking hooks are distribution and claim. `Undelegate` calls
  `BeforeDelegationSharesModified`, which withdraws rewards and errors with
  `ErrEmptyDelegationDistInfo` if the delegator starting info is missing. Undelegating a
  validator operator below its `min_self_delegation` jails the validator.
- The mainnet ICA host allow-list contains `MsgDelegate` and `MsgBeginRedelegate`.
- `removed_handlers_test.go` asserts the removed/kept rpc sets; staking messages are SDK
  rpcs and are not touched there.

## §3. Upgrade handler additions

Appended to `v35.CreateUpgradeHandler` after `ResetFailedLSMDeposit`, in this order. New
keepers passed in: `ConsensusParamsKeeper` (value; `ParamsStore` is the collections item),
`GovKeeper`, `StakingKeeper` (value), reusing the existing `*icahostkeeper.Keeper`.

1. **`SetConsensusAuthority(ctx, consensusParamsKeeper) error`** — `ParamsStore.Get`, set
   `Authority = &cmtproto.AuthorityParams{Authority: UpgradeAuthority}` (the field is nil on
   mainnet today), `ParamsStore.Set`. Returns the error: this step fails the upgrade,
   because skipping it would leave the chain with no upgrade path once stake is gone.
   `UpgradeAuthority` is a constant in `constants.go`.
2. **`CloseGovSubmission(ctx, govKeeper) error`** — set `MinDeposit` to
   `GovUnreachableDeposit` = 1e18 ustrd (total supply is ~3.78e13 ustrd) and
   `ExpeditedMinDeposit` to `GovUnreachableExpeditedDeposit` = 2e18 ustrd (gov params
   validation requires expedited to be strictly greater than min). Every other gov param is kept as read. Returns the error (param write).
3. **`RaiseMaxUnbondingEntries(ctx, stakingKeeper) error`** — set staking `MaxEntries` to
   100. A pair can hold 7 entries today and the handler adds one; with delegation blocked
   afterwards (§4) no pair can approach 100. Returns the error.
4. **`RemoveStakingFromICAHostAllowList(ctx, icaHostKeeper)`** — same filter-in-place shape
   as `RemoveStakeibcFromICAHostAllowList`, dropping the four `BlockedStakingMsgTypeUrls`
   (§4). Logs each removal; cannot fail.
5. **`UndelegateAllDelegations(ctx, stakingKeeper)`** — `GetAllDelegations`, then for each
   delegation `Undelegate(delegator, validator, delegation.Shares)`. On error, log
   `v35: skipping undelegation of <delegator> from <validator>: <err>` and continue. Track
   and log a final summary: undelegated count, skipped count, total ustrd returned to the
   not-bonded pool. No events, no per-delegation success log (tens of thousands of lines).
   Deterministic order comes from store iteration. Expected and accepted side effects:
   rewards withdrawn to each delegator by the distribution hook; operators jailed when
   self-delegation drops below minimum; every bonded validator left at zero tokens, moved to
   unbonding by the EndBlocker and removed after 14 days; every already-unbonded validator
   (125 on mainnet) removed on the spot by `Unbond` once its shares hit zero. Either way the
   distribution `AfterValidatorRemoved` hook pays out commission at removal. A failure to
   list the delegations fails the upgrade. In-flight unbondings and redelegations are
   untouched and complete on their own clocks.

Handler doc comment and the ordered step list get entries 10–14.

## §4. Ante decorator

`app/ante_blocked_msgs.go`: `BlockedMsgsDecorator` holding a `map[string]bool` of type
URLs, constructed from `v35.BlockedStakingMsgTypeUrls` = `MsgDelegate`, `MsgBeginRedelegate`,
`MsgCreateValidator`, `MsgCancelUnbondingDelegation` (the one list, in `constants.go`). `AnteHandle` walks `tx.GetMsgs()`;
for an authz `MsgExec` it unpacks `GetMessages()` and walks those too (recursively, since an
exec can nest). Any hit rejects the whole tx with `sdkerrors.ErrUnauthorized` wrapping
`"<type url> is disabled: the chain is winding down"`. Simulation and check are treated the
same as deliver. Wired in `NewAnteHandler` right after `NewValidateBasicDecorator`. The
same list drives §3 step 4 so the ICA host path is closed too. CosmWasm staking messages
from contracts are the one path left open; no Stride contract is known to stake.

## §5. Tests

v35 suite (`upgrades_test.go` plus a new `staking_test.go` / `authority_test.go`, following
the one-file-per-step pattern):

- `SetConsensusAuthority`: authority nil before, equals `UpgradeAuthority` after, every
  other consensus param unchanged; a second call is idempotent.
- `CloseGovSubmission`: both deposits equal 1e18 ustrd, voting period and thresholds unchanged.
- `RaiseMaxUnbondingEntries`: 100 after; unbonding time unchanged.
- `RemoveStakingFromICAHostAllowList`: the four URLs removed, `MsgSend` and
  `MsgUndelegate` kept, empty list is a no-op.
- `UndelegateAllDelegations`: seed three validators (one bonded, one unbonded, one with
  `min_self_delegation` above its post-undelegation self-stake) the way
  `v33/upgrades_test.go setupGovenatorState` builds validator records, but create the
  delegations through `StakingKeeper.Delegate` from funded accounts so the distribution and
  claim hooks run and the delegator starting info exists (a bare `SetDelegation` would make
  every undelegation fail the rewards withdrawal). Include an operator self-delegation.
  Assert: no delegations remain, one unbonding entry per pair with
  completion = block time + unbonding time and balance = the delegated tokens, validator
  tokens zero, the operator jailed, a pre-existing unbonding entry untouched (its
  completion time unchanged), and a delegation whose distribution starting info was deleted
  is skipped while the rest still complete. A pair with 7 existing entries succeeds after
  step 3 (run step 3 then step 5 in that test).
- `TestUpgrade` and `TestUpgrade_EmptyState` extended to run the whole handler with the new
  keepers and assert the authority, gov and staking params.

Mainnet export suite: after replay, assert the authority and the gov deposits. No staking
fixture (tens of thousands of delegations would bloat it).

Ante suite (`app/ante_blocked_msgs_test.go`): each of the four messages rejected at the top
level and inside `MsgExec`; `MsgUndelegate`, `MsgWithdrawDelegatorReward` and `MsgSend`
pass through.

## §6. Out of scope

- Shipping the multisig sign/multisign/broadcast scripts from the spike. v35 itself still
  passes through gov; the first direct upgrade is v36, and the scripts land before it.
- Force-completing in-flight unbondings.
- Blocking CosmWasm staking messages.
- Switching Stride's own modules to `sdk.ValidateAuthority`.

## §7. Build plan

Branch `v35-authority-undelegate` from `main` at `ecf073afd`.

**Chunk A — handler steps and tests.** Depends on: none.
- Files: `app/upgrades/v35/constants.go` (`UpgradeAuthority`, `GovUnreachableDeposit`,
  `StakingMaxEntries`), new `app/upgrades/v35/authority.go` (steps 1–2),
  `app/upgrades/v35/staking.go` (steps 3–5), `app/upgrades/v35/entry_points.go` (step 4),
  `app/upgrades/v35/upgrades.go` (signature, doc list, calls), `app/upgrades.go` (wiring:
  `app.ConsensusParamsKeeper`, `app.GovKeeper`, `app.StakingKeeper`), tests per §5,
  `mainnet_export_test.go` assertions, CHANGELOG entry under Unreleased item 3's stanza.
- Interfaces consumed: `consensusparamkeeper.Keeper.ParamsStore`, `govkeeper.Keeper.Params`,
  `stakingkeeper.Keeper.{GetParams,SetParams,GetAllDelegations,Undelegate}`,
  `icahostkeeper.Keeper.{GetParams,SetParams}`. Interface produced:
  `v35.BlockedStakingMsgTypeUrls`, one `[]string` in `constants.go` read by step 4 and by
  chunk B's decorator (the app package already imports v35, so there is no cycle).
- Key decisions: step 1–3 errors fail the upgrade; step 5 skips and logs per delegation;
  no per-delegation success logging.

**Chunk B — ante decorator and tests.** Depends on: chunk A's `BlockedStakingMsgTypeUrls`
constant (one slice; build B after A's constants file exists, or define it first).
- Files: `app/ante_blocked_msgs.go`, `app/ante_blocked_msgs_test.go`, `app/ante_handler.go`
  (one line in the decorator chain).
- Key decisions: reject on any blocked message in the tx, recurse into `MsgExec`, same
  behaviour in check/simulate/deliver.

Final: run `go build ./...`, `go test ./app/...`, confirm `TestMainnetExportTestSuite`
still passes with the committed fixture, open the PR against `main`, remove the spike
worktree `spike/direct-upgrade-authority`.
