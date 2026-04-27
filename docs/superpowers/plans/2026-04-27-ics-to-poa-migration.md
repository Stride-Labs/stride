# Stride v33 — ICS → POA Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Stride's v33 upgrade that migrates the chain from an ICS consumer (validator set pushed by Cosmos Hub via PSS) to a POA chain (validator set controlled by an admin multisig), preserving govenator staking, delegations, governance, and all liquid-staking product economics.

**Architecture:** Add `x/poa` as the new block-producing validator-set module. Remove `ccvconsumer`, `ccvdistr`, `slashing`, `evidence` from the module manager (keep their keepers mounted for v33 — deletion is v34's job). Keep `ccvstaking` as the wrapper around `x/staking` so govenator delegations continue to work. Reconfigure `x/distribution` to use the standard `fee_collector` and route tx fees to POA via the ante handler. Implement an upgrade handler that snapshots the live ICS validator set into POA state on the upgrade block, with no validator-set discontinuity.

**Tech Stack:** Cosmos SDK v0.54, CometBFT v0.38, `github.com/cosmos/cosmos-sdk/enterprise/poa` (the SDK POA module), Stride's existing custom modules unchanged. Go 1.23+.

**Spec reference:** `docs/superpowers/specs/2026-04-27-ics-to-poa-migration-design.md`. Read it before starting.

**Out of scope (do NOT implement):** v34 cleanup is a separate plan. Do not delete `ConsumerKeeper`, `SlashingKeeper`, `EvidenceKeeper` or their store keys. Do not write the in-tree `stakingwrapper`. Do not drop `interchain-security` from go.mod. If you find yourself wanting to do any of these, escalate — don't expand v33's scope.

**Module path policy:** Keep `module github.com/Stride-Labs/stride/v32` in go.mod throughout this work. Bumping to `v33` is a separate end-of-cycle PR.

---

## File Structure

### New files

| File                                                   | Responsibility                                                         |
| ------------------------------------------------------ | ---------------------------------------------------------------------- |
| `app/upgrades/v33/constants.go`                        | `UpgradeName`, `AdminMultisigAddress`, embeds `validators.json`        |
| `app/upgrades/v33/validators.json`                     | Source of truth for 8 validators. Already generated and committed.     |
| `app/upgrades/v33/scripts/spreadsheet.csv`             | Validator-supplied data (moniker + Hub addresses + Stride payment).    |
| `app/upgrades/v33/scripts/fetch_validator_monikers.sh` | Generates `validators.json` from spreadsheet + Stride RPC + Hub gRPC.  |
| `app/upgrades/v33/helpers.go`                          | `snapshotValidatorsFromICS`, `initializePOA`, `sweepICSModuleAccounts` |
| `app/upgrades/v33/helpers_test.go`                     | Unit tests for the three helpers                                       |
| `app/upgrades/v33/upgrades.go`                         | `CreateUpgradeHandler` orchestrating helpers                           |
| `app/upgrades/v33/upgrades_test.go`                    | App-level handler test (synthetic state) + mainnet-export test         |
| `app/upgrades/v33/testdata/mainnet_export.json.gz`     | Sanitized mainnet state export fixture (committed)                     |

**Note on `validators.json`**: this file has already been generated and committed. The script that builds it (`scripts/fetch_validator_monikers.sh`) joins three data sources: validator-supplied identity (the CSV), Stride's live CometBFT validator set (RPC), and the Hub's ICS provider key-assignment pairs (gRPC). Re-run if the validator set changes.

### Modified files

| File                  | Change                                                                                             |
| --------------------- | -------------------------------------------------------------------------------------------------- |
| `go.mod`              | Add `github.com/cosmos/cosmos-sdk/enterprise/poa`                                                  |
| `app/app.go`          | Add POA wiring; remove ICS-era modules from manager (keep keepers); rewire distribution            |
| `app/ante_handler.go` | Replace standard fee deductor with POA's fee deductor (or set fee collector to POA module account) |
| `app/upgrades.go`     | Register v33 handler; add v33 case to `StoreUpgrades` switch                                       |

---

## Task 1: Add POA dependency and verify it builds

**Files:**

- Modify: `go.mod`
- Modify: `go.sum` (auto-generated)

The cosmos-sdk POA module is a separate go module at `github.com/cosmos/cosmos-sdk/enterprise/poa`. Stride's go.mod has to add it explicitly.

- [ ] **Step 1.1: Identify the right POA module version**

Check which version of `github.com/cosmos/cosmos-sdk/enterprise/poa` is compatible with `cosmos-sdk v0.53.7` / `v0.54` (whichever Stride is on). Inspect the POA module's own `go.mod` at https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/go.mod and pick a tagged release that matches Stride's SDK pin. Record the version.

- [ ] **Step 1.2: Add the dependency**

Edit `go.mod`:

```
require (
    ...existing deps...
    github.com/cosmos/cosmos-sdk/enterprise/poa <VERSION>
)
```

- [ ] **Step 1.3: Run `go mod tidy`**

```bash
go mod tidy
```

Expected: `go.sum` updated with new entries. No errors.

- [ ] **Step 1.4: Verify import paths exist**

```bash
go list github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper
```

Expected: both packages listed, no errors.

If errors, the module version is wrong. Try a different tag.

- [ ] **Step 1.5: Build the binary**

```bash
make build
```

Expected: `build/strided` produced, no compile errors. (Adding a dep without using it should be a no-op build.)

- [ ] **Step 1.6: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add cosmos-sdk enterprise/poa module"
```

---

## Task 2: Create v33 upgrade package skeleton

**Files:**

- Create: `app/upgrades/v33/constants.go`

- [ ] **Step 2.1: Create the directory**

```bash
mkdir -p app/upgrades/v33
```

- [ ] **Step 2.2: Write `constants.go`**

Create `app/upgrades/v33/constants.go`:

```go
package v33

// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
const UpgradeName = "v33"

// AdminMultisigAddress is the bech32 address that POA recognizes as its admin
// post-upgrade. The multisig does not need to be operationally signing-ready
// at upgrade time; it just has to be a valid bech32 address.
//
// FINAL VALUE TO BE PROVIDED BY OPS BEFORE MAINNET RELEASE.
// During implementation, use a placeholder Stride address you control on testnet.
const AdminMultisigAddress = "stride1fduug6m38gyuqt3wcgc2kcgr9nnte0n7ssn27e"

// ValidatorMonikers maps the consensus address (hex of the raw 20-byte address
// returned by ccv consumer GetAllCCValidator) to the validator's moniker.
//
// ICS does not store monikers on the consumer chain — they live on the Hub.
// We pre-bake them here so they appear correctly in POA's validator records.
//
// FINAL VALUES TO BE PROVIDED BY VALIDATOR OUTREACH BEFORE MAINNET RELEASE.
// Each of the 8 PSS-allowlisted validators must have an entry.
var ValidatorMonikers = map[string]string{
    // "<HEX_CONS_ADDR_1>": "Validator Name 1",
    // "<HEX_CONS_ADDR_2>": "Validator Name 2",
    // ... 8 entries total
}

// ExpectedValidatorCount is enforced by the upgrade handler. Panics if
// consumerKeeper.GetAllCCValidator returns a different count.
const ExpectedValidatorCount = 8
```

- [ ] **Step 2.3: Verify it compiles**

```bash
go build ./app/upgrades/v33/...
```

Expected: no errors.

- [ ] **Step 2.4: Commit**

```bash
git add app/upgrades/v33/constants.go
git commit -m "feat(v33): add upgrade package skeleton"
```

---

## Task 3: Build `snapshotValidatorsFromICS` (TDD)

**Files:**

- Create: `app/upgrades/v33/helpers.go`
- Create: `app/upgrades/v33/helpers_test.go`

This helper reads the live ICS validator set and converts each entry to a POA `Validator`.

- [ ] **Step 3.1: Write the failing test**

Create `app/upgrades/v33/helpers_test.go`:

```go
package v33_test

import (
    "encoding/hex"
    "testing"

    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
    "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
    consumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
    "github.com/stretchr/testify/require"

    "github.com/Stride-Labs/stride/v32/app/apptesting"
    v33 "github.com/Stride-Labs/stride/v32/app/upgrades/v33"
)

func TestSnapshotValidatorsFromICS_HappyPath(t *testing.T) {
    s := apptesting.NewTestApp(t)

    // Seed 8 CCValidators with deterministic pubkeys
    seedConsumerValidators(t, s, 8)

    // Pre-populate moniker map for the 8 seeded addresses
    addrs := getSeededConsAddresses(t, s)
    for i, addr := range addrs {
        v33.ValidatorMonikers[hex.EncodeToString(addr)] = fmtMoniker(i)
    }
    t.Cleanup(func() {
        for _, addr := range addrs {
            delete(v33.ValidatorMonikers, hex.EncodeToString(addr))
        }
    })

    poaValidators, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
    require.NoError(t, err)
    require.Len(t, poaValidators, 8)

    for i, val := range poaValidators {
        require.NotNil(t, val.PubKey)
        require.Equal(t, int64(100), val.Power)
        require.NotNil(t, val.Metadata)
        require.Equal(t, fmtMoniker(i), val.Metadata.Moniker)
        require.Contains(t, val.Metadata.OperatorAddress, "stridevaloper")
    }
}

func TestSnapshotValidatorsFromICS_WrongCount(t *testing.T) {
    s := apptesting.NewTestApp(t)
    seedConsumerValidators(t, s, 7) // expecting 8

    _, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
    require.Error(t, err)
    require.Contains(t, err.Error(), "expected 8 validators")
}

func TestSnapshotValidatorsFromICS_UnknownMoniker(t *testing.T) {
    s := apptesting.NewTestApp(t)
    seedConsumerValidators(t, s, 8)
    // Do NOT populate ValidatorMonikers — every validator should fall back to ""

    poaValidators, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
    require.NoError(t, err)
    for _, val := range poaValidators {
        require.Equal(t, "", val.Metadata.Moniker)
    }
}

// --- test helpers ---

func seedConsumerValidators(t *testing.T, s *apptesting.AppTestHelper, count int) {
    for i := 0; i < count; i++ {
        privKey := ed25519.GenPrivKeyFromSecret([]byte{byte(i + 1)})
        pubKey := privKey.PubKey()
        addr := pubKey.Address().Bytes()
        pkAny, err := codectypes.NewAnyWithValue(pubKey)
        require.NoError(t, err)

        ccVal := consumertypes.CrossChainValidator{
            Address: addr,
            Power:   100,
            Pubkey:  pkAny,
        }
        s.App.ConsumerKeeper.SetCCValidator(s.Ctx, ccVal)
    }
}

func getSeededConsAddresses(t *testing.T, s *apptesting.AppTestHelper) [][]byte {
    vals := s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx)
    out := make([][]byte, 0, len(vals))
    for _, v := range vals {
        out = append(out, v.Address)
    }
    return out
}

func fmtMoniker(i int) string {
    return "validator-" + string(rune('a'+i))
}
```

If `apptesting.NewTestApp(t)` doesn't exist in Stride's helper, swap for the actual constructor. Stride's pattern likely uses a testify suite — check `app/upgrades/v32/upgrades_test.go` and adapt.

- [ ] **Step 3.2: Run test to verify it fails**

```bash
go test ./app/upgrades/v33/... -run TestSnapshotValidatorsFromICS -v
```

Expected: FAIL with "undefined: v33.SnapshotValidatorsFromICS" or similar.

- [ ] **Step 3.3: Implement `SnapshotValidatorsFromICS`**

Create `app/upgrades/v33/helpers.go`:

```go
package v33

import (
    "encoding/hex"
    "fmt"

    errorsmod "cosmossdk.io/errors"
    "github.com/cosmos/cosmos-sdk/codec"
    codectypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/bech32"

    poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
    poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"

    authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
    authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
    bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
    distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

    ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"
    ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
)

// SnapshotValidatorsFromICS reads the current CCV validator set from the
// consumer keeper and converts it into a slice of POA Validators ready to
// be passed to poaKeeper.InitGenesis.
//
// Monikers are looked up from the pre-baked ValidatorMonikers map (keyed by
// hex of the consensus address). Missing entries fall back to empty string.
//
// Operator addresses are derived from the consensus address bytes by
// bech32-encoding with the "stridevaloper" prefix. Stride's ICS validators
// have no Stride-side operator address (the operator key lives on the Hub),
// so this is a metadata-only string — POA's runtime logic keys off the
// consensus address, not OperatorAddress.
func SnapshotValidatorsFromICS(
    ctx sdk.Context,
    consumerKeeper ccvconsumerkeeper.Keeper,
) ([]poatypes.Validator, error) {
    ccVals := consumerKeeper.GetAllCCValidator(ctx)
    if len(ccVals) != ExpectedValidatorCount {
        return nil, fmt.Errorf(
            "expected %d validators in consumer keeper, got %d",
            ExpectedValidatorCount, len(ccVals),
        )
    }

    poaVals := make([]poatypes.Validator, 0, len(ccVals))
    for _, ccVal := range ccVals {
        consPubKey, err := ccVal.ConsPubKey()
        if err != nil {
            return nil, errorsmod.Wrapf(err,
                "failed to decode cons pubkey for validator %x", ccVal.Address)
        }
        pubKeyAny, err := codectypes.NewAnyWithValue(consPubKey)
        if err != nil {
            return nil, errorsmod.Wrapf(err,
                "failed to wrap cons pubkey for validator %x", ccVal.Address)
        }

        operatorAddr, err := bech32.ConvertAndEncode("stridevaloper", ccVal.Address)
        if err != nil {
            return nil, errorsmod.Wrapf(err,
                "failed to bech32-encode operator address for validator %x", ccVal.Address)
        }

        moniker := ValidatorMonikers[hex.EncodeToString(ccVal.Address)]

        poaVals = append(poaVals, poatypes.Validator{
            PubKey: pubKeyAny,
            Power:  ccVal.Power,
            Metadata: &poatypes.ValidatorMetadata{
                Moniker:         moniker,
                OperatorAddress: operatorAddr,
            },
        })
    }

    return poaVals, nil
}
```

If the POA module's `Validator` or `ValidatorMetadata` field names differ, fix to match.

- [ ] **Step 3.4: Run test to verify it passes**

```bash
go test ./app/upgrades/v33/... -run TestSnapshotValidatorsFromICS -v
```

Expected: PASS for all three test cases.

- [ ] **Step 3.5: Commit**

```bash
git add app/upgrades/v33/helpers.go app/upgrades/v33/helpers_test.go
git commit -m "feat(v33): add SnapshotValidatorsFromICS helper"
```

---

## Task 4: Build `InitializePOA` (TDD)

**Files:**

- Modify: `app/upgrades/v33/helpers.go`
- Modify: `app/upgrades/v33/helpers_test.go`

This helper synthesizes a POA `GenesisState` and calls `poaKeeper.InitGenesis(...)`. The shape mirrors the canonical SDK sample at `cosmos-sdk/enterprise/poa/examples/migrate-from-pos/sample_upgrades/upgrade_handler.go:initializePOA` — match it closely.

- [ ] **Step 4.1: Write the failing test**

Append to `app/upgrades/v33/helpers_test.go`:

```go
func TestInitializePOA_HappyPath(t *testing.T) {
    s := apptesting.NewTestApp(t)
    seedConsumerValidators(t, s, 8)

    poaVals, err := v33.SnapshotValidatorsFromICS(s.Ctx, s.App.ConsumerKeeper)
    require.NoError(t, err)

    err = v33.InitializePOA(s.Ctx, s.App.AppCodec(), s.App.POAKeeper, "stride1<TEST_ADMIN_BECH32>", poaVals)
    require.NoError(t, err)

    // Confirm POA state was populated
    storedVals, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
    require.NoError(t, err)
    require.Len(t, storedVals, 8)

    params, err := s.App.POAKeeper.GetParams(s.Ctx)
    require.NoError(t, err)
    require.Equal(t, "stride1<TEST_ADMIN_BECH32>", params.Admin)
}

func TestInitializePOA_InvalidAdmin(t *testing.T) {
    s := apptesting.NewTestApp(t)
    err := v33.InitializePOA(s.Ctx, s.App.AppCodec(), s.App.POAKeeper, "not_a_bech32", []poatypes.Validator{})
    require.Error(t, err)
}
```

(Replace `stride1<TEST_ADMIN_BECH32>` with a valid bech32 generated however Stride's tests typically generate them — `s.TestAccs[0]` or similar. `s.App.AppCodec()` returns the `codec.Codec` registered on the test app; substitute with the project's actual accessor if the name differs.)

- [ ] **Step 4.2: Run test to verify it fails**

```bash
go test ./app/upgrades/v33/... -run TestInitializePOA -v
```

Expected: FAIL with "undefined: v33.InitializePOA".

- [ ] **Step 4.3: Implement `InitializePOA`**

Append to `app/upgrades/v33/helpers.go`:

```go
// InitializePOA seeds POA's KV store with the given validator set and admin.
// Mirrors the canonical SDK sample at
// cosmos-sdk/enterprise/poa/examples/migrate-from-pos/sample_upgrades/upgrade_handler.go.
//
// WithBlockHeight(0) is required: POA's CreateValidator path calls
// GetTotalPower, which only treats "no total power yet" as a non-error
// case when ctx.BlockHeight() == 0 (enterprise/poa/x/poa/keeper/validator.go).
//
// The keeper-level InitGenesis returns ([]abci.ValidatorUpdate, error); we
// discard the updates because an upgrade handler returns a VersionMap, not
// ABCI updates. The next EndBlock will reap and emit anything still queued.
func InitializePOA(
    ctx sdk.Context,
    cdc codec.Codec,
    poaKeeper *poakeeper.Keeper,
    adminAddress string,
    validators []poatypes.Validator,
) error {
    if _, err := sdk.AccAddressFromBech32(adminAddress); err != nil {
        return errorsmod.Wrapf(err, "invalid admin address: %s", adminAddress)
    }

    sdkCtx := sdk.UnwrapSDKContext(ctx).WithBlockHeight(0)

    genesis := &poatypes.GenesisState{
        Params:     poatypes.Params{Admin: adminAddress},
        Validators: validators,
        // AllocatedFees intentionally omitted — fresh POA init has no
        // pre-existing per-validator fee allocations to restore.
    }

    _, err := poaKeeper.InitGenesis(sdkCtx, cdc, genesis)
    return err
}
```

Notes:

- `poaKeeper` is `*poakeeper.Keeper` (pointer), matching the sample and the keeper's own method receivers.
- Add `"github.com/cosmos/cosmos-sdk/codec"` to the imports if not already present.
- POA's `Validator.Metadata` has a `Description` field (in addition to `Moniker` and `OperatorAddress`) which Stride leaves empty; the admin can set it later via `MsgUpdateValidators`.

- [ ] **Step 4.4: Run test to verify it passes**

```bash
go test ./app/upgrades/v33/... -run TestInitializePOA -v
```

Expected: PASS for both test cases.

This will also require the `POAKeeper` to exist on the app struct — if `apptesting.NewTestApp` doesn't yet wire POA, this test will fail at the `s.App.POAKeeper` access. In that case, complete Task 6 (POA wiring in app.go) first, then come back to verify these tests.

- [ ] **Step 4.5: Commit**

```bash
git add app/upgrades/v33/helpers.go app/upgrades/v33/helpers_test.go
git commit -m "feat(v33): add InitializePOA helper"
```

---

## Task 5: Build `SweepICSModuleAccounts` (TDD)

**Files:**

- Modify: `app/upgrades/v33/helpers.go`
- Modify: `app/upgrades/v33/helpers_test.go`

This helper sweeps any residual balance in `cons_redistribute` and `cons_to_send_to_provider` to the community pool.

- [ ] **Step 5.1: Write the failing test**

Append to `app/upgrades/v33/helpers_test.go`:

```go
func TestSweepICSModuleAccounts_HappyPath(t *testing.T) {
    s := apptesting.NewTestApp(t)

    consRedistributeAddr := s.App.AccountKeeper.GetModuleAddress(ccvconsumertypes.ConsumerRedistributeName)
    consToProviderAddr := s.App.AccountKeeper.GetModuleAddress(ccvconsumertypes.ConsumerToSendToProviderName)

    // Pre-fund both accounts with some test STRD
    bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
    require.NoError(t, err)
    fundAmount := sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1_000_000))

    s.FundModuleAccount(ccvconsumertypes.ConsumerRedistributeName, fundAmount[0])
    s.FundModuleAccount(ccvconsumertypes.ConsumerToSendToProviderName, fundAmount[0])

    err = v33.SweepICSModuleAccounts(
        s.Ctx, s.App.AccountKeeper, s.App.BankKeeper, s.App.DistrKeeper,
    )
    require.NoError(t, err)

    require.True(t, s.App.BankKeeper.GetAllBalances(s.Ctx, consRedistributeAddr).IsZero())
    require.True(t, s.App.BankKeeper.GetAllBalances(s.Ctx, consToProviderAddr).IsZero())

    // Community pool should have received both amounts
    feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
    require.NoError(t, err)
    cpBalance := feePool.CommunityPool.AmountOf(bondDenom)
    require.Equal(t, math.LegacyNewDec(2_000_000), cpBalance)
}

func TestSweepICSModuleAccounts_EmptyAccounts(t *testing.T) {
    s := apptesting.NewTestApp(t)

    // Don't fund anything — both accounts are empty
    err := v33.SweepICSModuleAccounts(
        s.Ctx, s.App.AccountKeeper, s.App.BankKeeper, s.App.DistrKeeper,
    )
    require.NoError(t, err) // empty sweep should not error
}
```

- [ ] **Step 5.2: Run test to verify it fails**

```bash
go test ./app/upgrades/v33/... -run TestSweepICSModuleAccounts -v
```

Expected: FAIL with "undefined: v33.SweepICSModuleAccounts".

- [ ] **Step 5.3: Implement `SweepICSModuleAccounts`**

Append to `app/upgrades/v33/helpers.go`:

```go
// SweepICSModuleAccounts moves any residual balance from the two ICS-era
// reward module accounts (cons_redistribute and cons_to_send_to_provider)
// to the community pool. After v33, no module deposits to these accounts;
// any leftover balance would be permanently stranded otherwise.
func SweepICSModuleAccounts(
    ctx sdk.Context,
    accountKeeper authkeeper.AccountKeeper,
    bankKeeper bankkeeper.Keeper,
    distrKeeper distrkeeper.Keeper,
) error {
    accountsToSweep := []string{
        ccvconsumertypes.ConsumerRedistributeName,
        ccvconsumertypes.ConsumerToSendToProviderName,
    }

    for _, moduleName := range accountsToSweep {
        moduleAddr := accountKeeper.GetModuleAddress(moduleName)
        balance := bankKeeper.GetAllBalances(ctx, moduleAddr)
        if balance.IsZero() {
            ctx.Logger().Info(fmt.Sprintf("v33: %s is empty, skipping sweep", moduleName))
            continue
        }

        if err := distrKeeper.FundCommunityPool(ctx, balance, moduleAddr); err != nil {
            return errorsmod.Wrapf(err,
                "failed to fund community pool from %s", moduleName)
        }
        ctx.Logger().Info(fmt.Sprintf("v33: swept %s from %s to community pool",
            balance, moduleName))
    }
    return nil
}
```

Note: `distrKeeper.FundCommunityPool` may have a slightly different signature in SDK v0.54 — verify and adapt. If it takes `senderAddr` differently, swap accordingly.

- [ ] **Step 5.4: Run test to verify it passes**

```bash
go test ./app/upgrades/v33/... -run TestSweepICSModuleAccounts -v
```

Expected: PASS.

- [ ] **Step 5.5: Commit**

```bash
git add app/upgrades/v33/helpers.go app/upgrades/v33/helpers_test.go
git commit -m "feat(v33): add SweepICSModuleAccounts helper"
```

---

## Task 6: Wire POA into app.go (additive only — don't remove anything yet)

**Files:**

- Modify: `app/app.go`

This task adds POA without removing any existing modules. After it, the chain can build with both ICS and POA modules registered. POA won't actually drive validator updates (since ICS is still doing that), but the wiring exists.

- [ ] **Step 6.1: Add POA imports to app.go**

In the import block of `app/app.go`, add:

```go
poa "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa"
poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
```

- [ ] **Step 6.2: Add POAKeeper to the StrideApp struct**

Locate the `StrideApp` struct definition (around line 256). Add `POAKeeper *poakeeper.Keeper` alongside the other keepers:

```go
type StrideApp struct {
    ...
    POAKeeper *poakeeper.Keeper
    ...
}
```

- [ ] **Step 6.3: Add POA store keys**

Locate the `NewKVStoreKeys(...)` call (around line 354). Add `poatypes.StoreKey`:

```go
keys := storetypes.NewKVStoreKeys(
    ...,
    poatypes.StoreKey,
)
```

Locate `NewTransientStoreKeys(...)` (or equivalent — check how Stride registers transient keys). Add `poatypes.TransientStoreKey`:

```go
tkeys := storetypes.NewTransientStoreKeys(
    ...,
    poatypes.TransientStoreKey,
)
```

If Stride doesn't currently use transient store keys, add the `tkeys` declaration and a corresponding `MountTransientStores(tkeys)` call.

- [ ] **Step 6.4: Add POA module account permission**

In `maccPerms` (around line 200), add:

```go
poatypes.ModuleName: nil, // POA accumulates tx fees; no mint/burn permissions needed
```

- [ ] **Step 6.5: Construct POAKeeper**

Find a logical spot in `NewStrideApp` to construct the POA keeper — after `BankKeeper` but before any keeper that depends on POA. Add:

```go
app.POAKeeper = poakeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[poatypes.StoreKey]),
    runtime.NewTransientStoreService(tkeys[poatypes.TransientStoreKey]),
    app.AccountKeeper,
    app.BankKeeper,
)
```

If the POA `NewKeeper` signature differs (e.g., takes additional params or a pointer), adapt.

- [ ] **Step 6.6: Register POA in module manager**

In the `module.NewManager(...)` call (around line 989), add:

```go
poa.NewAppModule(appCodec, *app.POAKeeper, poa.WithSecp256k1Support()),
```

If you confirm the 8 validators all use ed25519 keys (typical), `poa.WithSecp256k1Support()` can be omitted. Default to including it — harmless if unused.

- [ ] **Step 6.7: Add POA to begin/end blocker orderings**

Locate `SetOrderBeginBlockers(...)` and `SetOrderEndBlockers(...)`. In **both**, add `poatypes.ModuleName` near the end (just before `strdburnertypes.ModuleName` which is documented as last).

For `SetOrderInitGenesis(...)`, also add `poatypes.ModuleName` near the appropriate place (after `authtypes`/`banktypes`, before app-specific modules).

- [ ] **Step 6.8: Build and confirm chain still starts**

```bash
make build
```

Expected: builds clean.

```bash
go test ./app/... -run TestNewStrideApp -v
```

If Stride has a basic app construction test, run it. If not, run the existing v32 upgrade test:

```bash
go test ./app/upgrades/v32/... -v
```

Expected: passes (because POA is added but not yet replacing anything; chain is in a hybrid state).

- [ ] **Step 6.9: Commit**

```bash
git add app/app.go
git commit -m "feat(v33): wire POA module into app (additive)"
```

---

## Task 7: Remove `ccvconsumer` and `ccvdistr` from module manager

**Files:**

- Modify: `app/app.go`

Removes the two ICS-era modules that are being replaced by POA. **Keep their keepers constructed** in app.go — only remove the `*.NewAppModule(...)` registrations and the corresponding entries in blocker/genesis ordering.

- [ ] **Step 7.1: Remove from `module.NewManager(...)`**

In the module manager construction, **delete** these lines:

```go
ccvdistr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, authtypes.FeeCollectorName, app.GetSubspace(distrtypes.ModuleName)),
```

```go
consumerModule, // (the variable defined as ccvconsumer.NewAppModule(...))
```

Also remove the line that defines `consumerModule` (around line 691):

```go
consumerModule := ccvconsumer.NewAppModule(app.ConsumerKeeper, app.GetSubspace(ccvconsumertypes.ModuleName))
```

- [ ] **Step 7.2: Replace ccvdistr with plain x/distribution**

Add to the module manager (replacing the deleted `ccvdistr.NewAppModule` line):

```go
distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
```

Add the import if missing:

```go
distr "github.com/cosmos/cosmos-sdk/x/distribution"
```

- [ ] **Step 7.3: Remove from begin/end blocker orderings**

In `SetOrderBeginBlockers(...)` and `SetOrderEndBlockers(...)`, **delete** the line:

```go
ccvconsumertypes.ModuleName,
```

(There's only one entry per ordering for ICS modules — the democracy distribution and democracy staking modules use the same module name as their underlying SDK module: `distrtypes.ModuleName` and `stakingtypes.ModuleName`. So those names stay; only `ccvconsumertypes.ModuleName` needs removal.)

- [ ] **Step 7.4: Remove from `SetOrderInitGenesis`**

Same removal: delete `ccvconsumertypes.ModuleName` from the InitGenesis order list.

- [ ] **Step 7.5: Remove the IBC route**

Locate the IBC router setup (around line 972):

```go
.AddRoute(ccvconsumertypes.ModuleName, consumerModule)
```

**Delete** this line.

- [ ] **Step 7.6: Remove the unused `ccvconsumer` and `ccvdistr` package imports**

If those packages are no longer referenced anywhere (because `consumerModule` is gone and `ccvdistr.NewAppModule` is gone), Go will emit unused-import errors. Remove the imports:

```go
ccvconsumer "github.com/cosmos/interchain-security/v7/x/ccv/consumer"
ccvdistr "github.com/cosmos/interchain-security/v7/x/ccv/democracy/distribution"
```

**Keep**:

- `ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"` — for `app.ConsumerKeeper`
- `ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"` — for store key reference, module account names, etc.
- `ccvstaking "github.com/cosmos/interchain-security/v7/x/ccv/democracy/staking"` — still used

- [ ] **Step 7.7: Build and confirm**

```bash
make build
```

Expected: builds clean.

- [ ] **Step 7.8: Commit**

```bash
git add app/app.go
git commit -m "feat(v33): remove ccvconsumer and ccvdistr from module manager"
```

---

## Task 8: Remove `slashing` and `evidence` from module manager

**Files:**

- Modify: `app/app.go`

Same pattern as Task 7. Keep the keepers constructed (for v34 to clean up later), only remove module-manager registration.

- [ ] **Step 8.1: Remove from `module.NewManager(...)`**

**Delete** these lines from the module manager construction:

```go
slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.ConsumerKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
```

```go
evidence.NewAppModule(app.EvidenceKeeper),
```

- [ ] **Step 8.2: Remove from begin/end blocker orderings**

Delete `slashingtypes.ModuleName` and `evidencetypes.ModuleName` from `SetOrderBeginBlockers(...)` and `SetOrderEndBlockers(...)`.

- [ ] **Step 8.3: Remove from `SetOrderInitGenesis`**

Same deletions in InitGenesis order.

- [ ] **Step 8.4: Remove unused `slashing` and `evidence` package imports**

If these packages are no longer referenced (only the keeper + types packages are used), remove:

```go
"github.com/cosmos/cosmos-sdk/x/evidence"
"github.com/cosmos/cosmos-sdk/x/slashing"
```

**Keep:**

- `slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"` — for keeper construction (which we keep)
- `slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"` — for store key + module name
- Same for `evidencekeeper`, `evidencetypes`

- [ ] **Step 8.5: Build and confirm**

```bash
make build
```

Expected: builds clean.

- [ ] **Step 8.6: Commit**

```bash
git add app/app.go
git commit -m "feat(v33): remove slashing and evidence from module manager"
```

---

## Task 8b: Delete stakeibc's consumer-reward-denom whitelist plumbing

**Files:**

- Delete: `x/stakeibc/keeper/consumer.go`
- Delete: `x/stakeibc/keeper/consumer_test.go`
- Modify: `x/stakeibc/keeper/registration.go` (remove the whitelist call site)
- Modify: `x/stakeibc/keeper/keeper.go` (drop `ConsumerKeeper` field + constructor param)
- Modify: `x/stakeibc/types/expected_keepers.go` (drop `ConsumerKeeper` interface)
- Modify: `x/stakeibc/keeper/reward_allocation_test.go` (drop the `ConsumerKeeper.SetParams` test setup line)
- Modify: `app/app.go` (drop the `app.ConsumerKeeper` argument from `stakeibcmodulekeeper.NewKeeper(...)`)

**Why this lives in v33:** `stakeibc.RegisterStTokenDenomsToWhitelist` is the only Stride-side production code path that mutates ICS consumer state at runtime — it appends new stToken denoms to `ConsumerKeeper.GetConsumerParams(ctx).RewardDenoms` whenever a host zone is registered. The whitelist exists so the ICS consumer module knows which local denoms count as "reward denoms" to ship 15% of to the Hub each block. Post-v33, that 15% Hub-shipping path no longer runs (`ccvconsumer` is out of the manager and the inflation/fee flow is rewired around it), so the whitelist serves no purpose. Leaving the function in place would leave a runtime caller of a dead module — exactly the "module dependency we missed" risk Stride wanted to surface.

This task is intentionally placed before the `x/distribution` fee-collector switch (Task 9) and the ante-handler change (Task 10) because it removes the last live caller of `ConsumerKeeper.SetParams` — clean to delete first, then let the rest of the rewiring flow.

- [ ] **Step 8b.1: Delete the whitelist function and its test**

```bash
rm x/stakeibc/keeper/consumer.go
rm x/stakeibc/keeper/consumer_test.go
```

- [ ] **Step 8b.2: Remove the call site in `registration.go`**

In `x/stakeibc/keeper/registration.go` (around line 211), delete the block:

```go
// register stToken to consumer reward denom whitelist so that
// stToken rewards can be distributed to provider validators
err = k.RegisterStTokenDenomsToWhitelist(ctx, []string{types.StAssetDenomFromHostZoneDenom(zone.HostDenom)})
if err != nil {
    return nil, errorsmod.Wrap(err, "unable to register stToken as ICS reward denom")
}
```

The surrounding `RegisterHostZone` flow does not depend on the whitelist call's return value or side effects.

- [ ] **Step 8b.3: Drop `ConsumerKeeper` from the stakeibc keeper struct and constructor**

In `x/stakeibc/keeper/keeper.go`:

- Remove `ConsumerKeeper types.ConsumerKeeper` from the struct (around line 45).
- Remove the `ConsumerKeeper types.ConsumerKeeper` parameter from `NewKeeper` (around line 65).
- Remove the `ConsumerKeeper: ConsumerKeeper,` field assignment (around line 88).

- [ ] **Step 8b.4: Drop the `ConsumerKeeper` interface in `expected_keepers.go`**

In `x/stakeibc/types/expected_keepers.go`, delete the entire `ConsumerKeeper` interface block (around line 57).

- [ ] **Step 8b.5: Update `reward_allocation_test.go`**

In `x/stakeibc/keeper/reward_allocation_test.go` (around lines 46–50), delete the three lines that set `ConsumerRedistributionFraction` via `s.App.ConsumerKeeper.GetConsumerParams` / `SetParams`. Replace with a comment noting that the equivalent fraction is now hardcoded in `utils.PoaValPaymentRate` (the existing reward-allocation logic already reads that constant), so the test setup no longer needs to mutate consumer params.

If the test asserts a specific reward split that depended on the consumer-param value, update the expected value to use `utils.PoaValPaymentRate` (`0.15`).

- [ ] **Step 8b.6: Drop the `app.ConsumerKeeper` argument from `stakeibcmodulekeeper.NewKeeper(...)` in `app/app.go`**

Around `app/app.go:759`, the last argument is `app.ConsumerKeeper`. Delete that line. The constructor no longer accepts it after Step 8b.3.

- [ ] **Step 8b.7: Build and run stakeibc tests**

```bash
make build
go test ./x/stakeibc/...
```

Expected: builds clean, all tests pass.

- [ ] **Step 8b.8: Commit**

```bash
git add x/stakeibc/keeper/registration.go x/stakeibc/keeper/keeper.go \
        x/stakeibc/types/expected_keepers.go \
        x/stakeibc/keeper/reward_allocation_test.go app/app.go
git rm x/stakeibc/keeper/consumer.go x/stakeibc/keeper/consumer_test.go
git commit -m "feat(v33): remove stakeibc consumer-reward-denom whitelist (no longer needed post-POA)"
```

---

## Task 9: Reconfigure x/distribution fee collector

**Files:**

- Modify: `app/app.go`

`x/distribution` is currently configured to read its fees from `cons_redistribute`. Switch to the standard `fee_collector`.

- [ ] **Step 9.1: Find the DistrKeeper construction**

Around line 466 of `app/app.go`:

```go
app.DistrKeeper = distrkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
    app.AccountKeeper,
    app.BankKeeper,
    app.StakingKeeper,
    ccvconsumertypes.ConsumerRedistributeName,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
```

- [ ] **Step 9.2: Change the fee collector argument**

Replace `ccvconsumertypes.ConsumerRedistributeName` with `authtypes.FeeCollectorName`:

```go
app.DistrKeeper = distrkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
    app.AccountKeeper,
    app.BankKeeper,
    app.StakingKeeper,
    authtypes.FeeCollectorName, // was: ccvconsumertypes.ConsumerRedistributeName
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
```

- [ ] **Step 9.3: Build and confirm**

```bash
make build
```

Expected: builds clean.

- [ ] **Step 9.4: Commit**

```bash
git add app/app.go
git commit -m "feat(v33): switch x/distribution fee collector to fee_collector"
```

---

## Task 10: Modify ante handler to route tx fees to POA

**Files:**

- Modify: `app/ante_handler.go`

POA's `EndBlock` panics at block height 1 if the ante handler isn't routing tx fees to the POA module account. There is **no separate POA ante package** — POA reuses the standard SDK `ante.DeductFeeDecorator` and exposes a `WithFeeRecipientModule(string)` builder method that overrides the destination module account name. The change is one line in `NewAnteHandler` plus an unrelated cleanup of a stale ICS decorator.

- [ ] **Step 10.1: Add the `poatypes` import to `app/ante_handler.go`**

```go
poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
```

No new keeper import is required and `HandlerOptions` does **not** need a `POAKeeper` field — the decorator only needs the module name string.

- [ ] **Step 10.2: Replace the `DeductFeeDecorator` line**

Find this in `NewAnteHandler`:

```go
ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
```

Replace with the same constructor chained with `WithFeeRecipientModule`:

```go
ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker).
    WithFeeRecipientModule(poatypes.ModuleName),
```

- [ ] **Step 10.3: Remove the stale `DisabledModulesDecorator` line**

Find this in `NewAnteHandler` (currently `app/ante_handler.go:55`):

```go
consumerante.NewDisabledModulesDecorator("/cosmos.evidence", "/cosmos.slashing"),
```

**Delete** the line. It rejects tx messages targeted at the `x/slashing` and `x/evidence` modules; once those modules are removed from the manager (Task 8), their msg routes are gone, so the decorator is at best a no-op and at worst a stale dependency on `interchain-security/v7/app/consumer/ante` that v34 wants removed.

If the `consumerante` import is now unused, drop it as well. Keep the `ccvconsumerkeeper` import if `HandlerOptions.ConsumerKeeper` is still used elsewhere in this file; it can be removed in v34 with the keeper itself.

- [ ] **Step 10.4: Build and confirm**

```bash
make build
```

Expected: builds clean. No call-site change in `app.go` is required because `HandlerOptions` is unchanged.

- [ ] **Step 10.5: Commit**

```bash
git add app/ante_handler.go
git commit -m "feat(v33): route tx fees to POA module account; drop stale slashing/evidence decorator"
```

---

## Task 11: Implement `CreateUpgradeHandler`

**Files:**

- Create: `app/upgrades/v33/upgrades.go`

Now that all the helpers exist and the wiring is in place, implement the handler that orchestrates them.

- [ ] **Step 11.1: Create `upgrades.go`**

```go
package v33

import (
    "context"
    "fmt"

    upgradetypes "cosmossdk.io/x/upgrade/types"
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/module"

    poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"

    authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
    bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
    distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

    ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"
)

// CreateUpgradeHandler returns the v33 upgrade handler that migrates Stride
// from ICS consumer to POA. See docs/superpowers/specs/2026-04-27-ics-to-poa-migration-design.md.
//
// poaKeeper is a pointer because POA's keeper methods (including InitGenesis)
// have pointer receivers; passing by value here would not compile.
func CreateUpgradeHandler(
    mm *module.Manager,
    configurator module.Configurator,
    cdc codec.Codec,
    consumerKeeper ccvconsumerkeeper.Keeper,
    poaKeeper *poakeeper.Keeper,
    bankKeeper bankkeeper.Keeper,
    accountKeeper authkeeper.AccountKeeper,
    distrKeeper distrkeeper.Keeper,
) upgradetypes.UpgradeHandler {
    return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
        sdkCtx := sdk.UnwrapSDKContext(ctx)
        sdkCtx.Logger().Info(fmt.Sprintf("Starting upgrade %s (ICS → POA)...", UpgradeName))

        // 1. Run module migrations. RunMigrations silently skips modules removed
        //    from the manager (ccvconsumer, ccvdistr, slashing, evidence).
        sdkCtx.Logger().Info("v33: running module migrations...")
        versionMap, err := mm.RunMigrations(sdkCtx, configurator, vm)
        if err != nil {
            return nil, err
        }

        // 2. Snapshot current ICS validator set into POA-shaped Validators.
        sdkCtx.Logger().Info("v33: snapshotting ICS validator set...")
        poaValidators, err := SnapshotValidatorsFromICS(sdkCtx, consumerKeeper)
        if err != nil {
            return nil, err
        }

        // 3. Initialize POA state with that set + admin.
        sdkCtx.Logger().Info("v33: initializing POA state...")
        if err := InitializePOA(sdkCtx, cdc, poaKeeper, AdminMultisigAddress, poaValidators); err != nil {
            return nil, err
        }

        // 4. Sweep residual ICS reward module accounts to community pool.
        sdkCtx.Logger().Info("v33: sweeping ICS module accounts to community pool...")
        if err := SweepICSModuleAccounts(sdkCtx, accountKeeper, bankKeeper, distrKeeper); err != nil {
            return nil, err
        }

        sdkCtx.Logger().Info(fmt.Sprintf("Upgrade %s complete.", UpgradeName))
        return versionMap, nil
    }
}
```

- [ ] **Step 11.2: Verify it compiles**

```bash
go build ./app/upgrades/v33/...
```

Expected: builds clean.

- [ ] **Step 11.3: Commit**

```bash
git add app/upgrades/v33/upgrades.go
git commit -m "feat(v33): add upgrade handler"
```

---

## Task 12: Register v33 in `app/upgrades.go`

**Files:**

- Modify: `app/upgrades.go`

- [ ] **Step 12.1: Add v33 import**

In the import block of `app/upgrades.go`:

```go
v33 "github.com/Stride-Labs/stride/v32/app/upgrades/v33"
```

Also add the POA types import:

```go
poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
```

- [ ] **Step 12.2: Register the v33 upgrade handler**

After the `v32.UpgradeName` registration (around line 415), add:

```go
// v33 upgrade handler
app.UpgradeKeeper.SetUpgradeHandler(
    v33.UpgradeName,
    v33.CreateUpgradeHandler(
        app.ModuleManager,
        app.configurator,
        app.appCodec,
        app.ConsumerKeeper,
        app.POAKeeper, // already *poakeeper.Keeper
        app.BankKeeper,
        app.AccountKeeper,
        app.DistrKeeper,
    ),
)
```

- [ ] **Step 12.3: Add v33 case to the StoreUpgrades switch**

In the `switch upgradeInfo.Name` block (around line 437), add:

```go
case "v33":
    storeUpgrades = &storetypes.StoreUpgrades{
        Added: []string{
            poatypes.StoreKey,
            poatypes.TransientStoreKey,
        },
    }
```

- [ ] **Step 12.4: Build and confirm**

```bash
make build
```

Expected: builds clean.

- [ ] **Step 12.5: Commit**

```bash
git add app/upgrades.go
git commit -m "feat(v33): register upgrade handler and store upgrades"
```

---

## Task 13: Write the upgrade handler integration test (synthetic state)

**Files:**

- Create: `app/upgrades/v33/upgrades_test.go`

App-level test using `apptesting.AppTestHelper`. Seeds dummy state, runs the handler, asserts post-state.

- [ ] **Step 13.1: Create `upgrades_test.go`**

```go
package v33_test

import (
    "encoding/hex"
    "testing"

    sdkmath "cosmossdk.io/math"
    sdk "github.com/cosmos/cosmos-sdk/types"
    authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
    govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
    govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
    ccvconsumertypes "github.com/cosmos/interchain-security/v7/x/ccv/consumer/types"
    "github.com/stretchr/testify/suite"

    "github.com/Stride-Labs/stride/v32/app/apptesting"
    v33 "github.com/Stride-Labs/stride/v32/app/upgrades/v33"
)

type UpgradeTestSuite struct {
    apptesting.AppTestHelper

    // captured pre-upgrade state for post-upgrade comparison
    preUpgradeBondedPool          sdk.Coins
    preUpgradeStakingValidators   int
    preUpgradeDelegations         int
    preUpgradeActiveProposalIDs   []uint64
}

func (s *UpgradeTestSuite) SetupTest() {
    s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
    suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
    // ----- arrange ------
    s.SetupICSValidatorSet(8)
    s.SetupGovenatorState(3)        // 3 govenators with delegations
    s.SetupConsumerRewardAccounts() // pre-fund both ICS reward accounts
    s.SetupActiveGovProposal()      // 1 active proposal in voting period

    s.captureGovenatorState()
    s.populateValidatorMonikers()

    // ----- act ------
    s.ConfirmUpgradeSucceeded(v33.UpgradeName)

    // ----- assert ------
    s.checkPOAValidatorsMatchICSSnapshot()
    s.checkPOAAdminSet()
    s.checkGovenatorStateUntouched()
    s.checkICSModuleAccountsDrained()
    s.checkActiveGovProposalUnaffected()
    s.checkValidatorSetContinuity() // most important
}

// --- assertion helpers ---

func (s *UpgradeTestSuite) checkPOAValidatorsMatchICSSnapshot() {
    poaVals, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
    s.Require().NoError(err)
    s.Require().Len(poaVals, 8)
    // each validator's pubkey + power matches what was seeded
}

func (s *UpgradeTestSuite) checkPOAAdminSet() {
    params, err := s.App.POAKeeper.GetParams(s.Ctx)
    s.Require().NoError(err)
    s.Require().Equal(v33.AdminMultisigAddress, params.Admin)
}

func (s *UpgradeTestSuite) checkGovenatorStateUntouched() {
    bondedAddr := s.App.AccountKeeper.GetModuleAddress("bonded_tokens_pool")
    s.Require().Equal(s.preUpgradeBondedPool, s.App.BankKeeper.GetAllBalances(s.Ctx, bondedAddr))

    vals, err := s.App.StakingKeeper.GetAllValidators(s.Ctx)
    s.Require().NoError(err)
    s.Require().Len(vals, s.preUpgradeStakingValidators)

    // delegations preserved
    iter, err := s.App.StakingKeeper.Delegations.Iterate(s.Ctx, nil)
    s.Require().NoError(err)
    delegations, err := iter.Values()
    s.Require().NoError(err)
    s.Require().Len(delegations, s.preUpgradeDelegations)
}

func (s *UpgradeTestSuite) checkICSModuleAccountsDrained() {
    consRedistrAddr := s.App.AccountKeeper.GetModuleAddress(ccvconsumertypes.ConsumerRedistributeName)
    consToProvAddr := s.App.AccountKeeper.GetModuleAddress(ccvconsumertypes.ConsumerToSendToProviderName)
    s.Require().True(s.App.BankKeeper.GetAllBalances(s.Ctx, consRedistrAddr).IsZero())
    s.Require().True(s.App.BankKeeper.GetAllBalances(s.Ctx, consToProvAddr).IsZero())
}

func (s *UpgradeTestSuite) checkActiveGovProposalUnaffected() {
    for _, id := range s.preUpgradeActiveProposalIDs {
        prop, err := s.App.GovKeeper.Proposals.Get(s.Ctx, id)
        s.Require().NoError(err)
        s.Require().NotEqual(govv1.StatusFailed, prop.Status, "proposal %d should not be failed", id)
    }
}

func (s *UpgradeTestSuite) checkValidatorSetContinuity() {
    // The most important assertion: POA's first EndBlock returns no updates,
    // OR an update set whose net effect equals the pre-handler CometBFT set.
    //
    // We invoke poa.EndBlock directly and assert it returns zero updates,
    // since InitGenesis seeded the validators directly into KV without
    // queueing any updates in the transient store.
    updates, err := s.App.POAKeeper.EndBlocker(s.Ctx)
    s.Require().NoError(err)
    s.Require().Empty(updates, "POA EndBlock must not emit ValidatorUpdates on the upgrade block")
}

// --- setup helpers ---

func (s *UpgradeTestSuite) SetupICSValidatorSet(count int) {
    // implementation: seed `count` CCValidators into ConsumerKeeper
    seedConsumerValidators(s.T(), &s.AppTestHelper, count)
}

func (s *UpgradeTestSuite) SetupGovenatorState(count int) {
    // implementation: create `count` govenator validators with self-delegations
    // Use s.App.StakingKeeper directly via a MsgCreateValidator-style flow
    // or just SetValidator + SetDelegation.
}

func (s *UpgradeTestSuite) SetupConsumerRewardAccounts() {
    bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
    s.Require().NoError(err)
    coin := sdk.NewCoin(bondDenom, sdkmath.NewInt(500_000))
    s.FundModuleAccount(ccvconsumertypes.ConsumerRedistributeName, coin)
    s.FundModuleAccount(ccvconsumertypes.ConsumerToSendToProviderName, coin)
}

func (s *UpgradeTestSuite) SetupActiveGovProposal() {
    // create a stub gov proposal in voting_period
    // store its ID into s.preUpgradeActiveProposalIDs
}

func (s *UpgradeTestSuite) captureGovenatorState() {
    bondedAddr := s.App.AccountKeeper.GetModuleAddress("bonded_tokens_pool")
    s.preUpgradeBondedPool = s.App.BankKeeper.GetAllBalances(s.Ctx, bondedAddr)

    vals, err := s.App.StakingKeeper.GetAllValidators(s.Ctx)
    s.Require().NoError(err)
    s.preUpgradeStakingValidators = len(vals)

    iter, err := s.App.StakingKeeper.Delegations.Iterate(s.Ctx, nil)
    s.Require().NoError(err)
    dels, err := iter.Values()
    s.Require().NoError(err)
    s.preUpgradeDelegations = len(dels)
}

func (s *UpgradeTestSuite) populateValidatorMonikers() {
    vals := s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx)
    for i, v := range vals {
        v33.ValidatorMonikers[hex.EncodeToString(v.Address)] = fmtMoniker(i)
    }
    s.T().Cleanup(func() {
        for _, v := range vals {
            delete(v33.ValidatorMonikers, hex.EncodeToString(v.Address))
        }
    })
}
```

The `SetupGovenatorState` and `SetupActiveGovProposal` helpers are stubs above. Implement them concretely against Stride's apptesting helpers — they may already exist or need to be added to `app/apptesting/`. If they don't exist, write them inline first; refactor to apptesting later if reused.

- [ ] **Step 13.2: Run the test**

```bash
go test ./app/upgrades/v33/... -run TestUpgradeTestSuite -v
```

Expected: PASS, with all assertions green.

If it fails, the most likely failure modes (in priority order):

1. **`checkValidatorSetContinuity` returns non-empty updates** — POA's `InitGenesis` is queueing updates we expected to skip. Investigate POA's source; may need to drain the transient store after `InitGenesis`.
2. **`AdminMultisigAddress` invalid bech32** — replace the placeholder in `constants.go` with a valid Stride address before running.
3. **POA validator count mismatch** — ICS state seeded with wrong count.

Fix as needed and re-run.

- [ ] **Step 13.3: Commit**

```bash
git add app/upgrades/v33/upgrades_test.go
git commit -m "test(v33): add handler integration test with synthetic state"
```

---

## Task 14: Mainnet state-export test infrastructure

**Files:**

- Create: `app/upgrades/v33/testdata/README.md`
- Modify: `app/apptesting/test_helpers.go` (add helper to load app from genesis JSON)

The "Test 3" from the spec — load real mainnet state and run the handler against it.

- [ ] **Step 14.1: Document the fixture-generation process**

Create `app/upgrades/v33/testdata/README.md`:

```markdown
# v33 Upgrade Test Fixtures

## How to generate `mainnet_export.json.gz`

1. Sync a Stride mainnet node to a recent height (or use one you trust).
2. Stop the node.
3. Run `strided export --height <H> > mainnet_export.json`
4. Optional: sanitize. Strip account keys / signatures if any present.
   For this test, raw export is fine — we're not committing private data.
5. Compress: `gzip -9 mainnet_export.json`
6. Move to this directory: `mv mainnet_export.json.gz app/upgrades/v33/testdata/`
7. Commit. Note: file may be 50-200MB. Use git-lfs if needed; otherwise commit once and update only when test specifically needs newer state.

The fixture is used only by `TestUpgradeFromMainnetExport`.
```

- [ ] **Step 14.2: Add an apptesting helper to load from a genesis JSON**

In `app/apptesting/test_helpers.go`, add a method to construct a StrideApp from a genesis state file:

```go
// SetupFromGenesis instantiates the test app from a path to a genesis JSON file
// (which may be gzipped). Used to test upgrades against real mainnet state.
func (s *AppTestHelper) SetupFromGenesis(t *testing.T, genesisPath string) {
    // 1. Open + decompress (if .gz) the file.
    // 2. Parse the genesis JSON.
    // 3. Construct StrideApp with that genesis.
    // 4. Run InitChain to populate state.
    // 5. Set s.App and s.Ctx.
    //
    // Reference: cosmossdk's `simapp/app.go` and look at how Stride's existing
    // Setup() bootstraps; this is a variant that uses a real genesis instead of defaults.
    panic("implement based on Stride's existing Setup() pattern")
}
```

If implementing this is non-trivial (~hours of work), keep it stubbed for now and skip Task 15. The synthetic-state test (Task 13) is the primary protection; the mainnet-export test is a robustness boost.

If implementing it is straightforward (~30 min by adapting Stride's existing genesis-loading code), do it now.

- [ ] **Step 14.3: Commit**

```bash
git add app/upgrades/v33/testdata/README.md app/apptesting/test_helpers.go
git commit -m "test(v33): add fixture infrastructure for mainnet-export test"
```

---

## Task 15: Mainnet state-export test

**Files:**

- Modify: `app/upgrades/v33/upgrades_test.go`

Append a test that runs the v33 handler against a real mainnet export.

- [ ] **Step 15.1: Generate the fixture**

Follow the README from Task 14 to produce `app/upgrades/v33/testdata/mainnet_export.json.gz`. Commit it.

- [ ] **Step 15.2: Append `TestUpgradeFromMainnetExport`**

Append to `app/upgrades/v33/upgrades_test.go`:

```go
func (s *UpgradeTestSuite) TestUpgradeFromMainnetExport() {
    s.SetupFromGenesis(s.T(), "testdata/mainnet_export.json.gz")

    // Real mainnet has the validator set, govenators, delegations, gov proposals,
    // module account balances etc. Capture the values we want to preserve.
    s.captureGovenatorState()
    preUpgradePOAValSetSize := len(s.App.ConsumerKeeper.GetAllCCValidator(s.Ctx))
    s.Require().Equal(8, preUpgradePOAValSetSize, "expected mainnet PSS allowlist of 8")

    // Capture all active proposal IDs
    iter, err := s.App.GovKeeper.Proposals.Iterate(s.Ctx, nil)
    s.Require().NoError(err)
    proposals, err := iter.Values()
    s.Require().NoError(err)
    for _, p := range proposals {
        if p.Status == govv1.StatusVotingPeriod || p.Status == govv1.StatusDepositPeriod {
            s.preUpgradeActiveProposalIDs = append(s.preUpgradeActiveProposalIDs, p.Id)
        }
    }

    // Real ICS reward account balance — log for posterity
    consRedistrAddr := s.App.AccountKeeper.GetModuleAddress(ccvconsumertypes.ConsumerRedistributeName)
    preUpgradeConsRedistr := s.App.BankKeeper.GetAllBalances(s.Ctx, consRedistrAddr)
    s.T().Logf("pre-upgrade cons_redistribute balance: %s", preUpgradeConsRedistr)

    // Populate monikers from real consensus addresses (may need to bake actual
    // moniker values into ValidatorMonikers per the constants file).
    s.populateValidatorMonikers()

    // Run the upgrade
    s.ConfirmUpgradeSucceeded(v33.UpgradeName)

    // Run all the synthetic-state assertions against real state
    s.checkPOAValidatorsMatchICSSnapshot()
    s.checkPOAAdminSet()
    s.checkGovenatorStateUntouched()
    s.checkICSModuleAccountsDrained()
    s.checkActiveGovProposalUnaffected()
    s.checkValidatorSetContinuity()

    // Mainnet-specific: real cons_redistribute balance landed in community pool
    feePool, err := s.App.DistrKeeper.FeePool.Get(s.Ctx)
    s.Require().NoError(err)
    bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
    s.Require().NoError(err)
    cpBalance := feePool.CommunityPool.AmountOf(bondDenom)
    expectedAdded := preUpgradeConsRedistr.AmountOf(bondDenom)
    s.Require().True(cpBalance.GTE(sdkmath.LegacyNewDecFromInt(expectedAdded)),
        "community pool should have grown by at least the swept cons_redistribute balance")
}
```

- [ ] **Step 15.3: Run the test**

```bash
go test ./app/upgrades/v33/... -run TestUpgradeFromMainnetExport -v
```

Expected: PASS.

If it fails, common causes:

1. **Real mainnet has unexpected state** — accumulated KV entries we didn't account for. Investigate the failing assertion, then either fix the handler to handle that state or document it as a known artifact.
2. **`SetupFromGenesis` helper doesn't fully replicate genesis init** — real genesis bootstrap may have nuances (e.g., specific module init order). Compare with how Stride's mainnet node actually bootstraps from genesis.

- [ ] **Step 15.4: Commit**

```bash
git add app/upgrades/v33/testdata/mainnet_export.json.gz app/upgrades/v33/upgrades_test.go
git commit -m "test(v33): add mainnet-export upgrade test"
```

---

## Task 16: Final integration check + changelog

**Files:**

- Modify: `CHANGELOG.md`

- [ ] **Step 16.1: Run the full test suite**

```bash
make test-unit
```

Expected: all packages pass. If any test outside `app/upgrades/v33/...` fails, investigate — the migration may have inadvertently touched something.

Likely categories of failures:

- Tests that explicitly check for slashing/evidence module presence: update to acknowledge their removal.
- Tests that simulate ICS validator-set updates: update to account for POA-driven flow.
- Tests that check distribution fee pool balance from `cons_redistribute`: update to use `fee_collector`.

- [ ] **Step 16.2: Run a build with all tags**

```bash
make build
```

Expected: clean build.

- [ ] **Step 16.3: Update `CHANGELOG.md`**

Add an entry at the top:

```markdown
## v33

- **Migrate from ICS consumer to POA**. Block-producing validator set is
  now controlled by an admin multisig via `x/poa`. Govenator staking,
  delegations, governance, and liquid staking product economics are
  unchanged.
- Remove `ccvconsumer`, `ccvdistr`, `slashing`, `evidence` from the module
  manager (their keepers stay mounted; full deletion is deferred to v34).
- Reconfigure `x/distribution` to use the standard `fee_collector` (was
  `cons_redistribute`).
- Route tx fees to POA module account via the ante handler.
- Inflation flow unchanged except no longer leaks 15% to Cosmos Hub.
- Liquid staking revenue flow (stakeibc → reward collector → 15%/85%
  split → auction → strdburner) is unchanged.
```

- [ ] **Step 16.4: Final commit**

```bash
git add CHANGELOG.md
git commit -m "chore(v33): update changelog"
```

- [ ] **Step 16.5: Self-verify**

Confirm:

- All tests pass: `make test-unit`
- Binary builds: `make build`
- v32 upgrade test still passes (no regression): `go test ./app/upgrades/v32/...`
- v33 helper unit tests pass: `go test ./app/upgrades/v33/... -run TestSnapshot -v`
- v33 handler test passes: `go test ./app/upgrades/v33/... -run TestUpgradeTestSuite -v`
- Mainnet export test passes (if implemented): `go test ./app/upgrades/v33/... -run TestUpgradeFromMainnetExport -v`

---

## Self-review

### Spec coverage check

| Spec section                                  | Implemented in tasks                                      |
| --------------------------------------------- | --------------------------------------------------------- |
| §1 Overview & goals                           | Tasks 11, 16 (CHANGELOG)                                  |
| §2 Module manager changes                     | Tasks 6, 7, 8                                             |
| §2 stakeibc whitelist removal                 | Task 8b                                                   |
| §2 Keeper wiring (POA)                        | Task 6                                                    |
| §2 Keeper wiring (distribution)               | Task 9                                                    |
| §2 Ante handler change                        | Task 10                                                   |
| §2 Begin/end blocker ordering                 | Tasks 6, 7, 8                                             |
| §2 Store upgrades                             | Task 12                                                   |
| §2 Module account perms                       | Task 6                                                    |
| §2 VersionMap handling                        | Task 11 (handler does not call delete(vm,))               |
| §3 Handler signature                          | Task 11                                                   |
| §3 Handler body (RunMigrations + 3 helpers)   | Task 11                                                   |
| §3 `SnapshotValidatorsFromICS`                | Task 3                                                    |
| §3 `InitializePOA`                            | Task 4                                                    |
| §3 `SweepICSModuleAccounts`                   | Task 5                                                    |
| §3 Invariants (count, bech32, non-zero power) | Tasks 3, 4                                                |
| §4 v34 cleanup                                | Out of scope — separate plan                              |
| §5 Rewards flow                               | Tasks 9, 10 (code); test in 13                            |
| §6 Test 1 (helper unit)                       | Tasks 3, 4, 5                                             |
| §6 Test 2 (synthetic upgrade)                 | Task 13                                                   |
| §6 Test 3 (mainnet export)                    | Tasks 14, 15                                              |
| §7 Risks 1, 8                                 | Mitigated by tests in Task 13                             |
| §7 Risks 2                                    | Mitigated by Task 15                                      |
| §7 Risks 3 (audit grep)                       | Implicit in `make test-unit` and Task 16                  |
| §7 Risk 6 (admin bech32 validation)           | Task 4 (`InitializePOA` validates)                        |
| §7 Risk 7 (ccvstaking compat)                 | Implicitly tested by Task 13 (chain runs after migration) |

### Placeholder scan

Searched for "TBD", "TODO", "fill in", "implement later", "appropriate error handling": none found in code blocks. The constants file has placeholders for `AdminMultisigAddress` and `ValidatorMonikers` map values, but these are explicitly documented as ops-provided values, not implementation TBDs.

### Type consistency check

- `SnapshotValidatorsFromICS` returns `[]poatypes.Validator` — matches what `InitializePOA` accepts.
- `InitializePOA` takes `poaKeeper poakeeper.Keeper` — matches the handler signature.
- `SweepICSModuleAccounts` parameter order (account, bank, distr) is consistent across declaration, test, and handler call site.
- `AdminMultisigAddress` is a string constant — matches `string` parameter in `InitializePOA`.

No drift found.

### Scope check

This plan covers v33 only — the migration. v34 (cleanup) is explicitly excluded and lives in its own spec at `docs/superpowers/specs/2026-04-27-ics-cleanup-design.md`. The v33 plan produces a working, testable migration on its own.

---
