# Wind-Down PR 4: Constants and the ICA-Side Admin Txs — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the wind-down constants, the proto for all four admin messages, and the three ICA-side txs (`MsgUndelegateFromValidators`, `MsgTransferFromIca`, `MsgTransferStaketiaClaimBalance`) with types, keeper logic, CLI and tests, per spec §4, §7 and §11.

**Architecture:** One new constants file in `x/stakeibc/types` holds the two operator addresses (empty until the release gate; every use fails closed), the Stride→Osmosis channel, the host-side `chain_id → channel` map and the sweep unwind whitelist. All four messages go into `proto/stride/stakeibc/tx.proto` in a single proto-gen; the sweep handler is stubbed so the generated `MsgServer` interface compiles, and PR 5 fills it in. Each tx's keeper logic lives in its own `wind_down_*.go` file behind a thin delegate in `msg_server_wind_down.go`; the message builders are exported pure functions so tests assert the exact ICA contents without decoding packets. Nothing here reads or writes host-zone accounting; amounts and validator lists are ops inputs and every destination is a constant.

**Tech Stack:** Go 1.25, Cosmos SDK v0.54.3, ibc-go v11.2.0, gogoproto via `make proto-gen` (docker), testify suites (`KeeperTestSuite` in `x/stakeibc/keeper`, `types_test` and `cli_test` packages).

**Spec reference:** `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` §4 (operator addresses), §7 (the four admin txs), §11 (tests), §12 item 4, §13 "Admin txs (PRs 4 and 5)". Read those before starting.

> **Branching:** PR 1 branches off `wind-down-design-consolidation`. Each later PR branches
> off the previous PR's branch (PR 2 off PR 1, PR 3 off PR 2, and so on) and the PRs are
> implemented and merged strictly in order: 1, 2, 3, 4, 5, 6. Branch names:
> `wind-down-pr1-remove-handlers`, `wind-down-pr2-freeze-by-code`,
> `wind-down-pr3-upgrade-handler`, `wind-down-pr4-admin-txs`, `wind-down-pr5-sweep-tx`,
> `wind-down-pr6-release-gate`.
> The Go module path stays `github.com/Stride-Labs/stride/v34` in every PR; the bump to
> `/v35` is a manual step after all six land and is out of scope for every plan.

This plan is PR 4: branch `wind-down-pr4-admin-txs` off `wind-down-pr3-upgrade-handler`.

## Global Constraints

- Module path stays `github.com/Stride-Labs/stride/v34`. Never edit `go.mod`'s module line or any import path's version.
- `SweepOperatorAddress` and `OsmosisVaultAddress` ship as empty package `var`s. `MsgTransferFromIca` must error with `ErrOsmosisVaultNotConfigured` while the vault is empty. The release gate (PR 6) fills both.
- `OsmosisChainId = "osmosis-1"`, `OsmosisBech32Prefix = "osmo"`, `StrideToOsmosisTransferChannelId = "channel-5"`, `WindDownTransferTimeout = 24 * time.Hour`, `MaxSweepAddressesPerTx = 100`.
- `HostToOsmosisTransferChannel` is exactly: celestia `channel-2`, cosmoshub-4 `channel-141`, dydx-mainnet-1 `channel-3`, haqq_11235-1 `channel-2`, injective-1 `channel-8`, juno-1 `channel-0`, laozi-mainnet `channel-83`, phoenix-1 `channel-1`, sommelier-3 `channel-0`, ssc-1 `channel-1`, osmosis-1 `""` (empty selects an ICA bank send). No deprecated zone (comdex-1, evmos_9001-2, stargaze-1, umee-1) appears in it.
- `SweepUnwindChannels` is exactly: channel-0 `cosmos`, channel-162 `celestia`, channel-5 `osmo`, channel-24 `juno`, channel-150 `somm`, channel-213 `saga`, channel-160 `dydx`.
- Proto shapes and amino names are fixed (Task 2): `stakeibc/MsgUndelegateFromValidators`, `stakeibc/MsgTransferFromIca`, `stakeibc/MsgTransferStaketiaClaimBalance`, `stakeibc/MsgSweepTokensOffStride`. All four are added to both `RegisterCodec` and `RegisterImplementations`.
- Error codes registered by this PR: `ErrOsmosisVaultNotConfigured` 1566, `ErrNoOsmosisChannelForHostZone` 1567, `ErrHostZoneUnbondingPending` 1568. PR 5 uses 1569 and 1570.
- The three txs are admin-gated in `ValidateBasic` with `utils.ValidateAdminAddress(msg.Creator)`.
- `MsgUndelegateFromValidators` rejects, before submitting anything: a deprecated zone, a zone with any `HostZoneUnbonding` in `UNBONDING_QUEUE` or `UNBONDING_RETRY_QUEUE` with a positive `NativeTokenAmount`, a listed validator with `DelegationChangesInProgress > 0`, and an amount (delegation minus offset) that is not positive. It never changes `Validator.Delegation` or `HostZone.TotalDelegations`.
- The drain submits through `BatchSubmitUndelegateICAMessages` with `nil` epoch unbonding record ids and then adds the batch count to `PendingUndelegationInFlight` for the zone.
- ICA transfers are submitted with `SubmitICATxWithoutCallback` and an absolute timeout of block time + 24h; ICS-20 transfers from Stride go through `k.RecordsKeeper.TransferKeeper.Transfer` with the same timeout and an empty memo.
- `MsgTransferStaketiaClaimBalance` moves only `staketiatypes.CelestiaNativeTokenIBCDenom` from `staketiatypes.ClaimAddress` to the celestia zone's `DelegationIcaAddress` over the zone's `TransferChannelId`; zero amount means the whole balance; zero balance or amount above balance is rejected.
- The sweep handler in this PR returns `errorsmod.Wrap(sdkerrors.ErrNotSupported, "MsgSweepTokensOffStride is delivered in the next PR")`. Do not implement any sweep logic here.
- Commit messages end with `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`.

---

## File Structure

### New files

| File | Responsibility |
| --- | --- |
| `x/stakeibc/types/wind_down.go` | Operator address vars, Osmosis constants, `HostToOsmosisTransferChannel`, `SweepUnwindChannels` |
| `x/stakeibc/types/wind_down_test.go` | Map coverage, channel-id validity, address parsing (skipped while empty) |
| `x/stakeibc/types/message_undelegate_from_validators.go` | Constructor, `ValidateBasic` (admin, chain id, validator list) |
| `x/stakeibc/types/message_undelegate_from_validators_test.go` | Table test for `ValidateBasic` |
| `x/stakeibc/types/message_transfer_from_ica.go` | Constructor, `ValidateBasic` (admin, chain id, ICA type, amount) |
| `x/stakeibc/types/message_transfer_from_ica_test.go` | Table test |
| `x/stakeibc/types/message_transfer_staketia_claim_balance.go` | Constructor, `ValidateBasic` (admin, non-negative amount) |
| `x/stakeibc/types/message_transfer_staketia_claim_balance_test.go` | Table test |
| `x/stakeibc/keeper/msg_server_wind_down.go` | Thin delegates for the three txs; `ErrNotSupported` stub for the sweep |
| `x/stakeibc/keeper/wind_down_undelegate.go` | `UndelegateFromValidators`, `BuildUndelegateFromValidatorsMsgs`, queued-record guard |
| `x/stakeibc/keeper/wind_down_undelegate_test.go` | Drain tests (§11) |
| `x/stakeibc/keeper/wind_down_transfer_from_ica.go` | `TransferFromIca`, `BuildTransferFromIcaMsg` |
| `x/stakeibc/keeper/wind_down_transfer_from_ica_test.go` | Transfer tests (§11) |
| `x/stakeibc/keeper/wind_down_staketia_claim.go` | `TransferStaketiaClaimBalance`, `BuildStaketiaClaimTransferMsg` |
| `x/stakeibc/keeper/wind_down_staketia_claim_test.go` | Claim-address tests (§11) |
| `x/stakeibc/client/cli/tx_wind_down.go` | `undelegate-from-validators`, `transfer-from-ica`, `transfer-staketia-claim-balance` |
| `x/stakeibc/client/cli/tx_wind_down_test.go` | Argument-parsing error tests |

### Modified files

| File | Change |
| --- | --- |
| `proto/stride/stakeibc/tx.proto` | Import `ica_account.proto`; four rpcs; `ValidatorUndelegation` and the eight message/response types |
| `x/stakeibc/types/tx.pb.go` | Regenerated by `make proto-gen` (the only generated file to commit) |
| `x/stakeibc/types/codec.go` | Amino names and `RegisterImplementations` for the four messages |
| `x/stakeibc/types/errors.go` | Codes 1566–1568 |
| `x/stakeibc/types/events.go` | `EventTypeTransferFromIca`, `EventTypeTransferStaketiaClaimBalance`, `AttributeKeyIcaType`, `AttributeKeyChannel`, `AttributeKeyAmount` |
| `x/stakeibc/client/cli/tx.go` | Three `cmd.AddCommand(...)` lines |

---

## Task 1: Wind-down constants

**Files:**
- Create: `x/stakeibc/types/wind_down.go`
- Test: `x/stakeibc/types/wind_down_test.go`

**Interfaces:**
- Produces: `types.SweepOperatorAddress`, `types.OsmosisVaultAddress` (vars), `types.OsmosisChainId`, `types.OsmosisBech32Prefix`, `types.StrideToOsmosisTransferChannelId`, `types.WindDownTransferTimeout`, `types.MaxSweepAddressesPerTx`, `types.HostToOsmosisTransferChannel map[string]string`, `types.SweepUnwindChannels map[string]string`.
- Review: yes (every destination in the wind-down is one of these constants)

- [ ] **Step 1: Write the failing tests**

```go
// x/stakeibc/types/wind_down_test.go
package types_test

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Every non-deprecated stakeibc zone (spec §2) and nothing else.
var inScopeChainIds = []string{
	"celestia", "cosmoshub-4", "dydx-mainnet-1", "haqq_11235-1", "injective-1", "juno-1",
	"laozi-mainnet", "osmosis-1", "phoenix-1", "sommelier-3", "ssc-1",
}

var deprecatedChainIds = []string{"comdex-1", "evmos_9001-2", "stargaze-1", "umee-1"}

func TestHostToOsmosisTransferChannel(t *testing.T) {
	for _, chainId := range inScopeChainIds {
		channelId, found := types.HostToOsmosisTransferChannel[chainId]
		require.True(t, found, "%s must have a host-side channel to osmosis", chainId)
		if chainId == types.OsmosisChainId {
			require.Empty(t, channelId, "osmosis-1 maps to an empty channel (bank send form)")
			continue
		}
		require.True(t, channeltypes.IsValidChannelID(channelId), "%s channel %q is not a channel id", chainId, channelId)
	}
	for _, chainId := range deprecatedChainIds {
		_, found := types.HostToOsmosisTransferChannel[chainId]
		require.False(t, found, "deprecated zone %s must not be sweepable", chainId)
	}
	require.Len(t, types.HostToOsmosisTransferChannel, len(inScopeChainIds), "no extra zones in the map")
}

func TestSweepUnwindChannels(t *testing.T) {
	expected := map[string]string{
		"channel-0":   "cosmos",
		"channel-162": "celestia",
		"channel-5":   "osmo",
		"channel-24":  "juno",
		"channel-150": "somm",
		"channel-213": "saga",
		"channel-160": "dydx",
	}
	require.Equal(t, expected, types.SweepUnwindChannels)
	for channelId := range types.SweepUnwindChannels {
		require.True(t, channeltypes.IsValidChannelID(channelId))
	}
	require.Equal(t, types.OsmosisBech32Prefix, types.SweepUnwindChannels[types.StrideToOsmosisTransferChannelId],
		"the Stride->Osmosis channel unwinds to the osmo prefix")
}

// The two operator addresses are empty until the release gate; once filled they must parse
// with the expected prefix. The test passes trivially while they are empty so CI stays green.
func TestOperatorAddressesParse(t *testing.T) {
	if types.OsmosisVaultAddress == "" {
		t.Log("OsmosisVaultAddress not configured yet")
	} else {
		_, err := sdk.GetFromBech32(types.OsmosisVaultAddress, types.OsmosisBech32Prefix)
		require.NoError(t, err, "osmosis vault must be an osmo bech32 address")
	}
	if types.SweepOperatorAddress == "" {
		t.Log("SweepOperatorAddress not configured yet")
	} else {
		_, err := sdk.GetFromBech32(types.SweepOperatorAddress, "stride")
		require.NoError(t, err, "sweep operator must be a stride bech32 address")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./x/stakeibc/types/... -run 'TestHostToOsmosisTransferChannel|TestSweepUnwindChannels|TestOperatorAddressesParse' -v`
Expected: build failure, `undefined: types.HostToOsmosisTransferChannel`

- [ ] **Step 3: Write the constants file**

```go
// x/stakeibc/types/wind_down.go
package types

import "time"

// Wind-down constants (spec §4, §7). The two addresses are vars so tests can set them; they
// ship empty in this PR and are filled by the release gate, and every use fails closed while
// they are empty (the transfer tx errors, the sweep gate rejects every signer).
var (
	SweepOperatorAddress = "" // stride1..., the only signer of MsgSweepTokensOffStride
	OsmosisVaultAddress  = "" // osmo1..., receiver of every MsgTransferFromIca
)

const (
	OsmosisChainId                   = "osmosis-1"
	OsmosisBech32Prefix              = "osmo"
	StrideToOsmosisTransferChannelId = "channel-5"
	WindDownTransferTimeout          = 24 * time.Hour
	MaxSweepAddressesPerTx           = 100
)

// Host-side transfer channel to osmosis-1 per in-scope zone (chain registry 2026-09-24, re-verified
// against each host before the proposal). osmosis-1 maps to "" which selects an ICA bank send.
var HostToOsmosisTransferChannel = map[string]string{
	"celestia":       "channel-2",
	"cosmoshub-4":    "channel-141",
	"dydx-mainnet-1": "channel-3",
	"haqq_11235-1":   "channel-2",
	"injective-1":    "channel-8",
	"juno-1":         "channel-0",
	"laozi-mainnet":  "channel-83",
	"phoenix-1":      "channel-1",
	"sommelier-3":    "channel-0",
	"ssc-1":          "channel-1",
	"osmosis-1":      "",
}

// Stride transfer channels a voucher may be unwound over, with the counterparty's bech32 prefix:
// exactly the chains whose wallets derive the same address bytes as Stride (spec §3, §7).
var SweepUnwindChannels = map[string]string{
	"channel-0":   "cosmos",
	"channel-162": "celestia",
	"channel-5":   "osmo",
	"channel-24":  "juno",
	"channel-150": "somm",
	"channel-213": "saga",
	"channel-160": "dydx",
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./x/stakeibc/types/... -run 'TestHostToOsmosisTransferChannel|TestSweepUnwindChannels|TestOperatorAddressesParse' -v`
Expected: `--- PASS` for all three (the address test logs "not configured yet" twice).

- [ ] **Step 5: Commit**

```bash
git add x/stakeibc/types/wind_down.go x/stakeibc/types/wind_down_test.go
git commit -m "feat(stakeibc): wind-down constants (operator addresses, osmosis channel maps)

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Task 2: Proto, codec, errors, events and the msg-server scaffold

**Files:**
- Modify: `proto/stride/stakeibc/tx.proto`
- Modify: `x/stakeibc/types/codec.go`
- Modify: `x/stakeibc/types/errors.go`
- Modify: `x/stakeibc/types/events.go`
- Create: `x/stakeibc/keeper/msg_server_wind_down.go`
- Generated: `x/stakeibc/types/tx.pb.go`

**Interfaces:**
- Produces: `types.MsgUndelegateFromValidators{Creator, ChainId, Validators []ValidatorUndelegation}`, `types.ValidatorUndelegation{Address string, Offset sdkmath.Int}`, `types.MsgUndelegateFromValidatorsResponse{NumBatchesSubmitted uint64}`, `types.MsgTransferFromIca{Creator, ChainId, IcaType ICAAccountType, Amount sdk.Coin}`, `types.MsgTransferFromIcaResponse{}`, `types.MsgTransferStaketiaClaimBalance{Creator, Amount sdkmath.Int}`, `types.MsgTransferStaketiaClaimBalanceResponse{Transferred sdk.Coin}`, `types.MsgSweepTokensOffStride{Creator, Denoms, Addresses []string}`, `types.MsgSweepTokensOffStrideResponse{NumTransfers, NumSkipped uint64}`; the four `MsgServer` methods; errors 1566–1568; event constants.
- Review: yes (the proto is the on-chain interface for all four txs and PR 5 depends on it)

- [ ] **Step 1: Edit the proto**

Add the import after `import "stride/stakeibc/validator.proto";`:

```proto
import "stride/stakeibc/ica_account.proto";
```

Add the four rpcs at the end of `service Msg` (after `DeprecateHostZone`):

```proto
  rpc UndelegateFromValidators(MsgUndelegateFromValidators)
      returns (MsgUndelegateFromValidatorsResponse);
  rpc TransferFromIca(MsgTransferFromIca) returns (MsgTransferFromIcaResponse);
  rpc TransferStaketiaClaimBalance(MsgTransferStaketiaClaimBalance)
      returns (MsgTransferStaketiaClaimBalanceResponse);
  rpc SweepTokensOffStride(MsgSweepTokensOffStride)
      returns (MsgSweepTokensOffStrideResponse);
```

Append the messages at the end of the file:

```proto
// ---------------------------------------------------------------------------
// Wind-down admin txs (spec §7)
// ---------------------------------------------------------------------------

// One validator to drain: the amount undelegated is the validator's recorded delegation minus
// offset (default zero), so a full drain is the common case and offset is the lever for a
// validator that drifted since the last refresh.
message ValidatorUndelegation {
  string address = 1;
  string offset = 2 [
    (gogoproto.customtype) = "cosmossdk.io/math.Int",
    (gogoproto.nullable) = false
  ];
}

// Admin drain of a host zone's delegations. An empty validators list means every validator
// with a positive recorded delegation.
message MsgUndelegateFromValidators {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name) = "stakeibc/MsgUndelegateFromValidators";

  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  string chain_id = 2;
  repeated ValidatorUndelegation validators = 3 [ (gogoproto.nullable) = false ];
}
message MsgUndelegateFromValidatorsResponse { uint64 num_batches_submitted = 1; }

// Admin ICA transfer of one of the four funded ICAs' balance (delegation, withdrawal, fee,
// redemption) to the Osmosis vault over the host's mapped channel to osmosis-1. The receiver
// and channel are hard-coded constants, not tx inputs.
message MsgTransferFromIca {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name) = "stakeibc/MsgTransferFromIca";

  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  string chain_id = 2;
  ICAAccountType ica_type = 3;
  cosmos.base.v1beta1.Coin amount = 4 [ (gogoproto.nullable) = false ];
}
message MsgTransferFromIcaResponse {}

// Admin transfer of the staketia claim address's TIA vouchers to the stakeibc celestia
// delegation ICA. amount is in utia; zero means the whole balance.
message MsgTransferStaketiaClaimBalance {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name) = "stakeibc/MsgTransferStaketiaClaimBalance";

  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  string amount = 2 [
    (gogoproto.customtype) = "cosmossdk.io/math.Int",
    (gogoproto.nullable) = false
  ];
}
message MsgTransferStaketiaClaimBalanceResponse {
  cosmos.base.v1beta1.Coin transferred = 1 [ (gogoproto.nullable) = false ];
}

// Batched sweep of holders' balances off Stride (implemented in the next PR).
message MsgSweepTokensOffStride {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name) = "stakeibc/MsgSweepTokensOffStride";

  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
  repeated string denoms = 2;
  repeated string addresses = 3;
}
message MsgSweepTokensOffStrideResponse {
  uint64 num_transfers = 1;
  uint64 num_skipped = 2;
}
```

- [ ] **Step 2: Regenerate and confirm only stakeibc's `tx.pb.go` changed**

Run: `make proto-gen && git status --short`
Expected: `M proto/stride/stakeibc/tx.proto` and `M x/stakeibc/types/tx.pb.go`. If any other `*.pb.go` shows as modified (descriptor churn), revert it with `git checkout -- <path>`.

Run: `go build ./... 2>&1 | head`
Expected: an error like `*msgServer does not implement types.MsgServer (missing method SweepTokensOffStride)` in `x/stakeibc/module.go` — the interface grew; Step 5 satisfies it.

- [ ] **Step 3: Register the messages in the codec**

In `x/stakeibc/types/codec.go`, append to `RegisterCodec` after the `MsgDeprecateHostZone` line:

```go
	legacy.RegisterAminoMsg(cdc, &MsgUndelegateFromValidators{}, "stakeibc/MsgUndelegateFromValidators")
	legacy.RegisterAminoMsg(cdc, &MsgTransferFromIca{}, "stakeibc/MsgTransferFromIca")
	legacy.RegisterAminoMsg(cdc, &MsgTransferStaketiaClaimBalance{}, "stakeibc/MsgTransferStaketiaClaimBalance")
	legacy.RegisterAminoMsg(cdc, &MsgSweepTokensOffStride{}, "stakeibc/MsgSweepTokensOffStride")
```

and to the `(*sdk.Msg)(nil)` `RegisterImplementations` list after `&MsgDeprecateHostZone{},`:

```go
		&MsgUndelegateFromValidators{},
		&MsgTransferFromIca{},
		&MsgTransferStaketiaClaimBalance{},
		&MsgSweepTokensOffStride{},
```

- [ ] **Step 4: Register errors and events**

Append inside the `var (...)` block of `x/stakeibc/types/errors.go`, after `ErrRedemptionsDisabled`:

```go
	ErrOsmosisVaultNotConfigured           = errorsmod.Register(ModuleName, 1566, "osmosis vault address is not configured")
	ErrNoOsmosisChannelForHostZone         = errorsmod.Register(ModuleName, 1567, "host zone has no transfer channel to osmosis")
	ErrHostZoneUnbondingPending            = errorsmod.Register(ModuleName, 1568, "host zone has an unbonding record queued or retrying")
```

Append to the event `const` block of `x/stakeibc/types/events.go` (next to `EventTypeRedemptionSweep`):

```go
	EventTypeTransferFromIca               = "transfer_from_ica"
	EventTypeTransferStaketiaClaimBalance  = "transfer_staketia_claim_balance"
```

and to the attribute block:

```go
	AttributeKeyIcaType = "ica_type"
	AttributeKeyChannel = "channel"
	AttributeKeyAmount  = "amount"
```

- [ ] **Step 5: Write the msg-server scaffold**

Every method is a stub for now so the package compiles; Tasks 3–5 replace the first three bodies, PR 5 replaces the fourth.

```go
// x/stakeibc/keeper/msg_server_wind_down.go
package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Wind-down admin txs (spec §7). Each handler is a thin delegate to the keeper function in the
// matching wind_down_*.go file. The sweep is delivered by the next PR.

func (k msgServer) UndelegateFromValidators(goCtx context.Context, msg *types.MsgUndelegateFromValidators) (*types.MsgUndelegateFromValidatorsResponse, error) {
	return nil, errorsmod.Wrap(sdkerrors.ErrNotSupported, "MsgUndelegateFromValidators is wired in a later task of this PR")
}

func (k msgServer) TransferFromIca(goCtx context.Context, msg *types.MsgTransferFromIca) (*types.MsgTransferFromIcaResponse, error) {
	return nil, errorsmod.Wrap(sdkerrors.ErrNotSupported, "MsgTransferFromIca is wired in a later task of this PR")
}

func (k msgServer) TransferStaketiaClaimBalance(goCtx context.Context, msg *types.MsgTransferStaketiaClaimBalance) (*types.MsgTransferStaketiaClaimBalanceResponse, error) {
	return nil, errorsmod.Wrap(sdkerrors.ErrNotSupported, "MsgTransferStaketiaClaimBalance is wired in a later task of this PR")
}

func (k msgServer) SweepTokensOffStride(goCtx context.Context, msg *types.MsgSweepTokensOffStride) (*types.MsgSweepTokensOffStrideResponse, error) {
	return nil, errorsmod.Wrap(sdkerrors.ErrNotSupported, "MsgSweepTokensOffStride is delivered in the next PR")
}
```

- [ ] **Step 6: Build and run the existing stakeibc suites**

Run: `go build ./... && go test ./x/stakeibc/... 2>&1 | tail -5`
Expected: build OK; `ok  github.com/Stride-Labs/stride/v34/x/stakeibc/...` for keeper, types and cli.

- [ ] **Step 7: Commit**

```bash
git add proto/stride/stakeibc/tx.proto x/stakeibc/types/tx.pb.go x/stakeibc/types/codec.go x/stakeibc/types/errors.go x/stakeibc/types/events.go x/stakeibc/keeper/msg_server_wind_down.go
git commit -m "feat(stakeibc): proto and registrations for the four wind-down admin txs

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Parallel-safe tasks

Tasks 3, 4 and 5 depend only on Tasks 1–2. Each replaces one stub body in
`msg_server_wind_down.go` and adds one `AddCommand` line in `cli/tx.go`; that is textual
overlap on a distinct function/line each, which the worktree merge handles. None consumes
another's interface.

---

## Task 3: `MsgUndelegateFromValidators`

**Files:**
- Create: `x/stakeibc/types/message_undelegate_from_validators.go`
- Test: `x/stakeibc/types/message_undelegate_from_validators_test.go`
- Create: `x/stakeibc/keeper/wind_down_undelegate.go`
- Test: `x/stakeibc/keeper/wind_down_undelegate_test.go`
- Modify: `x/stakeibc/keeper/msg_server_wind_down.go` (replace the `UndelegateFromValidators` stub)
- Modify: `x/stakeibc/client/cli/tx_wind_down.go` (create), `x/stakeibc/client/cli/tx.go` (one line)
- Test: `x/stakeibc/client/cli/tx_wind_down_test.go`

**Interfaces:**
- Consumes: `types.MsgUndelegateFromValidators`, `types.ValidatorUndelegation` (Task 2); `BatchSubmitUndelegateICAMessages`, `applySharesRoundingSafety`, `ValidatorUnbondCapacity`, `GetValidatorFromAddress`, `GetPendingUndelegationInFlight`/`SetPendingUndelegationInFlight`, `DefaultMaxMessagesPerIcaTx`, `CalculateTotalUnbondedInBatch`, `EmitUndelegationEvent` (existing keeper).
- Produces: `func (k Keeper) UndelegateFromValidators(ctx sdk.Context, msg *types.MsgUndelegateFromValidators) (numBatches uint64, err error)`, `func (k Keeper) BuildUndelegateFromValidatorsMsgs(hostZone types.HostZone, requested []types.ValidatorUndelegation) (msgs []proto.Message, splits []*types.SplitUndelegation, err error)`, `types.NewMsgUndelegateFromValidators(creator, chainId string, validators []types.ValidatorUndelegation)`, CLI `undelegate-from-validators`.
- Depends on: Tasks 1–2
- Review: yes (submits undelegations for the entire protocol stake)

- [ ] **Step 1: Write the failing `ValidateBasic` test**

```go
// x/stakeibc/types/message_undelegate_from_validators_test.go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgUndelegateFromValidators_ValidateBasic(t *testing.T) {
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	valA := types.ValidatorUndelegation{Address: "cosmosvaloper1aaa", Offset: sdkmath.ZeroInt()}
	valB := types.ValidatorUndelegation{Address: "cosmosvaloper1bbb", Offset: sdkmath.NewInt(5)}

	tests := []struct {
		name string
		msg  types.MsgUndelegateFromValidators
		err  error
	}{
		{
			name: "valid empty list (drain everything)",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4"},
		},
		{
			name: "valid explicit list with offsets",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{valA, valB}},
		},
		{
			name: "valid nil offset (treated as zero)",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{{Address: "cosmosvaloper1ccc"}}},
		},
		{
			name: "invalid creator",
			msg:  types.MsgUndelegateFromValidators{Creator: invalidAddress, ChainId: "cosmoshub-4"},
			err:  sdkerrors.ErrInvalidAddress,
		},
		{
			name: "not admin",
			msg:  types.MsgUndelegateFromValidators{Creator: validNotAdminAddress, ChainId: "cosmoshub-4"},
			err:  sdkerrors.ErrInvalidAddress,
		},
		{
			name: "missing chain id",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress},
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty validator address",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{{Address: ""}}},
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "address is not a valoper",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{{Address: "cosmos1notavaloper"}}},
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "duplicate validator",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{valA, valA}},
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "negative offset",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{{Address: "cosmosvaloper1aaa", Offset: sdkmath.NewInt(-1)}}},
			err:  sdkerrors.ErrInvalidRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.msg.Creator, tt.msg.GetSigners()[0].String())
		})
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./x/stakeibc/types/... -run TestMsgUndelegateFromValidators_ValidateBasic -v`
Expected: build failure, `tt.msg.GetSigners undefined` / `ValidateBasic undefined`.

- [ ] **Step 3: Write the message file**

```go
// x/stakeibc/types/message_undelegate_from_validators.go
package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
)

const TypeMsgUndelegateFromValidators = "undelegate_from_validators"

var _ sdk.Msg = &MsgUndelegateFromValidators{}

func NewMsgUndelegateFromValidators(creator, chainId string, validators []ValidatorUndelegation) *MsgUndelegateFromValidators {
	return &MsgUndelegateFromValidators{
		Creator:    creator,
		ChainId:    chainId,
		Validators: validators,
	}
}

func (msg *MsgUndelegateFromValidators) Route() string {
	return RouterKey
}

func (msg *MsgUndelegateFromValidators) Type() string {
	return TypeMsgUndelegateFromValidators
}

func (msg *MsgUndelegateFromValidators) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUndelegateFromValidators) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if len(msg.ChainId) == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "chain-id is required")
	}

	// An empty list is the full drain; a non-empty list must name distinct valopers with
	// non-negative offsets (a nil offset is the zero value from JSON and means zero)
	seen := map[string]bool{}
	for _, validator := range msg.Validators {
		if len(validator.Address) == 0 {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator address is required")
		}
		if !strings.Contains(validator.Address, "valoper") {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator address %s must contain 'valoper'", validator.Address)
		}
		if seen[validator.Address] {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s listed twice", validator.Address)
		}
		seen[validator.Address] = true
		if !validator.Offset.IsNil() && validator.Offset.IsNegative() {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "offset for %s must not be negative", validator.Address)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run the types test to verify it passes**

Run: `go test ./x/stakeibc/types/... -run TestMsgUndelegateFromValidators_ValidateBasic -v`
Expected: PASS for all ten cases.

- [ ] **Step 5: Write the failing keeper tests**

```go
// x/stakeibc/keeper/wind_down_undelegate_test.go
package keeper_test

import (
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v34/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v34/x/records/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

type undelegateFromValidatorsTestCase struct {
	hostZone            types.HostZone
	delegationPortID    string
	delegationChannelID string
}

// Four validators: two unslashed, one slashed (rate < 1, so a full drain gets the rounding
// buffer), one with no delegation (skipped by the empty-list drain). Batch size 2 so three
// messages take two ICAs.
func (s *KeeperTestSuite) SetupUndelegateFromValidators() undelegateFromValidatorsTestCase {
	delegationAccountOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	validators := []*types.Validator{
		{Address: "val1", Delegation: sdkmath.NewInt(1000), SharesToTokensRate: sdkmath.LegacyOneDec()},
		{Address: "val2", Delegation: sdkmath.NewInt(2000), SharesToTokensRate: sdkmath.LegacyOneDec()},
		{Address: "val3", Delegation: sdkmath.NewInt(3000), SharesToTokensRate: sdkmath.LegacyMustNewDecFromStr("0.9")},
		{Address: "val4", Delegation: sdkmath.ZeroInt(), SharesToTokensRate: sdkmath.LegacyOneDec()},
	}
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		MaxMessagesPerIcaTx:  2,
		TotalDelegations:     sdkmath.NewInt(6000),
		Validators:           validators,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// The undelegate ICA timeout is read from the day epoch tracker
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, types.EpochTracker{
		EpochIdentifier:    epochstypes.DAY_EPOCH,
		Duration:           10_000_000_000,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000),
	})

	return undelegateFromValidatorsTestCase{
		hostZone:            hostZone,
		delegationPortID:    delegationPortID,
		delegationChannelID: delegationChannelID,
	}
}

func (s *KeeperTestSuite) setHostZoneUnbondingStatus(status recordtypes.HostZoneUnbonding_Status, amount int64) {
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: 1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{{
			HostZoneId:        HostChainId,
			Status:            status,
			NativeTokenAmount: sdkmath.NewInt(amount),
			StTokenAmount:     sdkmath.NewInt(amount),
		}},
	})
}

// The recorded delegations must never move: the callback does that on ack
func (s *KeeperTestSuite) checkNoAccountingMutation(tc undelegateFromValidatorsTestCase) {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found)
	s.Require().Equal(tc.hostZone.TotalDelegations, hostZone.TotalDelegations, "total delegations unchanged")
	for i, validator := range hostZone.Validators {
		s.Require().Equal(tc.hostZone.Validators[i].Delegation, validator.Delegation, "%s delegation unchanged", validator.Address)
	}
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_EmptyListDrainsEveryFundedValidator() {
	tc := s.SetupUndelegateFromValidators()
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, nil)
	numBatches, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), numBatches, "3 messages in batches of 2")

	endSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)
	s.Require().Equal(startSequence+2, endSequence, "two ICAs submitted")

	// val1..3 are flagged, val4 (no delegation) is untouched
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().Equal(uint64(1), hostZone.Validators[0].DelegationChangesInProgress)
	s.Require().Equal(uint64(1), hostZone.Validators[1].DelegationChangesInProgress)
	s.Require().Equal(uint64(1), hostZone.Validators[2].DelegationChangesInProgress)
	s.Require().Equal(uint64(0), hostZone.Validators[3].DelegationChangesInProgress)

	// Both batches are registered in flight so the record-less callback's decrement is clean
	s.Require().Equal(uint64(2), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId))

	// The callbacks carry no epoch unbonding record ids and the exact splits
	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbackData, 2)
	splitsByValidator := map[string]sdkmath.Int{}
	for _, data := range callbackData {
		var callback types.UndelegateCallback
		s.Require().NoError(proto.Unmarshal(data.CallbackArgs, &callback))
		s.Require().Empty(callback.EpochUnbondingRecordIds, "record-less undelegation")
		s.Require().Equal(HostChainId, callback.HostZoneId)
		for _, split := range callback.SplitUndelegations {
			splitsByValidator[split.Validator] = split.NativeTokenAmount
		}
	}
	s.Require().Equal(sdkmath.NewInt(1000), splitsByValidator["val1"])
	s.Require().Equal(sdkmath.NewInt(2000), splitsByValidator["val2"])
	s.Require().Equal(sdkmath.NewInt(2999), splitsByValidator["val3"], "slashed validator drained with a 1 base unit buffer")
	s.Require().NotContains(splitsByValidator, "val4")

	s.checkNoAccountingMutation(tc)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyTotalUnbondAmount, "5999")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_ExplicitListWithOffset() {
	tc := s.SetupUndelegateFromValidators()
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{
		{Address: "val2", Offset: sdkmath.NewInt(500)},
	})
	numBatches, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), numBatches)
	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID))

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().Equal(uint64(0), hostZone.Validators[0].DelegationChangesInProgress, "val1 not listed")
	s.Require().Equal(uint64(1), hostZone.Validators[1].DelegationChangesInProgress, "val2 listed")

	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbackData, 1)
	var callback types.UndelegateCallback
	s.Require().NoError(proto.Unmarshal(callbackData[0].CallbackArgs, &callback))
	s.Require().Len(callback.SplitUndelegations, 1)
	s.Require().Equal(sdkmath.NewInt(1500), callback.SplitUndelegations[0].NativeTokenAmount, "2000 - 500 offset")
	s.Require().Equal(uint64(1), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId))
	s.checkNoAccountingMutation(tc)
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_InFlightAccumulates() {
	s.SetupUndelegateFromValidators()
	s.App.StakeibcKeeper.SetPendingUndelegationInFlight(s.Ctx, HostChainId, 3)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val1"}})
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "added to, not overwritten")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_HaltedZoneAccepted() {
	tc := s.SetupUndelegateFromValidators()
	tc.hostZone.Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, nil)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err, "the drain does not depend on the zone being active")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsFlaggedValidator() {
	tc := s.SetupUndelegateFromValidators()
	tc.hostZone.Validators[1].DelegationChangesInProgress = 1
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val1"}, {Address: "val2"}})
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrInvalidDelegationsInProgress)
	s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID), "nothing submitted")
	s.Require().Equal(uint64(0), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId))
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsDeprecatedZone() {
	tc := s.SetupUndelegateFromValidators()
	tc.hostZone.Deprecated = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, nil)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorContains(err, "deprecated")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsUnknownZone() {
	s.SetupUndelegateFromValidators()
	msg := types.NewMsgUndelegateFromValidators("admin", "unknown-1", nil)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrHostZoneNotFound)
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsUnknownValidator() {
	s.SetupUndelegateFromValidators()
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val9"}})
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrValidatorNotFound)
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsNonPositiveAmount() {
	s.SetupUndelegateFromValidators()

	// offset equal to the delegation
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val1", Offset: sdkmath.NewInt(1000)}})
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorContains(err, "not positive")

	// validator with no delegation, explicitly listed
	msg = types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val4"}})
	_, err = s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorContains(err, "not positive")
}

// STRIDE-07 guard: a record still queued or retrying could never be submitted once the zone
// is drained, so the drain refuses until the day epoch has moved it to IN_PROGRESS
func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsQueuedRecords() {
	tc := s.SetupUndelegateFromValidators()
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, nil)

	for _, status := range []recordtypes.HostZoneUnbonding_Status{
		recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		recordtypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE,
	} {
		s.setHostZoneUnbondingStatus(status, 100)
		startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)
		_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
		s.Require().ErrorIs(err, types.ErrHostZoneUnbondingPending, "status %s", status)
		s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID), "nothing submitted")
	}
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_AcceptsRecordsPastTheQueue() {
	s.SetupUndelegateFromValidators()
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val1"}})

	for _, status := range []recordtypes.HostZoneUnbonding_Status{
		recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
		recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		recordtypes.HostZoneUnbonding_CLAIMABLE,
	} {
		s.setHostZoneUnbondingStatus(status, 100)
		// clear the flag the previous iteration set so the validator is eligible again
		hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
		hostZone.Validators[0].DelegationChangesInProgress = 0
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

		_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
		s.Require().NoError(err, "status %s must not block the drain", status)
	}

	// A queued record with a zero amount is not a real record
	s.setHostZoneUnbondingStatus(recordtypes.HostZoneUnbonding_UNBONDING_QUEUE, 0)
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	hostZone.Validators[0].DelegationChangesInProgress = 0
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err, "zero-amount queued record does not block")
}

// Pure builder: message shape and rounding safety, including a large slashed delegation
// where the 1e17 divisor yields a non-trivial buffer
func (s *KeeperTestSuite) TestBuildUndelegateFromValidatorsMsgs() {
	big := sdkmath.NewInt(5).Mul(sdkmath.NewInt(1e17)) // 5e17
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		Validators: []*types.Validator{
			{Address: "val1", Delegation: sdkmath.NewInt(1000), SharesToTokensRate: sdkmath.LegacyOneDec()},
			{Address: "val2", Delegation: big, SharesToTokensRate: sdkmath.LegacyMustNewDecFromStr("0.95")},
		},
	}

	msgs, splits, err := s.App.StakeibcKeeper.BuildUndelegateFromValidatorsMsgs(hostZone, nil)
	s.Require().NoError(err)
	s.Require().Len(msgs, 2)
	s.Require().Len(splits, 2)

	undelegate1 := msgs[0].(*stakingtypes.MsgUndelegate)
	s.Require().Equal("cosmos_DELEGATION", undelegate1.DelegatorAddress)
	s.Require().Equal("val1", undelegate1.ValidatorAddress)
	s.Require().Equal(Atom, undelegate1.Amount.Denom)
	s.Require().Equal(sdkmath.NewInt(1000), undelegate1.Amount.Amount, "unslashed full drain has no buffer")

	undelegate2 := msgs[1].(*stakingtypes.MsgUndelegate)
	s.Require().Equal(big.Sub(sdkmath.NewInt(5)), undelegate2.Amount.Amount, "5e17 / 1e17 = 5 base unit buffer on a slashed full drain")
	s.Require().Equal(undelegate2.Amount.Amount, splits[1].NativeTokenAmount)

	// A partial drain of the slashed validator gets no buffer
	msgs, _, err = s.App.StakeibcKeeper.BuildUndelegateFromValidatorsMsgs(hostZone, []types.ValidatorUndelegation{{Address: "val2", Offset: sdkmath.NewInt(1)}})
	s.Require().NoError(err)
	s.Require().Equal(big.Sub(sdkmath.NewInt(1)), msgs[0].(*stakingtypes.MsgUndelegate).Amount.Amount)

	// Nothing to drain
	_, _, err = s.App.StakeibcKeeper.BuildUndelegateFromValidatorsMsgs(types.HostZone{ChainId: HostChainId}, nil)
	s.Require().ErrorContains(err, "no validator")
}

func (s *KeeperTestSuite) TestMsgServer_UndelegateFromValidators() {
	tc := s.SetupUndelegateFromValidators()
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	resp, err := s.GetMsgServer().UndelegateFromValidators(s.Ctx, types.NewMsgUndelegateFromValidators("admin", HostChainId, nil))
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), resp.NumBatchesSubmitted)
	s.Require().Equal(startSequence+2, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID))
}
```

- [ ] **Step 6: Run them to verify they fail**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestUndelegateFromValidators|TestKeeperTestSuite/TestBuildUndelegateFromValidatorsMsgs|TestKeeperTestSuite/TestMsgServer_UndelegateFromValidators' -v 2>&1 | tail -5`
Expected: build failure, `s.App.StakeibcKeeper.UndelegateFromValidators undefined`.

- [ ] **Step 7: Write the keeper file**

```go
// x/stakeibc/keeper/wind_down_undelegate.go
package keeper

import (
	"github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v34/utils"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// UndelegateFromValidators is the wind-down drain (spec §7): it submits one MsgUndelegate per
// listed validator (or per funded validator when the list is empty) for that validator's
// recorded delegation minus its offset, through the same record-less batch path the v34
// pending-undelegation pipeline uses. The callback decrements the recorded delegations on
// ack; nothing is changed here.
//
// It refuses, before anything is submitted: a deprecated zone, a zone that still has an
// unbonding record queued or retrying (the record-driven path could never submit it on a
// drained zone, STRIDE-07), and a listed validator with a delegation change in progress.
func (k Keeper) UndelegateFromValidators(ctx sdk.Context, msg *types.MsgUndelegateFromValidators) (numBatches uint64, err error) {
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return 0, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", msg.ChainId)
	}
	if hostZone.Deprecated {
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "host zone %s is deprecated", msg.ChainId)
	}
	if err := k.checkNoQueuedUnbondings(ctx, msg.ChainId); err != nil {
		return 0, err
	}

	msgs, splits, err := k.BuildUndelegateFromValidatorsMsgs(hostZone, msg.Validators)
	if err != nil {
		return 0, err
	}

	// Same batch size as the epochly unbonding; a zone that never set it gets the default
	batchSize := int(utils.UintToInt(hostZone.MaxMessagesPerIcaTx))
	if batchSize == 0 {
		batchSize = int(DefaultMaxMessagesPerIcaTx)
	}
	numBatches, err = k.BatchSubmitUndelegateICAMessages(ctx, hostZone, nil, msgs, splits, batchSize)
	if err != nil {
		return 0, err
	}

	// Register the batches in flight so the record-less callback's decrement finds a counter
	inFlight := k.GetPendingUndelegationInFlight(ctx, msg.ChainId)
	k.SetPendingUndelegationInFlight(ctx, msg.ChainId, inFlight+numBatches)

	totalUnbondAmount := k.CalculateTotalUnbondedInBatch(splits)
	k.Logger(ctx).Info(utils.LogWithHostZone(msg.ChainId,
		"Wind-down undelegation of %v%s across %d validator(s) in %d batch(es)",
		totalUnbondAmount, hostZone.HostDenom, len(msgs), numBatches))
	EmitUndelegationEvent(ctx, hostZone, totalUnbondAmount)

	return numBatches, nil
}

// checkNoQueuedUnbondings rejects the drain while any unbonding record for the zone is still
// waiting for the day epoch (queued or retrying) with a real amount
func (k Keeper) checkNoQueuedUnbondings(ctx sdk.Context, chainId string) error {
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, chainId)
		if !found {
			continue
		}
		queued := hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_QUEUE ||
			hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE
		hasAmount := !hostZoneUnbonding.NativeTokenAmount.IsNil() && hostZoneUnbonding.NativeTokenAmount.IsPositive()
		if queued && hasAmount {
			return types.ErrHostZoneUnbondingPending.Wrapf(
				"epoch %d record for %s is %s with %v; wait for the day epoch to submit it before draining",
				epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding.Status, hostZoneUnbonding.NativeTokenAmount)
		}
	}
	return nil
}

// BuildUndelegateFromValidatorsMsgs builds one MsgUndelegate and one SplitUndelegation per
// target validator. An empty request means every validator with a positive recorded
// delegation. Each amount is the recorded delegation minus the offset, passed through the same
// rounding safety the epochly unbonding applies to a full drain of a slashed validator.
func (k Keeper) BuildUndelegateFromValidatorsMsgs(
	hostZone types.HostZone,
	requested []types.ValidatorUndelegation,
) (msgs []proto.Message, splits []*types.SplitUndelegation, err error) {
	targets := requested
	if len(targets) == 0 {
		for _, validator := range hostZone.Validators {
			if !validator.Delegation.IsNil() && validator.Delegation.IsPositive() {
				targets = append(targets, types.ValidatorUndelegation{Address: validator.Address, Offset: sdkmath.ZeroInt()})
			}
		}
	}
	if len(targets) == 0 {
		return nil, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "no validator on %s has a delegation to undelegate", hostZone.ChainId)
	}

	for _, target := range targets {
		validator, _, found := GetValidatorFromAddress(hostZone.Validators, target.Address)
		if !found {
			return nil, nil, types.ErrValidatorNotFound.Wrapf("validator %s not found on %s", target.Address, hostZone.ChainId)
		}
		if validator.DelegationChangesInProgress > 0 {
			return nil, nil, types.ErrInvalidDelegationsInProgress.Wrapf(
				"validator %s has %d delegation change(s) in progress", target.Address, validator.DelegationChangesInProgress)
		}

		delegation := validator.Delegation
		if delegation.IsNil() {
			delegation = sdkmath.ZeroInt()
		}
		offset := target.Offset
		if offset.IsNil() {
			offset = sdkmath.ZeroInt()
		}
		amount := delegation.Sub(offset)
		if !amount.IsPositive() {
			return nil, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
				"validator %s: delegation %v minus offset %v is not positive", target.Address, delegation, offset)
		}

		capacity := ValidatorUnbondCapacity{ValidatorAddress: validator.Address, CurrentDelegation: delegation}
		amount = k.applySharesRoundingSafety(hostZone, capacity, amount)

		msgs = append(msgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: hostZone.DelegationIcaAddress,
			ValidatorAddress: validator.Address,
			Amount:           sdk.NewCoin(hostZone.HostDenom, amount),
		})
		splits = append(splits, &types.SplitUndelegation{
			Validator:         validator.Address,
			NativeTokenAmount: amount,
		})
	}
	return msgs, splits, nil
}
```

- [ ] **Step 8: Wire the msg-server delegate**

Replace the `UndelegateFromValidators` stub in `x/stakeibc/keeper/msg_server_wind_down.go` with:

```go
func (k msgServer) UndelegateFromValidators(goCtx context.Context, msg *types.MsgUndelegateFromValidators) (*types.MsgUndelegateFromValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	numBatches, err := k.Keeper.UndelegateFromValidators(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUndelegateFromValidatorsResponse{NumBatchesSubmitted: numBatches}, nil
}
```

and add `sdk "github.com/cosmos/cosmos-sdk/types"` to that file's imports.

- [ ] **Step 9: Run the keeper tests to verify they pass**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestUndelegateFromValidators|TestKeeperTestSuite/TestBuildUndelegateFromValidatorsMsgs|TestKeeperTestSuite/TestMsgServer_UndelegateFromValidators' -v 2>&1 | grep -E '^(=== RUN|--- (PASS|FAIL)|ok|FAIL)'`
Expected: `--- PASS` for all 13 subtests.

- [ ] **Step 10: Write the failing CLI test**

```go
// x/stakeibc/client/cli/tx_wind_down_test.go
package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/client/cli"
)

func TestCmdUndelegateFromValidators(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		cmd := cli.CmdUndelegateFromValidators()
		ExecuteCLIExpectError(t, cmd, []string{"cosmoshub-4", "/does/not/exist.json"}, "no such file")
	})

	t.Run("bad offset in file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "validators.json")
		require.NoError(t, os.WriteFile(path, []byte(`[{"address":"cosmosvaloper1abc","offset":"banana"}]`), 0o600))
		cmd := cli.CmdUndelegateFromValidators()
		ExecuteCLIExpectError(t, cmd, []string{"cosmoshub-4", path}, "can not convert string to int")
	})
}
```

- [ ] **Step 11: Run it to verify it fails**

Run: `go test ./x/stakeibc/client/cli/... -run TestCmdUndelegateFromValidators -v`
Expected: build failure, `undefined: cli.CmdUndelegateFromValidators`.

- [ ] **Step 12: Write the CLI command**

Create `x/stakeibc/client/cli/tx_wind_down.go`:

```go
package cli

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Wind-down admin txs (spec §7)

// validatorUndelegationInput is one entry of the validators file for undelegate-from-validators
type validatorUndelegationInput struct {
	Address string `json:"address"`
	Offset  string `json:"offset"` // base units, optional, default "0"
}

func CmdUndelegateFromValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate-from-validators [chain-id] [validators-file]",
		Short: "Wind-down: undelegate the recorded delegation from every validator, or from those in the file",
		Long: `Submits MsgUndelegateFromValidators (admin only). With no file, every validator with a
recorded delegation is drained in full. With a file, only the listed validators are drained,
each for its recorded delegation minus the offset. The file is a JSON list:
  [{"address": "cosmosvaloper1...", "offset": "0"}, ...]`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainId := args[0]

			validators := []types.ValidatorUndelegation{}
			if len(args) == 2 {
				validators, err = readValidatorUndelegations(args[1])
				if err != nil {
					return err
				}
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUndelegateFromValidators(clientCtx.GetFromAddress().String(), argChainId, validators)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func readValidatorUndelegations(path string) ([]types.ValidatorUndelegation, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var inputs []validatorUndelegationInput
	if err := json.Unmarshal(contents, &inputs); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to parse validators file: %s", err)
	}

	validators := make([]types.ValidatorUndelegation, 0, len(inputs))
	for _, input := range inputs {
		offset := sdkmath.ZeroInt()
		if input.Offset != "" {
			parsed, found := sdkmath.NewIntFromString(input.Offset)
			if !found {
				return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "can not convert string to int: offset %q for %s", input.Offset, input.Address)
			}
			offset = parsed
		}
		validators = append(validators, types.ValidatorUndelegation{Address: input.Address, Offset: offset})
	}
	return validators, nil
}
```

Register it in `x/stakeibc/client/cli/tx.go` after `cmd.AddCommand(CmdToggleTradeController())`:

```go
	cmd.AddCommand(CmdUndelegateFromValidators())
```

- [ ] **Step 13: Run the CLI test and the full stakeibc suite**

Run: `go test ./x/stakeibc/client/cli/... -run TestCmdUndelegateFromValidators -v && go test ./x/stakeibc/... 2>&1 | tail -4`
Expected: PASS for both subtests; `ok` for keeper, types, cli.

- [ ] **Step 14: Commit**

```bash
git add x/stakeibc/types/message_undelegate_from_validators.go x/stakeibc/types/message_undelegate_from_validators_test.go x/stakeibc/keeper/wind_down_undelegate.go x/stakeibc/keeper/wind_down_undelegate_test.go x/stakeibc/keeper/msg_server_wind_down.go x/stakeibc/client/cli/tx_wind_down.go x/stakeibc/client/cli/tx_wind_down_test.go x/stakeibc/client/cli/tx.go
git commit -m "feat(stakeibc): MsgUndelegateFromValidators, the wind-down drain

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Task 4: `MsgTransferFromIca`

**Files:**
- Create: `x/stakeibc/types/message_transfer_from_ica.go`
- Test: `x/stakeibc/types/message_transfer_from_ica_test.go`
- Create: `x/stakeibc/keeper/wind_down_transfer_from_ica.go`
- Test: `x/stakeibc/keeper/wind_down_transfer_from_ica_test.go`
- Modify: `x/stakeibc/keeper/msg_server_wind_down.go` (replace the `TransferFromIca` stub)
- Modify: `x/stakeibc/client/cli/tx_wind_down.go` (add the command; create the file with the same header as Task 3 if it does not exist yet in your worktree), `x/stakeibc/client/cli/tx.go` (one line)
- Test: `x/stakeibc/client/cli/tx_wind_down_test.go` (add a test function; create the file if absent)

**Interfaces:**
- Consumes: `types.MsgTransferFromIca` (Task 2), `types.OsmosisVaultAddress`, `types.HostToOsmosisTransferChannel`, `types.WindDownTransferTimeout` (Task 1), `SubmitICATxWithoutCallback`, `types.FormatHostZoneICAOwner`, `utils.IntToUint`.
- Produces: `func (k Keeper) TransferFromIca(ctx sdk.Context, msg *types.MsgTransferFromIca) error`, `func BuildTransferFromIcaMsg(hostZone types.HostZone, icaType types.ICAAccountType, amount sdk.Coin, timeoutTimestamp uint64) (proto.Message, error)`, `types.NewMsgTransferFromIca(creator, chainId string, icaType types.ICAAccountType, amount sdk.Coin)`, CLI `transfer-from-ica`.
- Depends on: Tasks 1–2
- Review: yes (moves the entire host-side balance to a hard-coded receiver)

- [ ] **Step 1: Write the failing `ValidateBasic` test**

```go
// x/stakeibc/types/message_transfer_from_ica_test.go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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
		err  error
	}{
		{name: "valid delegation", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}},
		{name: "valid withdrawal", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_WITHDRAWAL, Amount: amount}},
		{name: "valid fee", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_FEE, Amount: amount}},
		{name: "valid redemption", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_REDEMPTION, Amount: amount}},
		{name: "valid foreign denom", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "dydx-mainnet-1", IcaType: types.ICAAccountType_WITHDRAWAL, Amount: sdk.NewCoin("ibc/8E27BA2D5493AF5636760E354E46004562C46AB7EC0CC4C1CA14E9E20E2545B5", sdkmath.NewInt(3))}},
		{name: "invalid creator", msg: types.MsgTransferFromIca{Creator: invalidAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}, err: sdkerrors.ErrInvalidAddress},
		{name: "not admin", msg: types.MsgTransferFromIca{Creator: validNotAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}, err: sdkerrors.ErrInvalidAddress},
		{name: "missing chain id", msg: types.MsgTransferFromIca{Creator: validAdminAddress, IcaType: types.ICAAccountType_DELEGATION, Amount: amount}, err: sdkerrors.ErrInvalidRequest},
		{name: "community pool ICA not allowed", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, Amount: amount}, err: sdkerrors.ErrInvalidRequest},
		{name: "converter ICA not allowed", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_CONVERTER_TRADE, Amount: amount}, err: sdkerrors.ErrInvalidRequest},
		{name: "zero amount", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: sdk.NewCoin("uatom", sdkmath.ZeroInt())}, err: sdkerrors.ErrInvalidRequest},
		{name: "invalid denom", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: sdk.Coin{Denom: "", Amount: sdkmath.NewInt(1)}}, err: sdkerrors.ErrInvalidRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.msg.Creator, tt.msg.GetSigners()[0].String())
		})
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./x/stakeibc/types/... -run TestMsgTransferFromIca_ValidateBasic -v`
Expected: build failure on the missing methods.

- [ ] **Step 3: Write the message file**

```go
// x/stakeibc/types/message_transfer_from_ica.go
package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
)

const TypeMsgTransferFromIca = "transfer_from_ica"

var _ sdk.Msg = &MsgTransferFromIca{}

// The four ICAs that hold anything worth moving (spec §3, §7)
var TransferableIcaTypes = map[ICAAccountType]bool{
	ICAAccountType_DELEGATION: true,
	ICAAccountType_WITHDRAWAL: true,
	ICAAccountType_FEE:        true,
	ICAAccountType_REDEMPTION: true,
}

func NewMsgTransferFromIca(creator, chainId string, icaType ICAAccountType, amount sdk.Coin) *MsgTransferFromIca {
	return &MsgTransferFromIca{
		Creator: creator,
		ChainId: chainId,
		IcaType: icaType,
		Amount:  amount,
	}
}

func (msg *MsgTransferFromIca) Route() string {
	return RouterKey
}

func (msg *MsgTransferFromIca) Type() string {
	return TypeMsgTransferFromIca
}

func (msg *MsgTransferFromIca) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgTransferFromIca) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if len(msg.ChainId) == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "chain-id is required")
	}
	if !TransferableIcaTypes[msg.IcaType] {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ica type %s is not one of DELEGATION, WITHDRAWAL, FEE, REDEMPTION", msg.IcaType)
	}
	if err := msg.Amount.Validate(); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount: %s", err)
	}
	if !msg.Amount.IsPositive() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "amount must be greater than 0")
	}
	return nil
}
```

- [ ] **Step 4: Run the types test to verify it passes**

Run: `go test ./x/stakeibc/types/... -run TestMsgTransferFromIca_ValidateBasic -v`
Expected: PASS for all twelve cases.

- [ ] **Step 5: Write the failing keeper tests**

```go
// x/stakeibc/keeper/wind_down_transfer_from_ica_test.go
package keeper_test

import (
	"time"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const (
	testOsmosisVault = "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"
	hubChainId       = "cosmoshub-4" // must be a key of HostToOsmosisTransferChannel
)

// Sets the vault for the test and restores the empty default afterwards
func (s *KeeperTestSuite) withOsmosisVault() {
	previous := types.OsmosisVaultAddress
	types.OsmosisVaultAddress = testOsmosisVault
	s.T().Cleanup(func() { types.OsmosisVaultAddress = previous })
}

func (s *KeeperTestSuite) hubHostZone() types.HostZone {
	return types.HostZone{
		ChainId:              hubChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		WithdrawalIcaAddress: "cosmos_WITHDRAWAL",
		FeeIcaAddress:        "cosmos_FEE",
		RedemptionIcaAddress: "cosmos_REDEMPTION",
	}
}

func (s *KeeperTestSuite) TestTransferFromIca_EveryIcaType() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	for _, icaType := range []types.ICAAccountType{
		types.ICAAccountType_DELEGATION,
		types.ICAAccountType_WITHDRAWAL,
		types.ICAAccountType_FEE,
		types.ICAAccountType_REDEMPTION,
	} {
		owner := types.FormatHostZoneICAOwner(hubChainId, icaType)
		channelId, portId := s.CreateICAChannel(owner)

		msg := types.NewMsgTransferFromIca("admin", hubChainId, icaType, sdk.NewCoin(Atom, sdkmath.NewInt(1000)))
		s.CheckICATxSubmitted(portId, channelId, func() error {
			return s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
		})
		s.CheckEventValueEmitted(types.EventTypeTransferFromIca, types.AttributeKeyIcaType, icaType.String())
	}
}

func (s *KeeperTestSuite) TestTransferFromIca_ForeignDenom() {
	s.withOsmosisVault()
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, s.hubHostZone())
	channelId, portId := s.CreateICAChannel(types.FormatHostZoneICAOwner(hubChainId, types.ICAAccountType_WITHDRAWAL))

	usdcOnHub := "ibc/F663521BF1836B00F5F177680F74BFB9A8B5654A694D0D2BC249E03CF2509013"
	msg := types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_WITHDRAWAL, sdk.NewCoin(usdcOnHub, sdkmath.NewInt(3_800_000)))
	s.CheckICATxSubmitted(portId, channelId, func() error {
		return s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	})
}

func (s *KeeperTestSuite) TestTransferFromIca_VaultNotConfigured() {
	// The default is empty: the tx must fail closed
	s.Require().Empty(types.OsmosisVaultAddress)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, s.hubHostZone())
	channelId, portId := s.CreateICAChannel(types.FormatHostZoneICAOwner(hubChainId, types.ICAAccountType_DELEGATION))

	msg := types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	s.CheckICATxNotSubmitted(portId, channelId, func() error {
		return s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	})
	err := s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrOsmosisVaultNotConfigured)
}

func (s *KeeperTestSuite) TestTransferFromIca_ChainIdNotInMap() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()
	hostZone.ChainId = "comdex-1"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	msg := types.NewMsgTransferFromIca("admin", "comdex-1", types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	err := s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrNoOsmosisChannelForHostZone)
}

func (s *KeeperTestSuite) TestTransferFromIca_MissingZoneAndIca() {
	s.withOsmosisVault()

	msg := types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	err := s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrHostZoneNotFound)

	hostZone := s.hubHostZone()
	hostZone.FeeIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	msg = types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_FEE, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	err = s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrICAAccountNotFound)
}

// The built ICA message: mapped channel, vault receiver, the given timeout, empty memo
func (s *KeeperTestSuite) TestBuildTransferFromIcaMsg_Transfer() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()
	amount := sdk.NewCoin(Atom, sdkmath.NewInt(1000))
	timeout := uint64(s.Ctx.BlockTime().Add(24 * time.Hour).UnixNano())

	built, err := keeper.BuildTransferFromIcaMsg(hostZone, types.ICAAccountType_WITHDRAWAL, amount, timeout)
	s.Require().NoError(err)
	transfer, ok := built.(*transfertypes.MsgTransfer)
	s.Require().True(ok, "non-osmosis zones build an ICS-20 transfer")
	s.Require().Equal(transfertypes.PortID, transfer.SourcePort)
	s.Require().Equal("channel-141", transfer.SourceChannel, "cosmoshub-4's channel to osmosis")
	s.Require().Equal(amount, transfer.Token)
	s.Require().Equal("cosmos_WITHDRAWAL", transfer.Sender)
	s.Require().Equal(testOsmosisVault, transfer.Receiver)
	s.Require().Equal(timeout, transfer.TimeoutTimestamp)
	s.Require().Zero(transfer.TimeoutHeight.RevisionHeight)
	s.Require().Empty(transfer.Memo)
}

// osmosis-1 maps to an empty channel: the ICA is already on Osmosis so it is a bank send
func (s *KeeperTestSuite) TestBuildTransferFromIcaMsg_OsmosisBankSend() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()
	hostZone.ChainId = types.OsmosisChainId
	hostZone.DelegationIcaAddress = "osmo_DELEGATION"
	amount := sdk.NewCoin(Osmo, sdkmath.NewInt(500))

	built, err := keeper.BuildTransferFromIcaMsg(hostZone, types.ICAAccountType_DELEGATION, amount, 1)
	s.Require().NoError(err)
	send, ok := built.(*banktypes.MsgSend)
	s.Require().True(ok, "osmosis-1 builds a bank send")
	s.Require().Equal("osmo_DELEGATION", send.FromAddress)
	s.Require().Equal(testOsmosisVault, send.ToAddress)
	s.Require().Equal(sdk.NewCoins(amount), send.Amount)
}

func (s *KeeperTestSuite) TestBuildTransferFromIcaMsg_Rejections() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()

	_, err := keeper.BuildTransferFromIcaMsg(hostZone, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, sdk.NewCoin(Atom, sdkmath.NewInt(1)), 1)
	s.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest, "only the four funded ICAs")

	hostZone.ChainId = "evmos_9001-2"
	_, err = keeper.BuildTransferFromIcaMsg(hostZone, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)), 1)
	s.Require().ErrorIs(err, types.ErrNoOsmosisChannelForHostZone)
}

func (s *KeeperTestSuite) TestMsgServer_TransferFromIca() {
	s.withOsmosisVault()
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, s.hubHostZone())
	channelId, portId := s.CreateICAChannel(types.FormatHostZoneICAOwner(hubChainId, types.ICAAccountType_DELEGATION))

	msg := types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	s.CheckICATxSubmitted(portId, channelId, func() error {
		_, err := s.GetMsgServer().TransferFromIca(s.Ctx, msg)
		return err
	})
}
```

- [ ] **Step 6: Run them to verify they fail**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestTransferFromIca|TestKeeperTestSuite/TestBuildTransferFromIcaMsg|TestKeeperTestSuite/TestMsgServer_TransferFromIca' -v 2>&1 | tail -5`
Expected: build failure, `undefined: keeper.BuildTransferFromIcaMsg`.

- [ ] **Step 7: Write the keeper file**

```go
// x/stakeibc/keeper/wind_down_transfer_from_ica.go
package keeper

import (
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Stride-Labs/stride/v34/utils"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// TransferFromIca submits one ICA containing a transfer of `amount` from one of the zone's
// four funded ICAs to the Osmosis vault (spec §7). The receiver is the hard-coded vault and
// the channel is the zone's hard-coded host-side channel to osmosis-1; for osmosis-1 itself
// the ICA already lives on Osmosis, so the message is a bank send. There is no callback: a
// failed or timed-out transfer refunds the ICA on the host and ops resubmit.
func (k Keeper) TransferFromIca(ctx sdk.Context, msg *types.MsgTransferFromIca) error {
	if types.OsmosisVaultAddress == "" {
		return types.ErrOsmosisVaultNotConfigured
	}
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return types.ErrHostZoneNotFound.Wrapf("host zone %s not found", msg.ChainId)
	}

	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())
	icaMsg, err := BuildTransferFromIcaMsg(hostZone, msg.IcaType, msg.Amount, timeoutTimestamp)
	if err != nil {
		return err
	}

	owner := types.FormatHostZoneICAOwner(hostZone.ChainId, msg.IcaType)
	if err := k.SubmitICATxWithoutCallback(ctx, hostZone.ConnectionId, owner, []proto.Message{icaMsg}, timeoutTimestamp); err != nil {
		return errorsmod.Wrapf(err, "unable to submit %s ICA transfer for %s", msg.IcaType, msg.ChainId)
	}

	channelId := types.HostToOsmosisTransferChannel[hostZone.ChainId]
	k.Logger(ctx).Info(utils.LogWithHostZone(msg.ChainId,
		"Wind-down transfer of %v from the %s ICA to %s over %q", msg.Amount, msg.IcaType, types.OsmosisVaultAddress, channelId))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTransferFromIca,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyHostZone, msg.ChainId),
			sdk.NewAttribute(types.AttributeKeyIcaType, msg.IcaType.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyChannel, channelId),
			sdk.NewAttribute(types.AttributeKeyReceiver, types.OsmosisVaultAddress),
		),
	)
	return nil
}

// BuildTransferFromIcaMsg builds the message the ICA executes on the host: an ICS-20
// MsgTransfer over the zone's mapped channel to osmosis-1, or a bank MsgSend when the zone is
// osmosis-1 (mapped to an empty channel). It is exported so tests can assert every field.
func BuildTransferFromIcaMsg(
	hostZone types.HostZone,
	icaType types.ICAAccountType,
	amount sdk.Coin,
	timeoutTimestamp uint64,
) (proto.Message, error) {
	channelId, found := types.HostToOsmosisTransferChannel[hostZone.ChainId]
	if !found {
		return nil, types.ErrNoOsmosisChannelForHostZone.Wrapf("no channel to osmosis configured for %s", hostZone.ChainId)
	}
	icaAddress, err := windDownIcaAddress(hostZone, icaType)
	if err != nil {
		return nil, err
	}

	if channelId == "" {
		return &banktypes.MsgSend{
			FromAddress: icaAddress,
			ToAddress:   types.OsmosisVaultAddress,
			Amount:      sdk.NewCoins(amount),
		}, nil
	}
	return &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    channelId,
		Token:            amount,
		Sender:           icaAddress,
		Receiver:         types.OsmosisVaultAddress,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             "",
	}, nil
}

// windDownIcaAddress resolves the ICA address for one of the four funded ICA types
func windDownIcaAddress(hostZone types.HostZone, icaType types.ICAAccountType) (string, error) {
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
		return "", errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ica type %s cannot be transferred from", icaType)
	}
	if address == "" {
		return "", types.ErrICAAccountNotFound.Wrapf("%s ICA for %s has no address", icaType, hostZone.ChainId)
	}
	return address, nil
}
```

- [ ] **Step 8: Wire the msg-server delegate**

Replace the `TransferFromIca` stub in `msg_server_wind_down.go` with:

```go
func (k msgServer) TransferFromIca(goCtx context.Context, msg *types.MsgTransferFromIca) (*types.MsgTransferFromIcaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.Keeper.TransferFromIca(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgTransferFromIcaResponse{}, nil
}
```

(add the `sdk` import if Task 3 has not already).

- [ ] **Step 9: Run the keeper tests to verify they pass**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestTransferFromIca|TestKeeperTestSuite/TestBuildTransferFromIcaMsg|TestKeeperTestSuite/TestMsgServer_TransferFromIca' -v 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL)'`
Expected: `--- PASS` for all nine.

- [ ] **Step 10: Write the failing CLI test**

Add to `x/stakeibc/client/cli/tx_wind_down_test.go`:

```go
func TestCmdTransferFromIca(t *testing.T) {
	t.Run("bad ica type", func(t *testing.T) {
		cmd := cli.CmdTransferFromIca()
		ExecuteCLIExpectError(t, cmd, []string{"cosmoshub-4", "TREASURY", "1000uatom"}, "unknown ica type")
	})
	t.Run("bad amount", func(t *testing.T) {
		cmd := cli.CmdTransferFromIca()
		ExecuteCLIExpectError(t, cmd, []string{"cosmoshub-4", "DELEGATION", "banana"}, "invalid decimal coin expression")
	})
}
```

- [ ] **Step 11: Run it to verify it fails**

Run: `go test ./x/stakeibc/client/cli/... -run TestCmdTransferFromIca -v`
Expected: build failure, `undefined: cli.CmdTransferFromIca`.

- [ ] **Step 12: Write the CLI command**

Add to `x/stakeibc/client/cli/tx_wind_down.go` (imports needed beyond Task 3's: `fmt`, `strings`, `sdk "github.com/cosmos/cosmos-sdk/types"`):

```go
func CmdTransferFromIca() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-from-ica [chain-id] [ica-type] [amount]",
		Short: "Wind-down: transfer an ICA balance to the Osmosis vault",
		Long: `Submits MsgTransferFromIca (admin only). ica-type is one of DELEGATION, WITHDRAWAL, FEE,
REDEMPTION; amount is a coin in the denom as it exists on the host (e.g. 1000000uatom). The
receiver and channel are hard-coded in the binary.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainId := args[0]
			icaTypeValue, found := types.ICAAccountType_value[strings.ToUpper(args[1])]
			if !found {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unknown ica type %s", args[1])
			}
			argAmount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid amount %q: %w", args[2], err)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgTransferFromIca(clientCtx.GetFromAddress().String(), argChainId, types.ICAAccountType(icaTypeValue), argAmount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
```

Register in `tx.go`:

```go
	cmd.AddCommand(CmdTransferFromIca())
```

- [ ] **Step 13: Run the CLI test and the full stakeibc suite**

Run: `go test ./x/stakeibc/client/cli/... -run TestCmdTransferFromIca -v && go test ./x/stakeibc/... 2>&1 | tail -4`
Expected: PASS; `ok` for keeper, types, cli.

- [ ] **Step 14: Commit**

```bash
git add x/stakeibc/types/message_transfer_from_ica.go x/stakeibc/types/message_transfer_from_ica_test.go x/stakeibc/keeper/wind_down_transfer_from_ica.go x/stakeibc/keeper/wind_down_transfer_from_ica_test.go x/stakeibc/keeper/msg_server_wind_down.go x/stakeibc/client/cli/tx_wind_down.go x/stakeibc/client/cli/tx_wind_down_test.go x/stakeibc/client/cli/tx.go
git commit -m "feat(stakeibc): MsgTransferFromIca, ICA balance to the Osmosis vault

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Task 5: `MsgTransferStaketiaClaimBalance`

**Files:**
- Create: `x/stakeibc/types/message_transfer_staketia_claim_balance.go`
- Test: `x/stakeibc/types/message_transfer_staketia_claim_balance_test.go`
- Create: `x/stakeibc/keeper/wind_down_staketia_claim.go`
- Test: `x/stakeibc/keeper/wind_down_staketia_claim_test.go`
- Modify: `x/stakeibc/keeper/msg_server_wind_down.go` (replace the `TransferStaketiaClaimBalance` stub)
- Modify: `x/stakeibc/client/cli/tx_wind_down.go` (add the command; create the file with Task 3's header if absent), `x/stakeibc/client/cli/tx.go` (one line)
- Test: `x/stakeibc/client/cli/tx_wind_down_test.go` (add a function; create if absent)

**Interfaces:**
- Consumes: `types.MsgTransferStaketiaClaimBalance` (Task 2), `types.WindDownTransferTimeout` (Task 1), `staketiatypes.CelestiaChainId`, `staketiatypes.ClaimAddress`, `staketiatypes.CelestiaNativeTokenIBCDenom`, `k.RecordsKeeper.TransferKeeper.Transfer`, `k.bankKeeper.GetBalance`.
- Produces: `func (k Keeper) TransferStaketiaClaimBalance(ctx sdk.Context, msg *types.MsgTransferStaketiaClaimBalance) (sdk.Coin, error)`, `func BuildStaketiaClaimTransferMsg(hostZone types.HostZone, token sdk.Coin, timeoutTimestamp uint64) transfertypes.MsgTransfer`, `types.NewMsgTransferStaketiaClaimBalance(creator string, amount sdkmath.Int)`, CLI `transfer-staketia-claim-balance`.
- Depends on: Tasks 1–2
- Review: yes (the keeper signs for a multisig account holding the whole staketia stake)

- [ ] **Step 1: Write the failing `ValidateBasic` test**

```go
// x/stakeibc/types/message_transfer_staketia_claim_balance_test.go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgTransferStaketiaClaimBalance_ValidateBasic(t *testing.T) {
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	tests := []struct {
		name string
		msg  types.MsgTransferStaketiaClaimBalance
		err  error
	}{
		{name: "valid zero amount (whole balance)", msg: types.MsgTransferStaketiaClaimBalance{Creator: validAdminAddress, Amount: sdkmath.ZeroInt()}},
		{name: "valid nil amount (whole balance)", msg: types.MsgTransferStaketiaClaimBalance{Creator: validAdminAddress}},
		{name: "valid positive amount", msg: types.MsgTransferStaketiaClaimBalance{Creator: validAdminAddress, Amount: sdkmath.NewInt(1_000_000)}},
		{name: "invalid creator", msg: types.MsgTransferStaketiaClaimBalance{Creator: invalidAddress}, err: sdkerrors.ErrInvalidAddress},
		{name: "not admin", msg: types.MsgTransferStaketiaClaimBalance{Creator: validNotAdminAddress}, err: sdkerrors.ErrInvalidAddress},
		{name: "negative amount", msg: types.MsgTransferStaketiaClaimBalance{Creator: validAdminAddress, Amount: sdkmath.NewInt(-1)}, err: sdkerrors.ErrInvalidRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.msg.Creator, tt.msg.GetSigners()[0].String())
		})
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./x/stakeibc/types/... -run TestMsgTransferStaketiaClaimBalance_ValidateBasic -v`
Expected: build failure on the missing methods.

- [ ] **Step 3: Write the message file**

```go
// x/stakeibc/types/message_transfer_staketia_claim_balance.go
package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
)

const TypeMsgTransferStaketiaClaimBalance = "transfer_staketia_claim_balance"

var _ sdk.Msg = &MsgTransferStaketiaClaimBalance{}

// amount is in utia; zero means the whole balance
func NewMsgTransferStaketiaClaimBalance(creator string, amount sdkmath.Int) *MsgTransferStaketiaClaimBalance {
	return &MsgTransferStaketiaClaimBalance{
		Creator: creator,
		Amount:  amount,
	}
}

func (msg *MsgTransferStaketiaClaimBalance) Route() string {
	return RouterKey
}

func (msg *MsgTransferStaketiaClaimBalance) Type() string {
	return TypeMsgTransferStaketiaClaimBalance
}

func (msg *MsgTransferStaketiaClaimBalance) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgTransferStaketiaClaimBalance) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if !msg.Amount.IsNil() && msg.Amount.IsNegative() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "amount must not be negative")
	}
	return nil
}
```

- [ ] **Step 4: Run the types test to verify it passes**

Run: `go test ./x/stakeibc/types/... -run TestMsgTransferStaketiaClaimBalance_ValidateBasic -v`
Expected: PASS for all six.

- [ ] **Step 5: Write the failing keeper tests**

The TIA voucher's trace is `transfer/channel-162/utia`; registering that trace makes the test's IBC denom equal the real constant, and sending it over the test channel (`channel-0`) takes the escrow path, so the escrow balance is the assertion. The refund test drives the transfer module's timeout handler directly.

```go
// x/stakeibc/keeper/wind_down_staketia_claim_test.go
package keeper_test

import (
	"time"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

type staketiaClaimTestCase struct {
	claimAddress   sdk.AccAddress
	escrowAddress  sdk.AccAddress
	tiaDenom       string
	initialBalance sdkmath.Int
	strdBalance    sdkmath.Int
}

func (s *KeeperTestSuite) SetupTransferStaketiaClaimBalance() staketiaClaimTestCase {
	s.CreateTransferChannel(staketiatypes.CelestiaChainId)

	// The real voucher trace: registering it makes the hash equal the staketia constant
	tiaTrace := transfertypes.NewDenom("utia", transfertypes.NewHop(transfertypes.PortID, staketiatypes.StrideToCelestiaTransferChannelId))
	s.App.TransferKeeper.SetDenom(s.Ctx, tiaTrace)
	s.Require().Equal(staketiatypes.CelestiaNativeTokenIBCDenom, tiaTrace.IBCDenom(), "test trace must hash to the constant")

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:              staketiatypes.CelestiaChainId,
		TransferChannelId:    ibctesting.FirstChannelID,
		DelegationIcaAddress: "celestia_DELEGATION",
	})

	claimAddress := sdk.MustAccAddressFromBech32(staketiatypes.ClaimAddress)
	initialBalance := sdkmath.NewInt(1_000_000)
	strdBalance := sdkmath.NewInt(500)
	s.FundAccount(claimAddress, sdk.NewCoin(tiaTrace.IBCDenom(), initialBalance))
	s.FundAccount(claimAddress, sdk.NewCoin("ustrd", strdBalance))

	return staketiaClaimTestCase{
		claimAddress:   claimAddress,
		escrowAddress:  transfertypes.GetEscrowAddress(transfertypes.PortID, ibctesting.FirstChannelID),
		tiaDenom:       tiaTrace.IBCDenom(),
		initialBalance: initialBalance,
		strdBalance:    strdBalance,
	}
}

func (s *KeeperTestSuite) checkClaimTransferred(tc staketiaClaimTestCase, transferred sdkmath.Int) {
	s.Require().Equal(tc.initialBalance.Sub(transferred), s.App.BankKeeper.GetBalance(s.Ctx, tc.claimAddress, tc.tiaDenom).Amount, "claim address TIA")
	s.Require().Equal(transferred, s.App.BankKeeper.GetBalance(s.Ctx, tc.escrowAddress, tc.tiaDenom).Amount, "escrowed TIA")
	s.Require().Equal(tc.strdBalance, s.App.BankKeeper.GetBalance(s.Ctx, tc.claimAddress, "ustrd").Amount, "only TIA moves")
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_ZeroAmountSendsEverything() {
	tc := s.SetupTransferStaketiaClaimBalance()
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)

	transferred, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.ZeroInt()))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(tc.tiaDenom, tc.initialBalance), transferred)

	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID), "one transfer submitted")
	s.checkClaimTransferred(tc, tc.initialBalance)
	s.CheckEventValueEmitted(types.EventTypeTransferStaketiaClaimBalance, types.AttributeKeyAmount, transferred.String())
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_PositiveAmountSendsExactly() {
	tc := s.SetupTransferStaketiaClaimBalance()
	amount := sdkmath.NewInt(400)

	transferred, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", amount))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(tc.tiaDenom, amount), transferred)
	s.checkClaimTransferred(tc, amount)
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_Rejections() {
	tc := s.SetupTransferStaketiaClaimBalance()

	// above the balance
	_, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", tc.initialBalance.AddRaw(1)))
	s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)
	s.checkClaimTransferred(tc, sdkmath.ZeroInt())

	// zero balance
	s.Require().NoError(s.App.BankKeeper.SendCoins(s.Ctx, tc.claimAddress, s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(tc.tiaDenom, tc.initialBalance))))
	_, err = s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.ZeroInt()))
	s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)

	// no celestia host zone
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, staketiatypes.CelestiaChainId)
	_, err = s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.ZeroInt()))
	s.Require().ErrorIs(err, types.ErrHostZoneNotFound)
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_MissingDelegationIca() {
	s.SetupTransferStaketiaClaimBalance()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, staketiatypes.CelestiaChainId)
	hostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.ZeroInt()))
	s.Require().ErrorIs(err, types.ErrICAAccountNotFound)
}

func (s *KeeperTestSuite) TestBuildStaketiaClaimTransferMsg() {
	hostZone := types.HostZone{TransferChannelId: "channel-162", DelegationIcaAddress: "celestia_DELEGATION"}
	token := sdk.NewCoin(staketiatypes.CelestiaNativeTokenIBCDenom, sdkmath.NewInt(7))
	timeout := uint64(s.Ctx.BlockTime().Add(24 * time.Hour).UnixNano())

	msg := keeper.BuildStaketiaClaimTransferMsg(hostZone, token, timeout)
	s.Require().Equal(transfertypes.PortID, msg.SourcePort)
	s.Require().Equal("channel-162", msg.SourceChannel)
	s.Require().Equal(token, msg.Token)
	s.Require().Equal(staketiatypes.ClaimAddress, msg.Sender)
	s.Require().Equal("celestia_DELEGATION", msg.Receiver)
	s.Require().Equal(timeout, msg.TimeoutTimestamp)
	s.Require().Zero(msg.TimeoutHeight.RevisionHeight)
	s.Require().Empty(msg.Memo)
}

// A timed-out packet refunds the claim address through the normal ICS-20 path
func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_TimeoutRefundsClaimAddress() {
	tc := s.SetupTransferStaketiaClaimBalance()
	amount := sdkmath.NewInt(400)
	_, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", amount))
	s.Require().NoError(err)
	s.checkClaimTransferred(tc, amount)

	// Rebuild the packet's token from the registered trace (the same lookup lsm.go uses)
	hash, err := transfertypes.ParseHexHash(tc.tiaDenom[len("ibc/"):])
	s.Require().NoError(err)
	denom, found := s.App.TransferKeeper.GetDenom(s.Ctx, hash)
	s.Require().True(found)
	packetData := transfertypes.NewInternalTransferRepresentation(
		transfertypes.Token{Denom: denom, Amount: amount.String()},
		staketiatypes.ClaimAddress, "celestia_DELEGATION", "",
	)
	err = s.App.TransferKeeper.OnTimeoutPacket(s.Ctx, transfertypes.PortID, ibctesting.FirstChannelID, packetData)
	s.Require().NoError(err)

	s.checkClaimTransferred(tc, sdkmath.ZeroInt())
}

func (s *KeeperTestSuite) TestMsgServer_TransferStaketiaClaimBalance() {
	tc := s.SetupTransferStaketiaClaimBalance()
	resp, err := s.GetMsgServer().TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.NewInt(10)))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(tc.tiaDenom, sdkmath.NewInt(10)), resp.Transferred)
}
```

- [ ] **Step 6: Run them to verify they fail**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestTransferStaketiaClaimBalance|TestKeeperTestSuite/TestBuildStaketiaClaimTransferMsg|TestKeeperTestSuite/TestMsgServer_TransferStaketiaClaimBalance' -v 2>&1 | tail -5`
Expected: build failure, `undefined: keeper.BuildStaketiaClaimTransferMsg`.

- [ ] **Step 7: Write the keeper file**

```go
// x/stakeibc/keeper/wind_down_staketia_claim.go
package keeper

import (
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

// TransferStaketiaClaimBalance moves TIA vouchers from staketia's claim address (a multisig
// BaseAccount whose signers are not on the critical path) to the stakeibc celestia zone's
// delegation ICA, where they unwind to native TIA and leave with the ICA balance (spec §7).
// The keeper signs for the claim address the way staketia's keeper already does for its
// deposit address. amount zero means the whole balance; a timeout refunds the claim address.
func (k Keeper) TransferStaketiaClaimBalance(ctx sdk.Context, msg *types.MsgTransferStaketiaClaimBalance) (sdk.Coin, error) {
	hostZone, found := k.GetHostZone(ctx, staketiatypes.CelestiaChainId)
	if !found {
		return sdk.Coin{}, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", staketiatypes.CelestiaChainId)
	}
	if hostZone.DelegationIcaAddress == "" {
		return sdk.Coin{}, types.ErrICAAccountNotFound.Wrapf("delegation ICA for %s has no address", staketiatypes.CelestiaChainId)
	}
	if hostZone.TransferChannelId == "" {
		return sdk.Coin{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "host zone %s has no transfer channel", staketiatypes.CelestiaChainId)
	}

	claimAddress, err := sdk.AccAddressFromBech32(staketiatypes.ClaimAddress)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrapf(err, "invalid staketia claim address constant")
	}
	balance := k.bankKeeper.GetBalance(ctx, claimAddress, staketiatypes.CelestiaNativeTokenIBCDenom)
	if balance.IsZero() {
		return sdk.Coin{}, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "claim address %s holds no %s", staketiatypes.ClaimAddress, staketiatypes.CelestiaNativeTokenIBCDenom)
	}

	amount := msg.Amount
	if amount.IsNil() || amount.IsZero() {
		amount = balance.Amount
	}
	if amount.GT(balance.Amount) {
		return sdk.Coin{}, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "requested %v but the claim address holds %v", amount, balance.Amount)
	}
	token := sdk.NewCoin(balance.Denom, amount)

	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())
	transferMsg := BuildStaketiaClaimTransferMsg(hostZone, token, timeoutTimestamp)
	if _, err := k.RecordsKeeper.TransferKeeper.Transfer(ctx, &transferMsg); err != nil {
		return sdk.Coin{}, errorsmod.Wrapf(err, "unable to transfer %v from the staketia claim address", token)
	}

	k.Logger(ctx).Info(utils.LogWithHostZone(staketiatypes.CelestiaChainId,
		"Wind-down transfer of %v from the staketia claim address to %s over %s", token, hostZone.DelegationIcaAddress, hostZone.TransferChannelId))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTransferStaketiaClaimBalance,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyAmount, token.String()),
			sdk.NewAttribute(types.AttributeKeyChannel, hostZone.TransferChannelId),
			sdk.NewAttribute(types.AttributeKeyReceiver, hostZone.DelegationIcaAddress),
		),
	)
	return token, nil
}

// BuildStaketiaClaimTransferMsg is the ICS-20 transfer from the claim address to the celestia
// delegation ICA. Exported so tests can assert every field.
func BuildStaketiaClaimTransferMsg(hostZone types.HostZone, token sdk.Coin, timeoutTimestamp uint64) transfertypes.MsgTransfer {
	return transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostZone.TransferChannelId,
		Token:            token,
		Sender:           staketiatypes.ClaimAddress,
		Receiver:         hostZone.DelegationIcaAddress,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             "",
	}
}
```

- [ ] **Step 8: Wire the msg-server delegate**

Replace the `TransferStaketiaClaimBalance` stub in `msg_server_wind_down.go` with:

```go
func (k msgServer) TransferStaketiaClaimBalance(goCtx context.Context, msg *types.MsgTransferStaketiaClaimBalance) (*types.MsgTransferStaketiaClaimBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	transferred, err := k.Keeper.TransferStaketiaClaimBalance(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgTransferStaketiaClaimBalanceResponse{Transferred: transferred}, nil
}
```

- [ ] **Step 9: Run the keeper tests to verify they pass**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestTransferStaketiaClaimBalance|TestKeeperTestSuite/TestBuildStaketiaClaimTransferMsg|TestKeeperTestSuite/TestMsgServer_TransferStaketiaClaimBalance' -v 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL)'`
Expected: `--- PASS` for all seven. If the escrow path burns instead of escrowing (the test channel happens to match the trace's channel), the assertion on `escrowAddress` fails with a zero balance: in that case assert on total supply of `tc.tiaDenom` decreasing by `transferred` instead, as `x/staketia/keeper/delegation_test.go:70-72` does.

- [ ] **Step 10: Write the failing CLI test**

Add to `x/stakeibc/client/cli/tx_wind_down_test.go`:

```go
func TestCmdTransferStaketiaClaimBalance(t *testing.T) {
	cmd := cli.CmdTransferStaketiaClaimBalance()
	ExecuteCLIExpectError(t, cmd, []string{"banana"}, "can not convert string to int")
}
```

- [ ] **Step 11: Run it to verify it fails**

Run: `go test ./x/stakeibc/client/cli/... -run TestCmdTransferStaketiaClaimBalance -v`
Expected: build failure, `undefined: cli.CmdTransferStaketiaClaimBalance`.

- [ ] **Step 12: Write the CLI command**

Add to `x/stakeibc/client/cli/tx_wind_down.go`:

```go
func CmdTransferStaketiaClaimBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-staketia-claim-balance [amount]",
		Short: "Wind-down: move the staketia claim address's TIA to the celestia delegation ICA",
		Long: `Submits MsgTransferStaketiaClaimBalance (admin only). amount is in utia and is optional:
omitted or 0 moves the whole balance. Use a small amount first as the live test.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			amount := sdkmath.ZeroInt()
			if len(args) == 1 {
				parsed, found := sdkmath.NewIntFromString(args[0])
				if !found {
					return errorsmod.Wrap(sdkerrors.ErrInvalidType, "can not convert string to int")
				}
				amount = parsed
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgTransferStaketiaClaimBalance(clientCtx.GetFromAddress().String(), amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
```

Register in `tx.go`:

```go
	cmd.AddCommand(CmdTransferStaketiaClaimBalance())
```

- [ ] **Step 13: Run the CLI test and the full stakeibc suite**

Run: `go test ./x/stakeibc/client/cli/... -run TestCmdTransferStaketiaClaimBalance -v && go test ./x/stakeibc/... 2>&1 | tail -4`
Expected: PASS; `ok` for keeper, types, cli.

- [ ] **Step 14: Commit**

```bash
git add x/stakeibc/types/message_transfer_staketia_claim_balance.go x/stakeibc/types/message_transfer_staketia_claim_balance_test.go x/stakeibc/keeper/wind_down_staketia_claim.go x/stakeibc/keeper/wind_down_staketia_claim_test.go x/stakeibc/keeper/msg_server_wind_down.go x/stakeibc/client/cli/tx_wind_down.go x/stakeibc/client/cli/tx_wind_down_test.go x/stakeibc/client/cli/tx.go
git commit -m "feat(stakeibc): MsgTransferStaketiaClaimBalance, claim address TIA to the celestia delegation ICA

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Merge gate

After Tasks 3–5 are merged back onto `wind-down-pr4-admin-txs`:

Run: `go build ./... && go vet ./x/stakeibc/... && go test ./x/stakeibc/... ./x/staketia/... ./x/autopilot/... ./app/... 2>&1 | tail -8`
Expected: all `ok`; the only known pre-existing failure in the repo is `utils` `TestCreateModuleAccount` (fails on main too, not in this set).

Confirm no stub is left except the sweep: `grep -n "ErrNotSupported" x/stakeibc/keeper/msg_server_wind_down.go` prints exactly one line, the `SweepTokensOffStride` handler.

Confirm the module path was not touched: `git diff wind-down-pr3-upgrade-handler...HEAD -- go.mod` prints nothing.

---

## Self-review

**Spec coverage (§7, §11 for the three txs, §4, §13):**
- Operator addresses as fail-closed vars, channel map, unwind whitelist, channel-5, 24h timeout, batch bound → Task 1.
- Four protos in one proto-gen, both registrations, sweep stub → Task 2.
- Undelegate: empty list = every funded validator; per-validator delegation − offset; rounding safety; rejects flagged validator, deprecated zone, queued/retry record (STRIDE-07); no accounting mutation; batch submit with nil ids; in-flight registration; halted zone accepted → Task 3, every §11 bullet has a named test.
- Transfer: vault constant as receiver; mapped channel; osmosis-1 bank send; chain absent from map rejected; foreign denom; four ICA types; built `MsgTransfer` fields; no callback → Task 4. The "map covers every in-scope zone and no deprecated one" and "constants parse" tests are in Task 1.
- Claim balance: optional amount, zero = all, over-balance and zero-balance rejected, only TIA moves, built fields, refund on timeout, keeper signs for the claim address via the records transfer keeper → Task 5.
- CLI for all three, registered in `tx.go`.
- Out of scope and untouched: the sweep logic (PR 5), the release-gate address values (PR 6), the module-path bump.

**Placeholder scan:** none. The one "if this helper does not exist" note in Task 5 Step 5 names the exact substitute.

**Type consistency:** `BuildTransferFromIcaMsg` and `BuildStaketiaClaimTransferMsg` are package-level functions (tests call `keeper.Build...`); `BuildUndelegateFromValidatorsMsgs` and `UndelegateFromValidators`, `TransferFromIca`, `TransferStaketiaClaimBalance` are `Keeper` methods (tests call `s.App.StakeibcKeeper....`). Response field names (`NumBatchesSubmitted`, `Transferred`) match the proto in Task 2. Error names match Task 2. Event/attribute constants match Task 2.

**Review tags:** every task is `Review: yes`; Task 1 and 2 are foundation, 3–5 are the parallel wave.
