# v34 POA Validator Swap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** v34 chain upgrade that swaps Citadel.one and Cosmostation out of the POA validator set for cosmosrescue and Citizen Web3, updating both POA on-chain state and the `utils/poa.go` stToken payout registry.

**Architecture:** A minimal upgrade handler (`app/upgrades/v34/`) that, after `RunMigrations`, creates the two incoming validators (`poaKeeper.CreateValidator`, power 274523, `checkpoint=true`) and zeroes the two outgoing ones (`poaKeeper.UpdateValidator`, power 0 = removal). Incoming identity lives in v34 constants (base64 ed25519 pubkeys, placeholder sentinels until confirmed); payout addresses live only in `utils.PoaValidatorSet` (joined by moniker); outgoing validators are resolved from live POA state by moniker. v33's historical code is frozen against its own copy of the old validator set first so it stops tracking the live registry.

**Tech Stack:** Go, Cosmos SDK v0.50.x, `cosmos-sdk/enterprise/poa` v1.0.0, testify suites via `app/apptesting`.

**Spec:** `docs/superpowers/specs/2026-09-09-v34-validator-swap-design.md`

## Global Constraints

- **Each consensus address gets AT MOST ONE power-changing keeper call in the upgrade** (POA queues one un-deduped ABCI update per power change; CometBFT halts every node on a duplicate address in one block's update set). The handler achieves this structurally: 2 creates + 2 zero-outs, 6 continuing validators untouched → exactly 4 ABCI entries.
- **Incoming consensus keys are ed25519 by construction**: decode base64 into `ed25519.PubKey` directly (`keeper.CreateValidator` does not validate key type; consensus params allow only ed25519).
- **`CreateValidator` is always called with `checkpoint=true`.**
- **Do NOT use `poaKeeper.UpdateValidators` (plural)** — it calls `validatePubkeyType`, which reads `ctx.ConsensusParams()`; that is empty in the apptesting upgrade path and would fail tests. Use `UpdateValidator` (singular) with an explicitly resolved cons address.
- **Power stays 274523** (`ValidatorPower`); no normalization.
- Module path stays `github.com/Stride-Labs/stride/v33` — the `/v34` bump is a separate end-of-cycle PR. All new imports use `/v33`.
- Placeholder payout addresses must be **valid bech32** (stakeibc code and tests call `sdk.MustAccAddressFromBech32` on every operator) but detectable by exact equality.
- Run `go build ./...` before every commit.

## File Map

| File | Action | Responsibility |
|---|---|---|
| `app/upgrades/v33/frozen_validator_set.go` | Create | Frozen copy of the pre-v34 8-validator set for v33's historical join |
| `app/upgrades/v33/helpers.go` | Modify | Join against frozen set instead of `utils.PoaValidatorSet` |
| `app/upgrades/v33/{helpers,upgrades,mainnet_export}_test.go` | Modify | Reference frozen set |
| `utils/poa.go` | Modify | Drop `HubAddress`; swap 2 entries; placeholder machinery |
| `app/upgrades/v34/constants.go` | Create | UpgradeName, incoming/outgoing identity, power, sentinel |
| `app/upgrades/v34/upgrades.go` | Create | Handler + `SwapPoaValidators` |
| `app/upgrades/v34/upgrades_test.go` | Create | Synthetic upgrade suite + failure cases |
| `app/upgrades.go` | Modify | Register v34 handler |
| `app/upgrades/v34/mainnet_export_test.go` | Create | Replay handler against real export (release gate) |
| `app/upgrades/v34/testdata/README.md` | Create | Fixture generation instructions |
| `CHANGELOG.md` | Modify | Unreleased entry |

---

### Task 1: Freeze v33's validator-set dependency

v33's `SnapshotValidatorsFromICS` (production code, already executed on mainnet) and its tests join monikers against the **live** `utils.PoaValidatorSet`. Editing the registry in Task 2 would break the v33 mainnet-export test (real mainnet monikers `Citadel.one`/`Cosmostation` would no longer resolve). Freeze v33 against its own copy first.

**Files:**
- Create: `app/upgrades/v33/frozen_validator_set.go`
- Modify: `app/upgrades/v33/helpers.go` (join at lines ~56-58, error at ~84, comment at ~33)
- Modify: `app/upgrades/v33/helpers_test.go`, `app/upgrades/v33/upgrades_test.go`, `app/upgrades/v33/mainnet_export_test.go`

**Interfaces:**
- Produces: `v33.FrozenValidator{Moniker, Operator string}`, `v33.FrozenValidatorSet []FrozenValidator` (8 entries). Used only within the v33 package.

- [ ] **Step 1: Create the frozen set**

`app/upgrades/v33/frozen_validator_set.go`:

```go
package v33

// FrozenValidator pins the moniker → operator join the v33 upgrade handler
// used when it executed on mainnet. utils.PoaValidatorSet is a live registry
// that later upgrades edit (v34 swaps two entries); v33 is historical and must
// not track it. Copied verbatim from utils/poa.go as of the v33 release
// (HubAddress omitted — v33's join never read it).
type FrozenValidator struct {
	Moniker  string
	Operator string
}

var FrozenValidatorSet = []FrozenValidator{
	{Moniker: "Polkachu", Operator: "stride1gp957czryfgyvxwn3tfnyy2f0t9g2p4pxxdj7c"},
	{Moniker: "L5", Operator: "stride1wj9ckvakuzgvlgw3hwpmsfjxvsc7uke73ps4u8"},
	{Moniker: "Imperator", Operator: "stride13u4dsapth4m3hef3z8qgjtdnv06predefnndkw"},
	{Moniker: "Cosmostation", Operator: "stride1jj9z2xwxesuy65n90dujsak554eqkrr2ygyan2"},
	{Moniker: "Keplr", Operator: "stride1j79tw5chf34u88s30gxchzx2cu080elm4hqg5j"},
	{Moniker: "Stakecito", Operator: "stride1qe8uuf5x69c526h4nzxwv4ltftr73v7qr7y9vq"},
	{Moniker: "Citadel.one", Operator: "stride1rgwn0h67xmuluymk4vvhtl4tqtgfg39j9zuk2z"},
	{Moniker: "CryptoCrew", Operator: "stride1smuvvnjj6w7x6ytq9kdgvlj6er99y6648s3der"},
}
```

- [ ] **Step 2: Point helpers.go at the frozen set**

In `app/upgrades/v33/helpers.go`:

Replace (around line 56):
```go
	operatorByMoniker := make(map[string]string, len(utils.PoaValidatorSet))
	for _, v := range utils.PoaValidatorSet {
		operatorByMoniker[v.Moniker] = v.Operator
	}
```
with:
```go
	operatorByMoniker := make(map[string]string, len(FrozenValidatorSet))
	for _, v := range FrozenValidatorSet {
		operatorByMoniker[v.Moniker] = v.Operator
	}
```

Replace the error message (around line 84):
```go
			return nil, fmt.Errorf(
				"validator %s (moniker %q) has no entry in utils.PoaValidatorSet",
				hexAddr, moniker,
			)
```
with:
```go
			return nil, fmt.Errorf(
				"validator %s (moniker %q) has no entry in v33.FrozenValidatorSet",
				hexAddr, moniker,
			)
```

Update the doc comment around line 33 that says "joined to a Stride-side operator address via utils.PoaValidatorSet" to say "via v33.FrozenValidatorSet". Remove the now-unused `"github.com/Stride-Labs/stride/v33/utils"` import (verify with a grep that no other `utils.` reference remains in helpers.go).

- [ ] **Step 3: Mechanical replace in the v33 test files**

```bash
cd /Users/sampocs/Documents/Projects/stride
perl -pi -e 's/utils\.PoaValidatorSet/v33.FrozenValidatorSet/g' \
  app/upgrades/v33/helpers_test.go \
  app/upgrades/v33/upgrades_test.go \
  app/upgrades/v33/mainnet_export_test.go
```

This also rewrites the string literal asserted in `helpers_test.go` (~line 117) from `"no entry in utils.PoaValidatorSet"` to `"no entry in v33.FrozenValidatorSet"`, which now matches the Step 2 error text exactly — verify both strings are identical after the replace.

Then remove the `"github.com/Stride-Labs/stride/v33/utils"` import from each of the three test files **if** no other `utils.` reference remains (check per file: `grep -n 'utils\.' app/upgrades/v33/*_test.go`). Comment prose mentioning `utils.PoaValidatorSet` will have been rewritten by the perl command; that's desired.

- [ ] **Step 4: Verify**

```bash
go build ./... && go test ./app/upgrades/v33/... ./utils/...
```
Expected: PASS (the mainnet-export suite runs too — the fixture is present locally).

- [ ] **Step 5: Commit**

```bash
git add app/upgrades/v33/
git commit -m "refactor(v33): freeze validator set join against a package-local copy

v33 is historical; its moniker->operator join must not track the live
utils.PoaValidatorSet, which v34 is about to edit."
```

---

### Task 2: Rewrite the `utils/poa.go` registry

**Files:**
- Modify: `utils/poa.go` (full rewrite below)

**Interfaces:**
- Consumes: nothing (Task 1 removed all v33 references to this file's slice).
- Produces: `utils.PoaValidator{Moniker, Operator string}` (**`HubAddress` field deleted**), `utils.PoaValidatorSet` (8 entries: 6 continuing + 2 new), `utils.PlaceholderOperatorCosmosRescue`, `utils.PlaceholderOperatorCitizenWeb3` (const strings), `utils.IsPlaceholderOperator(operator string) bool`. Task 3's handler and tests rely on all of these exact names.
- Depends on: Task 1.

- [ ] **Step 1: Rewrite `utils/poa.go`**

Full new content:

```go
package utils

import sdkmath "cosmossdk.io/math"

// WARNING: DO NOT MODIFY outside of a coordinated validator-set upgrade.
// This registry drives the stToken reward payout in
// x/stakeibc/keeper/reward_allocation.go AND is the source of truth for POA
// operator (payout) addresses joined by upgrade handlers. Entries must stay in
// sync with the on-chain POA validator set.

// Validators are paid 15% of revenue
var PoaValPaymentRate = sdkmath.LegacyMustNewDecFromStr("0.15")

// Placeholder payout addresses for incoming validators whose payout address is
// not yet confirmed. Deterministic sha256-derived addresses with no known
// private key — valid bech32 (reward-allocation code parses every operator
// with MustAccAddressFromBech32) but unmistakably not a real wallet.
// The v34 upgrade handler refuses to run while any of these remain in the set.
//
// Derivation: bech32("stride", sha256("v34-placeholder-payout-<name>")[:20])
const (
	PlaceholderOperatorCosmosRescue = "stride1kddnkeu5ccca350thdhs2087268x4w3mfxy8s5"
	PlaceholderOperatorCitizenWeb3  = "stride1yqestx8f9z4sx9yct5ew4jkk45ntqevrwkkcpy"
)

// IsPlaceholderOperator reports whether the operator address is one of the
// not-yet-confirmed placeholder payout addresses.
func IsPlaceholderOperator(operator string) bool {
	return operator == PlaceholderOperatorCosmosRescue || operator == PlaceholderOperatorCitizenWeb3
}

type PoaValidator struct {
	Moniker  string
	Operator string // sdk.AccAddress bech32 — the payout + POA OperatorAddress
}

var PoaValidatorSet = []PoaValidator{
	{Moniker: "Polkachu", Operator: "stride1gp957czryfgyvxwn3tfnyy2f0t9g2p4pxxdj7c"},
	{Moniker: "L5", Operator: "stride1wj9ckvakuzgvlgw3hwpmsfjxvsc7uke73ps4u8"},
	{Moniker: "Imperator", Operator: "stride13u4dsapth4m3hef3z8qgjtdnv06predefnndkw"},
	{Moniker: "Keplr", Operator: "stride1j79tw5chf34u88s30gxchzx2cu080elm4hqg5j"},
	{Moniker: "Stakecito", Operator: "stride1qe8uuf5x69c526h4nzxwv4ltftr73v7qr7y9vq"},
	{Moniker: "CryptoCrew", Operator: "stride1smuvvnjj6w7x6ytq9kdgvlj6er99y6648s3der"},
	// v34 additions — placeholder payout addresses until the validators
	// confirm real ones (see PlaceholderOperator* above).
	{Moniker: "cosmosrescue", Operator: PlaceholderOperatorCosmosRescue},
	{Moniker: "Citizen Web3", Operator: PlaceholderOperatorCitizenWeb3},
}
```

Note the removals: `Citadel.one` and `Cosmostation` entries, the `HubAddress` field, and every `HubAddress:` value.

- [ ] **Step 2: Verify nothing else referenced `HubAddress` or the removed entries**

```bash
grep -rn "HubAddress" --include='*.go' . ; echo "exit: $?"
```
Expected: no matches (exit 1).

```bash
go build ./... && go test ./x/stakeibc/keeper/... ./app/upgrades/v33/... ./utils/...
```
Expected: PASS. (`reward_allocation_test.go` iterates the slice dynamically and the placeholders are valid, distinct bech32 addresses, so its per-validator balance assertions still hold.)

- [ ] **Step 3: Commit**

```bash
git add utils/poa.go
git commit -m "feat(v34): swap PoaValidatorSet entries and drop dead HubAddress field

Citadel.one and Cosmostation out; cosmosrescue and Citizen Web3 in with
placeholder payout addresses. HubAddress has been dead since the ICS
migration."
```

---

### Task 3: v34 upgrade handler, wiring, and synthetic test suite

**Files:**
- Create: `app/upgrades/v34/constants.go`
- Create: `app/upgrades/v34/upgrades.go`
- Create: `app/upgrades/v34/upgrades_test.go`
- Modify: `app/upgrades.go` (import block + after the v33 handler registration at ~line 446)

**Interfaces:**
- Consumes: `utils.PoaValidatorSet`, `utils.IsPlaceholderOperator`, `utils.PlaceholderOperatorCosmosRescue` (Task 2); POA keeper API: `CreateValidator(ctx, consAddr, validator, checkpoint) error`, `UpdateValidator(ctx, consAddr, updates) error`, `GetAllValidators(ctx) ([]poatypes.Validator, error)`, `GetTotalPower(ctx) (int64, error)`, `ReapValidatorUpdates(ctx) []abci.ValidatorUpdate`.
- Produces: `v34.UpgradeName`, `v34.CreateUpgradeHandler(mm, configurator, cdc, poaKeeper)`, `v34.SwapPoaValidators(ctx, cdc, poaKeeper) error`, `v34.IncomingValidators` (mutable var for tests), `v34.OutgoingMonikers`, `v34.ValidatorPower`, `v34.PlaceholderConsPubKey`. Task 4 relies on all of these.
- Depends on: Tasks 1-2.

- [ ] **Step 1: Write the failing test**

`app/upgrades/v34/upgrades_test.go`:

```go
package v34_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v33/app/apptesting"
	v34 "github.com/Stride-Labs/stride/v33/app/upgrades/v34"
	"github.com/Stride-Labs/stride/v33/utils"
)

// continuingMonikers are the POA validators the upgrade must not touch.
var continuingMonikers = []string{"Polkachu", "L5", "Imperator", "Keplr", "Stakecito", "CryptoCrew"}

var outgoingMonikers = []string{"Citadel.one", "Cosmostation"}

type UpgradeTestSuite struct {
	apptesting.AppTestHelper

	// consensus pubkeys generated for the incoming validators, injected into
	// the v34.IncomingValidators placeholders by fillPlaceholders
	incomingPubKeys map[string]cryptotypes.PubKey

	// consensus pubkeys and operator addresses of the seeded current set
	seededPubKeys   map[string]cryptotypes.PubKey
	seededOperators map[string]string

	preUpgradeTotalPower  int64
	preUpgradeUpdateCount int
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// fillPlaceholders substitutes test values for the release-time placeholders:
// generated ed25519 keys for the incoming consensus pubkeys, and random
// accounts for the placeholder payout addresses in utils.PoaValidatorSet.
// (Package-level vars are mutated; tests in this package run serially and
// every test that needs filled values calls this in arrange.)
func (s *UpgradeTestSuite) fillPlaceholders() {
	s.incomingPubKeys = map[string]cryptotypes.PubKey{}
	for i := range v34.IncomingValidators {
		moniker := v34.IncomingValidators[i].Moniker
		pubKey := ed25519.GenPrivKeyFromSecret([]byte("incoming-" + moniker)).PubKey()
		s.incomingPubKeys[moniker] = pubKey
		v34.IncomingValidators[i].ConsPubKeyBase64 = base64.StdEncoding.EncodeToString(pubKey.Bytes())
	}
	for i := range utils.PoaValidatorSet {
		if utils.IsPlaceholderOperator(utils.PoaValidatorSet[i].Operator) {
			utils.PoaValidatorSet[i].Operator = apptesting.CreateRandomAccounts(1)[0].String()
		}
	}
}

// seedPOASet seeds one POA validator per moniker at ValidatorPower, on top of
// the single power-1 genesis test validator.
func (s *UpgradeTestSuite) seedPOASet(monikers []string) {
	s.seededPubKeys = map[string]cryptotypes.PubKey{}
	s.seededOperators = map[string]string{}
	operators := apptesting.CreateRandomAccounts(len(monikers))

	for i, moniker := range monikers {
		pubKey := ed25519.GenPrivKeyFromSecret([]byte("existing-" + moniker)).PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		s.Require().NoError(err)

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  v34.ValidatorPower,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         moniker,
				OperatorAddress: operators[i].String(),
			},
		}
		err = s.App.POAKeeper.CreateValidator(s.Ctx, sdk.GetConsAddress(pubKey), validator, true)
		s.Require().NoError(err)

		s.seededPubKeys[moniker] = pubKey
		s.seededOperators[moniker] = operators[i].String()
	}
}

func (s *UpgradeTestSuite) seedCurrentPOASet() {
	s.seedPOASet(append(append([]string{}, continuingMonikers...), outgoingMonikers...))
}

func (s *UpgradeTestSuite) capturePreUpgradeState() {
	totalPower, err := s.App.POAKeeper.GetTotalPower(s.Ctx)
	s.Require().NoError(err)
	s.preUpgradeTotalPower = totalPower
	s.preUpgradeUpdateCount = len(s.App.POAKeeper.ReapValidatorUpdates(s.Ctx))
}

// validatorsByMoniker unpacks every POA validator into a moniker-keyed map.
func (s *UpgradeTestSuite) validatorsByMoniker() map[string]poatypes.Validator {
	validators, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)

	byMoniker := map[string]poatypes.Validator{}
	for _, validator := range validators {
		s.Require().NotNil(validator.Metadata)
		byMoniker[validator.Metadata.Moniker] = validator
	}
	return byMoniker
}

func (s *UpgradeTestSuite) TestUpgrade() {
	// ----- arrange -----
	s.fillPlaceholders()
	s.seedCurrentPOASet()
	s.capturePreUpgradeState()

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v34.UpgradeName)

	// ----- assert -----
	byMoniker := s.validatorsByMoniker()

	// Incoming: present at ValidatorPower with the generated pubkey and the
	// (test-filled) payout address from utils.PoaValidatorSet.
	operatorByMoniker := map[string]string{}
	for _, v := range utils.PoaValidatorSet {
		operatorByMoniker[v.Moniker] = v.Operator
	}
	for _, entry := range v34.IncomingValidators {
		validator, ok := byMoniker[entry.Moniker]
		s.Require().True(ok, "incoming validator %s missing from POA", entry.Moniker)
		s.Require().Equal(v34.ValidatorPower, validator.Power)
		s.Require().Equal(operatorByMoniker[entry.Moniker], validator.Metadata.OperatorAddress)

		var pubKey cryptotypes.PubKey
		s.Require().NoError(s.App.AppCodec().UnpackAny(validator.PubKey, &pubKey))
		s.Require().True(pubKey.Equals(s.incomingPubKeys[entry.Moniker]))
	}

	// Outgoing: soft-deleted — record retained at power 0 with metadata intact.
	for _, moniker := range outgoingMonikers {
		validator, ok := byMoniker[moniker]
		s.Require().True(ok, "outgoing validator %s record should be retained", moniker)
		s.Require().Zero(validator.Power)
		s.Require().Equal(s.seededOperators[moniker], validator.Metadata.OperatorAddress)
	}

	// Continuing: untouched.
	for _, moniker := range continuingMonikers {
		s.Require().Equal(v34.ValidatorPower, byMoniker[moniker].Power, "continuing validator %s power changed", moniker)
	}

	// Total power unchanged (+2*274523 from creates, -2*274523 from removals).
	totalPower, err := s.App.POAKeeper.GetTotalPower(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(s.preUpgradeTotalPower, totalPower)

	s.checkEmittedUpdates()
}

// checkEmittedUpdates is the direct regression guard against the chain-halt
// hazard: the upgrade must queue exactly 4 ABCI updates (2 creates + 2
// removals), each for a distinct consensus pubkey. A duplicate pubkey in one
// block's update set panics every CometBFT node.
func (s *UpgradeTestSuite) checkEmittedUpdates() {
	allUpdates := s.App.POAKeeper.ReapValidatorUpdates(s.Ctx)
	s.Require().GreaterOrEqual(len(allUpdates), s.preUpgradeUpdateCount)
	newUpdates := allUpdates[s.preUpgradeUpdateCount:]
	s.Require().Len(newUpdates, 4, "upgrade must emit exactly 4 validator updates")

	powersByPubKey := map[string]int64{}
	for _, update := range newUpdates {
		key := update.PubKey.String()
		_, duplicate := powersByPubKey[key]
		s.Require().False(duplicate, "duplicate consensus pubkey in emitted updates — this would halt every node")
		powersByPubKey[key] = update.Power
	}

	for moniker, pubKey := range s.incomingPubKeys {
		cmtKey, err := cryptocodec.ToCmtProtoPublicKey(pubKey)
		s.Require().NoError(err)
		s.Require().Equal(v34.ValidatorPower, powersByPubKey[cmtKey.String()],
			"incoming validator %s should be created at ValidatorPower", moniker)
	}
	for _, moniker := range outgoingMonikers {
		cmtKey, err := cryptocodec.ToCmtProtoPublicKey(s.seededPubKeys[moniker])
		s.Require().NoError(err)
		power, ok := powersByPubKey[cmtKey.String()]
		s.Require().True(ok, "outgoing validator %s should have a removal update", moniker)
		s.Require().Zero(power)
	}
}

func (s *UpgradeTestSuite) TestSwapFailsWithPlaceholderPubkey() {
	s.fillPlaceholders()
	v34.IncomingValidators[0].ConsPubKeyBase64 = v34.PlaceholderConsPubKey
	s.seedCurrentPOASet()

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, "placeholder consensus pubkey")
}

func (s *UpgradeTestSuite) TestSwapFailsWithPlaceholderPayoutAddress() {
	s.fillPlaceholders()
	for i := range utils.PoaValidatorSet {
		if utils.PoaValidatorSet[i].Moniker == "cosmosrescue" {
			utils.PoaValidatorSet[i].Operator = utils.PlaceholderOperatorCosmosRescue
		}
	}
	s.seedCurrentPOASet()

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, "placeholder payout address")
}

func (s *UpgradeTestSuite) TestSwapFailsWhenOutgoingValidatorMissing() {
	s.fillPlaceholders()
	// Seed everyone except Citadel.one.
	s.seedPOASet(append(append([]string{}, continuingMonikers...), "Cosmostation"))

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, `"Citadel.one" not found`)
}

func (s *UpgradeTestSuite) TestSwapFailsWhenIncomingMissingFromRegistry() {
	s.fillPlaceholders()
	// Break the moniker join for cosmosrescue.
	for i := range utils.PoaValidatorSet {
		if utils.PoaValidatorSet[i].Moniker == "cosmosrescue" {
			utils.PoaValidatorSet[i].Moniker = "not-cosmosrescue"
		}
	}
	s.seedCurrentPOASet()

	err := v34.SwapPoaValidators(s.Ctx, s.App.AppCodec(), s.App.POAKeeper)
	s.Require().ErrorContains(err, "no entry in utils.PoaValidatorSet")

	// Restore for subsequent tests.
	for i := range utils.PoaValidatorSet {
		if utils.PoaValidatorSet[i].Moniker == "not-cosmosrescue" {
			utils.PoaValidatorSet[i].Moniker = "cosmosrescue"
		}
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
go test ./app/upgrades/v34/... 2>&1 | head -20
```
Expected: FAIL — package `v34` does not exist / undefined symbols.

- [ ] **Step 3: Write `constants.go`**

`app/upgrades/v34/constants.go`:

```go
package v34

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v34"

	// PlaceholderConsPubKey marks a consensus pubkey the incoming validator has
	// not yet confirmed (via `strided tendermint show-validator`). The upgrade
	// handler refuses to run while any incoming validator still carries it.
	PlaceholderConsPubKey = "PLACEHOLDER"

	// ValidatorPower matches the uniform power of the existing POA set — the
	// value snapshotted from ICS at the v33 migration. POA only weighs relative
	// power, so incoming validators join at the same weight and total power is
	// unchanged by the swap.
	ValidatorPower = int64(274523)
)

// IncomingValidator identifies a validator added to the POA set by this
// upgrade. The payout/operator address deliberately does NOT live here — the
// handler joins it from utils.PoaValidatorSet by moniker so the address has a
// single source of truth.
type IncomingValidator struct {
	Moniker          string
	ConsPubKeyBase64 string // base64-encoded ed25519 consensus pubkey
}

// vars rather than consts so tests can substitute filled-in values.
var (
	IncomingValidators = []IncomingValidator{
		{Moniker: "cosmosrescue", ConsPubKeyBase64: PlaceholderConsPubKey},
		{Moniker: "Citizen Web3", ConsPubKeyBase64: PlaceholderConsPubKey},
	}

	// OutgoingMonikers are resolved against live POA state at upgrade time —
	// no hand-transcribed consensus addresses to typo.
	OutgoingMonikers = []string{"Citadel.one", "Cosmostation"}
)
```

- [ ] **Step 4: Write `upgrades.go`**

`app/upgrades/v34/upgrades.go`:

```go
package v34

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/Stride-Labs/stride/v33/utils"
)

// CreateUpgradeHandler returns the v34 upgrade handler, which swaps two POA
// validators. See docs/superpowers/specs/2026-09-09-v34-validator-swap-design.md.
//
// poaKeeper is a pointer because POA's keeper methods have pointer receivers.
// cdc unpacks the stored consensus-pubkey Anys when resolving the outgoing
// validators' consensus addresses.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	poaKeeper *poakeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s (POA validator swap)...", UpgradeName))

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		if err := SwapPoaValidators(ctx, cdc, poaKeeper); err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete", UpgradeName))
		return vm, nil
	}
}

type incomingValidator struct {
	consAddress sdk.ConsAddress
	validator   poatypes.Validator
}

type outgoingValidator struct {
	moniker     string
	consAddress sdk.ConsAddress
}

// SwapPoaValidators adds the incoming validators to the POA set and removes
// (power → 0) the outgoing ones.
//
// Consensus-safety invariant: each consensus address gets exactly ONE
// power-changing keeper call. POA queues one ABCI update per power change
// with no dedup, and CometBFT panics every node on a duplicate consensus
// address in a single block's update set. The 6 continuing validators are
// deliberately never touched.
func SwapPoaValidators(ctx sdk.Context, cdc codec.Codec, poaKeeper *poakeeper.Keeper) error {
	incoming, err := buildIncomingValidators()
	if err != nil {
		return err
	}
	outgoing, err := resolveOutgoingValidators(ctx, cdc, poaKeeper)
	if err != nil {
		return err
	}

	// Create before removing so total power never dips during the transition.
	// checkpoint=true so unallocated fees are checkpointed to the pre-existing
	// set before the new validator becomes eligible (genesis is the only
	// correct checkpoint=false caller).
	for _, validator := range incoming {
		ctx.Logger().Info(fmt.Sprintf("v34: adding POA validator %s", validator.validator.Metadata.Moniker))
		if err := poaKeeper.CreateValidator(ctx, validator.consAddress, validator.validator, true); err != nil {
			return fmt.Errorf("failed to create POA validator %s: %w", validator.validator.Metadata.Moniker, err)
		}
	}

	// Power 0 removes the validator from the active set; nil Metadata/PubKey
	// preserve the stored record so accrued fees stay withdrawable.
	for _, validator := range outgoing {
		ctx.Logger().Info(fmt.Sprintf("v34: removing POA validator %s", validator.moniker))
		if err := poaKeeper.UpdateValidator(ctx, validator.consAddress, poatypes.Validator{Power: 0}); err != nil {
			return fmt.Errorf("failed to remove POA validator %s: %w", validator.moniker, err)
		}
	}

	return nil
}

// buildIncomingValidators validates the incoming constants and joins each
// moniker to its payout address in utils.PoaValidatorSet.
//
// The pubkey is decoded into a concrete ed25519.PubKey (rather than accepting
// an arbitrary Any) because consensus params allow only ed25519 and
// keeper.CreateValidator does not validate key types — a wrong key type would
// reach CometBFT and halt the chain at EndBlock.
func buildIncomingValidators() ([]incomingValidator, error) {
	operatorByMoniker := make(map[string]string, len(utils.PoaValidatorSet))
	for _, v := range utils.PoaValidatorSet {
		operatorByMoniker[v.Moniker] = v.Operator
	}

	incoming := make([]incomingValidator, 0, len(IncomingValidators))
	for _, entry := range IncomingValidators {
		if entry.ConsPubKeyBase64 == PlaceholderConsPubKey {
			return nil, fmt.Errorf("incoming validator %q still has a placeholder consensus pubkey", entry.Moniker)
		}
		keyBytes, err := base64.StdEncoding.DecodeString(entry.ConsPubKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("incoming validator %q consensus pubkey is not valid base64: %w", entry.Moniker, err)
		}
		if len(keyBytes) != ed25519.PubKeySize {
			return nil, fmt.Errorf("incoming validator %q consensus pubkey has %d bytes, expected %d (ed25519)",
				entry.Moniker, len(keyBytes), ed25519.PubKeySize)
		}

		operator, ok := operatorByMoniker[entry.Moniker]
		if !ok {
			return nil, fmt.Errorf("incoming validator %q has no entry in utils.PoaValidatorSet", entry.Moniker)
		}
		if utils.IsPlaceholderOperator(operator) {
			return nil, fmt.Errorf("incoming validator %q still has a placeholder payout address", entry.Moniker)
		}
		if _, err := sdk.AccAddressFromBech32(operator); err != nil {
			return nil, fmt.Errorf("incoming validator %q payout address is invalid: %w", entry.Moniker, err)
		}

		pubKey := &ed25519.PubKey{Key: keyBytes}
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to pack pubkey for %q: %w", entry.Moniker, err)
		}

		incoming = append(incoming, incomingValidator{
			consAddress: sdk.GetConsAddress(pubKey),
			validator: poatypes.Validator{
				PubKey: pubKeyAny,
				Power:  ValidatorPower,
				Metadata: &poatypes.ValidatorMetadata{
					Moniker:         entry.Moniker,
					OperatorAddress: operator,
				},
			},
		})
	}
	return incoming, nil
}

// resolveOutgoingValidators maps each outgoing moniker to exactly one
// consensus address in live POA state, erroring loudly on a missing or
// ambiguous match.
func resolveOutgoingValidators(ctx sdk.Context, cdc codec.Codec, poaKeeper *poakeeper.Keeper) ([]outgoingValidator, error) {
	validators, err := poaKeeper.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	consAddressesByMoniker := make(map[string][]sdk.ConsAddress)
	for _, validator := range validators {
		if validator.Metadata == nil {
			continue
		}
		var pubKey cryptotypes.PubKey
		if err := cdc.UnpackAny(validator.PubKey, &pubKey); err != nil {
			return nil, fmt.Errorf("failed to unpack pubkey for POA validator %q: %w", validator.Metadata.Moniker, err)
		}
		moniker := validator.Metadata.Moniker
		consAddressesByMoniker[moniker] = append(consAddressesByMoniker[moniker], sdk.GetConsAddress(pubKey))
	}

	outgoing := make([]outgoingValidator, 0, len(OutgoingMonikers))
	for _, moniker := range OutgoingMonikers {
		matches := consAddressesByMoniker[moniker]
		if len(matches) == 0 {
			return nil, fmt.Errorf("outgoing validator %q not found in POA state", moniker)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("outgoing validator moniker %q matches %d POA validators, refusing to guess", moniker, len(matches))
		}
		outgoing = append(outgoing, outgoingValidator{moniker: moniker, consAddress: matches[0]})
	}
	return outgoing, nil
}
```

- [ ] **Step 5: Register the handler in `app/upgrades.go`**

Add to the import block (after the v33 import):
```go
	v34 "github.com/Stride-Labs/stride/v33/app/upgrades/v34"
```

Add immediately after the v33 `SetUpgradeHandler` block (after ~line 446):
```go
	// v34 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v34.UpgradeName,
		v34.CreateUpgradeHandler(
			app.ModuleManager,
			app.configurator,
			app.appCodec,
			app.POAKeeper,
		),
	)
```

- [ ] **Step 6: Run tests**

```bash
go build ./... && go test ./app/upgrades/v34/... -v 2>&1 | tail -20
```
Expected: PASS — `TestUpgrade`, `TestSwapFailsWithPlaceholderPubkey`, `TestSwapFailsWithPlaceholderPayoutAddress`, `TestSwapFailsWhenOutgoingValidatorMissing`, `TestSwapFailsWhenIncomingMissingFromRegistry` all pass. (The mainnet-export suite does not exist yet.)

- [ ] **Step 7: Commit**

```bash
git add app/upgrades/v34/ app/upgrades.go
git commit -m "feat(v34): upgrade handler swapping POA validators

Adds cosmosrescue and Citizen Web3 (placeholder pubkeys/payouts until
confirmed), removes Citadel.one and Cosmostation via power-0 soft delete.
Each consensus address is touched by exactly one power-changing call so the
upgrade block emits exactly 4 unique ABCI updates."
```

---

## Parallel-safe tasks

Tasks 4 and 5 are independent of each other (different files, no shared interfaces) and can run concurrently after Task 3.

### Task 4: v34 mainnet export test suite

Replays the handler against a real post-v33 `strided export`, mirroring the v33 harness. Skips when the fixture is absent (it won't be committed until release prep); once the fixture exists it fails while placeholders are unfilled — the release gate.

**Files:**
- Create: `app/upgrades/v34/mainnet_export_test.go`
- Create: `app/upgrades/v34/testdata/README.md`

**Interfaces:**
- Consumes: `v34.UpgradeName`, `v34.IncomingValidators`, `v34.ValidatorPower` (Task 3); `utils.PoaValidatorSet` (Task 2).
- Depends on: Tasks 1-3.

- [ ] **Step 1: Write the suite**

`app/upgrades/v34/mainnet_export_test.go`:

```go
package v34_test

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v33/app/apptesting"
	v34 "github.com/Stride-Labs/stride/v33/app/upgrades/v34"
	"github.com/Stride-Labs/stride/v33/utils"
)

// mainnetExportPath is relative to this package — read directly from the
// testdata/ checkout, not shipped in the binary.
const mainnetExportPath = "testdata/mainnet_export.json.gz"

// MainnetExportTestSuite replays the v34 handler against real post-v33
// mainnet POA state. Unlike the synthetic suite, it runs with the REAL
// constants — no placeholder substitution — so it doubles as the release
// gate: it fails until the incoming validators' pubkeys and payout addresses
// are filled in. The fixture is only committed during release prep; the
// suite skips when it is absent so CI stays green in the meantime.
type MainnetExportTestSuite struct {
	apptesting.AppTestHelper

	preUpgradeTotalPower  int64
	preUpgradeUpdateCount int
}

func (s *MainnetExportTestSuite) SetupTest() {
	s.Setup()
}

func TestMainnetExportTestSuite(t *testing.T) {
	if _, err := os.Stat(mainnetExportPath); errors.Is(err, os.ErrNotExist) {
		t.Skipf("skipping: mainnet export fixture not present at %s — see testdata/README.md to generate it", mainnetExportPath)
	}
	suite.Run(t, new(MainnetExportTestSuite))
}

// strideExport is a thin view over the trimmed `strided export` JSON shape.
type strideExport struct {
	AppState map[string]json.RawMessage `json:"app_state"`
}

func (s *MainnetExportTestSuite) TestUpgradeFromMainnetExport() {
	// ----- arrange: seed POA with the real mainnet validator set -----
	export := s.loadTrimmedExport()
	exportValidators := s.populatePOAFromExport(export)

	totalPower, err := s.App.POAKeeper.GetTotalPower(s.Ctx)
	s.Require().NoError(err)
	s.preUpgradeTotalPower = totalPower
	s.preUpgradeUpdateCount = len(s.App.POAKeeper.ReapValidatorUpdates(s.Ctx))

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v34.UpgradeName)

	// ----- assert -----
	byMoniker := s.validatorsByMoniker()

	operatorByMoniker := map[string]string{}
	for _, v := range utils.PoaValidatorSet {
		operatorByMoniker[v.Moniker] = v.Operator
	}
	for _, entry := range v34.IncomingValidators {
		validator, ok := byMoniker[entry.Moniker]
		s.Require().True(ok, "incoming validator %s missing from POA", entry.Moniker)
		s.Require().Equal(v34.ValidatorPower, validator.Power)
		s.Require().Equal(operatorByMoniker[entry.Moniker], validator.Metadata.OperatorAddress)

		// The real constant must decode to the real key.
		expectedKeyBytes, err := base64.StdEncoding.DecodeString(entry.ConsPubKeyBase64)
		s.Require().NoError(err)
		var pubKey cryptotypes.PubKey
		s.Require().NoError(s.App.AppCodec().UnpackAny(validator.PubKey, &pubKey))
		s.Require().True(pubKey.Equals(&ed25519.PubKey{Key: expectedKeyBytes}))
	}

	for _, moniker := range []string{"Citadel.one", "Cosmostation"} {
		validator, ok := byMoniker[moniker]
		s.Require().True(ok, "outgoing validator %s record should be retained", moniker)
		s.Require().Zero(validator.Power)
	}

	// Continuing mainnet validators untouched.
	for moniker, exportPower := range exportValidators {
		if moniker == "Citadel.one" || moniker == "Cosmostation" {
			continue
		}
		s.Require().Equal(exportPower, byMoniker[moniker].Power, "continuing validator %s power changed", moniker)
	}

	postTotalPower, err := s.App.POAKeeper.GetTotalPower(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(s.preUpgradeTotalPower, postTotalPower)

	// Exactly 4 new updates, unique pubkeys — the chain-halt guard.
	allUpdates := s.App.POAKeeper.ReapValidatorUpdates(s.Ctx)
	newUpdates := allUpdates[s.preUpgradeUpdateCount:]
	s.Require().Len(newUpdates, 4)
	seen := map[string]bool{}
	for _, update := range newUpdates {
		key := update.PubKey.String()
		s.Require().False(seen[key], "duplicate consensus pubkey in emitted updates")
		seen[key] = true
	}
}

// loadTrimmedExport reads testdata/mainnet_export.json.gz.
func (s *MainnetExportTestSuite) loadTrimmedExport() strideExport {
	f, err := os.Open(mainnetExportPath)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = f.Close() })

	gz, err := gzip.NewReader(f)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = gz.Close() })

	var export strideExport
	s.Require().NoError(json.NewDecoder(gz).Decode(&export))
	s.Require().NotEmpty(export.AppState, "trimmed export has no app_state — check testdata/README.md")
	return export
}

// populatePOAFromExport seeds POA with every validator from the export's
// app_state.poa section and returns moniker → power for later comparison.
// The test app's genesis POA validator (moniker "test-validator") is left in
// place; its consensus key cannot collide with mainnet keys.
func (s *MainnetExportTestSuite) populatePOAFromExport(export strideExport) map[string]int64 {
	raw, ok := export.AppState["poa"]
	s.Require().True(ok, "trimmed export missing poa section")

	var genesis poatypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &genesis))
	s.Require().Len(genesis.Validators, 8, "mainnet export should contain exactly 8 POA validators")

	powers := map[string]int64{}
	for _, validator := range genesis.Validators {
		var pubKey cryptotypes.PubKey
		s.Require().NoError(s.App.AppCodec().UnpackAny(validator.PubKey, &pubKey))
		err := s.App.POAKeeper.CreateValidator(s.Ctx, sdk.GetConsAddress(pubKey), validator, true)
		s.Require().NoError(err)
		powers[validator.Metadata.Moniker] = validator.Power
	}

	s.Require().Contains(powers, "Citadel.one", "export must contain the outgoing validators")
	s.Require().Contains(powers, "Cosmostation", "export must contain the outgoing validators")
	return powers
}

// validatorsByMoniker unpacks every POA validator into a moniker-keyed map.
func (s *MainnetExportTestSuite) validatorsByMoniker() map[string]poatypes.Validator {
	validators, err := s.App.POAKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)

	byMoniker := map[string]poatypes.Validator{}
	for _, validator := range validators {
		s.Require().NotNil(validator.Metadata)
		byMoniker[validator.Metadata.Moniker] = validator
	}
	return byMoniker
}
```

- [ ] **Step 2: Write `testdata/README.md`**

`app/upgrades/v34/testdata/README.md`:

```markdown
# v34 mainnet export fixture

`mainnet_export.json.gz` is a trimmed `strided export` of post-v33 mainnet
state, containing only the `poa` module section. It is NOT committed by
default — the v34 mainnet-export test suite skips when it is absent.

Generate during release prep (requires a synced post-v33 node):

​```bash
strided export --height <recent height> > full_export.json
jq '{app_state: {poa: .app_state.poa}}' full_export.json > trimmed.json
gzip -c trimmed.json > app/upgrades/v34/testdata/mainnet_export.json.gz
​```

The suite runs with the REAL v34 constants (no placeholder substitution), so
it fails until the incoming validators' consensus pubkeys and payout
addresses are filled in — run it as the final release gate before tagging.
```

(Remove the zero-width escapes around the inner code fence when writing the actual file — the inner block is a normal bash fence.)

- [ ] **Step 3: Verify**

```bash
go build ./... && go test ./app/upgrades/v34/... -run TestMainnetExportTestSuite -v
```
Expected: `SKIP` with the "fixture not present" message (no fixture committed), and the package still compiles.

- [ ] **Step 4: Commit**

```bash
git add app/upgrades/v34/mainnet_export_test.go app/upgrades/v34/testdata/
git commit -m "test(v34): mainnet export replay suite as release gate"
```

### Task 5: CHANGELOG entry

**Files:**
- Modify: `CHANGELOG.md`

**Interfaces:**
- Depends on: Tasks 1-3 (describes them; no code dependency).

- [ ] **Step 1: Add the Unreleased section**

Insert immediately above the `## [v33.0.0]` line (~line 45):

```markdown
## Unreleased

### State Machine Breaking

* (upgrades) v34: swap POA validators — remove Citadel.one and Cosmostation, add cosmosrescue and Citizen Web3 (POA set + `utils/poa.go` stToken payout registry).

```

- [ ] **Step 2: Commit**

```bash
git add CHANGELOG.md
git commit -m "docs(changelog): v34 POA validator swap entry"
```

---

## Post-plan release checklist (not tasks — tracked outside the repo)

1. Collect confirmed consensus pubkeys (validators run `strided tendermint show-validator` and echo the key in writing) and fresh payout addresses; replace `PlaceholderConsPubKey` values in `app/upgrades/v34/constants.go` and `PlaceholderOperator*` values in `utils/poa.go`.
2. Generate the mainnet export fixture (Task 4 README) and run the full suite — it must pass before tagging.
3. Both incoming nodes synced and signing-ready before the upgrade height (consensus needs 6-of-8 during the transition; the 6 continuing validators provide exactly that — zero margin until the new nodes sign at upgrade height + 2).
4. Module-path bump `/v33 → /v34` as its own PR at end of cycle.
5. Relayer coordination: counterparty clients updated near the upgrade, none close to expiry.
