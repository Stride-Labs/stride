# Stride v33 — ICS → POA Migration Design

## §1. Overview and objective

Migrate Stride from ICS consumer + govenator staking (the "democracy" pattern) to POA + govenator staking. The block-producing validator set moves from ICS-pushed to POA-admin-controlled; everything else about Stride (govenators, STRD delegation, staking rewards, governance, liquid staking product, tokenomics) stays functionally the same.

**This spec covers v33 only — the migration upgrade.** A follow-up cleanup binary (v34) is planned but **out of scope for this implementation plan** and lives in a separate spec (`2026-04-27-ics-cleanup-design.md`). v33 implementation work must not include any v34 changes.

At a high level, v33 does the migration with old ICS keepers left mounted but inert. v34 (separate plan) drops those keepers, deletes the store keys, replaces the `ccvstaking` wrapper with an in-tree equivalent, and removes the `interchain-security` go dependency.

**Pre-conditions:**

- Stride is on SDK v0.54 (assumed for this design).
- Cosmos Labs has granted Stride permission to use the cosmos-sdk POA module despite its Source Available license.
- An admin multisig bech32 address exists on Stride's address codec (must be a valid address; does not need to be operationally signing-ready until the first validator change, expected ≥1 year out).
- A consensus list of 8 validator monikers and their consensus addresses is finalized.

**Success criteria for v33:**

1. Block N (upgrade block) signs with the current 8 ICS validators.
2. Block N+1 signs with the same 8 validators, now produced by `x/poa`.
3. No interruption to block production.
4. Govenators, delegators, and their bonded STRD are undisturbed.
5. Staking rewards continue flowing from mint → `x/distribution` → govenators → delegators.
6. Governance proposals tally correctly based on bonded STRD (unchanged behavior).
7. Stride's liquid staking product (`stakeibc`, etc.) is unaffected.

**Non-goals for v33:**

- Changing tokenomics (inflation rate, distribution proportions).
- Introducing slashing for POA validators.
- Cleaning up ICS keepers or store state — deferred to v34.
- Closing the CCV IBC channel explicitly — allowed to time out naturally.
- Rotating validator consensus keys.

## §2. Architectural changes in v33 (app.go)

### Module manager changes

| Module | Today | v33 | Reason |
|---|---|---|---|
| `x/ccv/consumer` | registered | **removed** | No longer driving validator set |
| `x/ccv/democracy/distribution` (`ccvdistr`) | registered | **removed** | 85/15 split logic no longer needed; use plain `x/distribution` |
| `x/ccv/democracy/staking` (`ccvstaking`) | registered | **kept** | Still suppresses x/staking validator updates to CometBFT |
| `x/staking` | wrapped by ccvstaking | **unchanged** (still wrapped) | Govenators + delegators keep running |
| `x/distribution` | wrapped by ccvdistr | **used directly** (unwrapped) | Standard flow: `fee_collector` → validators |
| `x/slashing` | registered | **removed** | POA validators aren't slashed; govenators don't sign blocks |
| `x/evidence` | registered | **removed** | No slashing backend to route evidence to |
| `x/poa` | — | **added** | New block-producer module |
| Everything else (`x/bank`, `x/gov`, `x/mint`, `stakeibc`, `x/claim`, `x/wasm`, IBC stack, etc.) | registered | **unchanged** | Independent of validator-set mechanism |

### Keeper wiring changes

**`x/distribution` keeper** — change fee-collector-name argument:

```go
app.DistrKeeper = distrkeeper.NewKeeper(
    appCodec,
    ...,
    authtypes.FeeCollectorName, // was: ccvconsumertypes.ConsumerRedistributeName
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
```

**`x/poa` keeper** — new, requires KV + transient store + account + bank keepers:

```go
app.POAKeeper = poakeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[poatypes.StoreKey]),
    runtime.NewTransientStoreService(tkeys[poatypes.TransientStoreKey]),
    app.AccountKeeper,
    app.BankKeeper,
)
```

Add `poatypes.StoreKey` and `poatypes.TransientStoreKey` to store mounts. Use `poa.NewAppModule(appCodec, app.POAKeeper, poa.WithSecp256k1Support())` if any of the 8 validators use secp256k1 consensus keys (TBD during validator coordination).

**`x/slashing` and `x/evidence` keepers** — leave constructed in app.go for v33 (for handler access if needed) but don't register the modules. Both currently take `&app.ConsumerKeeper` as their staking-keeper interface argument; this is fine because the modules aren't running. Delete in v34.

**`x/ccv/consumer` keeper** — leave constructed, don't register. Handler reads `GetAllCCValidator`. Delete in v34.

**`x/gov` keeper** — **no change**. Still uses staking-based tally since govenators/delegators drive governance voting. Do NOT swap to POA's tally function.

**Staking hooks** — current hooks (`DistrKeeper.Hooks()`, `ClaimKeeper.Hooks()`) are unchanged. Slashing was never a hook on staking (it was driven off ICS events), so no hook removal needed.

**Standalone staking keeper** — `app.ConsumerKeeper.SetStandaloneStakingKeeper(app.StakingKeeper)` — keep for v33, remove in v34. This was for pre-CCV slashing of historical infractions, irrelevant post-migration.

### Ante handler change

Tx fee recipient changes from `fee_collector` → POA module account. POA's `EndBlock` at height 1 panics if this isn't done — it's a hard requirement of the POA module's design. Specific change is in `app/ante_handler.go` — replace the `DeductFeeDecorator`'s fee collector name.

### Begin/end blocker ordering

**Add** `poatypes.ModuleName` to `SetOrderBeginBlockers` and `SetOrderEndBlockers`. Per POA docs, POA should be **last** in EndBlock so gov tally runs before validator updates propagate.

**Remove** `ccvconsumertypes.ModuleName`, `slashingtypes.ModuleName`, `evidencetypes.ModuleName` from both ordering lists.

**Also update** `SetOrderInitGenesis`: add `poatypes.ModuleName`, remove the three deleted modules. Init genesis order isn't invoked during an upgrade (only during a fresh chain start), but keeping it consistent prevents surprises if someone ever rebuilds genesis from scratch.

### Store upgrades (`StoreUpgrades` in the upgrade plan)

```go
storetypes.StoreUpgrades{
    Added: []string{
        poatypes.StoreKey,
        poatypes.TransientStoreKey,
    },
    // Deleted: nothing in v33. All deletions in v34.
}
```

### Module account permissions

Add POA module account to `maccPerms`:

```go
poatypes.ModuleName: nil, // fee distribution account; no mint/burn
```

`cons_redistribute` and `cons_to_send_to_provider` stay in `maccPerms` through v33 (we drain their balances in the handler but keep the accounts registered until v34).

### VersionMap handling

The handler does **not** call `delete(vm, ...)` for removed modules. `RunMigrations` silently skips modules not in the manager — verified against SDK source and Neutron v6.0.0 precedent. The stale entries persist harmlessly in on-chain VersionMap state. Skipping the delete matches Stride's existing upgrade patterns (no other Stride upgrade has used this pattern except v5, which used it as a workaround for an authz state-sync bug).

## §3. The v33 upgrade handler

### Handler signature

```go
func CreateUpgradeHandler(
    mm *module.Manager,
    configurator module.Configurator,
    cdc codec.Codec,
    consumerKeeper ccvconsumerkeeper.Keeper,
    poaKeeper *poakeeper.Keeper,
    bankKeeper bankkeeper.Keeper,
    accountKeeper authkeeper.AccountKeeper,
    distrKeeper distrkeeper.Keeper,
) upgradetypes.UpgradeHandler
```

Keepers needed: `consumerKeeper` (snapshot + sweep), `poaKeeper` (init), `bankKeeper` + `accountKeeper` (balance moves), `distrKeeper` (community pool). Everything else — staking, slashing, gov — is not touched in the handler.

### Handler body

```
func(ctx, plan, vm) {
    logger.Info("Starting upgrade v33 (ICS → POA)...")

    // Step 1: run module migrations normally.
    //         RunMigrations silently skips modules removed from the manager.
    versionMap = mm.RunMigrations(ctx, configurator, vm)

    // Step 2: snapshot the current validator set from ccvconsumer.
    //         This is the source of truth for what CometBFT is currently using.
    poaValidators = snapshotValidatorsFromICS(ctx, consumerKeeper)

    // Step 3: seed POA state with that set + admin multisig.
    err = initializePOA(ctx, poaKeeper, AdminMultisigAddress, poaValidators)

    // Step 4: sweep residual ICS module-account balances.
    //         These accounts no longer have a consumer to drain them each block.
    err = sweepICSModuleAccounts(ctx, bankKeeper, accountKeeper, distrKeeper)

    logger.Info("v33 upgrade complete")
    return versionMap, nil
}
```

### Step 2 — `snapshotValidatorsFromICS`

Read the live CCV validator set and convert each entry to a POA `Validator`:

```
func snapshotValidatorsFromICS(ctx, consumerKeeper) []poatypes.Validator:
    ccValidators = consumerKeeper.GetAllCCValidator(ctx)

    result = []
    for each ccVal in ccValidators:
        pubkey = ccVal.Pubkey                          // already *codectypes.Any wrapping ed25519.PubKey
        operatorAddr = bech32.ConvertAndEncode(        // derive from cons-addr bytes
            "stridevaloper",
            ccVal.GetAddress(),
        )
        moniker = lookupMoniker(ccVal.GetAddress())    // pre-baked map; see below

        result.append(poatypes.Validator{
            PubKey: pubkey,
            Power: ccVal.Power,
            Metadata: &poatypes.ValidatorMetadata{
                Moniker: moniker,
                OperatorAddress: operatorAddr,
            },
        })

    // Invariant: exactly 8 validators (PSS allowlist size). Panic if not.
    require len(result) == 8

    return result
```

**Moniker lookup**: ICS doesn't store monikers. Pre-bake a map from consensus address → moniker in the v33 constants file. The 8 PSS allowlisted validators are known ahead of upgrade. If a consensus address isn't in the map, fall back to an empty moniker. Format roughly like the validator_weights files Stride already uses in upgrade dirs:

```go
var ValidatorMonikers = map[string]string{
    "<consaddrbytes_hex_1>": "Stride Labs",
    "<consaddrbytes_hex_2>": "...",
    // 8 entries
}
```

### Step 3 — `initializePOA`

Synthesize a `GenesisState` and call POA's `InitGenesis`:

```
func initializePOA(ctx, poaKeeper, adminAddr, validators) error:
    genesisState = poatypes.GenesisState{
        Params: poatypes.Params{
            Admin: adminAddr,  // multisig bech32
        },
        Validators: validators,
    }

    // POA's InitGenesis requires block height 0 — documented in the source.
    err = poaKeeper.InitGenesis(ctx.WithBlockHeight(0), genesisState)
    return err
```

**Why `WithBlockHeight(0)`**: POA's `InitGenesis` internally uses block-height-0-specific code paths for checkpoint init. Documented in `enterprise/poa/x/poa/keeper/genesis.go`. Easy to forget and silently corrupt state.

**On validator powers**: the handler seeds POA with each validator's *current* ICS-assigned power (whatever `GetAllCCValidator` returns). This guarantees POA's first `EndBlock` produces no diff against CometBFT's existing set — the cleanest possible handoff.

ICS-assigned powers reflect Hub-side bonded ATOM, which has no meaning on a permissioned POA chain. Long-term Stride probably wants all 8 validators at equal power (e.g., `1` each). Two paths to get there:

- **A — rebalance in the handler.** After `InitGenesis`, immediately submit synthesized `MsgUpdateValidators` (or directly modify keeper state) to set all powers to `1`. ~5 lines of code. Safe — CometBFT applies the change one block later via the standard ABCI lag; the old set still signs block N+1, the rebalanced set takes over at N+2. Avoids any post-upgrade multisig coordination.
- **B — rebalance later via multisig.** Handler leaves powers as inherited; multisig submits `MsgUpdateValidators` whenever convenient. Exercises POA's runtime path in production rather than just at upgrade time.

Either is fine; pick based on operational preference. **A** is recommended for simplicity (no MS signing needed and tests covered by the same upgrade test suite).

After this call, POA's KV store has the 8 validators with their power and keys. The next `EndBlock` will process them.

### Step 4 — `sweepICSModuleAccounts`

Two module accounts need handling:

```
func sweepICSModuleAccounts(ctx, bankKeeper, accountKeeper, distrKeeper) error:
    for each moduleName in [cons_to_send_to_provider, cons_redistribute]:
        moduleAddr = accountKeeper.GetModuleAddress(moduleName)
        balance = bankKeeper.GetAllBalances(ctx, moduleAddr)
        if balance.IsZero(): continue

        // Send to community pool via the distribution keeper.
        err = distrKeeper.FundCommunityPool(ctx, balance, moduleAddr)
```

**Policy**: residuals go to community pool. Two reasons: (1) `cons_to_send_to_provider` was already earmarked for external security, community pool is the neutral home; (2) `cons_redistribute` residuals are minuscule (distribution drains it every block).

**Note**: the exact API for `FundCommunityPool` in SDK v0.54 may differ slightly (module account vs address) — verify in implementation.

### What the handler deliberately does NOT do

- Does **not** call `delete(vm, ...)`. (Per §2 and the SDK research.)
- Does **not** drain `x/staking` pools. Govenators + delegators are preserved.
- Does **not** force-unbond delegations. Real user stake stays bonded.
- Does **not** fail active gov proposals. Gov tally still uses staking; proposals continue normally.
- Does **not** close the CCV IBC channel. Left to time out.
- Does **not** clear `ccvconsumer`'s KV store (pending VSC changes, outstanding downtime flags, etc.). State becomes dead weight; cleaned up in v34 via `Deleted`.
- Does **not** touch `x/slashing` or `x/evidence` state. Their stores become dead weight; cleaned up in v34.

### Key invariants the upgrade handler must preserve

1. `len(poaValidators) == 8` — exact PSS allowlist size; panic if not.
2. Every `poaValidator.PubKey` equals a pubkey currently in CometBFT's validator set. If POA emits a set different from what CometBFT has, the chain halts at N+1.
3. `AdminMultisigAddress` is a valid bech32 on Stride's address codec. Validate at handler entry.
4. All 8 validators have non-zero power after snapshot (a validator with power 0 would get dropped by POA's bonding logic).

## §4. Future v34 cleanup (separate plan, not part of this work)

After v33 ships and stabilizes (weeks-to-months later), a follow-up cleanup upgrade will:

- Delete `ConsumerKeeper`, `SlashingKeeper`, `EvidenceKeeper` from app.go.
- Delete the corresponding store keys (paired with `MountStores` removals).
- Replace `ccvstaking.NewAppModule(...)` with an in-tree wrapper (~50 lines, no ICS imports).
- Drop the `github.com/cosmos/interchain-security/v7` dep from `go.mod`.

**v34 is intentionally out of scope for this spec and the v33 implementation plan.** It has its own design document and will get its own implementation plan when v33 has stabilized on mainnet.

The reason for the split: v33 is the high-risk validator-set migration; v34 is pure housekeeping. Deferring v34 means the v33 implementation can stay tightly focused on the consensus-layer change without scope creep into module/dependency cleanup. It also means we get real-world post-v33 feedback before committing to v34's exact shape.

For the curious, see `2026-04-27-ics-cleanup-design.md` for the v34 plan. Implementers of v33 should not consult that document — it would only be a distraction.

## §5. Rewards flow before and after

### Before (Stride on ICS today)

```
┌──────────────────────────────────────────────────────────────────┐
│ x/mint  (hourly epoch)                                           │
│   mints STRD inflation, splits 4 ways:                           │
│     • 27.64% "Staking"           → fee_collector                 │
│     • 42.05% StrategicReserve    → stride1alnn79kh...            │
│     • 18.60% CommunityPoolGrowth → <submodule>                   │
│     • 11.71% SecurityBudget      → <submodule>                   │
└──────────────────────────────────────────────────────────────────┘

  fee_collector  ◄── tx fees (via ante handler)

       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ ccvconsumer.EndBlock                                             │
│   drains fee_collector, splits by ConsumerRedistributionFrac:    │
│     • 85% → cons_redistribute                                    │
│     • 15% → cons_to_send_to_provider  → IBC to Hub (every N blk) │
└──────────────────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ x/distribution (wrapped by ccvdistr, fee pool = cons_redistribute)│
│   BeginBlock: allocate cons_redistribute balance to govenators   │
│     → govenator commission → govenator operator account          │
│     → remainder → delegator rewards (claimable via Withdraw)     │
└──────────────────────────────────────────────────────────────────┘
```

### After (Stride on POA, post-v33)

```
┌──────────────────────────────────────────────────────────────────┐
│ x/mint  (hourly epoch) — UNCHANGED                               │
│   mints STRD inflation, splits 4 ways:                           │
│     • 27.64% "Staking"           → fee_collector                 │
│     • 42.05% StrategicReserve    → stride1alnn79kh...            │
│     • 18.60% CommunityPoolGrowth → <submodule>                   │
│     • 11.71% SecurityBudget      → <submodule>                   │
└──────────────────────────────────────────────────────────────────┘

  fee_collector                     poa_module_account ◄── tx fees
                                                          (via ante handler)
       │                                   │
       ▼                                   ▼
┌────────────────────────┐    ┌──────────────────────────────────┐
│ x/distribution         │    │ x/poa (lazy checkpoint model)    │
│   (unwrapped, standard)│    │   accrues tx fees in its account │
│   fee pool = fee_      │    │   allocates per-validator on:    │
│   collector            │    │     • power changes              │
│   → govenators         │    │     • MsgWithdrawFees            │
│   → delegators         │    │   → POA validators (8 of them)   │
└────────────────────────┘    └──────────────────────────────────┘
```

### Liquid-staking revenue flow (orthogonal to migration)

Stride's liquid staking product has its own revenue flow that is **completely independent** of the validator-set mechanism:

```
┌────────────────────────────────────────────────────────────────┐
│ x/stakeibc — LS fees from host zones (IBC denoms)              │
│   Land in RewardCollector module account                       │
└────────────────────────────────────────────────────────────────┘
       │
       ▼
┌────────────────────────────────────────────────────────────────┐
│ stakeibc.AuctionOffRewardCollectorBalance                      │
│   Splits by PoaValPaymentRate (0.15):                          │
│     • 15% → liquid-staked → stTokens → 8 hardcoded addresses   │
│             (utils.PoaValidatorSet)                            │
│     • 85% → x/auction module                                   │
└────────────────────────────────────────────────────────────────┘
       │
       ▼
┌────────────────────────────────────────────────────────────────┐
│ x/auction (FCFS) → buyers pay STRD → strdburner                │
└────────────────────────────────────────────────────────────────┘
       │
       ▼
┌────────────────────────────────────────────────────────────────┐
│ x/strdburner.EndBlocker — burns all STRD it holds              │
└────────────────────────────────────────────────────────────────┘
```

**Effect of migration: zero.** This flow uses hardcoded constants (`utils.PoaValidatorSet`, `utils.PoaValPaymentRate`) and doesn't touch any consensus-layer module.

**Non-blocking cleanup for later (not v33)**: replace `utils.PoaValidatorSet` with a dynamic lookup via `poaKeeper.GetAllValidators()` so multisig-driven validator changes propagate automatically. Today the list is manually maintained.

### Net economic picture

| Revenue stream | Recipient today | Recipient post-v33 |
|---|---|---|
| STRD inflation — 72% (strategic reserve, community pools) | Same fixed addresses | **Same, unchanged** |
| STRD inflation — 27.64% staking portion × 85% local | Govenators + delegators | Govenators + delegators (gets 100%, no 15% leak) |
| STRD inflation — 27.64% staking portion × 15% Hub | Hub via IBC | Govenators + delegators (gained) |
| Stride tx fees | Govenators + delegators via ccvconsumer split | POA validators via POA lazy distribution |
| LS revenue — 15% (via stakeibc) | 8 hardcoded addresses (as stTokens) | **Same, unchanged** |
| LS revenue — 85% (via stakeibc → auction) | Buyback-and-burn (STRD) | **Same, unchanged** |

### Concrete code changes

1. `x/mint`: no change. Keep routing 27.64% staking-portion to `fee_collector`.
2. `x/distribution` keeper construction: change `feeCollectorName` param from `ccvconsumertypes.ConsumerRedistributeName` → `authtypes.FeeCollectorName`.
3. Drop `ccvdistr.NewAppModule(...)` from module manager; use plain `distr.NewAppModule(...)`.
4. Ante handler: change tx-fee destination from `authtypes.FeeCollectorName` → `poatypes.ModuleName`.
5. Upgrade handler step 4: sweep residual balances in `cons_to_send_to_provider` and `cons_redistribute` to community pool.

## §6. Testing strategy

All tests are standard go tests run via `make test-unit`. Broader integration testing (k8s network, multi-node, public testnet rehearsals) is intentionally deferred — those are a separate, much larger workstream and outside this plan.

### Test 1 — helper unit tests

Pure-function helpers from §3, tested in isolation with mock keepers. Same pattern as Stride's existing keeper tests.

- **`snapshotValidatorsFromICS`** — fixtures cover: 8-validator happy path, validator with power=0 (assert behavior), unknown consensus address not in moniker map (assert empty moniker fallback), bech32-encoding of operator address.
- **`initializePOA`** — assert `ctx.WithBlockHeight(0)` is used, POA state populated correctly, admin set.
- **`sweepICSModuleAccounts`** — pre-fund both module accounts, assert post-sweep balances zero and community pool received the right amount.

### Test 2 — upgrade handler test with synthetic state

Same shape as `app/upgrades/v14/upgrades_test.go` and other existing upgrade tests. Embed `apptesting.AppTestHelper`, seed dummy state, run upgrade, assert.

```go
type UpgradeTestSuite struct {
    apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) TestUpgrade() {
    s.SetupICSValidatorSet()        // seed 8 CCValidators in ConsumerKeeper
    s.SetupGovenatorState()          // seed staking validators + delegations
    s.SetupConsumerRewards()         // pre-fund cons_redistribute, cons_to_send_to_provider
    s.SetupActiveGovProposal()       // confirm proposals continue post-upgrade

    s.ConfirmUpgradeSucceeded(v33.UpgradeName)

    s.CheckPOAValidatorsMatchICSSnapshot()
    s.CheckPOAAdminSet()
    s.CheckGovenatorStateUntouched()
    s.CheckICSModuleAccountsDrained()
    s.CheckActiveGovProposalUnaffected()
    s.CheckDistrFeeCollectorRewired()
    s.CheckValidatorSetContinuity()  // most important
}
```

The `CheckValidatorSetContinuity` assertion is load-bearing: post-handler POA `EndBlock` produces a `ValidatorUpdate` set that, when applied, equals the pre-handler CometBFT set. This catches the chain-halt bug in CI.

Should also include a scenario that triggers `AuctionOffRewardCollectorBalance` post-upgrade and asserts the full LS revenue flow (including final STRD burn).

### Test 3 — upgrade handler test seeded from a mainnet state export

Same test framework as Test 2, but instead of synthetic setup, the suite loads a real mainnet state export.

Mechanics:

1. Run `strided export --height N` against a recent mainnet node to produce `genesis.json`.
2. Commit a sanitized version of that file (or a small helper that loads it from a fixture path) into the test fixtures.
3. In `SetupTest()`, load the full StrideApp from the real export rather than from defaults.
4. Run `s.ConfirmUpgradeSucceeded(v33.UpgradeName)`.
5. Run all the post-upgrade assertions from Test 2, plus a few that only make sense against real state:
   - **All real govenators still have their full delegation totals.**
   - **Every active mainnet gov proposal still exists post-upgrade.**
   - **Real `cons_redistribute` mainnet balance moved to community pool.**
   - **Stride's vesting accounts, claim records, stakeibc host zones, etc. are all unaffected.**

This is the test that catches "I forgot mainnet has weird state from upgrades 1-31" bugs. The fixture file may be 100MB+; commit or fetch at test time.

Stride doesn't appear to do this pattern today — it's worth flagging as new test infrastructure to build for v33, reusable for future risky upgrades.

### Specific assertions to bake into CI

These should fail loudly before any human reviews a v33 PR:

1. **Validator set continuity**: post-handler POA `EndBlock` produces a set whose `ValidatorsHash` equals the pre-handler CometBFT `ValidatorsHash`.
2. **No accidental govenator drain**: total bonded STRD in `bonded_tokens_pool` post-handler == pre-handler.
3. **No accidental delegator drain**: delegation count post-handler == pre-handler.
4. **POA admin is set**: `poaKeeper.GetParams(ctx).Admin == ExpectedAdminMultisigAddress`.
5. **ICS module accounts drained**: `bankKeeper.GetAllBalances(consRedistributeAddr).IsZero()` and same for `consToSendToProviderAddr`.
6. **Distribution rewired correctly**: `distrKeeper.feeCollectorName == authtypes.FeeCollectorName`.
7. **Module manager invariants**: `ccvconsumer`, `slashing`, `evidence` not in `mm.Modules`; `poa` is in `mm.Modules`; `ccvstaking` (for now) is in `mm.Modules`.

### Things deliberately NOT in this test plan

- Multi-node localnet rehearsal (dockernet or otherwise).
- k8s integration tests (`./integration-tests`) — separate workstream.
- Public testnet rehearsal — Stride doesn't have one.
- Multisig signing rehearsal — operational, not in test plan.

## §7. Risks and mitigations

### Risk 1: Validator-set discontinuity at upgrade block

**The single biggest risk.** If POA's `EndBlock` at block N+1 emits a `ValidatorUpdate` set that differs from CometBFT's current set, CometBFT halts. Recovery requires a coordinated hotfix.

- **Likelihood**: Low if snapshot logic is correct.
- **Impact**: Catastrophic — chain halt, hours-to-days of downtime.
- **Mitigations**:
  - `CheckValidatorSetContinuity` assertion in Test 2 + Test 3.
  - Handler defensive comparison: `len(poaValidators) == 8`, every pubkey non-nil, every power > 0. Panic at handler time is recoverable; chain halt at N+1 is not.
  - Each of the 8 validators runs the v33 binary against the mainnet state export on their own infrastructure before the upgrade height.

### Risk 2: Real mainnet state surprises

We don't know exactly what's in `x/staking`, `x/distribution`, `x/slashing`, `x/evidence` stores from accumulated state across 31 upgrades.

- **Likelihood**: Moderate. Mainnet state is messy by definition.
- **Impact**: Cosmetic to severe.
- **Mitigations**:
  - **Test 3 (mainnet state export) is the primary mitigation.**
  - Inventory exercise as a v33 prerequisite: dump each module's KV store from a mainnet node, document what's there.
  - Handler is defensive: every keeper read wraps in error handling and logs; no naked panics on unexpected state.

### Risk 3: Module dependency we missed

Some Stride or third-party module references `SlashingKeeper` / `EvidenceKeeper` / `ConsumerKeeper` in a way we didn't catch.

- **Likelihood**: Low-moderate. Stride has many custom modules.
- **Impact**: Build error (best case) or runtime panic post-upgrade (worse).
- **Mitigations**:
  - **Pre-implementation audit**: grep every module for references to `SlashingKeeper`, `EvidenceKeeper`, `ConsumerKeeper`, `ccvconsumer`, `ccvdistr`, `ccvstaking`. Document each finding and decide: keep, rewire, or remove.
  - Test 3 against mainnet state catches runtime issues triggerable from existing state.

### Risk 4: ICS keeper state we left in place causes trouble

We're keeping `ConsumerKeeper` mounted with all its data.

- **Likelihood**: Low. We're removing ccvconsumer from the module manager; nothing should be calling its keeper methods.
- **Impact**: Mild.
- **Mitigations**: Audit grep step (Risk 3) covers this. v34 deletes the store entirely.

### Risk 5: CCV IBC channel not closing causes operational confusion

Letting it time out rather than explicitly closing.

- **Likelihood**: Certain (by design).
- **Impact**: Cosmetic; ops noise.
- **Mitigations**: Document expected channel state for ops/validators. v34 (or later) can call `ChanCloseInit` if cleanup desired.

### Risk 6: Multisig admin address invalid at upgrade time

POA admin must be a valid bech32 address at upgrade time. The multisig does **not** need to be operationally signing-ready until the first validator change (≥1 year out per current understanding). The address itself just needs to exist and be valid.

- **Likelihood**: Low.
- **Impact**: Severe if invalid — can't change admin without another upgrade.
- **Mitigations**:
  - v33 handler panics if `AdminMultisigAddress` fails bech32 validation — fail fast at handler time rather than silently.
  - Multisig address finalized and reviewed in v33 constants file before mainnet release.
  - Operational readiness of the multisig is a separate workstream not blocking v33.

### Risk 7: ccvstaking incompatibility with ccvconsumer removed (DOWNGRADED)

Initial concern: ccvstaking is part of the ICS package and might have hidden dependencies on ccvconsumer.

**Investigation finding (low risk)**: `ccvstaking` has zero imports from `x/ccv/consumer`. It's a ~80-line wrapper that just discards `InitGenesis` and `EndBlock` validator updates. No IBC dependencies, no CCV channel reads, no consumer-state reads. Compiles and runs correctly with or without `ccvconsumer` in the module manager.

- **Likelihood**: Low.
- **Impact**: If hidden incompatibility is discovered during implementation, the fallback (writing an in-tree wrapper as part of v33 instead of v34) is a small mitigation rather than a blocker. **However**, attempting that fallback should trigger a re-scoping conversation, not a silent expansion of v33 — the discipline of "v34 is separate" matters more than saving a few weeks.
- **Mitigations**:
  - Ship v33 with `ccvstaking` as planned.
  - If implementation surfaces issues, escalate before implementing the fallback.

### Risk 8: Liquid staking revenue flow breaks

stakeibc → reward collector → 15% to validators / 85% to auction → strdburner. Independent of migration but interacts with bank module accounting in non-obvious ways.

- **Likelihood**: Low.
- **Impact**: Severe to product economics.
- **Mitigations**:
  - Test 2 includes a scenario triggering `AuctionOffRewardCollectorBalance` post-upgrade with full assertion of the auction → strdburner pipeline.

### Residual risks accepted

- No multi-node testing in this plan.
- No public-testnet rehearsal (Stride doesn't have one).
- No automated rollback. If upgrade fails at block N, recovery requires multi-validator coordination on a hotfix binary. Mitigation: invest in pre-flight verification so the failure mode is "validators decline to ship the binary" rather than "validators ship and chain halts."

## §8. Summary

**v33 (this plan):**

- Snapshot 8 ICS validators → seed POA state → sweep ICS module accounts → done.
- Modules removed from manager: `ccvconsumer`, `ccvdistr`, `slashing`, `evidence`. Keepers stay mounted (deleted in v34).
- Modules kept: `staking` (wrapped by `ccvstaking`), `distribution` (unwrapped now), `gov`, `mint`, `poa` (new), all Stride-specific.
- Govenators + delegators completely undisturbed.
- Tx fees → POA module account; mint/inflation flow unchanged except no longer leaks 15% to Hub.
- Liquid staking revenue flow untouched.

**Testing:** three layers of go tests — helpers, synthetic upgrade, mainnet-export upgrade. Multi-node and k8s integration deferred.

**Top risk:** validator-set discontinuity at upgrade block. Primary mitigation: continuity assertion in Test 2 + Test 3, plus handler defensive checks.

**Out of scope:** v34 cleanup (separate spec). Implementation should not include any v34 work.
