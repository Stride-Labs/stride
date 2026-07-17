# Stride v34 — ICS Cleanup Design

## §1. Overview

This is the cleanup binary that follows v33 (the ICS → POA migration). After v33 ships, ICS keepers and store keys are still mounted but inert — kept around to preserve rollback optionality during the high-risk migration. v34 deletes them.

**Pre-conditions:**

- v33 has been deployed to mainnet.
- v33 has stabilized for at least several weeks; no rollback is being considered.
- Mainnet has produced enough blocks post-v33 to make residual ICS state safely abandonable.

**Goals:**

1. Remove all ICS keepers and store keys from Stride's binary.
2. Replace the borrowed `ccvstaking` wrapper with an in-tree equivalent.
3. Drop the `github.com/cosmos/interchain-security/v7` dependency from `go.mod`.

**Note**: the `ccvdistr` replacement was already done in v33 as
`app/distrwrapper`, so this plan only needs to handle `ccvstaking`. v33 itself
is therefore self-contained from a chain-safety perspective — v34 is purely
about pruning the ICS module from the dep graph.

**Non-goals:**

- Re-enabling slashing.
- Changing inflation, mint params, or any tokenomics.
- Modifying govenators, delegations, or the staking module.
- **Renormalizing POA validator powers.** All 8 validators inherit the same opaque uniform power (`275925`, a Hub-side scaling factor) from the v33 snapshot. Since the set is already balanced, there is no functional reason to rewrite it — POA only cares about relative power, and equal is equal. If the multisig ever wants a cleaner display value, that's a one-off `MsgUpdateValidators` outside this binary.
- Any new business logic.

## §2. Why a separate plan

v33 is the high-risk migration: validator-set handoff, module manager rewiring, ante handler change. v34 is pure housekeeping: deletion of keepers and dropping a go-mod dep.

The cost of running them as separate plans:

- One additional upgrade ceremony (validator coordination, governance proposal, etc.).
- Brief period where Stride's go.mod still has `interchain-security` listed.
- Brief period where `ccvconsumer`, `slashing`, `evidence` keepers exist as instantiated-but-unused objects.

The benefit:

- v33 implementation stays tightly focused on the consensus-layer migration.
- We get real-world post-v33 feedback before committing to v34's exact shape (e.g., we learn whether anything actually depends on the residual ICS state).
- If v33 reveals a need to roll back (rare, but possible for a migration this large), the rollback path is preserved because we haven't yet deleted any ICS keepers or store data.

The trade is firmly in favor of splitting them.

## §3. What v34 does in app.go

### Keeper removals

```diff
- ConsumerKeeper        ccvconsumerkeeper.Keeper
- SlashingKeeper        slashingkeeper.Keeper
- EvidenceKeeper        evidencekeeper.Keeper
```

Remove all corresponding `app.X = Xkeeper.NewKeeper(...)` blocks. Audit every place where `&app.SlashingKeeper`, `&app.ConsumerKeeper`, `&app.EvidenceKeeper` is passed to other constructors and remove the references. Some specific known references in v32/v33:

- `slashingkeeper.NewKeeper(..., &app.ConsumerKeeper, ...)` → entire SlashingKeeper construction is removed.
- `evidencekeeper.NewKeeper(..., &app.ConsumerKeeper, ...)` → entire EvidenceKeeper construction is removed.
- `app.ConsumerKeeper.SetStandaloneStakingKeeper(app.StakingKeeper)` → entire line removed.
- `slashing.NewAppModule(..., app.ConsumerKeeper, ...)` → already removed in v33 (slashing module isn't in the manager), so just confirming.

### Store key removals

```diff
- ccvconsumertypes.StoreKey,
- slashingtypes.StoreKey,
- evidencetypes.StoreKey,
```

Remove from `NewKVStoreKeys(...)`. Remove from `MountStores(...)` if explicitly listed there.

### Module account permission removals

```diff
- ccvconsumertypes.ConsumerRedistributeName:     nil,
- ccvconsumertypes.ConsumerToSendToProviderName: nil,
```

Remove from `maccPerms`.

### Module manager — replace `ccvstaking` with in-tree wrapper

```diff
- ccvstaking.NewAppModule(appCodec, &app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
+ stakingwrapper.NewAppModule(appCodec, &app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
```

### Drop the ICS dependency

```diff
- github.com/cosmos/interchain-security/v7 v7.0.0-...
```

Run `go mod tidy` after the import removal. Verify no transitive dependents remain.

## §4. The in-tree wrapper module

New package: `x/stakingwrapper/`

### Why it's needed

`x/staking` is still active in v34 (govenators, delegations, distribution rewards). Without a wrapper around it, `x/staking`'s `EndBlock` would emit `ValidatorUpdate`s that conflict with POA's. The `ccvstaking` wrapper currently in use (from `interchain-security/v7`) suppresses these updates. To drop the ICS dependency, we need our own wrapper.

### File layout

```
x/stakingwrapper/
├── module.go    (~50 lines — the AppModule wrapper)
└── module_test.go  (basic test that EndBlock returns empty []ValidatorUpdate)
```

### Module source

```go
package stakingwrapper

import (
    "context"
    "encoding/json"

    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/staking"
    "github.com/cosmos/cosmos-sdk/x/staking/exported"
    "github.com/cosmos/cosmos-sdk/x/staking/keeper"
    "github.com/cosmos/cosmos-sdk/x/staking/types"

    abci "github.com/cometbft/cometbft/abci/types"
)

// AppModule wraps the standard x/staking AppModule and suppresses validator
// set updates so that CometBFT is driven by x/poa instead. All staking
// bookkeeping (delegations, unbonding queue, distribution hooks) still runs.
type AppModule struct {
    staking.AppModule
    keeper keeper.Keeper
}

func NewAppModule(
    cdc codec.Codec,
    k *keeper.Keeper,
    ak types.AccountKeeper,
    bk types.BankKeeper,
    subspace exported.Subspace,
) AppModule {
    return AppModule{
        AppModule: staking.NewAppModule(cdc, k, ak, bk, subspace),
        keeper:    *k,
    }
}

// InitGenesis runs the full staking genesis but returns no validator updates;
// CometBFT validators come from x/poa, not x/staking.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
    var gs types.GenesisState
    cdc.MustUnmarshalJSON(data, &gs)
    _ = am.keeper.InitGenesis(ctx, &gs)
    return []abci.ValidatorUpdate{}
}

// EndBlock runs all staking unbonding/redelegation logic but returns no
// validator updates to CometBFT.
func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
    _, _ = am.keeper.BlockValidatorUpdates(ctx)
    return []abci.ValidatorUpdate{}, nil
}
```

Behaviorally identical to `ccvstaking`; just lives in Stride's tree.

## §5. Upgrade plan StoreUpgrades

```go
storetypes.StoreUpgrades{
    Deleted: []string{
        ccvconsumertypes.StoreKey,
        slashingtypes.StoreKey,
        evidencetypes.StoreKey,
    },
}
```

These three store keys had data accumulated through v33 but became orphans after v33 removed their modules from the manager. v34 clears them and removes the substores from the Merkle tree.

**Critical paired-change**: every key listed in `Deleted` must also have its `MountStores` call and `NewKVStoreKeys` entry removed in the same v34 binary. Listing a key in `Deleted` while leaving it mounted causes `"version of store X mismatch root store's version"` panic at startup.

## §6. The v34 upgrade handler

Minimal — possibly empty-bodied other than running migrations:

```go
func CreateUpgradeHandler(
    mm *module.Manager,
    configurator module.Configurator,
) upgradetypes.UpgradeHandler {
    return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
        sdkCtx := sdk.UnwrapSDKContext(ctx)
        sdkCtx.Logger().Info("Starting upgrade v34 (ICS cleanup)...")
        return mm.RunMigrations(sdkCtx, configurator, vm)
    }
}
```

No state reads, no state writes, no balance moves, no validator changes. The store-key deletion happens via `StoreUpgrades`, not via the handler.

## §7. Paired-change invariants for v34

The kind of bug that causes a panic on chain restart:

1. For every entry in `StoreUpgrades.Deleted`, there must be **no** `MountStores(keys[...])` call for that key.
2. For every removed `MountStores` call, the corresponding store key must be in `StoreUpgrades.Deleted`.
3. For every removed keeper, all `app.go` references to that keeper must be gone (no lingering `&app.SlashingKeeper` passed to anyone).

PR checklist must verify all three.

## §8. Testing strategy

### Test 1 — basic upgrade test

Standard go test using `apptesting.AppTestHelper`. Embed it, run `s.ConfirmUpgradeSucceeded(v34.UpgradeName)`, assert:

- `bonded_tokens_pool` balance unchanged (govenators undisturbed).
- POA validator state unchanged.
- Active gov proposals still exist.
- Mint params unchanged.
- Liquid staking flow still operates (call `AuctionOffRewardCollectorBalance` and assert).

### Test 2 — wrapper module unit test

Test `stakingwrapper.AppModule.EndBlock` returns `[]abci.ValidatorUpdate{}` regardless of underlying staking state. Test that `BlockValidatorUpdates` is still called (delegations and unbondings still process).

### Test 3 — chain restart simulation

Most important test for v34. Run the v34 binary against a v33 mainnet state export. Specifically verify:

- Chain starts up without panic.
- All removed store keys are no longer in the Merkle tree.
- All retained store keys still hash correctly.
- POA validator set is unchanged.
- A test transaction goes through.

This is the test that would have caught the "I deleted from `StoreUpgrades` but forgot to unmount" panic — invest in it heavily.

### Things NOT in v34's test plan

- No multi-node testing.
- No public testnet rehearsal.
- No multisig testing (multisig untouched by v34).

## §9. Risks

### Risk 1: Mount/Delete pairing bug

Listing a store key in `StoreUpgrades.Deleted` while leaving its `MountStores` call in app.go panics at chain restart with `"version of store X mismatch"`.

- **Likelihood**: Moderate if reviewers don't enforce the invariant strictly.
- **Impact**: Severe — chain can't start. All nodes need a hotfix binary.
- **Mitigations**:
  - Test 3 (chain restart simulation) catches this in CI.
  - PR checklist explicitly verifies §7 invariants.
  - Pre-flight by spinning up v34 against a v33 state export on every reviewer's machine.

### Risk 2: Hidden reference to a deleted keeper

A module or function still has a code path that reaches into `ConsumerKeeper`, `SlashingKeeper`, or `EvidenceKeeper`. After deletion, that path causes a build error (best case) or runtime nil-dereference (worst case).

- **Likelihood**: Low. Build errors catch most cases. Runtime nil-derefs would have surfaced post-v33.
- **Impact**: Build error is annoying but recoverable. Runtime panic post-upgrade is severe.
- **Mitigations**:
  - Audit grep for keeper references before merging.
  - Test 1 exercises a broad set of operations post-upgrade.

### Risk 3: ICS dependency removal breaks transitive imports

`go mod tidy` post-removal might fail if some other package depends on `interchain-security` transitively.

- **Likelihood**: Low.
- **Impact**: Build error, not runtime.
- **Mitigations**:
  - Run `go mod tidy` and `go build ./...` early in implementation.
  - If conflicts surface, isolate the offender and decide module-by-module.

### Risk 4: stakingwrapper subtly differs from ccvstaking

The in-tree wrapper is a copy, but Go's interface conformance and module.AppModule wiring is fiddly. A subtle behavior difference could surface.

- **Likelihood**: Low (50 lines, trivially auditable).
- **Impact**: Could cause validator-update emissions to break (CometBFT halt) or silently skip staking bookkeeping (slow govenator drain over time).
- **Mitigations**:
  - Test 2 (wrapper unit test) explicitly compares behavior against `ccvstaking`.
  - Manual code review comparing the wrappers line-by-line.

## §10. Summary

v34 is a small upgrade. Three things:

1. Delete `ConsumerKeeper`, `SlashingKeeper`, `EvidenceKeeper` from app.go (and their store keys).
2. Add a ~50-line in-tree wrapper at `x/stakingwrapper/`; swap `ccvstaking` for it in app.go.
3. `go.mod`: drop `interchain-security/v7`, run `go mod tidy`.

Top risk: forgetting to pair `StoreUpgrades.Deleted` with `MountStores` removal, causing chain restart panic. Test 3 (chain restart simulation) is the primary mitigation.
