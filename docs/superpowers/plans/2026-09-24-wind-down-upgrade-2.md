# Wind-Down Upgrade 2 (v36): Halt, Unbond, Migrate — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the v36 binary that halts every in-scope zone, removes the last operational messages, drops every IBC rate limit, and adds the four admin txs that move the protocol's native tokens from the host ICAs to Osmosis and users' stTokens from Stride to Osmosis.

**Architecture:** One SDK upgrade package `app/upgrades/v36` (handler + helpers + tests, v34/v35 style) whose handler only flips state: halts, oracle deactivation, ICA-host allow-list, rate-limit removal. Four new stakeibc messages, each a thin msg-server delegate to a keeper method in its own `wind_down_*.go` file, reusing `BatchSubmitUndelegateICAMessages`, `SubmitICATxWithoutCallback` and `RecordsKeeper.TransferKeeper.Transfer`. Every address, channel and bound the txs need is a constant in `x/stakeibc/types/wind_down.go`; the txs take only a zone, an ICA type, an amount, or an address list.

**Tech Stack:** Go 1.2x, cosmos-sdk v0.54.3, ibc-go v11.2.0 (rate-limiting middleware, ICA host, transfer), gogoproto via `make proto-gen` (docker), testify suites via `app/apptesting`, Python 3 for the two ops scripts.

Spec: `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` (§3a operator addresses, §6 this upgrade, §7 Osmosis side, §8 windows and checklists, §9 accounting, §10 testing).

## Global Constraints

- Builds on the merged upgrade 1 plan (`docs/superpowers/plans/2026-09-21-wind-down-upgrade-1.md`): the v35 removals are already in the tree, the ICA host allow-list already lacks `MsgLiquidStake`/`MsgRedeemStake`, comdex-1 is `Deprecated`.
- The Go module path stays `github.com/Stride-Labs/stride/v34` for every file this plan touches; the bump is manual, after both plans land. The upgrade package is `app/upgrades/v36`, plan name `"v36"`.
- Removed message types STAY registered in the interface registry (historical tx decoding): remove only the `rpc`, the msg-server handler, the `legacy.RegisterAminoMsg` line and the CLI command, exactly as upgrade 1 did. New submissions are rejected by the router ("can't route message").
- Only `rpc` lines are removed from `.proto` files; new messages are added to `proto/stride/stakeibc/tx.proto` only (all four new txs live in stakeibc). After a `.proto` edit run `make proto-gen` and commit only the `tx.pb.go` of the modules whose proto changed; revert any descriptor-only churn elsewhere with `git checkout <base> -- <files>`.
- Gating, verbatim from spec §3a and §6: `MsgUndelegateFromValidators`, `MsgTransferFromIca` and `MsgTransferStaketiaClaimBalance` check `utils.ValidateAdminAddress(msg.Creator)` in `ValidateBasic`; `MsgSweepTokensOffStride` checks `msg.Creator == types.SweepOperatorAddress`. `MsgUpdateValidatorSharesExchRate` and `MsgCalibrateDelegation` gain `utils.ValidateAdminAddress`.
- Constants, verbatim from spec §3a and §6, all in `x/stakeibc/types/wind_down.go`: `SweepOperatorAddress` and `OsmosisVaultAddress` (empty until the release-gate task fills them; every use fails closed while empty), `OsmosisChainId = "osmosis-1"`, `OsmosisBech32Prefix = "osmo"`, `StrideToOsmosisTransferChannelId = "channel-5"`, `MaxSweepBatchSize = 100`, `WindDownTransferTimeout = 24 * time.Hour`, and `HostToOsmosisTransferChannel` mapping each of the eleven in-scope chain ids to the transfer channel on that host that leads to Osmosis (osmosis-1 maps to `""`, which selects the ICA bank-send form).
- The transfer tx accepts exactly `ica_type ∈ {DELEGATION, WITHDRAWAL, FEE, REDEMPTION}`.
- The sweep takes any valid bank denom and picks its destination from the denom, once per tx (spec §6): a Stride-native denom (no `ibc/` prefix: every stToken, `ustrd`) goes to Osmosis over `StrideToOsmosisTransferChannelId` with the `osmo` prefix; an `ibc/` voucher must have a single-hop trace whose channel is a key of `SweepUnwindChannels` (channel → counterparty bech32 prefix) and goes back over that channel with that prefix. Any other voucher rejects the whole tx. No memo, ever.
- `SweepUnwindChannels`, verbatim from spec §6: channel-0 `cosmos`, channel-162 `celestia`, channel-5 `osmo`, channel-24 `juno`, channel-150 `somm`, channel-213 `saga`, channel-160 `dydx`, plus Stride's channel to noble-1 → `noble` (confirmed from the USDC voucher's denom trace before the PR is cut). Nothing for phoenix-1, laozi-mainnet, injective-1 or haqq_11235-1: their wallets derive different address bytes.
- The sweep's on-chain rule set (spec §6): an address is sweepable iff it decodes to 20 bytes, is not a transfer escrow address, and its account is a `BaseAccount` or one of `ContinuousVestingAccount`, `DelayedVestingAccount`, `PeriodicVestingAccount`, `StridePeriodicVestingAccount`. Anything else, including module accounts and `InterchainAccount`, rejects the whole batch. A zero balance is skipped with an event, not an error.
- Upgrade handler helpers never return an error for a missing-state case; they log and continue (v34 convention). Only `RunMigrations` errors propagate.
- macOS host: use `sed -i ''` (BSD sed). Every commit message ends with the attribution lines from the session's system reminder. Do not push. Branch from `wind-down-design` per task in worktrees as the sub-skill directs.
- Run `go build ./...` before every commit; run the named package tests in each task.

---

## Foundation tasks (serial)

### Task 1: v36 upgrade package skeleton, wiring, and the wind-down constants

**Files:**
- Create: `app/upgrades/v36/constants.go`
- Create: `app/upgrades/v36/upgrades.go`
- Create: `app/upgrades/v36/upgrades_test.go`
- Create: `x/stakeibc/types/wind_down.go`
- Create: `x/stakeibc/types/wind_down_test.go`
- Modify: `app/upgrades.go` (import next to the v35 one; handler registration directly after the v35 `SetUpgradeHandler` block)

**Interfaces:**
- Produces: `v36.UpgradeName = "v36"`; `v36.CreateUpgradeHandler(mm, configurator, stakeibcKeeper stakeibckeeper.Keeper, stakedymKeeper stakedymkeeper.Keeper, icaoracleKeeper icaoraclekeeper.Keeper, icaHostKeeper *icahostkeeper.Keeper, ratelimitKeeper *ratelimitkeeper.Keeper)`; test suite `UpgradeTestSuite` with `s.Setup()` and `s.ConfirmUpgradeSucceeded(v36.UpgradeName)`; every constant and the map named in Global Constraints, plus `types.WindDownAllowedIcaTypes`.
- Review: no

- [ ] **Step 1: Write the failing tests**

`app/upgrades/v36/upgrades_test.go`:

```go
package v36_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v36 "github.com/Stride-Labs/stride/v34/app/upgrades/v36"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// TestUpgradeRuns is the smoke test: the handler is registered under the plan name and completes
// on an empty app. Task 3 adds a focused test per handler step.
func (s *UpgradeTestSuite) TestUpgradeRuns() {
	s.ConfirmUpgradeSucceeded(v36.UpgradeName)
}
```

`x/stakeibc/types/wind_down_test.go`:

```go
package types_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// The eleven in-scope zones of spec §2, and nothing else, must be in the channel map
var inScopeChainIds = []string{
	"celestia", "cosmoshub-4", "dydx-mainnet-1", "haqq_11235-1", "injective-1", "juno-1",
	"laozi-mainnet", "osmosis-1", "phoenix-1", "sommelier-3", "ssc-1",
}

func TestHostToOsmosisTransferChannel(t *testing.T) {
	require.Len(t, types.HostToOsmosisTransferChannel, len(inScopeChainIds))
	channelPattern := regexp.MustCompile(`^channel-[0-9]+$`)
	for _, chainId := range inScopeChainIds {
		channel, found := types.HostToOsmosisTransferChannel[chainId]
		require.True(t, found, "missing channel for %s", chainId)
		if chainId == types.OsmosisChainId {
			require.Equal(t, "", channel, "osmosis-1 must map to the bank-send form")
			continue
		}
		require.Regexp(t, channelPattern, channel, "channel for %s", chainId)
	}
	for _, deprecated := range []string{"comdex-1", "evmos_9001-2", "stargaze-1", "umee-1"} {
		_, found := types.HostToOsmosisTransferChannel[deprecated]
		require.False(t, found, "deprecated zone %s must not be in the map", deprecated)
	}
}

func TestSweepUnwindChannels(t *testing.T) {
	channelPattern := regexp.MustCompile(`^channel-[0-9]+$`)
	for channel, prefix := range types.SweepUnwindChannels {
		require.Regexp(t, channelPattern, channel)
		require.Regexp(t, `^[a-z]+$`, prefix)
	}
	require.Equal(t, "osmo", types.SweepUnwindChannels[types.StrideToOsmosisTransferChannelId], "Osmosis vouchers unwind over the same channel the natives use")
	require.Equal(t, "cosmos", types.SweepUnwindChannels["channel-0"])
	// The four host zones whose wallets derive different address bytes must never be listed
	for _, channel := range []string{"channel-52", "channel-258", "channel-6", "channel-240"} {
		_, found := types.SweepUnwindChannels[channel]
		require.False(t, found, "%s (phoenix-1/laozi-mainnet/injective-1/haqq_11235-1) must not be sweepable", channel)
	}
}

func TestWindDownConstants(t *testing.T) {
	require.Equal(t, "channel-5", types.StrideToOsmosisTransferChannelId)
	require.Equal(t, 100, types.MaxSweepBatchSize)
	require.Equal(t, "osmo", types.OsmosisBech32Prefix)
	require.ElementsMatch(t, []types.ICAAccountType{
		types.ICAAccountType_DELEGATION, types.ICAAccountType_WITHDRAWAL,
		types.ICAAccountType_FEE, types.ICAAccountType_REDEMPTION,
	}, types.WindDownAllowedIcaTypes)
}

// TestWindDownAddressesConfigured is the release gate for the two operator addresses (spec §3a).
// It skips while they are empty and fails if either is set to something that does not parse.
func TestWindDownAddressesConfigured(t *testing.T) {
	if types.SweepOperatorAddress == "" && types.OsmosisVaultAddress == "" {
		t.Skip("wind-down addresses not configured yet; filled in the release-gate task")
	}
	_, err := sdk.AccAddressFromBech32(types.SweepOperatorAddress)
	require.NoError(t, err, "sweep operator must be a stride address")
	_, err = sdk.GetFromBech32(types.OsmosisVaultAddress, types.OsmosisBech32Prefix)
	require.NoError(t, err, "osmosis vault must be an osmo address")
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./app/upgrades/v36/... ./x/stakeibc/types/... -run 'TestUpgradeTestSuite|TestHostToOsmosis|TestWindDown|TestSweepUnwind' 2>&1 | tail -3`
Expected: build failures (`undefined: v36.UpgradeName`, `undefined: types.HostToOsmosisTransferChannel`)

- [ ] **Step 3: Write the constants**

`x/stakeibc/types/wind_down.go`:

```go
package types

import "time"

// Wind-down constants (spec §3a operator addresses, §6 upgrade 2). The two addresses are vars
// so tests can set them; on mainnet they are filled by the release-gate task once the accounts
// exist, and every code path that needs them fails closed while they are empty.
var (
	// SweepOperatorAddress is the only signer of MsgSweepTokensOffStride (spec §3a).
	SweepOperatorAddress = ""
	// OsmosisVaultAddress receives every ICA transfer and funds the pools (spec §3a).
	OsmosisVaultAddress = ""
	// StrideToOsmosisTransferChannelId is Stride's canonical transfer channel to Osmosis
	// (osmosis-1 side: channel-326); the sweep sends over it.
	StrideToOsmosisTransferChannelId = "channel-5"
)

const (
	OsmosisChainId      = "osmosis-1"
	OsmosisBech32Prefix = "osmo"
	// MaxSweepBatchSize bounds MsgSweepTokensOffStride.Addresses so one tx stays inside the block gas limit.
	MaxSweepBatchSize = 100
	// WindDownTransferTimeout is the ICS-20 (and ICA) timeout used by every wind-down transfer.
	WindDownTransferTimeout = 24 * time.Hour

	EventTypeUndelegateFromValidators   = "undelegate_from_validators"
	EventTypeTransferFromIca            = "transfer_from_ica"
	EventTypeTransferStaketiaClaim      = "transfer_staketia_claim_balance"
	EventTypeSweepTokensOffStride              = "sweep_tokens_off_stride"
	AttributeKeyValidator               = "validator"
	AttributeKeyAmount                  = "amount"
	AttributeKeyIcaType                 = "ica_type"
	AttributeKeyChannel                 = "channel"
	AttributeKeyReceiver                = "receiver"
	AttributeKeyHolder                  = "holder"
	AttributeKeySequence                = "sequence"
	AttributeKeySkipped                 = "skipped"
)

// SweepUnwindChannels maps a Stride transfer channel to the bech32 prefix of its counterparty
// chain, for the chains whose wallets derive the same address bytes as Stride (secp256k1, coin
// type 118). MsgSweepTokensOffStride sends a single-hop voucher back over its channel only when
// the channel is here; on any other chain the address with the same bytes is not the holder's.
// Ops confirm the noble-1 channel from the USDC voucher's denom trace before the upgrade PR.
var SweepUnwindChannels = map[string]string{
	"channel-0":   "cosmos",   // cosmoshub-4
	"channel-162": "celestia", // celestia
	"channel-5":   "osmo",     // osmosis-1
	"channel-24":  "juno",     // juno-1
	"channel-150": "somm",     // sommelier-3
	"channel-213": "saga",     // ssc-1
	"channel-160": "dydx",     // dydx-mainnet-1
}

// WindDownAllowedIcaTypes are the four ICAs MsgTransferFromIca may drain (spec §6).
var WindDownAllowedIcaTypes = []ICAAccountType{
	ICAAccountType_DELEGATION,
	ICAAccountType_WITHDRAWAL,
	ICAAccountType_FEE,
	ICAAccountType_REDEMPTION,
}

// HostToOsmosisTransferChannel maps each in-scope zone to the ICS-20 channel ON THAT HOST that
// leads to osmosis-1, i.e. the channel that mints the canonical denom on Osmosis. Values are the
// chain-registry preferred channels on 2026-09-24 and are re-verified against each host before
// the upgrade proposal (spec §8, window 1 step 5). osmosis-1 maps to "" because its ICAs are
// already on Osmosis and the transfer tx uses a bank send there.
var HostToOsmosisTransferChannel = map[string]string{
	"celestia":       "channel-2",
	"cosmoshub-4":    "channel-141",
	"dydx-mainnet-1": "channel-3",
	"haqq_11235-1":   "channel-2",
	"injective-1":    "channel-8",
	"juno-1":         "channel-0",
	"laozi-mainnet":  "channel-83",
	OsmosisChainId:   "",
	"phoenix-1":      "channel-1",
	"sommelier-3":    "channel-0",
	"ssc-1":          "channel-1",
}
```

- [ ] **Step 4: Create the upgrade package and wire it**

`app/upgrades/v36/constants.go`:

```go
package v36

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v36"
)
```

`app/upgrades/v36/upgrades.go`:

```go
package v36

import (
	"context"
	"fmt"

	icahostkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/keeper"
	ratelimitkeeper "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icaoraclekeeper "github.com/Stride-Labs/stride/v34/x/icaoracle/keeper"
	stakedymkeeper "github.com/Stride-Labs/stride/v34/x/stakedym/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

// CreateUpgradeHandler returns the v36 upgrade handler: the second wind-down upgrade
// (docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md §6). The message removals and
// the four admin txs live in the binary itself; the handler halts every in-scope zone and
// stakedym, deactivates the ICA oracles, drops the claim message from the ICA host allow-list,
// and removes every rate limit. Every step logs and skips on missing state rather than erroring.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	stakedymKeeper stakedymkeeper.Keeper,
	icaoracleKeeper icaoraclekeeper.Keeper,
	icaHostKeeper *icahostkeeper.Keeper,
	ratelimitKeeper *ratelimitkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s (wind-down 2: halt, unbond, migrate)...", UpgradeName))

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete", UpgradeName))
		return vm, nil
	}
}
```

In `app/upgrades.go` add the import next to the v35 one:

```go
	v36 "github.com/Stride-Labs/stride/v34/app/upgrades/v36"
```

and directly after the v35 `SetUpgradeHandler(...)` call:

```go
	// v36 upgrade handler
	app.UpgradeKeeper.SetUpgradeHandler(
		v36.UpgradeName,
		v36.CreateUpgradeHandler(
			app.ModuleManager,
			app.configurator,
			app.StakeibcKeeper,
			app.StakedymKeeper,
			app.ICAOracleKeeper,
			app.ICAHostKeeper,
			&app.RatelimitKeeper,
		),
	)
```

(`app.RatelimitKeeper` is a value whose methods have pointer receivers, hence `&`.) The unused keeper parameters are used by Task 3.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go build ./... && go test ./app/upgrades/v36/... ./x/stakeibc/types/... -run 'TestUpgradeTestSuite|TestHostToOsmosis|TestWindDown|TestSweepUnwind' 2>&1 | tail -3`
Expected: `ok` for both packages (`TestWindDownAddressesConfigured` reports SKIP)

- [ ] **Step 6: Commit**

```bash
git add app/upgrades/v36 app/upgrades.go x/stakeibc/types/wind_down.go x/stakeibc/types/wind_down_test.go
git commit -m "feat(upgrade): v36 handler skeleton, wiring and wind-down constants"
```

### Task 2: The four messages — proto, types, codec, keeper stubs, msg-server delegates, CLI

**Files:**
- Modify: `proto/stride/stakeibc/tx.proto` (four `rpc` lines in `service Msg`; the message definitions appended at the end; add `import "stride/stakeibc/ica_account.proto";` if `ICAAccountType` is not already imported)
- Create: `x/stakeibc/types/message_undelegate_from_validators.go` (+`_test.go`), `message_transfer_from_ica.go` (+`_test.go`), `message_transfer_staketia_claim_balance.go` (+`_test.go`), `message_sweep_tokens_off_stride.go` (+`_test.go`)
- Modify: `x/stakeibc/types/codec.go` (four `legacy.RegisterAminoMsg` lines in `RegisterCodec`; four entries in the `RegisterImplementations` list in `RegisterInterfaces`)
- Create: `x/stakeibc/keeper/wind_down_undelegate.go`, `wind_down_transfer_from_ica.go`, `wind_down_claim_balance.go`, `wind_down_sweep.go` (keeper stubs with the final signatures; Tasks 5-8 replace their bodies)
- Create: `x/stakeibc/keeper/msg_server_wind_down.go` (the four handlers, thin delegates)
- Create: `x/stakeibc/client/cli/tx_wind_down.go`; Modify: `x/stakeibc/client/cli/tx.go` (`GetTxCmd` gains four `cmd.AddCommand(...)` lines)

**Interfaces:**
- Produces (keeper signatures Tasks 5-8 implement):
  - `func (k Keeper) UndelegateFromValidators(ctx sdk.Context, msg *types.MsgUndelegateFromValidators) (numTxsSubmitted uint64, err error)`
  - `func (k Keeper) TransferFromIca(ctx sdk.Context, msg *types.MsgTransferFromIca) error`
  - `func (k Keeper) TransferStaketiaClaimBalance(ctx sdk.Context) (sequence uint64, amount sdk.Coin, err error)`
  - `func (k Keeper) SweepTokensOffStride(ctx sdk.Context, msg *types.MsgSweepTokensOffStride) (numSwept uint64, numSkipped uint64, err error)`
- Produces: `types.NewMsgUndelegateFromValidators`, `types.NewMsgTransferFromIca`, `types.NewMsgTransferStaketiaClaimBalance`, `types.NewMsgSweepTokensOffStride` and their `ValidateBasic`; `types.ValidatorUndelegation{Address, Offset}`.
- Depends on: Task 1
- Review: yes (the gates are the security boundary of every later task)

- [ ] **Step 1: Write the failing ValidateBasic tests**

`x/stakeibc/types/message_undelegate_from_validators_test.go`:

```go
package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgUndelegateFromValidators_ValidateBasic(t *testing.T) {
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validators := []types.ValidatorUndelegation{
		{Address: "cosmosvaloper1abc", Offset: sdkmath.ZeroInt()},
		{Address: "cosmosvaloper1def", Offset: sdkmath.NewInt(5)},
	}

	tests := []struct {
		name string
		msg  types.MsgUndelegateFromValidators
		err  string
	}{
		{name: "valid, explicit validators", msg: types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: validators}},
		{name: "valid, empty validators means all", msg: types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4"}},
		{name: "invalid creator", msg: types.MsgUndelegateFromValidators{Creator: invalidAddress, ChainId: "cosmoshub-4"}, err: "invalid creator address"},
		{name: "not admin", msg: types.MsgUndelegateFromValidators{Creator: validNotAdminAddress, ChainId: "cosmoshub-4"}, err: "is not an admin"},
		{name: "missing chain id", msg: types.MsgUndelegateFromValidators{Creator: validAdminAddress}, err: "chain id"},
		{name: "empty validator address", msg: types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4",
			Validators: []types.ValidatorUndelegation{{Address: "", Offset: sdkmath.ZeroInt()}}}, err: "validator address"},
		{name: "duplicate validator", msg: types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4",
			Validators: []types.ValidatorUndelegation{validators[0], validators[0]}}, err: "duplicate"},
		{name: "negative offset", msg: types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4",
			Validators: []types.ValidatorUndelegation{{Address: "cosmosvaloper1abc", Offset: sdkmath.NewInt(-1)}}}, err: "offset"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.err == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.err)
		})
	}
	_ = sdkerrors.ErrInvalidAddress
}
```

`x/stakeibc/types/message_transfer_from_ica_test.go`:

```go
package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgTransferFromIca_ValidateBasic(t *testing.T) {
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)
	amount := sdk.NewCoin("uatom", sdkmath.NewInt(1000))

	tests := []struct {
		name string
		msg  types.MsgTransferFromIca
		err  string
	}{
		{name: "valid delegation", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}},
		{name: "valid withdrawal foreign denom", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "dydx-mainnet-1", IcaType: types.ICAAccountType_WITHDRAWAL, Amount: sdk.NewCoin("ibc/8E27", sdkmath.NewInt(1))}},
		{name: "valid osmosis bank send form", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "osmosis-1", IcaType: types.ICAAccountType_FEE, Amount: sdk.NewCoin("uosmo", sdkmath.NewInt(1))}},
		{name: "invalid creator", msg: types.MsgTransferFromIca{Creator: invalidAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}, err: "invalid creator address"},
		{name: "not admin", msg: types.MsgTransferFromIca{Creator: validNotAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}, err: "is not an admin"},
		{name: "chain not in map", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "evmos_9001-2", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}, err: "not an in-scope wind-down zone"},
		{name: "community pool ica rejected", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, Amount: amount}, err: "ica type"},
		{name: "converter ica rejected", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_CONVERTER_TRADE, Amount: amount}, err: "ica type"},
		{name: "zero amount", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: sdk.NewCoin("uatom", sdkmath.ZeroInt())}, err: "amount"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.err == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.err)
		})
	}
}
```

`x/stakeibc/types/message_transfer_staketia_claim_balance_test.go`:

```go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgTransferStaketiaClaimBalance_ValidateBasic(t *testing.T) {
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	require.NoError(t, types.NewMsgTransferStaketiaClaimBalance(validAdminAddress).ValidateBasic())
	require.ErrorContains(t, types.NewMsgTransferStaketiaClaimBalance(invalidAddress).ValidateBasic(), "invalid creator address")
	require.ErrorContains(t, types.NewMsgTransferStaketiaClaimBalance(validNotAdminAddress).ValidateBasic(), "is not an admin")
}
```

`x/stakeibc/types/message_sweep_tokens_off_stride_test.go`:

```go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgSweepTokensOffStride_ValidateBasic(t *testing.T) {
	operator, other := apptesting.GenerateTestAddrs()
	_, invalidAddress := apptesting.GenerateTestAddrs()
	holder1, holder2 := apptesting.GenerateTestAddrs()

	// The gate is a package var; set it for this test and restore it after
	previous := types.SweepOperatorAddress
	types.SweepOperatorAddress = operator
	defer func() { types.SweepOperatorAddress = previous }()

	tooMany := make([]string, types.MaxSweepBatchSize+1)
	for i := range tooMany {
		tooMany[i] = holder1 // duplicates are caught after the bound; the bound fires first
	}

	tests := []struct {
		name string
		msg  types.MsgSweepTokensOffStride
		err  string
	}{
		{name: "valid", msg: types.MsgSweepTokensOffStride{Creator: operator, Denom: "stuatom", Addresses: []string{holder1, holder2}}},
		{name: "not the sweep operator", msg: types.MsgSweepTokensOffStride{Creator: other, Denom: "stuatom", Addresses: []string{holder1}}, err: "sweep operator"},
		{name: "invalid creator", msg: types.MsgSweepTokensOffStride{Creator: invalidAddress, Denom: "stuatom", Addresses: []string{holder1}}, err: "invalid creator address"},
		{name: "valid strd", msg: types.MsgSweepTokensOffStride{Creator: operator, Denom: "ustrd", Addresses: []string{holder1}}},
		{name: "valid ibc voucher", msg: types.MsgSweepTokensOffStride{Creator: operator, Denom: "ibc/065DE8EC5E26C3798B1292446DBB5DAB9A490EEC9986ED9955328C38", Addresses: []string{holder1}}},
		{name: "invalid denom string", msg: types.MsgSweepTokensOffStride{Creator: operator, Denom: "not a denom!", Addresses: []string{holder1}}, err: "invalid denom"},
		{name: "no addresses", msg: types.MsgSweepTokensOffStride{Creator: operator, Denom: "stuatom"}, err: "at least one"},
		{name: "over the batch bound", msg: types.MsgSweepTokensOffStride{Creator: operator, Denom: "stuatom", Addresses: tooMany}, err: "at most"},
		{name: "invalid holder address", msg: types.MsgSweepTokensOffStride{Creator: operator, Denom: "stuatom", Addresses: []string{invalidAddress}}, err: "invalid address"},
		{name: "duplicate holder", msg: types.MsgSweepTokensOffStride{Creator: operator, Denom: "stuatom", Addresses: []string{holder1, holder1}}, err: "duplicate"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.err == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.err)
		})
	}
}

func TestMsgSweepTokensOffStride_UnconfiguredOperatorRejectsEveryone(t *testing.T) {
	operator, _ := apptesting.GenerateTestAddrs()
	previous := types.SweepOperatorAddress
	types.SweepOperatorAddress = ""
	defer func() { types.SweepOperatorAddress = previous }()

	err := (&types.MsgSweepTokensOffStride{Creator: operator, Denom: "stuatom", Addresses: []string{operator}}).ValidateBasic()
	require.ErrorContains(t, err, "sweep operator")
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./x/stakeibc/types/... -run 'TestMsgUndelegateFromValidators|TestMsgTransferFromIca|TestMsgTransferStaketia|TestMsgSweepTokensOffStride' 2>&1 | tail -3`
Expected: build failure (`undefined: types.MsgUndelegateFromValidators` and the others)

- [ ] **Step 3: Add the messages to the proto and regenerate**

In `proto/stride/stakeibc/tx.proto`, inside `service Msg` (after the last existing rpc):

```proto
  // Wind-down admin txs (spec §6). Added in v36.
  rpc UndelegateFromValidators(MsgUndelegateFromValidators)
      returns (MsgUndelegateFromValidatorsResponse);
  rpc TransferFromIca(MsgTransferFromIca) returns (MsgTransferFromIcaResponse);
  rpc TransferStaketiaClaimBalance(MsgTransferStaketiaClaimBalance)
      returns (MsgTransferStaketiaClaimBalanceResponse);
  rpc SweepTokensOffStride(MsgSweepTokensOffStride) returns (MsgSweepTokensOffStrideResponse);
```

At the end of the file (and add `import "stride/stakeibc/ica_account.proto";` next to the `validator.proto` import if it is not already there):

```proto
// One validator to undelegate from during the wind-down: the full stored delegation minus offset
message ValidatorUndelegation {
  string address = 1;
  string offset = 2 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "cosmossdk.io/math.Int",
    (gogoproto.nullable) = false
  ];
}

// Submits an undelegate ICA for every listed validator (or every validator with a delegation
// when the list is empty) on a halted host zone, with no epoch unbonding records attached
message MsgUndelegateFromValidators {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name) = "stakeibc/MsgUndelegateFromValidators";

  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  string chain_id = 2;
  repeated ValidatorUndelegation validators = 3 [ (gogoproto.nullable) = false ];
}
message MsgUndelegateFromValidatorsResponse { uint64 num_txs_submitted = 1; }

// Sends `amount` from one of a host zone's ICAs to the Osmosis vault over the host's channel to
// Osmosis (a bank send when the host is Osmosis itself). Receiver and channel are constants.
message MsgTransferFromIca {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name) = "stakeibc/MsgTransferFromIca";

  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  string chain_id = 2;
  ICAAccountType ica_type = 3;
  cosmos.base.v1beta1.Coin amount = 4 [ (gogoproto.nullable) = false ];
}
message MsgTransferFromIcaResponse {}

// Moves the staketia claim address's whole TIA voucher balance to the celestia zone's delegation
// ICA, so it leaves for Osmosis with the zone's balance. Everything is a constant or state.
message MsgTransferStaketiaClaimBalance {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name) = "stakeibc/MsgTransferStaketiaClaimBalance";

  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
}
message MsgTransferStaketiaClaimBalanceResponse {
  uint64 sequence = 1;
  cosmos.base.v1beta1.Coin amount = 2 [ (gogoproto.nullable) = false ];
}

// Sends each listed holder's full balance of one denom off Stride to the same address bytes
// on the destination chain: Osmosis for a Stride-native denom (stTokens, ustrd), the source
// chain for a single-hop voucher on a whitelisted channel. Signed only by the sweep operator;
// the batch is rejected if the denom has no destination or any address is not sweepable.
message MsgSweepTokensOffStride {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name) = "stakeibc/MsgSweepTokensOffStride";

  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  string denom = 2;
  repeated string addresses = 3;
}
message MsgSweepTokensOffStrideResponse {
  uint64 num_swept = 1;
  uint64 num_skipped = 2;
}
```

Run: `make proto-gen`
Expected: `x/stakeibc/types/tx.pb.go` regenerated; `grep -c "func (m \*MsgSweepTokensOffStride)" x/stakeibc/types/tx.pb.go` is non-zero. Revert descriptor churn in any other `*.pb.go` (Global Constraints).

- [ ] **Step 4: Write the message types**

`x/stakeibc/types/message_undelegate_from_validators.go`:

```go
package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
)

const TypeMsgUndelegateFromValidators = "undelegate_from_validators"

var _ sdk.Msg = &MsgUndelegateFromValidators{}

func NewMsgUndelegateFromValidators(creator, chainId string, validators []ValidatorUndelegation) *MsgUndelegateFromValidators {
	return &MsgUndelegateFromValidators{Creator: creator, ChainId: chainId, Validators: validators}
}

func (msg *MsgUndelegateFromValidators) Route() string { return RouterKey }
func (msg *MsgUndelegateFromValidators) Type() string  { return TypeMsgUndelegateFromValidators }

func (msg *MsgUndelegateFromValidators) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUndelegateFromValidators) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if msg.ChainId == "" {
		return errorsmod.Wrap(ErrRequiredFieldEmpty, "chain id is required")
	}
	seen := map[string]bool{}
	for _, validator := range msg.Validators {
		if validator.Address == "" {
			return errorsmod.Wrap(ErrRequiredFieldEmpty, "validator address is required")
		}
		if seen[validator.Address] {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate validator %s", validator.Address)
		}
		seen[validator.Address] = true
		if !validator.Offset.IsNil() && validator.Offset.IsNegative() {
			return errorsmod.Wrapf(ErrInvalidAmount, "offset for %s must not be negative", validator.Address)
		}
	}
	return nil
}

// OffsetOrZero returns the offset, treating an unset (nil) value as zero
func (v ValidatorUndelegation) OffsetOrZero() sdkmath.Int {
	if v.Offset.IsNil() {
		return sdkmath.ZeroInt()
	}
	return v.Offset
}
```

`x/stakeibc/types/message_transfer_from_ica.go`:

```go
package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
)

const TypeMsgTransferFromIca = "transfer_from_ica"

var _ sdk.Msg = &MsgTransferFromIca{}

func NewMsgTransferFromIca(creator, chainId string, icaType ICAAccountType, amount sdk.Coin) *MsgTransferFromIca {
	return &MsgTransferFromIca{Creator: creator, ChainId: chainId, IcaType: icaType, Amount: amount}
}

func (msg *MsgTransferFromIca) Route() string { return RouterKey }
func (msg *MsgTransferFromIca) Type() string  { return TypeMsgTransferFromIca }

func (msg *MsgTransferFromIca) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgTransferFromIca) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if _, found := HostToOsmosisTransferChannel[msg.ChainId]; !found {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s is not an in-scope wind-down zone", msg.ChainId)
	}
	if !IsWindDownAllowedIcaType(msg.IcaType) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ica type %s cannot be drained; allowed: %v", msg.IcaType, WindDownAllowedIcaTypes)
	}
	if err := msg.Amount.Validate(); err != nil || !msg.Amount.IsPositive() {
		return errorsmod.Wrap(ErrInvalidAmount, "amount must be a positive coin")
	}
	return nil
}

// IsWindDownAllowedIcaType reports whether MsgTransferFromIca may drain this ICA
func IsWindDownAllowedIcaType(icaType ICAAccountType) bool {
	for _, allowed := range WindDownAllowedIcaTypes {
		if icaType == allowed {
			return true
		}
	}
	return false
}
```

`x/stakeibc/types/message_transfer_staketia_claim_balance.go`:

```go
package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
)

const TypeMsgTransferStaketiaClaimBalance = "transfer_staketia_claim_balance"

var _ sdk.Msg = &MsgTransferStaketiaClaimBalance{}

func NewMsgTransferStaketiaClaimBalance(creator string) *MsgTransferStaketiaClaimBalance {
	return &MsgTransferStaketiaClaimBalance{Creator: creator}
}

func (msg *MsgTransferStaketiaClaimBalance) Route() string { return RouterKey }
func (msg *MsgTransferStaketiaClaimBalance) Type() string  { return TypeMsgTransferStaketiaClaimBalance }

func (msg *MsgTransferStaketiaClaimBalance) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgTransferStaketiaClaimBalance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return utils.ValidateAdminAddress(msg.Creator)
}
```

`x/stakeibc/types/message_sweep_tokens_off_stride.go`:

```go
package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSweepTokensOffStride = "sweep_tokens_off_stride"

var _ sdk.Msg = &MsgSweepTokensOffStride{}

func NewMsgSweepTokensOffStride(creator, denom string, addresses []string) *MsgSweepTokensOffStride {
	return &MsgSweepTokensOffStride{Creator: creator, Denom: denom, Addresses: addresses}
}

func (msg *MsgSweepTokensOffStride) Route() string { return RouterKey }
func (msg *MsgSweepTokensOffStride) Type() string  { return TypeMsgSweepTokensOffStride }

func (msg *MsgSweepTokensOffStride) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic gates on the sweep operator (spec §3a), not the admin set. An empty
// SweepOperatorAddress rejects every signer, which is the fail-closed default. The denom is
// any valid bank denom; the keeper decides where it goes (Osmosis or the voucher's source).
func (msg *MsgSweepTokensOffStride) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if SweepOperatorAddress == "" || msg.Creator != SweepOperatorAddress {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the sweep operator", msg.Creator)
	}
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid denom %s: %s", msg.Denom, err.Error())
	}
	if len(msg.Addresses) == 0 {
		return errorsmod.Wrap(ErrRequiredFieldEmpty, "at least one address is required")
	}
	if len(msg.Addresses) > MaxSweepBatchSize {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "at most %d addresses per sweep, got %d", MaxSweepBatchSize, len(msg.Addresses))
	}
	seen := map[string]bool{}
	for _, address := range msg.Addresses {
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address %s (%s)", address, err)
		}
		if seen[address] {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate address %s", address)
		}
		seen[address] = true
	}
	return nil
}
```

`x/stakeibc/types/codec.go`: in `RegisterCodec` add

```go
	legacy.RegisterAminoMsg(cdc, &MsgUndelegateFromValidators{}, "stakeibc/MsgUndelegateFromValidators")
	legacy.RegisterAminoMsg(cdc, &MsgTransferFromIca{}, "stakeibc/MsgTransferFromIca")
	legacy.RegisterAminoMsg(cdc, &MsgTransferStaketiaClaimBalance{}, "stakeibc/MsgTransferStaketiaClaimBalance")
	legacy.RegisterAminoMsg(cdc, &MsgSweepTokensOffStride{}, "stakeibc/MsgSweepTokensOffStride")
```

and in `RegisterInterfaces` append `&MsgUndelegateFromValidators{}, &MsgTransferFromIca{}, &MsgTransferStaketiaClaimBalance{}, &MsgSweepTokensOffStride{}` to the `RegisterImplementations((*sdk.Msg)(nil), ...)` list.

- [ ] **Step 5: Keeper stubs and msg-server delegates**

Each stub carries the final signature so Tasks 5-8 only replace bodies. `x/stakeibc/keeper/wind_down_undelegate.go`:

```go
package keeper

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// UndelegateFromValidators is implemented in Task 5 of the upgrade 2 plan
func (k Keeper) UndelegateFromValidators(ctx sdk.Context, msg *types.MsgUndelegateFromValidators) (numTxsSubmitted uint64, err error) {
	return 0, errorsmod.Wrap(sdkerrors.ErrNotSupported, "UndelegateFromValidators not implemented")
}
```

`x/stakeibc/keeper/wind_down_transfer_from_ica.go`:

```go
package keeper

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// TransferFromIca is implemented in Task 6 of the upgrade 2 plan
func (k Keeper) TransferFromIca(ctx sdk.Context, msg *types.MsgTransferFromIca) error {
	return errorsmod.Wrap(sdkerrors.ErrNotSupported, "TransferFromIca not implemented")
}
```

`x/stakeibc/keeper/wind_down_claim_balance.go`:

```go
package keeper

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TransferStaketiaClaimBalance is implemented in Task 7 of the upgrade 2 plan
func (k Keeper) TransferStaketiaClaimBalance(ctx sdk.Context) (sequence uint64, amount sdk.Coin, err error) {
	return 0, sdk.Coin{}, errorsmod.Wrap(sdkerrors.ErrNotSupported, "TransferStaketiaClaimBalance not implemented")
}
```

`x/stakeibc/keeper/wind_down_sweep.go`:

```go
package keeper

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// SweepTokensOffStride is implemented in Task 8 of the upgrade 2 plan
func (k Keeper) SweepTokensOffStride(ctx sdk.Context, msg *types.MsgSweepTokensOffStride) (numSwept uint64, numSkipped uint64, err error) {
	return 0, 0, errorsmod.Wrap(sdkerrors.ErrNotSupported, "SweepTokensOffStride not implemented")
}
```

`x/stakeibc/keeper/msg_server_wind_down.go`:

```go
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// The four wind-down admin txs (spec §6). Each handler is a thin delegate; the logic and its
// tests live in the wind_down_*.go keeper files.

func (k msgServer) UndelegateFromValidators(goCtx context.Context, msg *types.MsgUndelegateFromValidators) (*types.MsgUndelegateFromValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	numTxsSubmitted, err := k.Keeper.UndelegateFromValidators(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUndelegateFromValidatorsResponse{NumTxsSubmitted: numTxsSubmitted}, nil
}

func (k msgServer) TransferFromIca(goCtx context.Context, msg *types.MsgTransferFromIca) (*types.MsgTransferFromIcaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.Keeper.TransferFromIca(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgTransferFromIcaResponse{}, nil
}

func (k msgServer) TransferStaketiaClaimBalance(goCtx context.Context, msg *types.MsgTransferStaketiaClaimBalance) (*types.MsgTransferStaketiaClaimBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sequence, amount, err := k.Keeper.TransferStaketiaClaimBalance(ctx)
	if err != nil {
		return nil, err
	}
	return &types.MsgTransferStaketiaClaimBalanceResponse{Sequence: sequence, Amount: amount}, nil
}

func (k msgServer) SweepTokensOffStride(goCtx context.Context, msg *types.MsgSweepTokensOffStride) (*types.MsgSweepTokensOffStrideResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	numSwept, numSkipped, err := k.Keeper.SweepTokensOffStride(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgSweepTokensOffStrideResponse{NumSwept: numSwept, NumSkipped: numSkipped}, nil
}
```

(`msgServer` embeds `Keeper` in `msg_server.go`; if the field is named differently, use that name.)

- [ ] **Step 6: CLI**

`x/stakeibc/client/cli/tx_wind_down.go`:

```go
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// undelegate-from-validators [chain-id] [validator[:offset],...]
// The second argument is optional; without it every validator with a delegation is drained.
func CmdUndelegateFromValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate-from-validators [chain-id] [validator[:offset],...]",
		Short: "Wind-down: undelegate the full stored delegation (minus an optional per-validator offset) from the listed validators, or from all when none are listed",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			validators := []types.ValidatorUndelegation{}
			if len(args) == 2 {
				for _, entry := range strings.Split(args[1], ",") {
					parts := strings.SplitN(strings.TrimSpace(entry), ":", 2)
					offset := sdkmath.ZeroInt()
					if len(parts) == 2 {
						parsed, ok := sdkmath.NewIntFromString(parts[1])
						if !ok {
							return fmt.Errorf("invalid offset %q for %s", parts[1], parts[0])
						}
						offset = parsed
					}
					validators = append(validators, types.ValidatorUndelegation{Address: parts[0], Offset: offset})
				}
			}
			msg := types.NewMsgUndelegateFromValidators(clientCtx.GetFromAddress().String(), args[0], validators)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// transfer-from-ica [chain-id] [DELEGATION|WITHDRAWAL|FEE|REDEMPTION] [amount]
func CmdTransferFromIca() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-from-ica [chain-id] [ica-type] [amount]",
		Short: "Wind-down: send amount (denom as on the host) from the zone's ICA to the Osmosis vault",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			icaTypeValue, found := types.ICAAccountType_value[strings.ToUpper(args[1])]
			if !found {
				return fmt.Errorf("unknown ica type %s", args[1])
			}
			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}
			msg := types.NewMsgTransferFromIca(clientCtx.GetFromAddress().String(), args[0], types.ICAAccountType(icaTypeValue), amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdTransferStaketiaClaimBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-staketia-claim-balance",
		Short: "Wind-down: move the staketia claim address's whole TIA balance to the celestia delegation ICA",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgTransferStaketiaClaimBalance(clientCtx.GetFromAddress().String())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// sweep-tokens-off-stride [denom] [addresses-file]: one stride address per line, blank lines ignored.
// The file is a batch written by scripts/wind-down/build_sweep_batches.py.
func CmdSweepTokensOffStride() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sweep-tokens-off-stride [denom] [addresses-file]",
		Short: "Wind-down: send each listed holder's full stToken balance to the same address on Osmosis",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			addresses, err := readAddressesFile(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgSweepTokensOffStride(clientCtx.GetFromAddress().String(), args[0], addresses)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func readAddressesFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	addresses := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			addresses = append(addresses, line)
		}
	}
	return addresses, scanner.Err()
}
```

In `x/stakeibc/client/cli/tx.go` `GetTxCmd`, add before the closing `return cmd`:

```go
	cmd.AddCommand(CmdUndelegateFromValidators())
	cmd.AddCommand(CmdTransferFromIca())
	cmd.AddCommand(CmdTransferStaketiaClaimBalance())
	cmd.AddCommand(CmdSweepTokensOffStride())
```

- [ ] **Step 7: Build and test**

Run: `go build ./... && go test ./x/stakeibc/types/... 2>&1 | tail -3`
Expected: `ok`

Run: `go test ./x/stakeibc/keeper/... 2>&1 | tail -3`
Expected: `ok` (the stubs are not exercised yet)

- [ ] **Step 8: Commit**

```bash
git add proto/stride/stakeibc/tx.proto x/stakeibc/types x/stakeibc/keeper/wind_down_*.go x/stakeibc/keeper/msg_server_wind_down.go x/stakeibc/client/cli
git commit -m "feat(stakeibc): wind-down admin message types, gates, stubs and CLI"
```

## Parallel-safe tasks

Every task below depends only on Tasks 1-2. Task 3 edits `app/upgrades/v36` only. Task 4 is the only task that edits `.proto` files after the foundation (three rpc removals) and the only one that touches `msg_server.go`, `claim.go`, the ICQ message types and the calibration callback. Tasks 5-8 each own exactly one `wind_down_*.go` keeper file and its test file. Task 9 is Python only.

### Task 3: Handler steps — halts, oracle deactivation, ICA host allow-list, rate-limit removal

**Files:**
- Create: `app/upgrades/v36/wind_down.go`
- Create: `app/upgrades/v36/wind_down_test.go`
- Modify: `app/upgrades/v36/upgrades.go` (four calls between the two `Logger` lines, after `RunMigrations`)

**Interfaces:**
- Produces: `HaltInScopeHostZones(ctx, k stakeibckeeper.Keeper)`, `HaltStakedym(ctx, k stakedymkeeper.Keeper)`, `DeactivateIcaOracles(ctx, k icaoraclekeeper.Keeper)`, `RemoveIcaHostClaimMessage(ctx, k *icahostkeeper.Keeper)`, `RemoveAllRateLimits(ctx, k *ratelimitkeeper.Keeper)`.
- Depends on: Task 1
- Review: yes (the halt is what freezes the redemption rate; a zone left un-halted keeps reinvesting)

- [ ] **Step 1: Write the failing tests**

`app/upgrades/v36/wind_down_test.go`:

```go
package v36_test

import (
	sdkmath "cosmossdk.io/math"
	ratelimittypes "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v36 "github.com/Stride-Labs/stride/v34/app/upgrades/v36"
	icaoracletypes "github.com/Stride-Labs/stride/v34/x/icaoracle/types"
	stakedymtypes "github.com/Stride-Labs/stride/v34/x/stakedym/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func (s *UpgradeTestSuite) TestHaltInScopeHostZones() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: "cosmoshub-4"})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: "celestia", Halted: false})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: "comdex-1", Deprecated: true, Halted: false})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: "evmos_9001-2", Deprecated: true, Halted: true})

	v36.HaltInScopeHostZones(s.Ctx, s.App.StakeibcKeeper)

	for _, chainId := range []string{"cosmoshub-4", "celestia"} {
		hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
		s.Require().True(hostZone.Halted, "%s must be halted", chainId)
	}
	comdex, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "comdex-1")
	s.Require().False(comdex.Halted, "deprecated zones are not touched (spec §2)")
	evmos, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "evmos_9001-2")
	s.Require().True(evmos.Halted, "already-halted deprecated zone stays halted")
}

func (s *UpgradeTestSuite) TestHaltStakedym() {
	s.App.StakedymKeeper.SetHostZone(s.Ctx, stakedymtypes.HostZone{ChainId: "dymension_1100-1", Halted: false})
	v36.HaltStakedym(s.Ctx, s.App.StakedymKeeper)
	hostZone, err := s.App.StakedymKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err)
	s.Require().True(hostZone.Halted)
}

func (s *UpgradeTestSuite) TestHaltStakedym_NoHostZone() {
	s.Require().NotPanics(func() { v36.HaltStakedym(s.Ctx, s.App.StakedymKeeper) })
}

func (s *UpgradeTestSuite) TestDeactivateIcaOracles() {
	for _, chainId := range []string{"injective-1", "neutron-1", "osmosis-1"} {
		s.App.ICAOracleKeeper.SetOracle(s.Ctx, icaoracletypes.Oracle{ChainId: chainId, Active: true})
	}
	v36.DeactivateIcaOracles(s.Ctx, s.App.ICAOracleKeeper)
	for _, oracle := range s.App.ICAOracleKeeper.GetAllOracles(s.Ctx) {
		s.Require().False(oracle.Active, "%s must be inactive", oracle.ChainId)
	}
}

func (s *UpgradeTestSuite) TestRemoveIcaHostClaimMessage() {
	params := s.App.ICAHostKeeper.GetParams(s.Ctx)
	params.AllowMessages = []string{
		"/cosmos.bank.v1beta1.MsgSend",
		sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
		"/ibc.applications.transfer.v1.MsgTransfer",
	}
	s.App.ICAHostKeeper.SetParams(s.Ctx, params)

	v36.RemoveIcaHostClaimMessage(s.Ctx, s.App.ICAHostKeeper)

	after := s.App.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().Equal([]string{"/cosmos.bank.v1beta1.MsgSend", "/ibc.applications.transfer.v1.MsgTransfer"}, after.AllowMessages)
	s.Require().True(after.HostEnabled)
}

func (s *UpgradeTestSuite) TestRemoveAllRateLimits() {
	k := &s.App.RatelimitKeeper
	k.SetRateLimit(s.Ctx, ratelimittypes.RateLimit{
		Path:  &ratelimittypes.Path{Denom: "stuatom", ChannelOrClientId: "channel-5"},
		Quota: &ratelimittypes.Quota{MaxPercentSend: sdkmath.NewInt(10), MaxPercentRecv: sdkmath.NewInt(10), DurationHours: 24},
		Flow:  &ratelimittypes.Flow{Inflow: sdkmath.ZeroInt(), Outflow: sdkmath.ZeroInt(), ChannelValue: sdkmath.NewInt(1000)},
	})
	k.AddDenomToBlacklist(s.Ctx, "stuosmo")
	k.SetWhitelistedAddressPair(s.Ctx, ratelimittypes.WhitelistedAddressPair{Sender: "stride1a", Receiver: "cosmos1b"})

	v36.RemoveAllRateLimits(s.Ctx, k)

	s.Require().Empty(k.GetAllRateLimits(s.Ctx))
	s.Require().Empty(k.GetAllBlacklistedDenoms(s.Ctx))
	s.Require().Empty(k.GetAllWhitelistedAddressPairs(s.Ctx))
}
```

(`Quota` and `Flow` field types are `sdkmath.Int` and `uint` in ibc-go v11.2.0's `rate_limiting.pb.go`.)

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./app/upgrades/v36/... 2>&1 | tail -3`
Expected: build failure `undefined: v36.HaltInScopeHostZones` (and the others)

- [ ] **Step 3: Write the helpers**

`app/upgrades/v36/wind_down.go`:

```go
package v36

import (
	"fmt"

	icahostkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/keeper"
	ratelimitkeeper "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icaoraclekeeper "github.com/Stride-Labs/stride/v34/x/icaoracle/keeper"
	stakedymkeeper "github.com/Stride-Labs/stride/v34/x/stakedym/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// HaltInScopeHostZones sets Halted on every stakeibc host zone that is not deprecated (spec §6
// step 1). The halt is what stops every epoch flow and freezes the redemption rate (§9).
// Deprecated zones are left exactly as they are (§2).
func HaltInScopeHostZones(ctx sdk.Context, k stakeibckeeper.Keeper) {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		if hostZone.Deprecated {
			ctx.Logger().Info(fmt.Sprintf("v36: %s is deprecated, not touched", hostZone.ChainId))
			continue
		}
		hostZone.Halted = true
		k.SetHostZone(ctx, hostZone)
		ctx.Logger().Info(fmt.Sprintf("v36: halted %s", hostZone.ChainId))
	}
}

// HaltStakedym halts the stakedym host zone (its operator flushed every record in window 1)
func HaltStakedym(ctx sdk.Context, k stakedymkeeper.Keeper) {
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("v36: stakedym host zone not found, skipping halt: %s", err.Error()))
		return
	}
	hostZone.Halted = true
	k.SetHostZone(ctx, hostZone)
	ctx.Logger().Info("v36: halted stakedym")
}

// DeactivateIcaOracles turns every ICA oracle off; nothing will push a redemption rate again
func DeactivateIcaOracles(ctx sdk.Context, k icaoraclekeeper.Keeper) {
	for _, oracle := range k.GetAllOracles(ctx) {
		if err := k.ToggleOracle(ctx, oracle.ChainId, false); err != nil {
			ctx.Logger().Error(fmt.Sprintf("v36: unable to deactivate oracle %s: %s", oracle.ChainId, err.Error()))
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v36: deactivated oracle %s", oracle.ChainId))
	}
}

// RemoveIcaHostClaimMessage drops MsgClaimUndelegatedTokens from the ICA host allow-list; the
// handler is gone in this binary, so the entry would only produce failed ICAs from other chains
func RemoveIcaHostClaimMessage(ctx sdk.Context, k *icahostkeeper.Keeper) {
	params := k.GetParams(ctx)
	claimUrl := sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{})
	kept := []string{}
	for _, allowed := range params.AllowMessages {
		if allowed == claimUrl {
			continue
		}
		kept = append(kept, allowed)
	}
	params.AllowMessages = kept
	k.SetParams(ctx, params)
	ctx.Logger().Info(fmt.Sprintf("v36: ICA host allow-list now %v", kept))
}

// RemoveAllRateLimits deletes every rate limit, blacklisted denom and whitelisted address pair
// (spec §6 step 3): the sweep sends most of each stToken's on-Stride supply out over channel-5
// in a few days, and there is no mint path left for a limit to protect
func RemoveAllRateLimits(ctx sdk.Context, k *ratelimitkeeper.Keeper) {
	for _, rateLimit := range k.GetAllRateLimits(ctx) {
		k.RemoveRateLimit(ctx, rateLimit.Path.Denom, rateLimit.Path.ChannelOrClientId)
		ctx.Logger().Info(fmt.Sprintf("v36: removed rate limit %s on %s", rateLimit.Path.Denom, rateLimit.Path.ChannelOrClientId))
	}
	for _, denom := range k.GetAllBlacklistedDenoms(ctx) {
		k.RemoveDenomFromBlacklist(ctx, denom)
		ctx.Logger().Info(fmt.Sprintf("v36: removed blacklisted denom %s", denom))
	}
	for _, pair := range k.GetAllWhitelistedAddressPairs(ctx) {
		k.RemoveWhitelistedAddressPair(ctx, pair.Sender, pair.Receiver)
		ctx.Logger().Info(fmt.Sprintf("v36: removed whitelisted pair %s -> %s", pair.Sender, pair.Receiver))
	}
}
```

In `app/upgrades/v36/upgrades.go`, after `RunMigrations`:

```go
		HaltInScopeHostZones(ctx, stakeibcKeeper)
		HaltStakedym(ctx, stakedymKeeper)
		DeactivateIcaOracles(ctx, icaoracleKeeper)
		RemoveIcaHostClaimMessage(ctx, icaHostKeeper)
		RemoveAllRateLimits(ctx, ratelimitKeeper)
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go build ./... && go test ./app/upgrades/v36/... 2>&1 | tail -3`
Expected: `ok`

- [ ] **Step 5: Commit**

```bash
git add app/upgrades/v36
git commit -m "feat(upgrade): v36 - halt in-scope zones and stakedym, deactivate oracles, drop claim from ICA host, remove all rate limits"
```

### Task 4: Remove the last operational handlers, admin-gate the ICQ messages, lift the calibration cap

**Files:**
- Modify: `proto/stride/stakeibc/tx.proto` (remove the `rpc` lines for `ClaimUndelegatedTokens`, `RebalanceValidators`, `ClearBalance`, `ResumeHostZone`), `proto/stride/staketia/tx.proto` (remove `rpc ResumeHostZone`), `proto/stride/stakedym/tx.proto` (remove `rpc ResumeHostZone`)
- Modify: `x/stakeibc/keeper/claim.go` (delete the `msgServer.ClaimUndelegatedTokens` handler; keep `GetClaimableRedemptionRecord` and the other helpers only if `grep -rn` finds a caller outside the deleted handler and its tests, else delete them), `x/stakeibc/keeper/msg_server.go` (delete `RebalanceValidators` ~156-164, `ClearBalance` ~166-206, `ResumeHostZone` ~744-767), `x/stakeibc/handler.go` (delete the four cases), `x/stakeibc/types/codec.go` (four `legacy.RegisterAminoMsg` lines; `RegisterImplementations` entries STAY), `x/stakeibc/client/cli/tx.go` (`CmdClaimUndelegatedTokens`, `CmdRebalanceValidators`, `CmdClearBalance`, `CmdResumeHostZone` and their `AddCommand` lines)
- Modify: `x/staketia/keeper/msg_server.go` (delete `ResumeHostZone`), `x/staketia/types/codec.go` (amino line only), `x/staketia/client/cli/tx.go` (command + `AddCommand`); same three in `x/stakedym`
- Delete: `x/stakeibc/types/message_claim_undelegated_tokens.go` (+`_test`), `message_rebalance_validators.go` (+`_test`), `message_clear_balance.go`, `message_resume_host_zone.go` — after `grep -rn "NewMsg<Name>\|Msg<Name>{" x/ app/ --include='*.go' | grep -v pb.go` shows only the deleted handler/CLI/tests. In `x/staketia/types/msgs.go` and `x/stakedym/types/msgs.go` delete the `MsgResumeHostZone` constructor and methods (the `TypeMsgResumeHostZone` constant and the `_ sdk.Msg` line) but keep the generated type.
- Modify: `x/stakeibc/types/message_update_delegation.go` (`ValidateBasic` gains `utils.ValidateAdminAddress`), `x/stakeibc/types/message_calibrate_delegation.go` (same); create `x/stakeibc/types/message_update_delegation_test.go` and `message_calibrate_delegation_test.go` if absent, else extend
- Modify: `x/stakeibc/keeper/icqcallbacks_callibrate_delegation.go` (delete `CalibrationThreshold` and the `delegationChange.Abs().GT(CalibrationThreshold)` block at ~71-75) and its test (`grep -rln CalibrationThreshold x/` names it)
- Test: `x/stakeibc/keeper/claim_test.go` (delete the nine `TestClaimUndelegatedTokens_*` handler tests; keep any that exercise a surviving helper), `x/stakeibc/keeper/msg_server_test.go` (delete `TestClearBalance*`, `TestRebalanceValidators*`, `TestResumeHostZone*` if present), `x/stakeibc/client/cli/tx_test.go` (delete the four commands' tests), `x/stakedym/keeper/msg_server_test.go` and `x/stakedym/types/msgs_test.go`, `x/staketia/types/msgs_test.go` (delete the `ResumeHostZone` cases)

**Interfaces:**
- Produces: no new interfaces. `Keeper.RebalanceDelegationsForHostZone` (used by the epoch hook) and the claim ICA callback stay.
- Depends on: Tasks 1-2
- Review: yes (the ICQ gates are what keep the frozen rate frozen; spec §9)

- [ ] **Step 1: Write the failing gate tests**

`x/stakeibc/types/message_update_delegation_test.go` (create, or add these cases if the file exists):

```go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgUpdateValidatorSharesExchRate_AdminGate(t *testing.T) {
	validNotAdminAddress, _ := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	require.NoError(t, (&types.MsgUpdateValidatorSharesExchRate{Creator: validAdminAddress, ChainId: "cosmoshub-4", Valoper: "cosmosvaloper1abc"}).ValidateBasic())
	require.ErrorContains(t, (&types.MsgUpdateValidatorSharesExchRate{Creator: validNotAdminAddress, ChainId: "cosmoshub-4", Valoper: "cosmosvaloper1abc"}).ValidateBasic(), "is not an admin")
}
```

`x/stakeibc/types/message_calibrate_delegation_test.go` (same shape):

```go
func TestMsgCalibrateDelegation_AdminGate(t *testing.T) {
	validNotAdminAddress, _ := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	require.NoError(t, (&types.MsgCalibrateDelegation{Creator: validAdminAddress, ChainId: "cosmoshub-4", Valoper: "cosmosvaloper1abc"}).ValidateBasic())
	require.ErrorContains(t, (&types.MsgCalibrateDelegation{Creator: validNotAdminAddress, ChainId: "cosmoshub-4", Valoper: "cosmosvaloper1abc"}).ValidateBasic(), "is not an admin")
}
```

(Use the message's real field names if they differ from `ChainId`/`Valoper`; check the existing `ValidateBasic` in each file.)

Run: `go test ./x/stakeibc/types/... -run 'AdminGate' 2>&1 | tail -3`
Expected: FAIL on the not-admin cases (`ValidateBasic` currently accepts any creator)

- [ ] **Step 2: Add the gates**

In both `ValidateBasic` bodies, directly after the creator address parse:

```go
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
```

adding `"github.com/Stride-Labs/stride/v34/utils"` to the imports.

Run: `go test ./x/stakeibc/types/... -run 'AdminGate' 2>&1 | tail -3`
Expected: `ok`

- [ ] **Step 3: Lift the calibration cap**

In `x/stakeibc/keeper/icqcallbacks_callibrate_delegation.go` delete the `var CalibrationThreshold = sdkmath.NewInt(5000)` line and the block

```go
	if delegationChange.Abs().GT(CalibrationThreshold) {
		k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_Calibrate,
			"Delegation change is GT CalibrationThreshold, failing calibration callback"))
		return nil
	}
```

Add a comment where the block was: `// No cap: the message is admin-gated since v36 (spec §6), so ops may correct any size of drift.` Remove the `sdkmath` import if it becomes unused. In the callback's test file, find the case that asserts a change above 5,000 is rejected and invert it: seed a delegation 1,000,000 base units off, run the callback, assert the validator's `Delegation` and the host zone's `TotalDelegations` moved by the full amount.

Run: `go test ./x/stakeibc/keeper/... -run 'Calibrat' 2>&1 | tail -3`
Expected: `ok`

- [ ] **Step 4: Remove the six rpcs and regenerate**

Delete from `proto/stride/stakeibc/tx.proto`:

```
rpc ClaimUndelegatedTokens(...)
rpc RebalanceValidators(...)
rpc ClearBalance(...)
rpc ResumeHostZone(...)
```

and `rpc ResumeHostZone(MsgResumeHostZone) returns (MsgResumeHostZoneResponse);` from both `proto/stride/staketia/tx.proto` and `proto/stride/stakedym/tx.proto`. Message definitions stay.

Run: `make proto-gen`
Expected: the three `tx.pb.go` files regenerated; the `_Msg_serviceDesc` in each no longer lists the removed methods. Revert descriptor churn elsewhere.

- [ ] **Step 5: Delete handlers, amino lines, CLI commands, dead message helpers**

Follow the **Files** list. For stakeibc: the `ClaimUndelegatedTokens` handler in `claim.go`; the three handlers in `msg_server.go`; four `case` blocks in `handler.go`; four `legacy.RegisterAminoMsg` lines in `codec.go` (add the same "kept for historical decoding" comment upgrade 1 used above the `RegisterImplementations` block if it is not already there); four `Cmd*` functions and their `AddCommand` lines. For staketia and stakedym: the handler, the amino line, the CLI command and `AddCommand` line, and the constructor/methods in `msgs.go`. Delete the `message_*.go` files listed once the grep shows no other reference. Remove imports that become unused (`utils` in stakedym's msg server, for one).

- [ ] **Step 6: Fix the tests**

Delete the tests named under **Files**. Every remaining test in `x/stakeibc/keeper` that used `GetMsgServer().ClaimUndelegatedTokens` must go (the handler no longer exists and nothing replaces it); keeper tests of the claim ICA callback (`icacallbacks_claim_test.go`) stay, they exercise the callback directly.

Run: `go build ./... && go test ./x/stakeibc/... ./x/staketia/... ./x/stakedym/... 2>&1 | tail -6`
Expected: all `ok`

Run: `grep -c "^\s*rpc " proto/stride/stakeibc/tx.proto`
Expected: the Task 2 count minus 4 (record the number in the commit message)

Run: `git diff --stat HEAD -- 'x/*/types/*.pb.go' | grep -v tx.pb.go`
Expected: nothing

- [ ] **Step 7: Commit**

```bash
git add -A proto x/stakeibc x/staketia x/stakedym
git commit -m "feat: remove claim, rebalance, clear-balance and resume handlers; admin-gate the ICQ messages; lift the calibration cap (wind-down 2)"
```

### Task 5: `MsgUndelegateFromValidators` keeper logic

**Files:**
- Modify: `x/stakeibc/keeper/wind_down_undelegate.go` (replace the stub body)
- Create: `x/stakeibc/keeper/wind_down_undelegate_test.go`

**Interfaces:**
- Consumes: `BatchSubmitUndelegateICAMessages` (`unbonding.go`), `applySharesRoundingSafety` and `ValidatorUnbondCapacity` (same file), `GetValidatorFromAddress` (`validator.go`), `GetPendingUndelegationInFlight`/`SetPendingUndelegationInFlight` (`pending_undelegation.go`).
- Produces: the `UndelegateFromValidators` body behind the Task 2 signature.
- Depends on: Tasks 1-2
- Review: yes (money path; per-validator amounts feed an atomic ICA)

- [ ] **Step 1: Write the failing tests**

`x/stakeibc/keeper/wind_down_undelegate_test.go`:

```go
package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	epochstypes "github.com/Stride-Labs/stride/v34/x/epochs/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

type UndelegateFromValidatorsTestCase struct {
	hostZone            types.HostZone
	delegationChannelId string
	delegationPortId    string
}

// Three validators: an unslashed full drain, a slashed one (rate below one, so the safety
// buffer applies), and one with a change in progress
func (s *KeeperTestSuite) SetupUndelegateFromValidators() UndelegateFromValidatorsTestCase {
	owner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	channelId, portId := s.CreateICAChannel(owner)

	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		Halted:               true,
		MaxMessagesPerIcaTx:  2,
		Validators: []*types.Validator{
			{Address: "val1", Delegation: sdkmath.NewInt(1_000_000), SharesToTokensRate: sdkmath.LegacyOneDec()},
			{Address: "val2", Delegation: sdkmath.NewInt(500_000), SharesToTokensRate: sdkmath.LegacyMustNewDecFromStr("0.95")},
			{Address: "val3", Delegation: sdkmath.NewInt(250_000), SharesToTokensRate: sdkmath.LegacyOneDec(), DelegationChangesInProgress: 1},
			{Address: "val4", Delegation: sdkmath.ZeroInt(), SharesToTokensRate: sdkmath.LegacyOneDec()},
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, types.EpochTracker{
		EpochIdentifier:    epochstypes.DAY_EPOCH,
		Duration:           10_000_000_000,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000),
	})
	return UndelegateFromValidatorsTestCase{hostZone: hostZone, delegationChannelId: channelId, delegationPortId: portId}
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_ExplicitList() {
	tc := s.SetupUndelegateFromValidators()
	msg := types.MsgUndelegateFromValidators{ChainId: HostChainId, Validators: []types.ValidatorUndelegation{
		{Address: "val1", Offset: sdkmath.ZeroInt()},
		{Address: "val2", Offset: sdkmath.NewInt(10)},
	}}

	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortId, tc.delegationChannelId)
	numTxs, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, &msg)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), numTxs, "two messages fit in one batch of MaxMessagesPerIcaTx")
	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(tc.delegationPortId, tc.delegationChannelId))

	// Both validators are flagged, val3 and val4 untouched, accounting untouched
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().Equal(int64(1), hostZone.Validators[0].DelegationChangesInProgress)
	s.Require().Equal(int64(1), hostZone.Validators[1].DelegationChangesInProgress)
	s.Require().Equal(int64(1), hostZone.Validators[2].DelegationChangesInProgress)
	s.Require().Equal(sdkmath.NewInt(1_000_000), hostZone.Validators[0].Delegation, "no accounting mutation before the ack")
	s.Require().Equal(uint64(1), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "batch registered as in flight")

	// val1 amount is the full delegation; val2 is delegation - offset, then buffered by the
	// rounding safety because it is a full drain of a slashed validator... except the offset
	// already makes it a partial drain, so no buffer applies
	s.CheckEventValueEmitted(types.EventTypeUndelegateFromValidators, types.AttributeKeyAmount, "1000000")
	s.CheckEventValueEmitted(types.EventTypeUndelegateFromValidators, types.AttributeKeyAmount, "499990")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_EmptyListMeansAllWithDelegation() {
	tc := s.SetupUndelegateFromValidators()
	// val3 has a change in progress and would fail the batch; clear it for this case
	tc.hostZone.Validators[2].DelegationChangesInProgress = 0
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	numTxs, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, &types.MsgUndelegateFromValidators{ChainId: HostChainId})
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), numTxs, "three validators with a delegation, batches of two")
	for _, event := range s.CheckEventTypeEmitted(types.EventTypeUndelegateFromValidators) {
		for _, attribute := range event.Attributes {
			s.Require().NotEqual("val4", attribute.Value, "a validator with no delegation is not submitted")
		}
	}
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_FullDrainOfSlashedValidatorIsBuffered() {
	s.SetupUndelegateFromValidators()
	msg := types.MsgUndelegateFromValidators{ChainId: HostChainId, Validators: []types.ValidatorUndelegation{{Address: "val2"}}}
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, &msg)
	s.Require().NoError(err)
	// applySharesRoundingSafety trims amount/UndelegationSharesSafetyDivisor (1e17) off a full
	// drain of a validator whose rate is below one; that floors to zero, so the buffer is one unit
	s.CheckEventValueEmitted(types.EventTypeUndelegateFromValidators, types.AttributeKeyAmount, "499999")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_Rejections() {
	tc := s.SetupUndelegateFromValidators()
	cases := []struct {
		name string
		msg  types.MsgUndelegateFromValidators
		err  string
	}{
		{name: "unknown zone", msg: types.MsgUndelegateFromValidators{ChainId: "nope"}, err: "host zone not found"},
		{name: "unknown validator", msg: types.MsgUndelegateFromValidators{ChainId: HostChainId, Validators: []types.ValidatorUndelegation{{Address: "val9"}}}, err: "validator not found"},
		{name: "change in progress", msg: types.MsgUndelegateFromValidators{ChainId: HostChainId, Validators: []types.ValidatorUndelegation{{Address: "val3"}}}, err: "delegation changes in progress"},
		{name: "offset consumes everything", msg: types.MsgUndelegateFromValidators{ChainId: HostChainId, Validators: []types.ValidatorUndelegation{{Address: "val1", Offset: sdkmath.NewInt(1_000_000)}}}, err: "not positive"},
		{name: "zero delegation", msg: types.MsgUndelegateFromValidators{ChainId: HostChainId, Validators: []types.ValidatorUndelegation{{Address: "val4"}}}, err: "not positive"},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, &c.msg)
			s.Require().ErrorContains(err, c.err)
		})
	}

	tc.hostZone.Halted = false
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, &types.MsgUndelegateFromValidators{ChainId: HostChainId})
	s.Require().ErrorContains(err, "not halted")
}
```

(`Atom` and `HostChainId` are the existing constants in `keeper_test.go`. `CheckEventTypeEmitted` and `CheckEventValueEmitted` exist in `app/apptesting/test_helpers.go`.)

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestUndelegateFromValidators' 2>&1 | tail -5`
Expected: FAIL with `UndelegateFromValidators not implemented`

- [ ] **Step 3: Implement**

Replace the body of `x/stakeibc/keeper/wind_down_undelegate.go`:

```go
package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v34/utils"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// UndelegateFromValidators submits undelegate ICAs for the listed validators (or every
// validator with a delegation when the list is empty) on a halted zone, with no epoch
// unbonding records attached (spec §6). Per validator the amount is the stored delegation
// minus the offset, passed through applySharesRoundingSafety. The existing undelegate
// callback then decrements the delegation balances and burns nothing. Never touches
// accounting itself.
func (k Keeper) UndelegateFromValidators(ctx sdk.Context, msg *types.MsgUndelegateFromValidators) (numTxsSubmitted uint64, err error) {
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return 0, errorsmod.Wrapf(types.ErrHostZoneNotFound, "host zone not found: %s", msg.ChainId)
	}
	if !hostZone.Halted {
		return 0, errorsmod.Wrapf(types.ErrHostZoneNotHalted, "%s is not halted; the wind-down undelegation only runs on halted zones", msg.ChainId)
	}

	// An empty list means every validator that still has a delegation
	targets := msg.Validators
	if len(targets) == 0 {
		for _, validator := range hostZone.Validators {
			if validator.Delegation.IsPositive() {
				targets = append(targets, types.ValidatorUndelegation{Address: validator.Address})
			}
		}
	}
	if len(targets) == 0 {
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s has no validator with a delegation", msg.ChainId)
	}

	// Build one MsgUndelegate per validator; any rejection fails the whole tx before anything is submitted
	msgs := []proto.Message{}
	splits := []*types.SplitUndelegation{}
	for _, target := range targets {
		validator, _, found := GetValidatorFromAddress(hostZone.Validators, target.Address)
		if !found {
			return 0, errorsmod.Wrapf(types.ErrValidatorNotFound, "validator not found on %s: %s", msg.ChainId, target.Address)
		}
		if validator.DelegationChangesInProgress > 0 {
			return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s has %d delegation changes in progress",
				validator.Address, validator.DelegationChangesInProgress)
		}
		amount := validator.Delegation.Sub(target.OffsetOrZero())
		capacity := ValidatorUnbondCapacity{ValidatorAddress: validator.Address, CurrentDelegation: validator.Delegation}
		amount = k.applySharesRoundingSafety(hostZone, capacity, amount)
		if !amount.IsPositive() {
			return 0, errorsmod.Wrapf(types.ErrInvalidAmount, "undelegation for %s is not positive (delegation %s, offset %s)",
				validator.Address, validator.Delegation, target.OffsetOrZero())
		}

		msgs = append(msgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: hostZone.DelegationIcaAddress,
			ValidatorAddress: validator.Address,
			Amount:           sdk.NewCoin(hostZone.HostDenom, amount),
		})
		splits = append(splits, &types.SplitUndelegation{Validator: validator.Address, NativeTokenAmount: amount})

		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeUndelegateFromValidators,
			sdk.NewAttribute(types.AttributeKeyHostZone, msg.ChainId),
			sdk.NewAttribute(types.AttributeKeyValidator, validator.Address),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		))
	}

	// Submit in the zone's usual batch size, with no epoch unbonding record ids
	batchSize := int(utils.UintToInt(hostZone.MaxMessagesPerIcaTx))
	if batchSize <= 0 {
		batchSize = len(msgs)
	}
	numTxsSubmitted, err = k.BatchSubmitUndelegateICAMessages(ctx, hostZone, []uint64{}, msgs, splits, batchSize)
	if err != nil {
		return 0, err
	}

	// A record-less batch acks through the pending-undelegation branch of UndelegateCallback,
	// which decrements the in-flight counter; register the batches so that accounting is clean
	// (there is no pending amount, so the day-epoch hook has nothing to resubmit)
	inFlight := k.GetPendingUndelegationInFlight(ctx, msg.ChainId)
	k.SetPendingUndelegationInFlight(ctx, msg.ChainId, inFlight+numTxsSubmitted)

	k.Logger(ctx).Info(fmt.Sprintf("wind-down undelegation submitted for %s: %d validators in %d ICA txs", msg.ChainId, len(msgs), numTxsSubmitted))
	return numTxsSubmitted, nil
}
```


- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestUndelegateFromValidators' 2>&1 | tail -5`
Expected: `ok`

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestUndelegateCallback' 2>&1 | tail -3`
Expected: `ok` (the existing callback tests still pass; the new tx changes nothing in the callback)

- [ ] **Step 5: Commit**

```bash
git add x/stakeibc/keeper/wind_down_undelegate.go x/stakeibc/keeper/wind_down_undelegate_test.go
git commit -m "feat(stakeibc): MsgUndelegateFromValidators - per-validator wind-down undelegation with no records"
```

### Task 6: `MsgTransferFromIca` keeper logic

**Files:**
- Modify: `x/stakeibc/keeper/wind_down_transfer_from_ica.go` (replace the stub body)
- Create: `x/stakeibc/keeper/wind_down_transfer_from_ica_test.go`

**Interfaces:**
- Consumes: `SubmitICATxWithoutCallback` (`interchainaccounts.go`), `types.FormatHostZoneICAOwner`, `types.HostToOsmosisTransferChannel`, `types.OsmosisVaultAddress`.
- Produces: the `TransferFromIca` body; helper `func (k Keeper) getWindDownIcaAddress(hostZone types.HostZone, icaType types.ICAAccountType) (string, error)`.
- Depends on: Tasks 1-2
- Review: yes (money path; the receiver constant is the whole protocol's destination)

- [ ] **Step 1: Write the failing tests**

`x/stakeibc/keeper/wind_down_transfer_from_ica_test.go`:

```go
package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// testOsmosisVault is set per test from a generated 20-byte address re-encoded with the osmo prefix
var testOsmosisVault string

type TransferFromIcaTestCase struct {
	hostZone  types.HostZone
	channelId string
	portId    string
}

func (s *KeeperTestSuite) SetupTransferFromIca(chainId string, icaType types.ICAAccountType) TransferFromIcaTestCase {
	previous := types.OsmosisVaultAddress
	testOsmosisVault = sdk.MustBech32ifyAddressBytes(types.OsmosisBech32Prefix, s.TestAccs[0])
	types.OsmosisVaultAddress = testOsmosisVault
	s.T().Cleanup(func() { types.OsmosisVaultAddress = previous })

	owner := types.FormatHostZoneICAOwner(chainId, icaType)
	channelId, portId := s.CreateICAChannel(owner)
	hostZone := types.HostZone{
		ChainId:              chainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "host_DELEGATION",
		WithdrawalIcaAddress: "host_WITHDRAWAL",
		FeeIcaAddress:        "host_FEE",
		RedemptionIcaAddress: "host_REDEMPTION",
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	return TransferFromIcaTestCase{hostZone: hostZone, channelId: channelId, portId: portId}
}

func (s *KeeperTestSuite) TestTransferFromIca_EveryAllowedIcaType() {
	for _, icaType := range types.WindDownAllowedIcaTypes {
		s.Run(icaType.String(), func() {
			s.SetupTest()
			tc := s.SetupTransferFromIca("cosmoshub-4", icaType)
			msg := types.MsgTransferFromIca{ChainId: "cosmoshub-4", IcaType: icaType, Amount: sdk.NewCoin(Atom, sdkmath.NewInt(1000))}
			s.CheckICATxSubmitted(tc.portId, tc.channelId, func() error {
				return s.App.StakeibcKeeper.TransferFromIca(s.Ctx, &msg)
			})
			s.CheckEventValueEmitted(types.EventTypeTransferFromIca, types.AttributeKeyChannel, types.HostToOsmosisTransferChannel["cosmoshub-4"])
			s.CheckEventValueEmitted(types.EventTypeTransferFromIca, types.AttributeKeyReceiver, testOsmosisVault)
		})
	}
}

func (s *KeeperTestSuite) TestTransferFromIca_ForeignDenom() {
	tc := s.SetupTransferFromIca("dydx-mainnet-1", types.ICAAccountType_WITHDRAWAL)
	msg := types.MsgTransferFromIca{ChainId: "dydx-mainnet-1", IcaType: types.ICAAccountType_WITHDRAWAL, Amount: sdk.NewCoin("ibc/USDC", sdkmath.NewInt(50))}
	s.CheckICATxSubmitted(tc.portId, tc.channelId, func() error {
		return s.App.StakeibcKeeper.TransferFromIca(s.Ctx, &msg)
	})
}

func (s *KeeperTestSuite) TestTransferFromIca_OsmosisUsesBankSend() {
	tc := s.SetupTransferFromIca(types.OsmosisChainId, types.ICAAccountType_DELEGATION)
	msg := types.MsgTransferFromIca{ChainId: types.OsmosisChainId, IcaType: types.ICAAccountType_DELEGATION, Amount: sdk.NewCoin("uosmo", sdkmath.NewInt(7))}
	s.CheckICATxSubmitted(tc.portId, tc.channelId, func() error {
		return s.App.StakeibcKeeper.TransferFromIca(s.Ctx, &msg)
	})
	s.CheckEventValueEmitted(types.EventTypeTransferFromIca, types.AttributeKeyChannel, "bank-send")
}

func (s *KeeperTestSuite) TestTransferFromIca_Rejections() {
	tc := s.SetupTransferFromIca("cosmoshub-4", types.ICAAccountType_FEE)
	amount := sdk.NewCoin(Atom, sdkmath.NewInt(1))

	err := s.App.StakeibcKeeper.TransferFromIca(s.Ctx, &types.MsgTransferFromIca{ChainId: "juno-1", IcaType: types.ICAAccountType_FEE, Amount: amount})
	s.Require().ErrorContains(err, "host zone not found")

	tc.hostZone.FeeIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	err = s.App.StakeibcKeeper.TransferFromIca(s.Ctx, &types.MsgTransferFromIca{ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_FEE, Amount: amount})
	s.Require().ErrorContains(err, "ICA acccount not found")

	types.OsmosisVaultAddress = ""
	err = s.App.StakeibcKeeper.TransferFromIca(s.Ctx, &types.MsgTransferFromIca{ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount})
	s.Require().ErrorContains(err, "osmosis vault is not configured")
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestTransferFromIca' 2>&1 | tail -5`
Expected: FAIL with `TransferFromIca not implemented`

- [ ] **Step 3: Implement**

Replace `x/stakeibc/keeper/wind_down_transfer_from_ica.go`:

```go
package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Stride-Labs/stride/v34/utils"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const bankSendChannelLabel = "bank-send"

// TransferFromIca sends `amount` from one of the zone's ICAs to the Osmosis vault (spec §6):
// an ICS-20 transfer over the host's channel to Osmosis from the constant map, or, when the
// host is Osmosis, an ICA bank send. No callback: a failed or timed-out transfer refunds the
// ICA on the host and ops resubmit. Never reads or writes host zone accounting.
func (k Keeper) TransferFromIca(ctx sdk.Context, msg *types.MsgTransferFromIca) error {
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "host zone not found: %s", msg.ChainId)
	}
	icaAddress, err := k.getWindDownIcaAddress(hostZone, msg.IcaType)
	if err != nil {
		return err
	}
	if _, err := sdk.GetFromBech32(types.OsmosisVaultAddress, types.OsmosisBech32Prefix); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "osmosis vault is not configured: %s", err.Error())
	}
	channelId, found := types.HostToOsmosisTransferChannel[msg.ChainId]
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s is not an in-scope wind-down zone", msg.ChainId)
	}

	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())
	var icaMsg proto.Message
	channelLabel := channelId
	if msg.ChainId == types.OsmosisChainId {
		icaMsg = &banktypes.MsgSend{FromAddress: icaAddress, ToAddress: types.OsmosisVaultAddress, Amount: sdk.NewCoins(msg.Amount)}
		channelLabel = bankSendChannelLabel
	} else {
		icaMsg = &transfertypes.MsgTransfer{
			SourcePort:       transfertypes.PortID,
			SourceChannel:    channelId,
			Token:            msg.Amount,
			Sender:           icaAddress,
			Receiver:         types.OsmosisVaultAddress,
			TimeoutTimestamp: timeoutTimestamp,
			Memo:             "",
		}
	}

	owner := types.FormatHostZoneICAOwner(hostZone.ChainId, msg.IcaType)
	if err := k.SubmitICATxWithoutCallback(ctx, hostZone.ConnectionId, owner, []proto.Message{icaMsg}, timeoutTimestamp); err != nil {
		return errorsmod.Wrapf(err, "unable to submit wind-down transfer from %s %s", msg.ChainId, msg.IcaType)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransferFromIca,
		sdk.NewAttribute(types.AttributeKeyHostZone, msg.ChainId),
		sdk.NewAttribute(types.AttributeKeyIcaType, msg.IcaType.String()),
		sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyChannel, channelLabel),
		sdk.NewAttribute(types.AttributeKeyReceiver, types.OsmosisVaultAddress),
	))
	k.Logger(ctx).Info(fmt.Sprintf("wind-down transfer submitted: %s %s %s -> %s via %s", msg.ChainId, msg.IcaType, msg.Amount, types.OsmosisVaultAddress, channelLabel))
	return nil
}

// getWindDownIcaAddress resolves one of the four drainable ICAs; the others are rejected in ValidateBasic
func (k Keeper) getWindDownIcaAddress(hostZone types.HostZone, icaType types.ICAAccountType) (string, error) {
	var address string
	switch icaType {
	case types.ICAAccountType_DELEGATION:
		address = hostZone.DelegationIcaAddress
	case types.ICAAccountType_WITHDRAWAL:
		address = hostZone.WithdrawalIcaAddress
	case types.ICAAccountType_FEE:
		address = hostZone.FeeIcaAddress
	case types.ICAAccountType_REDEMPTION:
		address = hostZone.RedemptionIcaAddress
	default:
		return "", errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ica type %s cannot be drained", icaType)
	}
	if address == "" {
		return "", errorsmod.Wrapf(types.ErrICAAccountNotFound, "%s has no %s ICA address", hostZone.ChainId, icaType)
	}
	return address, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestTransferFromIca' 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 5: Commit**

```bash
git add x/stakeibc/keeper/wind_down_transfer_from_ica.go x/stakeibc/keeper/wind_down_transfer_from_ica_test.go
git commit -m "feat(stakeibc): MsgTransferFromIca - ICA to Osmosis vault over the mapped channel"
```

### Task 7: `MsgTransferStaketiaClaimBalance` keeper logic

**Files:**
- Modify: `x/stakeibc/keeper/wind_down_claim_balance.go` (replace the stub body)
- Create: `x/stakeibc/keeper/wind_down_claim_balance_test.go`

**Interfaces:**
- Consumes: `staketiatypes.ClaimAddress`, `staketiatypes.CelestiaNativeTokenIBCDenom`, `staketiatypes.CelestiaChainId` (`x/staketia/types/celestia.go`; `x/stakeibc` importing `x/staketia/types` is cycle-free, staketia types only import stakeibc types), `k.RecordsKeeper.TransferKeeper.Transfer`, the celestia stakeibc host zone's `TransferChannelId` and `DelegationIcaAddress`.
- Produces: the `TransferStaketiaClaimBalance` body.
- Depends on: Tasks 1-2
- Review: yes (moves the whole staketia balance in one tx)

- [ ] **Step 1: Write the failing tests**

`x/stakeibc/keeper/wind_down_claim_balance_test.go`:

```go
package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

func (s *KeeperTestSuite) SetupTransferStaketiaClaimBalance(balance sdkmath.Int) (claimAddress sdk.AccAddress) {
	s.CreateTransferChannel(staketiatypes.CelestiaChainId)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:              staketiatypes.CelestiaChainId,
		TransferChannelId:    ibctesting.FirstChannelID,
		DelegationIcaAddress: "celestia_DELEGATION",
		HostDenom:            staketiatypes.CelestiaNativeTokenDenom,
		IbcDenom:             staketiatypes.CelestiaNativeTokenIBCDenom,
	})
	claimAddress = sdk.MustAccAddressFromBech32(staketiatypes.ClaimAddress)
	if balance.IsPositive() {
		s.FundAccount(claimAddress, sdk.NewCoin(staketiatypes.CelestiaNativeTokenIBCDenom, balance))
	}
	// Something else in the claim address must be left alone
	s.FundAccount(claimAddress, sdk.NewCoin("ustrd", sdkmath.NewInt(5)))
	return claimAddress
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_Successful() {
	claimAddress := s.SetupTransferStaketiaClaimBalance(sdkmath.NewInt(700_000_000_000))
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)

	sequence, amount, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(startSequence, sequence)
	s.Require().Equal(sdk.NewCoin(staketiatypes.CelestiaNativeTokenIBCDenom, sdkmath.NewInt(700_000_000_000)), amount)
	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID))

	// The voucher left the claim address (escrowed for the outbound transfer); ustrd stayed
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, claimAddress, staketiatypes.CelestiaNativeTokenIBCDenom).IsZero())
	s.Require().Equal(int64(5), s.App.BankKeeper.GetBalance(s.Ctx, claimAddress, "ustrd").Amount.Int64())
	s.CheckEventValueEmitted(stakeibctypes.EventTypeTransferStaketiaClaim, stakeibctypes.AttributeKeyReceiver, "celestia_DELEGATION")
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_ZeroBalance() {
	s.SetupTransferStaketiaClaimBalance(sdkmath.ZeroInt())
	_, _, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx)
	s.Require().ErrorContains(err, "claim address holds no")
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_NoCelestiaZone() {
	s.SetupTransferStaketiaClaimBalance(sdkmath.NewInt(1))
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, staketiatypes.CelestiaChainId)
	_, _, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx)
	s.Require().ErrorContains(err, "host zone not found")
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_NoDelegationIca() {
	s.SetupTransferStaketiaClaimBalance(sdkmath.NewInt(1))
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, staketiatypes.CelestiaChainId)
	hostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	_, _, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx)
	s.Require().ErrorContains(err, "ICA acccount not found")
}
```

(`RemoveHostZone` exists in `host_zone.go`; if named differently use that. `FundAccount` mints to the address.)

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestTransferStaketiaClaimBalance' 2>&1 | tail -5`
Expected: FAIL with `TransferStaketiaClaimBalance not implemented`

- [ ] **Step 3: Implement**

Replace `x/stakeibc/keeper/wind_down_claim_balance.go`:

```go
package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/utils"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

// TransferStaketiaClaimBalance moves the staketia claim address's whole TIA voucher balance to
// the celestia zone's delegation ICA (spec §6), where it unwinds to native TIA and later leaves
// for Osmosis with the zone's balance via MsgTransferFromIca. Everything is a constant or read
// from state; there is nothing for ops to type. Sending from a plain account with keeper code
// is the pattern staketia already uses for its deposit address (x/staketia/keeper/delegation.go).
func (k Keeper) TransferStaketiaClaimBalance(ctx sdk.Context) (sequence uint64, amount sdk.Coin, err error) {
	claimAddress, err := sdk.AccAddressFromBech32(staketiatypes.ClaimAddress)
	if err != nil {
		return 0, amount, errorsmod.Wrap(err, "invalid staketia claim address constant")
	}
	amount = k.bankKeeper.GetBalance(ctx, claimAddress, staketiatypes.CelestiaNativeTokenIBCDenom)
	if !amount.IsPositive() {
		return 0, amount, errorsmod.Wrapf(types.ErrInsufficientFunds, "staketia claim address holds no %s", staketiatypes.CelestiaNativeTokenIBCDenom)
	}

	hostZone, found := k.GetHostZone(ctx, staketiatypes.CelestiaChainId)
	if !found {
		return 0, amount, errorsmod.Wrapf(types.ErrHostZoneNotFound, "host zone not found: %s", staketiatypes.CelestiaChainId)
	}
	if hostZone.DelegationIcaAddress == "" {
		return 0, amount, errorsmod.Wrapf(types.ErrICAAccountNotFound, "%s has no delegation ICA address", hostZone.ChainId)
	}

	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())
	msg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostZone.TransferChannelId,
		Token:            amount,
		Sender:           staketiatypes.ClaimAddress,
		Receiver:         hostZone.DelegationIcaAddress,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             "",
	}
	response, err := k.RecordsKeeper.TransferKeeper.Transfer(ctx, &msg)
	if err != nil {
		return 0, amount, errorsmod.Wrapf(err, "unable to transfer %s from the staketia claim address", amount)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeTransferStaketiaClaim,
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		sdk.NewAttribute(types.AttributeKeyChannel, hostZone.TransferChannelId),
		sdk.NewAttribute(types.AttributeKeyReceiver, hostZone.DelegationIcaAddress),
		sdk.NewAttribute(types.AttributeKeySequence, fmt.Sprintf("%d", response.Sequence)),
	))
	k.Logger(ctx).Info(fmt.Sprintf("staketia claim balance %s sent to %s (sequence %d)", amount, hostZone.DelegationIcaAddress, response.Sequence))
	return response.Sequence, amount, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestTransferStaketiaClaimBalance' 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 5: Commit**

```bash
git add x/stakeibc/keeper/wind_down_claim_balance.go x/stakeibc/keeper/wind_down_claim_balance_test.go
git commit -m "feat(stakeibc): MsgTransferStaketiaClaimBalance - claim address TIA to the celestia delegation ICA"
```

### Task 8: `MsgSweepTokensOffStride` keeper logic

**Files:**
- Modify: `x/stakeibc/keeper/wind_down_sweep.go` (replace the stub body)
- Create: `x/stakeibc/keeper/wind_down_sweep_test.go`

**Interfaces:**
- Consumes: `k.IBCKeeper.ChannelKeeper.GetAllChannelsWithPortPrefix`, `transfertypes.GetEscrowAddress`, `transfertypes.ParseHexHash`, `k.RecordsKeeper.TransferKeeper.GetDenom(ctx, hash) (transfertypes.Denom, bool)` (the pattern in `lsm.go`), `k.AccountKeeper.GetAccount`, `k.bankKeeper.GetBalance`, `k.RecordsKeeper.TransferKeeper.Transfer`, `sdk.MustBech32ifyAddressBytes`, `types.StrideToOsmosisTransferChannelId`, `types.SweepUnwindChannels`.
- Produces: the `SweepTokensOffStride` body; helpers `func (k Keeper) resolveSweepDestination(ctx sdk.Context, denom string) (channelId string, bech32Prefix string, err error)` and `func (k Keeper) isSweepableAccount(ctx sdk.Context, address sdk.AccAddress, escrowAddresses map[string]bool) error`.
- Depends on: Tasks 1-2
- Review: yes (the highest-review item of the spec: moves user balances)

- [ ] **Step 1: Write the failing tests**

`x/stakeibc/keeper/wind_down_sweep_test.go`:

```go
package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	claimvestingtypes "github.com/Stride-Labs/stride/v34/x/claim/vesting/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const stAtom = "stuatom"

type SweepTestCase struct {
	base     sdk.AccAddress
	vesting  sdk.AccAddress
	empty    sdk.AccAddress
	module   sdk.AccAddress
	ica      sdk.AccAddress
	escrow   sdk.AccAddress
	contract sdk.AccAddress
	unknown  sdk.AccAddress
}

func (s *KeeperTestSuite) SetupSweep() SweepTestCase {
	s.CreateTransferChannel(HostChainId)
	previousChannel := types.StrideToOsmosisTransferChannelId
	types.StrideToOsmosisTransferChannelId = ibctesting.FirstChannelID
	s.T().Cleanup(func() { types.StrideToOsmosisTransferChannelId = previousChannel })

	newAccount := func(account sdk.AccountI) sdk.AccAddress {
		s.App.AccountKeeper.SetAccount(s.Ctx, account)
		return account.GetAddress()
	}
	fund := func(address sdk.AccAddress, amount int64) {
		s.FundAccount(address, sdk.NewCoin(stAtom, sdkmath.NewInt(amount)))
		s.FundAccount(address, sdk.NewCoin("ustrd", sdkmath.NewInt(1)))
	}

	tc := SweepTestCase{}
	tc.base = s.TestAccs[0]
	fund(tc.base, 1_000_000)

	vestingBase := authtypes.NewBaseAccountWithAddress(s.TestAccs[1])
	tc.vesting = newAccount(claimvestingtypes.NewStridePeriodicVestingAccount(vestingBase, sdk.NewCoins(sdk.NewCoin("ustrd", sdkmath.NewInt(100))), claimvestingtypes.Periods{}))
	fund(tc.vesting, 2_000_000)

	tc.empty = s.TestAccs[2] // exists once funded with ustrd only
	s.FundAccount(tc.empty, sdk.NewCoin("ustrd", sdkmath.NewInt(1)))

	tc.module = s.App.AccountKeeper.GetModuleAccount(s.Ctx, types.RewardCollectorName).GetAddress()
	fund(tc.module, 10)

	icaBase := authtypes.NewBaseAccountWithAddress(authtypes.NewModuleAddress("ica-host-test"))
	tc.ica = newAccount(icatypes.NewInterchainAccount(icaBase, "owner"))
	fund(tc.ica, 10)

	tc.escrow = transfertypes.GetEscrowAddress(transfertypes.PortID, ibctesting.FirstChannelID)
	fund(tc.escrow, 10)

	tc.contract = sdk.AccAddress(make([]byte, 32)) // 32-byte, contract-style
	newAccount(authtypes.NewBaseAccountWithAddress(tc.contract))
	fund(tc.contract, 10)

	tc.unknown = sdk.AccAddress([]byte("unknown-account-20by")) // 20 bytes, never created
	return tc
}

func (s *KeeperTestSuite) TestSweepTokensOffStride_BaseAndVestingSweptZeroSkipped() {
	tc := s.SetupSweep()
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)

	msg := types.MsgSweepTokensOffStride{Denom: stAtom, Addresses: []string{tc.base.String(), tc.vesting.String(), tc.empty.String()}}
	numSwept, numSkipped, err := s.App.StakeibcKeeper.SweepTokensOffStride(s.Ctx, &msg)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), numSwept)
	s.Require().Equal(uint64(1), numSkipped)
	s.Require().Equal(startSequence+2, s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID))

	// Full stToken balances left; ustrd untouched; receivers are the same bytes with the osmo prefix
	for _, address := range []sdk.AccAddress{tc.base, tc.vesting} {
		s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, address, stAtom).IsZero())
		s.Require().Equal(int64(1), s.App.BankKeeper.GetBalance(s.Ctx, address, "ustrd").Amount.Int64())
		s.CheckEventValueEmitted(types.EventTypeSweepTokensOffStride, types.AttributeKeyReceiver, sdk.MustBech32ifyAddressBytes("osmo", address))
	}
	s.CheckEventValueEmitted(types.EventTypeSweepTokensOffStride, types.AttributeKeySkipped, tc.empty.String())
	// Supply unchanged: swept tokens are escrowed, not burned (spec §9)
	s.Require().Equal(int64(3_000_040), s.App.BankKeeper.GetSupply(s.Ctx, stAtom).Amount.Int64()) // 1M + 2M + 4×10 in the non-sweepable accounts
}

func (s *KeeperTestSuite) TestSweepTokensOffStride_RejectsNonSweepableAddresses() {
	tc := s.SetupSweep()
	cases := []struct {
		name    string
		address sdk.AccAddress
		err     string
	}{
		{name: "module account", address: tc.module, err: "not sweepable"},
		{name: "interchain account", address: tc.ica, err: "not sweepable"},
		{name: "transfer escrow", address: tc.escrow, err: "escrow"},
		{name: "32-byte address", address: tc.contract, err: "20 bytes"},
		{name: "unknown account", address: tc.unknown, err: "no account"},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)
			// A good address in the same batch must not be swept when the batch is rejected
			msg := types.MsgSweepTokensOffStride{Denom: stAtom, Addresses: []string{tc.base.String(), c.address.String()}}
			_, _, err := s.App.StakeibcKeeper.SweepTokensOffStride(s.Ctx, &msg)
			s.Require().ErrorContains(err, c.err)
			s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID), "nothing submitted")
		})
	}
}

// Native denoms (stTokens, ustrd) go to Osmosis over channel-5 with the osmo prefix (spec §6)
func (s *KeeperTestSuite) TestSweepTokensOffStride_NativeDenomsGoToOsmosis() {
	tc := s.SetupSweep()
	for _, denom := range []string{stAtom, "ustrd"} {
		startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)
		numSwept, _, err := s.App.StakeibcKeeper.SweepTokensOffStride(s.Ctx, &types.MsgSweepTokensOffStride{Denom: denom, Addresses: []string{tc.base.String()}})
		s.Require().NoError(err, denom)
		s.Require().Equal(uint64(1), numSwept, denom)
		s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID), denom)
		s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, tc.base, denom).IsZero(), denom)
		s.CheckEventValueEmitted(types.EventTypeSweepTokensOffStride, types.AttributeKeyReceiver, sdk.MustBech32ifyAddressBytes("osmo", tc.base))
		s.CheckEventValueEmitted(types.EventTypeSweepTokensOffStride, types.AttributeKeyChannel, ibctesting.FirstChannelID)
	}
}

// A single-hop voucher on a whitelisted channel goes back over that channel with that
// chain's prefix. The test channel is channel-0, which the real whitelist maps to cosmos.
func (s *KeeperTestSuite) TestSweepTokensOffStride_VoucherUnwindsToSourceChain() {
	tc := s.SetupSweep()
	uatomVoucher := transfertypes.NewDenom("uatom", transfertypes.NewHop(transfertypes.PortID, ibctesting.FirstChannelID))
	s.App.TransferKeeper.SetDenom(s.Ctx, uatomVoucher)
	s.FundAccount(tc.base, sdk.NewCoin(uatomVoucher.IBCDenom(), sdkmath.NewInt(400)))

	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)
	numSwept, _, err := s.App.StakeibcKeeper.SweepTokensOffStride(s.Ctx, &types.MsgSweepTokensOffStride{Denom: uatomVoucher.IBCDenom(), Addresses: []string{tc.base.String()}})
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), numSwept)
	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID))
	s.CheckEventValueEmitted(types.EventTypeSweepTokensOffStride, types.AttributeKeyReceiver, sdk.MustBech32ifyAddressBytes("cosmos", tc.base))
	// Sending a voucher back over its own channel burns it (ICS-20 unwind), it is not escrowed
	s.Require().True(s.App.BankKeeper.GetSupply(s.Ctx, uatomVoucher.IBCDenom()).IsZero())
}

// Vouchers with no destination reject the whole batch before anything is sent
func (s *KeeperTestSuite) TestSweepTokensOffStride_VoucherRejections() {
	tc := s.SetupSweep()
	notWhitelisted := transfertypes.NewDenom("uluna", transfertypes.NewHop(transfertypes.PortID, "channel-999"))
	twoHops := transfertypes.NewDenom("uusdc", transfertypes.NewHop(transfertypes.PortID, ibctesting.FirstChannelID), transfertypes.NewHop(transfertypes.PortID, "channel-7"))
	s.App.TransferKeeper.SetDenom(s.Ctx, notWhitelisted)
	s.App.TransferKeeper.SetDenom(s.Ctx, twoHops)
	unknownHash := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"

	cases := []struct {
		name  string
		denom string
		err   string
	}{
		{name: "channel not whitelisted", denom: notWhitelisted.IBCDenom(), err: "not whitelisted"},
		{name: "two hops", denom: twoHops.IBCDenom(), err: "single-hop"},
		{name: "unknown voucher", denom: unknownHash, err: "no denom trace"},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)
			_, _, err := s.App.StakeibcKeeper.SweepTokensOffStride(s.Ctx, &types.MsgSweepTokensOffStride{Denom: c.denom, Addresses: []string{tc.base.String()}})
			s.Require().ErrorContains(err, c.err)
			s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID))
		})
	}
}

func (s *KeeperTestSuite) TestSweepTokensOffStride_OnlyThatDenomMoves() {
	tc := s.SetupSweep()
	s.FundAccount(tc.base, sdk.NewCoin("stuosmo", sdkmath.NewInt(77)))

	_, _, err := s.App.StakeibcKeeper.SweepTokensOffStride(s.Ctx, &types.MsgSweepTokensOffStride{Denom: stAtom, Addresses: []string{tc.base.String()}})
	s.Require().NoError(err)
	s.Require().Equal(int64(77), s.App.BankKeeper.GetBalance(s.Ctx, tc.base, "stuosmo").Amount.Int64())
}
```

The SDK vesting types are exercised through `vestingtypes.NewContinuousVestingAccount` in one more case: add `TestSweepTokensOffStride_SdkVestingTypes` that creates a `ContinuousVestingAccount`, a `DelayedVestingAccount` and a `PeriodicVestingAccount` (constructors in `vestingtypes`, each from a `BaseAccount` and an `ustrd` vesting amount, end time in the future), funds each with stATOM, sweeps all three and asserts `numSwept == 3`.

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestSweepTokensOffStride' 2>&1 | tail -5`
Expected: FAIL with `SweepTokensOffStride not implemented`

- [ ] **Step 3: Implement**

Replace `x/stakeibc/keeper/wind_down_sweep.go`:

```go
package keeper

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/Stride-Labs/stride/v34/utils"
	claimvestingtypes "github.com/Stride-Labs/stride/v34/x/claim/vesting/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// sweepableAddressLength is the length of a key-derived account address; only those have an
// owner who controls the same bytes on Osmosis (spec §3)
const sweepableAddressLength = 20

// SweepTokensOffStride sends each listed holder's full balance of `denom` off Stride to the
// same address bytes on the destination chain (spec §6): Osmosis for a Stride-native denom,
// the source chain for a single-hop voucher on a whitelisted channel. The destination is
// resolved once, before any address is looked at, and a denom with no destination rejects
// the tx. The batch is also rejected if any address is not sweepable, so the on-chain rule
// set, not the batch-building script, is the safety net. A zero balance is skipped. No memo
// is ever attached. A timeout or a rejected receive refunds the holder through the normal
// ICS-20 path.
func (k Keeper) SweepTokensOffStride(ctx sdk.Context, msg *types.MsgSweepTokensOffStride) (numSwept uint64, numSkipped uint64, err error) {
	channelId, bech32Prefix, err := k.resolveSweepDestination(ctx, msg.Denom)
	if err != nil {
		return 0, 0, err
	}

	escrowAddresses := k.transferEscrowAddresses(ctx)
	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())

	for _, holder := range msg.Addresses {
		address, err := sdk.AccAddressFromBech32(holder)
		if err != nil {
			return 0, 0, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address %s", holder)
		}
		if err := k.isSweepableAccount(ctx, address, escrowAddresses); err != nil {
			return 0, 0, err
		}

		balance := k.bankKeeper.GetBalance(ctx, address, msg.Denom)
		if balance.IsZero() {
			numSkipped++
			ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeSweepTokensOffStride,
				sdk.NewAttribute(types.AttributeKeySkipped, holder)))
			continue
		}

		receiver := sdk.MustBech32ifyAddressBytes(bech32Prefix, address)
		transfer := transfertypes.MsgTransfer{
			SourcePort:       transfertypes.PortID,
			SourceChannel:    channelId,
			Token:            balance,
			Sender:           holder,
			Receiver:         receiver,
			TimeoutTimestamp: timeoutTimestamp,
			Memo:             "",
		}
		response, err := k.RecordsKeeper.TransferKeeper.Transfer(ctx, &transfer)
		if err != nil {
			return 0, 0, errorsmod.Wrapf(err, "unable to sweep %s from %s", balance, holder)
		}
		numSwept++

		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSweepTokensOffStride,
			sdk.NewAttribute(types.AttributeKeyHolder, holder),
			sdk.NewAttribute(types.AttributeKeyReceiver, receiver),
			sdk.NewAttribute(types.AttributeKeyChannel, channelId),
			sdk.NewAttribute(types.AttributeKeyAmount, balance.String()),
			sdk.NewAttribute(types.AttributeKeySequence, fmt.Sprintf("%d", response.Sequence)),
		))
	}

	k.Logger(ctx).Info(fmt.Sprintf("sweep of %s over %s (%s prefix): %d holders swept, %d skipped", msg.Denom, channelId, bech32Prefix, numSwept, numSkipped))
	return numSwept, numSkipped, nil
}

// resolveSweepDestination picks the channel and address prefix for a denom (spec §6): a
// Stride-native denom goes to Osmosis; an ibc/ voucher goes back over its single hop when
// that channel is whitelisted, which is exactly the set of chains whose wallets derive the
// same address bytes as Stride. Everything else has no destination.
func (k Keeper) resolveSweepDestination(ctx sdk.Context, denom string) (channelId string, bech32Prefix string, err error) {
	if !strings.HasPrefix(denom, transfertypes.DenomPrefix+"/") {
		return types.StrideToOsmosisTransferChannelId, types.OsmosisBech32Prefix, nil
	}

	hash, err := transfertypes.ParseHexHash(strings.TrimPrefix(denom, transfertypes.DenomPrefix+"/"))
	if err != nil {
		return "", "", errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid ibc denom %s: %s", denom, err.Error())
	}
	trace, found := k.RecordsKeeper.TransferKeeper.GetDenom(ctx, hash)
	if !found {
		return "", "", errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "no denom trace for %s", denom)
	}
	if len(trace.Trace) != 1 || trace.Trace[0].PortId != transfertypes.PortID {
		return "", "", errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s (%s) is not a single-hop transfer voucher; only those unwind to their source", denom, trace.Path())
	}
	channelId = trace.Trace[0].ChannelId
	bech32Prefix, whitelisted := types.SweepUnwindChannels[channelId]
	if !whitelisted {
		return "", "", errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%s came over %s, which is not whitelisted for the sweep (its chain derives different address bytes)", denom, channelId)
	}
	return channelId, bech32Prefix, nil
}

// transferEscrowAddresses returns every ICS-20 escrow address on the chain, one per transfer
// channel; those hold the stTokens that back vouchers on other chains and must never move
func (k Keeper) transferEscrowAddresses(ctx sdk.Context) map[string]bool {
	escrow := map[string]bool{}
	for _, channel := range k.IBCKeeper.ChannelKeeper.GetAllChannelsWithPortPrefix(ctx, transfertypes.PortID) {
		escrow[transfertypes.GetEscrowAddress(channel.PortId, channel.ChannelId).String()] = true
	}
	return escrow
}

// isSweepableAccount applies the spec §6 rule set: 20-byte address, not an escrow, and an
// account type a key controls (base or vesting). Module accounts, interchain accounts and any
// other type reject the batch.
func (k Keeper) isSweepableAccount(ctx sdk.Context, address sdk.AccAddress, escrowAddresses map[string]bool) error {
	if len(address) != sweepableAddressLength {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "%s is not 20 bytes; contract-style addresses have no owner on Osmosis", address)
	}
	if escrowAddresses[address.String()] {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is a transfer escrow address", address)
	}
	account := k.AccountKeeper.GetAccount(ctx, address)
	if account == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "no account at %s", address)
	}
	switch account.(type) {
	case *authtypes.BaseAccount,
		*vestingtypes.ContinuousVestingAccount,
		*vestingtypes.DelayedVestingAccount,
		*vestingtypes.PeriodicVestingAccount,
		*claimvestingtypes.StridePeriodicVestingAccount:
		return nil
	default:
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is a %T and not sweepable", address, account)
	}
}
```


- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestSweepTokensOffStride' 2>&1 | tail -5`
Expected: `ok`

- [ ] **Step 5: Commit**

```bash
git add x/stakeibc/keeper/wind_down_sweep.go x/stakeibc/keeper/wind_down_sweep_test.go
git commit -m "feat(stakeibc): MsgSweepTokensOffStride - natives to Osmosis, whitelisted vouchers back to their source, on-chain skip rules"
```

### Task 9: Ops scripts — coverage check and sweep batch builder

**Files:**
- Create: `scripts/wind-down/coverage_check.py`
- Create: `scripts/wind-down/build_sweep_batches.py`
- Create: `scripts/wind-down/test_wind_down_scripts.py`
- Create: `scripts/wind-down/README.md`

**Interfaces:**
- Consumes: a Stride export (`strided export` JSON, `app_state.bank.balances`, `app_state.bank.supply`, `app_state.stakeibc.host_zone_list`, `app_state.auth.accounts`, `app_state.ibc.channel_genesis.channels`), prices as a JSON `{denom: usd_per_native_token}`, and Osmosis vault balances as a JSON `{native_denom_on_osmosis: amount}` (both files written by ops).
- Produces: `coverage_check.py EXPORT OSMOSIS_BALANCES` prints one row per in-scope stToken (supply, frozen rate, required native, held native, surplus or shortfall) and exits non-zero on any shortfall; `build_sweep_batches.py EXPORT PRICES --floor-usd 10 --batch-size 100 --out DIR [--extra-denom ustrd=0.05:6 ...]` writes `DIR/<denom>/batch-NNN.txt` files (one address per line) that `strided tx stakeibc sweep-tokens-off-stride` consumes, applying the same skip rules as the chain plus the dollar floor, for every stToken plus each listed extra denom.
- Depends on: nothing in Go (Python only). Complements `scripts/wind-down/check_transmuter_pool.py` (already on the branch), which checks a created pool's factors, roles and limiters; these two scripts decide how much goes in and who gets swept.
- Review: yes (its output decides how much native goes into each pool)

- [ ] **Step 1: Write the failing tests**

`scripts/wind-down/test_wind_down_scripts.py`:

```python
"""Unit tests for the wind-down ops scripts against a tiny synthetic export."""

import json
import pathlib
import tempfile
import unittest
from decimal import Decimal

import build_sweep_batches
import coverage_check

ESCROW_CHANNEL_5 = "stride1cvygnqhawy2t2pf9d5vm2mpvazyuw6h4pt8y37"  # escrow of transfer/channel-5, ADR-028
BASE = "stride1base000000000000000000000000000000000"
VESTING = "stride1vest000000000000000000000000000000000"
MODULE = "stride1modu000000000000000000000000000000000"
ICA = "stride1icaa000000000000000000000000000000000"
DUST = "stride1dust000000000000000000000000000000000"


def synthetic_export() -> dict:
    return {
        "app_state": {
            "auth": {"accounts": [
                {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": BASE},
                {"@type": "/stride.vesting.StridePeriodicVestingAccount", "base_vesting_account": {"base_account": {"address": VESTING}}},
                {"@type": "/cosmos.auth.v1beta1.ModuleAccount", "base_account": {"address": MODULE}, "name": "distribution"},
                {"@type": "/ibc.applications.interchain_accounts.v1.InterchainAccount", "base_account": {"address": ICA}},
                {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": DUST},
                {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": ESCROW_CHANNEL_5},
            ]},
            "bank": {
                "balances": [
                    {"address": BASE, "coins": [{"denom": "stuatom", "amount": "1000000000"}]},
                    {"address": VESTING, "coins": [{"denom": "stuatom", "amount": "500000000"}]},
                    {"address": MODULE, "coins": [{"denom": "stuatom", "amount": "300000000"}]},
                    {"address": ICA, "coins": [{"denom": "stuatom", "amount": "200000000"}]},
                    {"address": DUST, "coins": [{"denom": "stuatom", "amount": "1000"}]},
                    {"address": ESCROW_CHANNEL_5, "coins": [{"denom": "stuatom", "amount": "4000000000"}]},
                ],
                "supply": [{"denom": "stuatom", "amount": "6000001000"}],
            },
            "stakeibc": {"host_zone_list": [
                {"chain_id": "cosmoshub-4", "host_denom": "uatom", "redemption_rate": "2.000000000000000000", "halted": True, "deprecated": False},
                {"chain_id": "comdex-1", "host_denom": "ucmdx", "redemption_rate": "1.4", "halted": False, "deprecated": True},
            ]},
            "ibc": {"channel_genesis": {"channels": [{"port_id": "transfer", "channel_id": "channel-5", "state": "STATE_OPEN"}]}},
            "transfer": {"denoms": [
                {"base": "uusdc", "trace": [{"port_id": "transfer", "channel_id": "channel-999"}]},
                {"base": "uatom", "trace": [{"port_id": "transfer", "channel_id": "channel-0"}]},
            ]},
        }
    }


class CoverageCheckTest(unittest.TestCase):
    def test_required_native_is_supply_times_rate(self) -> None:
        rows = coverage_check.compute_rows(export=synthetic_export(), osmosis_balances={"uatom": "12000002000"})
        self.assertEqual(len(rows), 1, "deprecated zones are excluded")
        row = rows[0]
        self.assertEqual(row.st_denom, "stuatom")
        self.assertEqual(row.required_native, Decimal("12000002000"))
        self.assertEqual(row.surplus, Decimal("0"))

    def test_shortfall_is_reported(self) -> None:
        rows = coverage_check.compute_rows(export=synthetic_export(), osmosis_balances={"uatom": "1"})
        self.assertTrue(rows[0].surplus < 0)
        self.assertTrue(coverage_check.has_shortfall(rows))


class BuildSweepBatchesTest(unittest.TestCase):
    def test_skip_rules_and_floor(self) -> None:
        holders = build_sweep_batches.sweepable_holders(
            export=synthetic_export(), prices={"uatom": Decimal("1.8")}, floor_usd=Decimal("10")
        )
        self.assertEqual({h.address for h in holders["stuatom"]}, {BASE, VESTING})

    def test_extra_denoms_are_swept_when_listed(self) -> None:
        export = synthetic_export()
        export["app_state"]["bank"]["balances"][0]["coins"].append({"denom": "ustrd", "amount": "500000000"})
        without = build_sweep_batches.sweepable_holders(export=export, prices={"uatom": Decimal("1.8")}, floor_usd=Decimal("10"))
        self.assertNotIn("ustrd", without)
        with_strd = build_sweep_batches.sweepable_holders(
            export=export, prices={"uatom": Decimal("1.8")}, floor_usd=Decimal("10"),
            extra_denoms={"ustrd": build_sweep_batches.ExtraDenom(usd_per_token=Decimal("0.05"), decimals=6)},
        )
        self.assertEqual([h.address for h in with_strd["ustrd"]], [BASE])

    def test_voucher_off_whitelist_is_rejected(self) -> None:
        export = synthetic_export()
        usdc = build_sweep_batches.ibc_denom({"base": "uusdc", "trace": [{"port_id": "transfer", "channel_id": "channel-999"}]})
        atom = build_sweep_batches.ibc_denom({"base": "uatom", "trace": [{"port_id": "transfer", "channel_id": "channel-0"}]})
        extra = build_sweep_batches.ExtraDenom(usd_per_token=Decimal("1"), decimals=6)
        build_sweep_batches.reject_unwindable_vouchers(export=export, extra_denoms={atom: extra})
        with self.assertRaises(SystemExit):
            build_sweep_batches.reject_unwindable_vouchers(export=export, extra_denoms={usdc: extra})

    def test_batches_are_bounded_and_written(self) -> None:
        holders = build_sweep_batches.sweepable_holders(
            export=synthetic_export(), prices={"uatom": Decimal("1.8")}, floor_usd=Decimal("0")
        )
        with tempfile.TemporaryDirectory() as out_dir:
            written = build_sweep_batches.write_batches(holders=holders, batch_size=2, out_dir=pathlib.Path(out_dir))
            self.assertEqual(written, {"stuatom": 2})
            lines = (pathlib.Path(out_dir) / "stuatom" / "batch-000.txt").read_text().splitlines()
            self.assertEqual(len(lines), 2)


if __name__ == "__main__":
    unittest.main()
```

Replace `ESCROW_CHANNEL_5` with the real escrow address of `transfer/channel-5`, computed once with the same function the script uses (`python3 -c "import build_sweep_batches; print(build_sweep_batches.escrow_address('transfer', 'channel-5'))"` after Step 3), so the test proves the derivation matches ADR-028.

Run: `cd scripts/wind-down && python3 -m unittest test_wind_down_scripts 2>&1 | tail -3`
Expected: `ModuleNotFoundError: No module named 'build_sweep_batches'`

- [ ] **Step 2: Write the batch builder**

`scripts/wind-down/build_sweep_batches.py`:

```python
"""Build MsgSweepTokensOffStride batches from a Stride export (wind-down spec §6, §8 window 2 step 6).

Applies the chain's skip rules (20-byte address, base or vesting account, not an escrow) plus
the dollar floor ops chose, and writes one file per batch that the CLI consumes:

    python3 build_sweep_batches.py export.json prices.json --floor-usd 10 --batch-size 100 --out batches/

Every in-scope stToken is on the sweep list automatically, priced through its zone's redemption
rate. Any other denom (ustrd, a USDC voucher) is added with --extra-denom DENOM=USD_PER_TOKEN:DECIMALS,
e.g. --extra-denom ustrd=0.05:6. An ibc/ extra denom is accepted only if its trace in the export
is a single hop over a channel in SWEEP_UNWIND_CHANNELS (mirrors x/stakeibc/types/wind_down.go),
which is the same rule the chain applies. The chain re-checks every rule; this script only
decides who is above the floor.
"""

import argparse
import collections
import dataclasses
import hashlib
import json
import pathlib
from decimal import Decimal

import bech32

SWEEPABLE_ACCOUNT_TYPES = {
    "/cosmos.auth.v1beta1.BaseAccount",
    "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
    "/cosmos.vesting.v1beta1.DelayedVestingAccount",
    "/cosmos.vesting.v1beta1.PeriodicVestingAccount",
    "/stride.vesting.StridePeriodicVestingAccount",
}
EIGHTEEN_DECIMAL_HOST_DENOMS = {"adydx", "aISLM", "inj"}
ST_PREFIX = "st"
ADDRESS_BYTES = 20
IBC_PREFIX = "ibc/"
# Mirror of types.SweepUnwindChannels; keep the two in step
SWEEP_UNWIND_CHANNELS = {"channel-0", "channel-162", "channel-5", "channel-24", "channel-150", "channel-213", "channel-160"}


@dataclasses.dataclass
class Holder:
    address: str
    st_denom: str
    amount: int
    value_usd: Decimal


@dataclasses.dataclass(frozen=True)
class ExtraDenom:
    """A non-stToken denom on the sweep list, priced directly rather than through a zone."""

    usd_per_token: Decimal
    decimals: int


def parse_extra_denom(entry: str) -> tuple[str, ExtraDenom]:
    denom, spec = entry.split("=", 1)
    price, decimals = spec.split(":", 1)
    return denom, ExtraDenom(usd_per_token=Decimal(price), decimals=int(decimals))


def main() -> None:
    args = parse_args()
    export = json.load(open(args.export))
    prices = {denom: Decimal(str(price)) for denom, price in json.load(open(args.prices)).items()}

    extra_denoms = {denom: spec for denom, spec in (parse_extra_denom(entry) for entry in args.extra_denom)}
    reject_unwindable_vouchers(export=export, extra_denoms=extra_denoms)
    holders = sweepable_holders(export=export, prices=prices, floor_usd=Decimal(str(args.floor_usd)), extra_denoms=extra_denoms)
    written = write_batches(holders=holders, batch_size=args.batch_size, out_dir=pathlib.Path(args.out))
    for st_denom, batches in sorted(written.items()):
        print(f"{st_denom}: {len(holders[st_denom])} holders in {batches} batches")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("export")
    parser.add_argument("prices", help='JSON {host_denom: usd per whole native token}, e.g. {"uatom": 1.8}')
    parser.add_argument("--floor-usd", type=float, default=10.0)
    parser.add_argument("--batch-size", type=int, default=100)
    parser.add_argument("--out", default="batches")
    parser.add_argument("--extra-denom", action="append", default=[], metavar="DENOM=USD_PER_TOKEN:DECIMALS",
                        help="a non-stToken denom to sweep, with its price and exponent")
    return parser.parse_args()


def sweepable_holders(
    export: dict, prices: dict[str, Decimal], floor_usd: Decimal, extra_denoms: dict[str, ExtraDenom] | None = None
) -> dict[str, list[Holder]]:
    """Every (holder, denom) pair that passes the chain's rules and is worth at least floor_usd."""
    zones = in_scope_zones(export)
    extra_denoms = extra_denoms or {}
    account_types = account_type_by_address(export)
    escrows = escrow_addresses(export)

    holders: dict[str, list[Holder]] = collections.defaultdict(list)
    for balance in export["app_state"]["bank"]["balances"]:
        address = balance["address"]
        if not is_sweepable_address(address=address, account_types=account_types, escrows=escrows):
            continue
        for coin in balance["coins"]:
            value_usd = coin_value_usd(denom=coin["denom"], amount=int(coin["amount"]), zones=zones, prices=prices, extra_denoms=extra_denoms)
            if value_usd is None or value_usd < floor_usd:
                continue
            holders[coin["denom"]].append(Holder(address=address, st_denom=coin["denom"], amount=int(coin["amount"]), value_usd=value_usd))

    # Largest first, so a batch that fails late in the run is the cheapest one to redo
    for st_denom in holders:
        holders[st_denom].sort(key=lambda holder: -holder.amount)
    return holders


def write_batches(holders: dict[str, list[Holder]], batch_size: int, out_dir: pathlib.Path) -> dict[str, int]:
    written: dict[str, int] = {}
    for st_denom, denom_holders in holders.items():
        denom_dir = out_dir / st_denom
        denom_dir.mkdir(parents=True, exist_ok=True)
        batches = [denom_holders[i : i + batch_size] for i in range(0, len(denom_holders), batch_size)]
        for index, batch in enumerate(batches):
            (denom_dir / f"batch-{index:03d}.txt").write_text("\n".join(holder.address for holder in batch) + "\n")
        written[st_denom] = len(batches)
    return written


def reject_unwindable_vouchers(export: dict, extra_denoms: dict[str, ExtraDenom]) -> None:
    """Fail fast on an ibc/ extra denom the chain would reject: not single-hop, or not on a whitelisted channel."""
    traces = {IBC_PREFIX + entry["hash"] if "hash" in entry else ibc_denom(entry): entry for entry in export["app_state"]["transfer"]["denoms"]}
    for denom in extra_denoms:
        if not denom.startswith(IBC_PREFIX):
            continue
        trace = traces.get(denom)
        if trace is None:
            raise SystemExit(f"{denom}: no denom trace in the export")
        hops = trace["trace"]
        if len(hops) != 1 or hops[0]["channel_id"] not in SWEEP_UNWIND_CHANNELS:
            raise SystemExit(f"{denom}: trace {hops} is not a single hop over a whitelisted channel; the chain would reject it")


def ibc_denom(entry: dict) -> str:
    """ibc/<sha256 of the full trace path> for a transfer-genesis denom entry."""
    path = "/".join(f"{hop['port_id']}/{hop['channel_id']}" for hop in entry["trace"]) + "/" + entry["base"]
    return IBC_PREFIX + hashlib.sha256(path.encode()).hexdigest().upper()


def in_scope_zones(export: dict) -> dict[str, dict]:
    """stToken denom -> host zone, for zones that are not deprecated."""
    return {
        ST_PREFIX + zone["host_denom"]: zone
        for zone in export["app_state"]["stakeibc"]["host_zone_list"]
        if not zone.get("deprecated", False)
    }


def account_type_by_address(export: dict) -> dict[str, str]:
    types: dict[str, str] = {}
    for account in export["app_state"]["auth"]["accounts"]:
        types[account_address(account)] = account["@type"]
    return types


def account_address(account: dict) -> str:
    """Every SDK account type nests the address differently; walk down to it."""
    if "address" in account:
        return account["address"]
    if "base_account" in account:
        return account_address(account["base_account"])
    return account_address(account["base_vesting_account"])


def escrow_addresses(export: dict) -> set[str]:
    channels = export["app_state"]["ibc"]["channel_genesis"]["channels"]
    return {escrow_address(channel["port_id"], channel["channel_id"]) for channel in channels if channel["port_id"] == "transfer"}


def escrow_address(port_id: str, channel_id: str) -> str:
    """ICS-20 escrow address per ADR-028: sha256("ics20-1" || 0x00 || "port/channel")[:20]."""
    digest = hashlib.sha256(b"ics20-1" + b"\x00" + f"{port_id}/{channel_id}".encode()).digest()[:ADDRESS_BYTES]
    return bech32.bech32_encode("stride", bech32.convertbits(digest, 8, 5))


def is_sweepable_address(address: str, account_types: dict[str, str], escrows: set[str]) -> bool:
    _, data = bech32.bech32_decode(address)
    if data is None or len(bech32.convertbits(data, 5, 8, False)) != ADDRESS_BYTES:
        return False
    if address in escrows:
        return False
    return account_types.get(address) in SWEEPABLE_ACCOUNT_TYPES


def coin_value_usd(
    denom: str, amount: int, zones: dict[str, dict], prices: dict[str, Decimal], extra_denoms: dict[str, ExtraDenom]
) -> Decimal | None:
    """USD value of a balance, or None when the denom is not on the sweep list."""
    zone = zones.get(denom)
    if zone is not None:
        return usd_value(zone=zone, amount=amount, prices=prices)
    extra = extra_denoms.get(denom)
    if extra is None:
        return None
    return Decimal(amount) / Decimal(10**extra.decimals) * extra.usd_per_token


def usd_value(zone: dict, amount: int, prices: dict[str, Decimal]) -> Decimal:
    decimals = 18 if zone["host_denom"] in EIGHTEEN_DECIMAL_HOST_DENOMS else 6
    native = Decimal(amount) * Decimal(zone["redemption_rate"]) / Decimal(10**decimals)
    return native * prices.get(zone["host_denom"], Decimal("0"))


if __name__ == "__main__":
    main()
```

- [ ] **Step 3: Write the coverage check**

`scripts/wind-down/coverage_check.py`:

```python
"""Coverage check for the transmuter pools (wind-down spec §9).

Per in-scope stToken: required native = Stride bank supply × HostZone.RedemptionRate, compared
with what the Osmosis vault holds of that native denom. Run against a fresh export before each
pool is funded and again before the halt:

    python3 coverage_check.py export.json osmosis_balances.json

osmosis_balances.json is {host_denom: amount} in base units as held by the Osmosis vault, keyed
by the host denom (uatom, utia, ...) rather than the ibc/ hash so the file is readable. Exits 1
on any shortfall.
"""

import dataclasses
import json
import sys
from decimal import Decimal

ST_PREFIX = "st"


@dataclasses.dataclass
class CoverageRow:
    st_denom: str
    host_denom: str
    supply: Decimal
    redemption_rate: Decimal
    required_native: Decimal
    held_native: Decimal

    @property
    def surplus(self) -> Decimal:
        return self.held_native - self.required_native


def main() -> None:
    export = json.load(open(sys.argv[1]))
    osmosis_balances = json.load(open(sys.argv[2]))

    rows = compute_rows(export=export, osmosis_balances=osmosis_balances)
    print(f"{'stToken':10s} {'supply':>22s} {'rate':>22s} {'required':>22s} {'held':>22s} {'surplus':>22s}")
    for row in rows:
        print(f"{row.st_denom:10s} {row.supply:22.0f} {row.redemption_rate:22.18f} {row.required_native:22.0f} {row.held_native:22.0f} {row.surplus:22.0f}")

    if has_shortfall(rows):
        print("SHORTFALL: at least one pool would be under-funded", file=sys.stderr)
        sys.exit(1)


def compute_rows(export: dict, osmosis_balances: dict[str, str]) -> list[CoverageRow]:
    supply = {coin["denom"]: Decimal(coin["amount"]) for coin in export["app_state"]["bank"]["supply"]}
    rows = []
    for zone in export["app_state"]["stakeibc"]["host_zone_list"]:
        if zone.get("deprecated", False):
            continue
        st_denom = ST_PREFIX + zone["host_denom"]
        rate = Decimal(zone["redemption_rate"])
        st_supply = supply.get(st_denom, Decimal("0"))
        rows.append(CoverageRow(
            st_denom=st_denom,
            host_denom=zone["host_denom"],
            supply=st_supply,
            redemption_rate=rate,
            required_native=(st_supply * rate).to_integral_value(rounding="ROUND_CEILING"),
            held_native=Decimal(osmosis_balances.get(zone["host_denom"], "0")),
        ))
    return sorted(rows, key=lambda row: row.st_denom)


def has_shortfall(rows: list[CoverageRow]) -> bool:
    return any(row.surplus < 0 for row in rows)


if __name__ == "__main__":
    main()
```

`scripts/wind-down/README.md`: one paragraph per script (purpose, invocation as in the docstrings, where the inputs come from: `strided export --height H`, the prices file from the day's quotes, the Osmosis balances from `osmosisd q bank balances <vault>` translated to host denoms), and the note that the chain re-applies the skip rules so a wrong batch fails loudly rather than sweeping the wrong account.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd scripts/wind-down && pip3 show bech32 >/dev/null 2>&1 || pip3 install bech32; python3 -m unittest test_wind_down_scripts 2>&1 | tail -3`
Expected: `OK`

- [ ] **Step 5: Run both against the real export and check the output in**

Run (needs a fresh `strided export` as `export.json`, and the two input files):

```bash
python3 scripts/wind-down/build_sweep_batches.py export.json prices.json --floor-usd 10 --extra-denom ustrd=<price>:6 --out /tmp/batches | tee scripts/wind-down/sweep_batches_summary.txt
python3 scripts/wind-down/coverage_check.py export.json osmosis_balances.json | tee scripts/wind-down/coverage_report.txt || true
```

Expected: the summary lists the eleven stTokens plus `ustrd` with holder counts in the thousands (5,663 pairs at the $10 floor on 2026-09-22 prices); the coverage report shows a shortfall on every row today (nothing is on Osmosis yet), which is the expected pre-migration state. Commit the two text files as the fixture of what the scripts produce.

- [ ] **Step 6: Commit**

```bash
git add scripts/wind-down
git commit -m "ops(wind-down): coverage check and sweep batch builder with tests"
```

## Integration tasks (serial, after every parallel task has merged)

### Task 10: Release gate — operator constants, removed-message guard, mainnet export suite, changelog

**Files:**
- Modify: `x/stakeibc/types/wind_down.go` (fill `SweepOperatorAddress` and `OsmosisVaultAddress`)
- Create: `app/upgrades/v36/removed_messages_test.go`
- Create: `app/upgrades/v36/mainnet_export_test.go`
- Create: `app/upgrades/v36/testdata/README.md`
- Create: `app/upgrades/v36/testdata/mainnet_export.json.gz` (generated, committed)
- Modify: `CHANGELOG.md`

**Interfaces:**
- Consumes: every helper from Task 3; the four message types from Task 2; `types.HostToOsmosisTransferChannel`.
- Depends on: Tasks 1-9
- Review: yes (release gate)

- [ ] **Step 1: Fill the operator constants**

Once the sweep operator key and the Osmosis vault multisig exist and have each signed a spend (spec §8, upgrade 2 checklist), set the two values in `x/stakeibc/types/wind_down.go`:

```go
	SweepOperatorAddress = "<stride1... from the spec §3a table>"
	OsmosisVaultAddress  = "<osmo1... from the spec §3a table>"
```

and record both in the spec's §3a table in the same commit. Run: `go test ./x/stakeibc/types/... -run TestWindDownAddressesConfigured -v 2>&1 | tail -3`
Expected: `--- PASS` (no longer SKIP). This is the only step of the plan that may land after the code review, but it must land before the upgrade proposal.

- [ ] **Step 2: Write the removed-message guard**

`app/upgrades/v36/removed_messages_test.go`:

```go
package v36_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	stakedymtypes "github.com/Stride-Labs/stride/v34/x/stakedym/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

// Every message spec §6 removes must have no tx handler; every message it adds or keeps must have one
func (s *UpgradeTestSuite) TestRemovedAndAddedMessages() {
	removed := []sdk.Msg{
		&stakeibctypes.MsgClaimUndelegatedTokens{}, &stakeibctypes.MsgRebalanceValidators{},
		&stakeibctypes.MsgClearBalance{}, &stakeibctypes.MsgResumeHostZone{},
		&staketiatypes.MsgResumeHostZone{}, &stakedymtypes.MsgResumeHostZone{},
	}
	for _, msg := range removed {
		s.Require().Nil(s.App.MsgServiceRouter().Handler(msg), "%s must have no handler", sdk.MsgTypeURL(msg))
	}

	kept := []sdk.Msg{
		&stakeibctypes.MsgUndelegateFromValidators{}, &stakeibctypes.MsgTransferFromIca{},
		&stakeibctypes.MsgTransferStaketiaClaimBalance{}, &stakeibctypes.MsgSweepTokensOffStride{},
		&stakeibctypes.MsgUpdateValidatorSharesExchRate{}, &stakeibctypes.MsgCalibrateDelegation{},
		&stakeibctypes.MsgRestoreInterchainAccount{},
		&staketiatypes.MsgConfirmUndelegation{}, &staketiatypes.MsgConfirmUnbondedTokenSweep{},
		&stakedymtypes.MsgConfirmUndelegation{}, &stakedymtypes.MsgConfirmUnbondedTokenSweep{},
	}
	for _, msg := range kept {
		s.Require().NotNil(s.App.MsgServiceRouter().Handler(msg), "%s must have a handler", sdk.MsgTypeURL(msg))
	}
}
```

Add `TestRemovedMessagesStillDecode` exactly as upgrade 1's Task 8 did, for `&stakeibctypes.MsgClaimUndelegatedTokens{Creator: <valid bech32>, ChainId: "cosmoshub-4", EpochNumber: 1}` and `&stakedymtypes.MsgResumeHostZone{Creator: <valid bech32>}`: build, encode and decode a tx and assert one message of the right type comes back.

Run: `go test ./app/upgrades/v36/... -run 'TestUpgradeTestSuite/TestRemoved' 2>&1 | tail -3`
Expected: `ok`

- [ ] **Step 3: Assemble the mainnet export fixture**

`app/upgrades/v36/testdata/README.md`:

```markdown
# v36 mainnet export fixture

`mainnet_export.json.gz` is post-v35 mainnet state trimmed to what the v36 mainnet-export
suite consumes:

- `app_state.stakeibc.host_zone_list` — every stakeibc host zone (halt and deprecated flags).
- `app_state.stakedym.host_zone` — the stakedym host zone.
- `app_state.icaoracle.oracles` — every ICA oracle.
- `app_state.ratelimit` — rate limits, blacklisted denoms, whitelisted address pairs.
- `app_state.interchainaccounts.host_genesis_state.params` — the ICA host allow-list.

The suite skips when the file is absent. Regenerate right before the proposal:

```bash
API=https://stride-api.polkachu.com
H="x-cosmos-block-height: <HEIGHT>"
curl -s -A curl/8.0 -H "$H" $API/Stride-Labs/stride/stakeibc/host_zone                > hz.json
curl -s -A curl/8.0 -H "$H" $API/Stride-Labs/stride/stakedym/host_zone                > dym.json
curl -s -A curl/8.0 -H "$H" $API/Stride-Labs/stride/icaoracle/oracles                 > oracles.json
curl -s -A curl/8.0 -H "$H" $API/ibc/apps/rate-limiting/v1/ratelimits                 > rl.json
curl -s -A curl/8.0 -H "$H" $API/ibc/apps/rate-limiting/v1/blacklisted_denoms         > bl.json
curl -s -A curl/8.0 -H "$H" $API/ibc/apps/rate-limiting/v1/whitelisted_addresses      > wl.json
curl -s -A curl/8.0 -H "$H" $API/ibc/apps/interchain_accounts/host/v1/params          > icahost.json

jq -n --slurpfile h hz.json --slurpfile d dym.json --slurpfile o oracles.json \
      --slurpfile rl rl.json --slurpfile bl bl.json --slurpfile wl wl.json --slurpfile p icahost.json \
  '{app_state: {
      stakeibc: {host_zone_list: $h[0].host_zone},
      stakedym: {host_zone: $d[0].host_zone},
      icaoracle: {oracles: $o[0].oracles},
      ratelimit: {rate_limits: $rl[0].rate_limits, blacklisted_denoms: $bl[0].denoms, whitelisted_address_pairs: $wl[0].address_pairs},
      interchainaccounts: {host_genesis_state: {params: $p[0].params}}}}' \
  | gzip -9 > mainnet_export.json.gz
```

(The REST paths for the rate-limiting queries are those served by ibc-go v11's rate-limiting
module; confirm them against `osmosisd`-style `--help` on a full node if any 404s.)
```

Generate the fixture with those commands at the current height.

- [ ] **Step 4: Write the mainnet export suite**

`app/upgrades/v36/mainnet_export_test.go`, modelled on `app/upgrades/v34/mainnet_export_test.go` (copy its `strideExport` type, the gzip loader and the `populateHostZonesFromExport` helper; add `populateStakedymFromExport`, `populateOraclesFromExport`, `populateRateLimitsFromExport`, `populateIcaHostParamsFromExport` that unmarshal the fixture's JSON into the module types with `s.App.AppCodec()` and call the keepers' `Set*` methods):

```go
func (s *MainnetExportTestSuite) TestUpgradeAgainstMainnetExport() {
	export := s.loadTrimmedExport()
	s.populateHostZonesFromExport(export)
	s.populateStakedymFromExport(export)
	s.populateOraclesFromExport(export)
	s.populateRateLimitsFromExport(export)
	s.populateIcaHostParamsFromExport(export)

	// Sanity: the fixture is pre-upgrade state
	s.Require().NotEmpty(s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx), "fixture has rate limits")

	s.ConfirmUpgradeSucceeded(v36.UpgradeName)

	halted := map[string]bool{}
	for _, hostZone := range s.App.StakeibcKeeper.GetAllHostZone(s.Ctx) {
		if hostZone.Deprecated {
			s.Require().Equal(preUpgradeHalted[hostZone.ChainId], hostZone.Halted, "deprecated %s untouched", hostZone.ChainId)
			continue
		}
		s.Require().True(hostZone.Halted, "%s halted", hostZone.ChainId)
		halted[hostZone.ChainId] = true
	}
	// The channel map covers exactly the zones that were halted, no more and no fewer
	s.Require().Len(halted, len(stakeibctypes.HostToOsmosisTransferChannel))
	for chainId := range stakeibctypes.HostToOsmosisTransferChannel {
		s.Require().True(halted[chainId], "channel map entry %s is not a halted in-scope zone", chainId)
	}

	dym, err := s.App.StakedymKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err)
	s.Require().True(dym.Halted)
	for _, oracle := range s.App.ICAOracleKeeper.GetAllOracles(s.Ctx) {
		s.Require().False(oracle.Active, "oracle %s", oracle.ChainId)
	}
	s.Require().NotContains(s.App.ICAHostKeeper.GetParams(s.Ctx).AllowMessages, sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}))
	s.Require().Empty(s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx))
	s.Require().Empty(s.App.RatelimitKeeper.GetAllBlacklistedDenoms(s.Ctx))
	s.Require().Empty(s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx))
}
```

where `preUpgradeHalted` is a `map[string]bool` the test fills from the export before running the upgrade.

Run: `go test ./app/upgrades/v36/... -run TestMainnetExportTestSuite -v 2>&1 | tail -5`
Expected: `--- PASS`

- [ ] **Step 5: Changelog**

Under `## Unreleased` → `### On-Chain changes` in `CHANGELOG.md`, add (numbering continues the list):

```
N. v36: halt every in-scope stakeibc host zone and stakedym; deactivate the ICA oracles; remove `MsgClaimUndelegatedTokens` from the ICA host allow-list; remove every IBC rate limit, blacklisted denom and whitelisted address pair (wind-down spec §6)
N+1. v36: remove the `ClaimUndelegatedTokens`, `RebalanceValidators`, `ClearBalance` and `ResumeHostZone` tx handlers (stakeibc), and `ResumeHostZone` (staketia, stakedym); admin-gate `UpdateValidatorSharesExchRate` and `CalibrateDelegation`; lift the 5,000 base-unit calibration cap
N+2. v36: add the wind-down admin txs `MsgUndelegateFromValidators`, `MsgTransferFromIca`, `MsgTransferStaketiaClaimBalance` (protocol admin) and `MsgSweepTokensOffStride` (sweep operator; natives to Osmosis, whitelisted vouchers to their source), with the Osmosis vault, the host-to-Osmosis channel map and channel-5 as constants
```

- [ ] **Step 6: Full suite and commit**

Run: `go build ./... && go test ./... 2>&1 | grep -v "^ok\|no test files" | tail -20`
Expected: only the pre-existing `utils` `TestCreateModuleAccount` failure (fails on main too; see the upgrade 1 plan's dry-run notes)

```bash
git add app/upgrades/v36 x/stakeibc/types/wind_down.go CHANGELOG.md docs/superpowers/specs
git commit -m "test(upgrade): v36 removed-message guard and mainnet export suite; changelog; operator constants"
```

### Task 11: Localstride dry run (manual, documents the release checklist)

Not a subagent task: the user runs it (see the memory note on localstride upgrade-trigger binaries for the old-release-plus-uncached-context trick, and on POA validator changes). Record the outcome in this file under "Notes from a dry run".

- [ ] **Step 1: Local overrides, never committed.** Localstride's host chain id is not in `HostToOsmosisTransferChannel` and the vault is empty, so for the dry run only, patch `x/stakeibc/types/wind_down.go`: add the dockernet host chain id (`GAIA`) mapped to its transfer channel to the second dockernet chain, set `OsmosisVaultAddress` to an address on that second chain, set `SweepOperatorAddress` to a dockernet key, and set `StrideToOsmosisTransferChannelId` to Stride's channel to that chain. `git stash` the patch before committing anything.
- [ ] **Step 2: Upgrade.** Start localstride on the pre-upgrade release, submit and pass the `v36` proposal, swap the binary at the halt height, confirm the handler log lines: every zone halted, stakedym halted, oracles deactivated, allow-list without claim, rate limits removed.
- [ ] **Step 3: Undelegate one validator.** `strided tx stakeibc undelegate-from-validators GAIA <valoper>` from the admin key; watch the ICA ack; confirm the validator's `delegation_changes_in_progress` returns to 0 and its delegation drops by the amount, the host zone's `total_delegations` likewise, and no stTokens were burned (supply unchanged). Then `undelegate-from-validators GAIA` with no list for the rest.
- [ ] **Step 4: Transfer.** After the local unbonding period, `transfer-from-ica GAIA DELEGATION <amount>` and `transfer-from-ica GAIA WITHDRAWAL <amount>`; confirm the funds arrive at the vault address on the second chain as the expected denom.
- [ ] **Step 5: Sweep.** Fund two dockernet accounts with stATOM, ustrd and the GAIA ATOM voucher, build a batch file by hand, run `sweep-tokens-off-stride` for each of the three denoms from the sweep operator key (add the dockernet GAIA channel to `SweepUnwindChannels` in the local patch of Step 1); confirm the stATOM and ustrd packets land at the derived addresses on the second chain and the ATOM voucher lands as native ATOM on GAIA; confirm a batch containing a module account is rejected whole and a voucher over a non-whitelisted channel is rejected.
- [ ] **Step 6: Claim balance.** Fund the (dockernet) claim address with the celestia IBC denom and run `transfer-staketia-claim-balance`; this needs a dockernet celestia zone, so skip if the local network has none and note it.
- [ ] **Step 7: Record.** Append the findings under a "Notes from a dry run" heading at the top of this plan, as the upgrade 1 plan did.
