# Wind-Down PR 5: Sweep Tx (`MsgSweepTokensOffStride`) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver `MsgSweepTokensOffStride`, the batched token sweep that sends each listed holder's stTokens and STRD to the same key on Osmosis and unwinds whitelisted IBC vouchers one hop, with its skip rules, destination resolution, CLI and the off-chain batch builder (spec §7, §11, §12 item 5, §13).

**Architecture:** PR 4 already generated the proto for all four wind-down messages (including `MsgSweepTokensOffStride{creator, denoms, addresses}` and its response `{num_transfers, num_skipped}`), registered them in the codec, created `x/stakeibc/types/wind_down.go` (constants and the two address vars) and left a stub `SweepTokensOffStride` handler in `msg_server_wind_down.go`. This PR adds the message's `ValidateBasic`, one keeper file with the sweep algorithm and its three helpers, replaces the stub handler, adds the CLI command, and ships the batch-builder script. Every ICS-20 transfer goes through the real transfer keeper (`k.RecordsKeeper.TransferKeeper.Transfer`) with the holder as sender, so timeouts and rejected receives refund the holder through the normal transfer path with no state of our own.

**Tech Stack:** Go 1.25, Cosmos SDK v0.54.3, ibc-go v11.2.0 (ICS-20 v1 channels), gogoproto (no proto changes in this PR), testify suites via `app/apptesting`, Python 3 (stdlib plus a vendored bech32 reference module) for the ops script.

> **Branching:** PR 1 branches off `wind-down-design-consolidation`. Each later PR branches
> off the previous PR's branch (PR 2 off PR 1, PR 3 off PR 2, and so on) and the PRs are
> implemented and merged strictly in order: 1, 2, 3, 4, 5, 6. Branch names:
> `wind-down-pr1-remove-handlers`, `wind-down-pr2-freeze-by-code`,
> `wind-down-pr3-upgrade-handler`, `wind-down-pr4-admin-txs`, `wind-down-pr5-sweep-tx`,
> `wind-down-pr6-release-gate`.
> The Go module path stays `github.com/Stride-Labs/stride/v34` in every PR; the bump to
> `/v35` is a manual step after all six land and is out of scope for every plan.

This plan's branch is `wind-down-pr5-sweep-tx`, created from `wind-down-pr4-admin-txs`.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` §7 "MsgSweepTokensOffStride", §11 "Sweep, the highest-review item", §13 "Admin txs" (sweep skip rules) and "Ops scripts".
- Only the sweep operator signs: `ValidateBasic` rejects any creator other than `types.SweepOperatorAddress`, and rejects everyone while that var is empty (fail closed; it is filled in by the release gate PR, not here).
- `denoms` non-empty, each passing `sdk.ValidateDenom`, no duplicates. `addresses` between 1 and `types.MaxSweepAddressesPerTx` (100) valid `stride` bech32 addresses, no duplicates.
- Destination per denom, decided once per tx before any address is read: a non-`ibc/` denom goes to `types.StrideToOsmosisTransferChannelId` (`channel-5`) with prefix `types.OsmosisBech32Prefix` (`osmo`); an `ibc/` denom goes over `Denom.Trace[0].ChannelId` (its outermost hop) with the prefix `types.SweepUnwindChannels[channel]`, and a channel absent from that map rejects the whole tx with `ErrSweepDestinationUnavailable`. An `ibc/` denom with no trace in the transfer store rejects the whole tx.
- Per address, skipped with event `sweep_skipped` (attributes `address`, `reason`) and counted in `num_skipped`, in this order: not 20 bytes; no account in the auth store; a transfer escrow address; account type not one of `*authtypes.BaseAccount`, `*vestingtypes.ContinuousVestingAccount`, `*vestingtypes.DelayedVestingAccount`, `*vestingtypes.PeriodicVestingAccount`, `*claimvestingtypes.StridePeriodicVestingAccount` (module accounts and `*icatypes.InterchainAccount` therefore skip). Skipping moves nothing.
- Per (address, denom): zero balance is skipped silently (no event, not counted). Otherwise one `MsgTransfer` of the full balance: `SourcePort` `transfer`, `SourceChannel` the destination channel, `Sender` the holder, `Receiver` `sdk.MustBech32ifyAddressBytes(prefix, holderBytes)`, `TimeoutTimestamp` block time + `types.WindDownTransferTimeout` (24h) in unix nanos, empty memo. A transfer error rejects the whole tx.
- Events: `types.EventTypeSweepSkipped = "sweep_skipped"` (`address`, `reason`), `types.EventTypeSweepTransfer = "sweep_transfer"` (`address`, `denom`, `amount`, `channel`, `receiver`). Errors: `ErrSweepDestinationUnavailable` (code 1569), `ErrSweepOperatorNotConfigured` (code 1570).
- No new keeper dependencies: escrow addresses from `k.IBCKeeper.ChannelKeeper.GetAllChannelsWithPortPrefix(ctx, transfertypes.PortID)` + `transfertypes.GetEscrowAddress`; denom traces from `k.RecordsKeeper.TransferKeeper.GetDenom`; accounts from `k.AccountKeeper.GetAccount`; balances from `k.bankKeeper.GetBalance`; transfers from `k.RecordsKeeper.TransferKeeper.Transfer`.
- No host-zone accounting is read or written. No proto change. No module-path change.
- Import path prefix `github.com/Stride-Labs/stride/v34`. Commit messages end with `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`.

---

## File structure

| File | Responsibility |
|---|---|
| `x/stakeibc/types/errors.go` | Modify: register `ErrSweepDestinationUnavailable` (1569), `ErrSweepOperatorNotConfigured` (1570). |
| `x/stakeibc/types/events.go` | Modify: add the two sweep event types and their attribute keys. |
| `x/stakeibc/types/message_sweep_tokens_off_stride.go` | Create: constructor, `Route/Type/GetSigners`, `ValidateBasic`. |
| `x/stakeibc/types/message_sweep_tokens_off_stride_test.go` | Create: table-driven `ValidateBasic` tests. |
| `x/stakeibc/keeper/wind_down_sweep.go` | Create: `sweepDestination`, `resolveSweepDestination`, `transferEscrowAddresses`, `sweepSkipReason`, `SweepTokensOffStride`, event emitters. |
| `x/stakeibc/keeper/wind_down_sweep_test.go` | Create: helper unit tests and the end-to-end sweep tests over real transfer channels. |
| `x/stakeibc/keeper/msg_server_wind_down.go` | Modify: replace the PR 4 stub body with the delegate. |
| `x/stakeibc/client/cli/tx_wind_down.go` | Modify: add `CmdSweepTokensOffStride` and register it. |
| `x/stakeibc/client/cli/tx_wind_down_test.go` | Modify (or create if PR 4 named it differently): argument-parsing tests for the new command. |
| `scripts/wind-down/build_sweep_batches.py` | Create: export → batch files, mirroring the on-chain skip rules plus a USD floor. |
| `scripts/wind-down/test_build_sweep_batches.py` | Create: unittest over a synthetic export. |

Before starting, confirm the PR 4 names this plan relies on exist on the branch: `types.SweepOperatorAddress`, `types.OsmosisBech32Prefix`, `types.StrideToOsmosisTransferChannelId`, `types.SweepUnwindChannels`, `types.MaxSweepAddressesPerTx`, `types.WindDownTransferTimeout` in `x/stakeibc/types/wind_down.go`; `types.MsgSweepTokensOffStride`, `types.MsgSweepTokensOffStrideResponse` in `tx.pb.go`; the stub `SweepTokensOffStride` in `x/stakeibc/keeper/msg_server_wind_down.go`; `GetTxCmd` registrations in `x/stakeibc/client/cli/tx_wind_down.go` (PR 4 added its three commands there; check how it registered them, because the new command is registered the same way).

```bash
grep -n "SweepOperatorAddress\|OsmosisBech32Prefix\|StrideToOsmosisTransferChannelId\|SweepUnwindChannels\|MaxSweepAddressesPerTx\|WindDownTransferTimeout" x/stakeibc/types/wind_down.go
grep -n "SweepTokensOffStride" x/stakeibc/keeper/msg_server_wind_down.go x/stakeibc/types/tx.pb.go | head
grep -n "AddCommand\|func Cmd" x/stakeibc/client/cli/tx_wind_down.go x/stakeibc/client/cli/tx.go | head -30
```

Expected: every name found. If the stub or a constant is missing, stop and report; the PR 4 branch is the base.

---

### Task 1: Message type, errors and events

**Files:**
- Modify: `x/stakeibc/types/errors.go` (append after `ErrRedemptionsDisabled`, code 1565)
- Modify: `x/stakeibc/types/events.go`
- Create: `x/stakeibc/types/message_sweep_tokens_off_stride.go`
- Test: `x/stakeibc/types/message_sweep_tokens_off_stride_test.go`

**Interfaces:**
- Consumes: `types.SweepOperatorAddress` (var), `types.MaxSweepAddressesPerTx` (const) from PR 4; generated `MsgSweepTokensOffStride`.
- Produces: `NewMsgSweepTokensOffStride(creator string, denoms, addresses []string) *MsgSweepTokensOffStride`; `(*MsgSweepTokensOffStride).ValidateBasic() error`; `types.ErrSweepDestinationUnavailable`, `types.ErrSweepOperatorNotConfigured`; `types.EventTypeSweepSkipped`, `types.EventTypeSweepTransfer`, `types.AttributeKeySweepAddress`, `types.AttributeKeySweepReason`, `types.AttributeKeySweepDenom`, `types.AttributeKeySweepAmount`, `types.AttributeKeySweepChannel`, `types.AttributeKeySweepReceiver`.
- Review: yes (the gate on the only tx that moves user balances).

- [ ] **Step 1: Write the failing ValidateBasic test**

`x/stakeibc/types/message_sweep_tokens_off_stride_test.go`:

```go
package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// withSweepOperator sets the sweep operator constant for one test and restores it after
func withSweepOperator(t *testing.T, address string) {
	t.Helper()
	previous := types.SweepOperatorAddress
	types.SweepOperatorAddress = address
	t.Cleanup(func() { types.SweepOperatorAddress = previous })
}

func randomStrideAddresses(n int) []string {
	addresses := make([]string, 0, n)
	for _, account := range apptesting.CreateRandomAccounts(n) {
		addresses = append(addresses, account.String())
	}
	return addresses
}

func TestMsgSweepTokensOffStride_ValidateBasic(t *testing.T) {
	operator := apptesting.SampleStrideAddress()
	withSweepOperator(t, operator)

	holders := randomStrideAddresses(3)
	tooMany := randomStrideAddresses(types.MaxSweepAddressesPerTx + 1)
	atCap := randomStrideAddresses(types.MaxSweepAddressesPerTx)

	tests := []struct {
		name string
		msg  types.MsgSweepTokensOffStride
		err  error
	}{
		{
			name: "valid: one denom, three holders",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"}, holders),
		},
		{
			name: "valid: three denoms incl. an ibc voucher, at the address cap",
			msg: *types.NewMsgSweepTokensOffStride(operator,
				[]string{"stuatom", "ustrd", "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"}, atCap),
		},
		{
			name: "invalid creator address",
			msg:  *types.NewMsgSweepTokensOffStride("invalid_address", []string{"stuatom"}, holders),
			err:  sdkerrors.ErrInvalidAddress,
		},
		{
			name: "creator is not the sweep operator",
			msg:  *types.NewMsgSweepTokensOffStride(holders[0], []string{"stuatom"}, holders),
			err:  sdkerrors.ErrUnauthorized,
		},
		{
			name: "empty denom list",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{}, holders),
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid denom string",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"st uatom"}, holders),
			err:  sdkerrors.ErrInvalidCoins,
		},
		{
			name: "duplicate denom",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom", "stuatom"}, holders),
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty address list",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"}, []string{}),
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "batch over the cap",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"}, tooMany),
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "address with the wrong bech32 prefix",
			msg: *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"},
				[]string{"osmo1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz"}),
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "duplicate address",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"}, []string{holders[0], holders[0]}),
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
			require.Equal(t, types.TypeMsgSweepTokensOffStride, tt.msg.Type())
			require.Equal(t, []sdk.AccAddress{sdk.MustAccAddressFromBech32(operator)}, tt.msg.GetSigners())
		})
	}
}

// While the operator constant is empty (as shipped by PR 4) the gate rejects everyone
func TestMsgSweepTokensOffStride_ValidateBasic_OperatorNotConfigured(t *testing.T) {
	withSweepOperator(t, "")

	msg := types.NewMsgSweepTokensOffStride(apptesting.SampleStrideAddress(), []string{"stuatom"}, randomStrideAddresses(1))
	err := msg.ValidateBasic()
	require.ErrorIs(t, err, types.ErrSweepOperatorNotConfigured, fmt.Sprintf("got %v", err))
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./x/stakeibc/types/... -run 'TestMsgSweepTokensOffStride' -v`
Expected: FAIL to compile with `undefined: types.NewMsgSweepTokensOffStride` (and the error/type names).

- [ ] **Step 3: Register the errors and events**

Append to the `var (...)` block in `x/stakeibc/types/errors.go`, directly after `ErrRedemptionsDisabled`:

```go
	ErrSweepDestinationUnavailable         = errorsmod.Register(ModuleName, 1569, "sweep destination unavailable for denom")
	ErrSweepOperatorNotConfigured          = errorsmod.Register(ModuleName, 1570, "sweep operator address is not configured")
```

Append to the stakeibc `const (...)` block in `x/stakeibc/types/events.go` (the block that holds `EventTypeRedemptionSweep`):

```go
	EventTypeSweepSkipped  = "sweep_skipped"
	EventTypeSweepTransfer = "sweep_transfer"

	AttributeKeySweepAddress  = "address"
	AttributeKeySweepReason   = "reason"
	AttributeKeySweepDenom    = "denom"
	AttributeKeySweepAmount   = "amount"
	AttributeKeySweepChannel  = "channel"
	AttributeKeySweepReceiver = "receiver"
```

- [ ] **Step 4: Write the message type**

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

func NewMsgSweepTokensOffStride(creator string, denoms []string, addresses []string) *MsgSweepTokensOffStride {
	return &MsgSweepTokensOffStride{
		Creator:   creator,
		Denoms:    denoms,
		Addresses: addresses,
	}
}

func (msg *MsgSweepTokensOffStride) Route() string {
	return RouterKey
}

func (msg *MsgSweepTokensOffStride) Type() string {
	return TypeMsgSweepTokensOffStride
}

func (msg *MsgSweepTokensOffStride) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic gates the sweep on the sweep operator (spec §4, §7) and bounds the batch.
// The operator var ships empty and is filled by the release gate; while it is empty the gate
// rejects every signer, so an unconfigured binary can never sweep.
func (msg *MsgSweepTokensOffStride) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if SweepOperatorAddress == "" {
		return ErrSweepOperatorNotConfigured
	}
	if msg.Creator != SweepOperatorAddress {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "creator %s is not the sweep operator", msg.Creator)
	}

	if len(msg.Denoms) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one denom is required")
	}
	seenDenoms := map[string]bool{}
	for _, denom := range msg.Denoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "invalid denom %s: %s", denom, err)
		}
		if seenDenoms[denom] {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate denom %s", denom)
		}
		seenDenoms[denom] = true
	}

	if len(msg.Addresses) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one address is required")
	}
	if len(msg.Addresses) > MaxSweepAddressesPerTx {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%d addresses exceeds the batch bound of %d",
			len(msg.Addresses), MaxSweepAddressesPerTx)
	}
	seenAddresses := map[string]bool{}
	for _, address := range msg.Addresses {
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address %s: %s", address, err)
		}
		if seenAddresses[address] {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate address %s", address)
		}
		seenAddresses[address] = true
	}
	return nil
}
```

Note: `sdk.AccAddressFromBech32` enforces the `stride` prefix (the tests' `SetupConfig` sets it) and accepts 20- and 32-byte addresses; the 20-byte rule is a keeper skip, not a `ValidateBasic` rejection, so a 32-byte holder in a batch never fails the good addresses beside it.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./x/stakeibc/types/... -run 'TestMsgSweepTokensOffStride' -v`
Expected: PASS for both tests (13 sub-tests).

- [ ] **Step 6: Commit**

```bash
git add x/stakeibc/types/errors.go x/stakeibc/types/events.go x/stakeibc/types/message_sweep_tokens_off_stride.go x/stakeibc/types/message_sweep_tokens_off_stride_test.go
git commit -m "feat(stakeibc): MsgSweepTokensOffStride type, operator gate, batch bound

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 2: Destination resolution and skip rules (keeper helpers)

**Files:**
- Create: `x/stakeibc/keeper/wind_down_sweep.go` (helpers only in this task; the sweep itself comes in Task 3)
- Test: `x/stakeibc/keeper/wind_down_sweep_test.go`

**Interfaces:**
- Consumes: `types.SweepUnwindChannels`, `types.StrideToOsmosisTransferChannelId`, `types.OsmosisBech32Prefix`, `k.RecordsKeeper.TransferKeeper.GetDenom`, `k.IBCKeeper.ChannelKeeper.GetAllChannelsWithPortPrefix`, `k.AccountKeeper.GetAccount`.
- Produces (package-private, used by Task 3):
  - `type sweepDestination struct { ChannelId string; Bech32Prefix string }`
  - `func (k Keeper) resolveSweepDestination(ctx sdk.Context, denom string) (sweepDestination, error)`
  - `func (k Keeper) transferEscrowAddresses(ctx sdk.Context) map[string]bool` (keyed by bech32 string)
  - `func (k Keeper) sweepSkipReason(ctx sdk.Context, address sdk.AccAddress, escrows map[string]bool) (reason string, skip bool)`
- Review: yes.

- [ ] **Step 1: Write the failing helper tests**

`x/stakeibc/keeper/wind_down_sweep_test.go` (first part; Task 3 appends to this file):

```go
package keeper_test

import (
	"fmt"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	claimvestingtypes "github.com/Stride-Labs/stride/v34/x/claim/vesting/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const (
	sweepTestStToken = "stuatom"
	sweepTestStrd    = "ustrd"
)

// registerVoucher stores a denom trace (outermost hop first) and returns its ibc/ denom
func (s *KeeperTestSuite) registerVoucher(base string, hops ...transfertypes.Hop) string {
	denom := transfertypes.NewDenom(base, hops...)
	s.App.TransferKeeper.SetDenom(s.Ctx, denom)
	return denom.IBCDenom()
}

func (s *KeeperTestSuite) TestResolveSweepDestination() {
	singleHopAtom := s.registerVoucher("uatom", transfertypes.NewHop(transfertypes.PortID, "channel-0"))
	twoHopStAtomViaOsmosis := s.registerVoucher(sweepTestStToken,
		transfertypes.NewHop(transfertypes.PortID, "channel-5"),
		transfertypes.NewHop(transfertypes.PortID, "channel-326"))
	unwhitelistedLuna := s.registerVoucher("uluna", transfertypes.NewHop(transfertypes.PortID, "channel-52"))

	testCases := []struct {
		name        string
		denom       string
		expected    keeper.SweepDestinationForTest
		expectedErr error
	}{
		{
			name:     "stToken goes to osmosis over channel-5",
			denom:    sweepTestStToken,
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
		},
		{
			name:     "ustrd goes to osmosis over channel-5",
			denom:    sweepTestStrd,
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
		},
		{
			name:     "single-hop atom voucher unwinds over channel-0 to cosmos",
			denom:    singleHopAtom,
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-0", Bech32Prefix: "cosmos"},
		},
		{
			name:     "two-hop voucher unwinds one hop over its outer channel",
			denom:    twoHopStAtomViaOsmosis,
			expected: keeper.SweepDestinationForTest{ChannelId: "channel-5", Bech32Prefix: "osmo"},
		},
		{
			name:        "voucher whose outer channel is not whitelisted is rejected",
			denom:       unwhitelistedLuna,
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
		{
			name:        "ibc denom with no trace in the store is rejected",
			denom:       "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
		{
			name:        "ibc denom with a malformed hash is rejected",
			denom:       "ibc/NOTAHASH",
			expectedErr: types.ErrSweepDestinationUnavailable,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			destination, err := keeper.ResolveSweepDestinationForTest(s.App.StakeibcKeeper, s.Ctx, tc.denom)
			if tc.expectedErr != nil {
				s.Require().ErrorIs(err, tc.expectedErr)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expected, destination)
		})
	}
}

// Sets up one account of each kind the skip rules distinguish and returns them keyed by label
func (s *KeeperTestSuite) setupSweepAccountKinds() map[string]sdk.AccAddress {
	accounts := apptesting.CreateRandomAccounts(8)
	kinds := map[string]sdk.AccAddress{}
	newBase := func(address sdk.AccAddress) *authtypes.BaseAccount {
		return authtypes.NewBaseAccountWithAddress(address)
	}
	vesting := sdk.NewCoins(sdk.NewInt64Coin(sweepTestStrd, 1))
	now := s.Ctx.BlockTime().Unix()

	s.App.AccountKeeper.SetAccount(s.Ctx, newBase(accounts[0]))
	kinds["base"] = accounts[0]

	continuous, err := vestingtypes.NewContinuousVestingAccount(newBase(accounts[1]), vesting, now, now+1000)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, continuous)
	kinds["continuous_vesting"] = accounts[1]

	delayed, err := vestingtypes.NewDelayedVestingAccount(newBase(accounts[2]), vesting, now+1000)
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, delayed)
	kinds["delayed_vesting"] = accounts[2]

	periodic, err := vestingtypes.NewPeriodicVestingAccount(newBase(accounts[3]), vesting, now,
		vestingtypes.Periods{{Length: 1000, Amount: vesting}})
	s.Require().NoError(err)
	s.App.AccountKeeper.SetAccount(s.Ctx, periodic)
	kinds["periodic_vesting"] = accounts[3]

	stridePeriodic := claimvestingtypes.NewStridePeriodicVestingAccount(newBase(accounts[4]), vesting,
		claimvestingtypes.Periods{{StartTime: now, Length: 1000, Amount: vesting}})
	s.App.AccountKeeper.SetAccount(s.Ctx, stridePeriodic)
	kinds["stride_periodic_vesting"] = accounts[4]

	ica := icatypes.NewInterchainAccount(newBase(accounts[5]), "cosmos1owner")
	s.App.AccountKeeper.SetAccount(s.Ctx, ica)
	kinds["interchain_account"] = accounts[5]

	kinds["unknown"] = accounts[6]

	kinds["module"] = authtypes.NewModuleAddress(distrtypes.ModuleName)
	kinds["thirty_two_bytes"] = sdk.AccAddress(make([]byte, 32))
	kinds["escrow"] = transfertypes.GetEscrowAddress(transfertypes.PortID, "channel-0")
	return kinds
}

func (s *KeeperTestSuite) TestSweepSkipReason() {
	kinds := s.setupSweepAccountKinds()

	// channel-0 must exist for its escrow address to be in the set
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, "channel-0", channeltypes.Channel{
		State:          channeltypes.OPEN,
		Ordering:       channeltypes.UNORDERED,
		Counterparty:   channeltypes.NewCounterparty(transfertypes.PortID, "channel-0"),
		ConnectionHops: []string{ibctesting.FirstConnectionID},
		Version:        transfertypes.V1,
	})
	escrows := keeper.TransferEscrowAddressesForTest(s.App.StakeibcKeeper, s.Ctx)
	s.Require().True(escrows[kinds["escrow"].String()], "channel-0 escrow should be in the set")

	testCases := []struct {
		kind           string
		expectedSkip   bool
		expectedReason string
	}{
		{kind: "base", expectedSkip: false},
		{kind: "continuous_vesting", expectedSkip: false},
		{kind: "delayed_vesting", expectedSkip: false},
		{kind: "periodic_vesting", expectedSkip: false},
		{kind: "stride_periodic_vesting", expectedSkip: false},
		{kind: "thirty_two_bytes", expectedSkip: true, expectedReason: "address is not 20 bytes"},
		{kind: "unknown", expectedSkip: true, expectedReason: "account not found"},
		{kind: "escrow", expectedSkip: true, expectedReason: "transfer escrow address"},
		{kind: "module", expectedSkip: true, expectedReason: "account type *types.ModuleAccount is not sweepable"},
		{kind: "interchain_account", expectedSkip: true, expectedReason: "account type *types.InterchainAccount is not sweepable"},
	}
	for _, tc := range testCases {
		s.Run(tc.kind, func() {
			reason, skip := keeper.SweepSkipReasonForTest(s.App.StakeibcKeeper, s.Ctx, kinds[tc.kind], escrows)
			s.Require().Equal(tc.expectedSkip, skip, fmt.Sprintf("skip for %s (reason %q)", tc.kind, reason))
			if tc.expectedSkip {
				s.Require().Equal(tc.expectedReason, reason)
			}
		})
	}
}
```

The `...ForTest` names are export shims in `export_test.go` (Step 3) so the helpers stay unexported.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/(TestResolveSweepDestination|TestSweepSkipReason)' -v`
Expected: FAIL to compile with `undefined: keeper.ResolveSweepDestinationForTest` (and the other shims).

- [ ] **Step 3: Write the helpers and the export shims**

`x/stakeibc/keeper/wind_down_sweep.go` (Task 3 appends `SweepTokensOffStride` to this file):

```go
package keeper

import (
	"fmt"
	"strings"

	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	claimvestingtypes "github.com/Stride-Labs/stride/v34/x/claim/vesting/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// sweepDestination is where a denom leaves Stride: the transfer channel and the bech32 prefix
// the holder's own address bytes are encoded with on the other side (spec §7)
type sweepDestination struct {
	ChannelId    string
	Bech32Prefix string
}

// resolveSweepDestination decides a denom's destination once per tx. A Stride-native denom
// (every stToken, ustrd) goes to Osmosis. An ibc/ voucher goes back over the channel it
// arrived on (the outermost hop of its trace) so it unwinds exactly one hop, and only if that
// channel leads to a chain whose wallets derive the same address bytes as Stride
// (types.SweepUnwindChannels); anything else has no safe destination and rejects the batch
func (k Keeper) resolveSweepDestination(ctx sdk.Context, denom string) (sweepDestination, error) {
	ibcPrefix := transfertypes.DenomPrefix + "/"
	if !strings.HasPrefix(denom, ibcPrefix) {
		return sweepDestination{
			ChannelId:    types.StrideToOsmosisTransferChannelId,
			Bech32Prefix: types.OsmosisBech32Prefix,
		}, nil
	}

	hash, err := transfertypes.ParseHexHash(denom[len(ibcPrefix):])
	if err != nil {
		return sweepDestination{}, errorsmod.Wrapf(types.ErrSweepDestinationUnavailable, "invalid ibc denom %s: %s", denom, err)
	}
	trace, found := k.RecordsKeeper.TransferKeeper.GetDenom(ctx, hash)
	if !found || len(trace.Trace) == 0 {
		return sweepDestination{}, errorsmod.Wrapf(types.ErrSweepDestinationUnavailable, "no denom trace for %s", denom)
	}

	outerChannel := trace.Trace[0].ChannelId
	prefix, whitelisted := types.SweepUnwindChannels[outerChannel]
	if !whitelisted {
		return sweepDestination{}, errorsmod.Wrapf(types.ErrSweepDestinationUnavailable,
			"denom %s arrived over %s, which is not a whitelisted unwind channel", denom, outerChannel)
	}
	return sweepDestination{ChannelId: outerChannel, Bech32Prefix: prefix}, nil
}

// transferEscrowAddresses returns the escrow address of every transfer channel, keyed by
// bech32 string. Escrows hold the supply of every stToken that lives on another chain and are
// never swept
func (k Keeper) transferEscrowAddresses(ctx sdk.Context) map[string]bool {
	escrows := map[string]bool{}
	for _, channel := range k.IBCKeeper.ChannelKeeper.GetAllChannelsWithPortPrefix(ctx, transfertypes.PortID) {
		escrows[transfertypes.GetEscrowAddress(channel.PortId, channel.ChannelId).String()] = true
	}
	return escrows
}

// sweepSkipReason applies the per-address rules of spec §7: only a 20-byte address whose
// account is a plain or vesting account has a counterpart the same key controls on the
// destination chain. Everything else (escrows, module accounts, interchain accounts owned by
// other chains, 32-byte contract-style addresses, addresses with no account) is skipped, and
// the reason is what the event carries so the off-chain builder learns why it disagreed
func (k Keeper) sweepSkipReason(ctx sdk.Context, address sdk.AccAddress, escrows map[string]bool) (reason string, skip bool) {
	if len(address) != 20 {
		return "address is not 20 bytes", true
	}
	account := k.AccountKeeper.GetAccount(ctx, address)
	if account == nil {
		return "account not found", true
	}
	if escrows[address.String()] {
		return "transfer escrow address", true
	}

	// The concrete types are listed on purpose: an interchain account embeds a BaseAccount, so
	// an interface check would admit it
	switch account.(type) {
	case *authtypes.BaseAccount,
		*vestingtypes.ContinuousVestingAccount,
		*vestingtypes.DelayedVestingAccount,
		*vestingtypes.PeriodicVestingAccount,
		*claimvestingtypes.StridePeriodicVestingAccount:
		return "", false
	case *icatypes.InterchainAccount:
		return fmt.Sprintf("account type %T is not sweepable", account), true
	default:
		return fmt.Sprintf("account type %T is not sweepable", account), true
	}
}
```

`x/stakeibc/keeper/export_test.go` (create if absent; if a file of that name already exists, append):

```go
package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// Test shims for the unexported sweep helpers
type SweepDestinationForTest = sweepDestination

func ResolveSweepDestinationForTest(k Keeper, ctx sdk.Context, denom string) (SweepDestinationForTest, error) {
	return k.resolveSweepDestination(ctx, denom)
}

func TransferEscrowAddressesForTest(k Keeper, ctx sdk.Context) map[string]bool {
	return k.transferEscrowAddresses(ctx)
}

func SweepSkipReasonForTest(k Keeper, ctx sdk.Context, address sdk.AccAddress, escrows map[string]bool) (string, bool) {
	return k.sweepSkipReason(ctx, address, escrows)
}
```

The `*icatypes.InterchainAccount` case is listed explicitly (even though `default` gives the same answer) so a reader sees the type the rule is written against; `%T` prints `*types.InterchainAccount` and `*types.ModuleAccount`, which is what the test expects.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/(TestResolveSweepDestination|TestSweepSkipReason)' -v`
Expected: PASS (7 + 10 sub-tests).

- [ ] **Step 5: Commit**

```bash
git add x/stakeibc/keeper/wind_down_sweep.go x/stakeibc/keeper/export_test.go x/stakeibc/keeper/wind_down_sweep_test.go
git commit -m "feat(stakeibc): sweep destination resolution and skip rules

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 3: The sweep itself, the handler, and the end-to-end tests

**Files:**
- Modify: `x/stakeibc/keeper/wind_down_sweep.go` (append `SweepTokensOffStride` and the event emitters)
- Modify: `x/stakeibc/keeper/msg_server_wind_down.go` (replace the stub)
- Test: `x/stakeibc/keeper/wind_down_sweep_test.go` (append)

**Interfaces:**
- Consumes: Task 2 helpers; `types.WindDownTransferTimeout`; `k.bankKeeper.GetBalance`; `k.RecordsKeeper.TransferKeeper.Transfer`.
- Produces: `func (k Keeper) SweepTokensOffStride(ctx sdk.Context, msg *types.MsgSweepTokensOffStride) (numTransfers, numSkipped uint64, err error)`; the `SweepTokensOffStride` msg-server handler returning `&types.MsgSweepTokensOffStrideResponse{NumTransfers, NumSkipped}`.
- Depends on: Tasks 1, 2.
- Review: yes (moves user balances).

**Test infrastructure decision.** `CreateTransferChannel` opens exactly one Stride transfer channel, `channel-0`, which is whitelisted (`cosmos`) so the voucher-unwind path runs over it end to end. The Osmosis path needs `channel-5`, and `StrideToOsmosisTransferChannelId` is a `const`, so the test opens five more transfer channels on the same connection (a helper below, mirroring the handshake in `apptesting.CreateTransferChannel` with `RunWithDifferentBechPrefix` around the host-side steps) until `channel-5` exists. The map `types.SweepUnwindChannels` is a package var and is not touched by these tests: `channel-0` and `channel-5` are both real mainnet entries. The one test that needs an unwhitelisted-but-existing channel uses `channel-1` (opened by the same helper, absent from the map).

- [ ] **Step 1: Write the failing end-to-end tests**

Append to `x/stakeibc/keeper/wind_down_sweep_test.go`:

```go
// openExtraTransferChannels opens n more transfer channels between Stride and the host chain
// on the connection CreateTransferChannel established, so tests can address channel-1 .. channel-n
func (s *KeeperTestSuite) openExtraTransferChannels(n int) {
	for i := 0; i < n; i++ {
		path := ibctesting.NewPath(s.StrideChain, s.HostChain).DisableUniqueChannelIDs()
		path.EndpointA.ClientID = s.TransferPath.EndpointA.ClientID
		path.EndpointA.ConnectionID = s.TransferPath.EndpointA.ConnectionID
		path.EndpointB.ClientID = s.TransferPath.EndpointB.ClientID
		path.EndpointB.ConnectionID = s.TransferPath.EndpointB.ConnectionID
		for _, endpoint := range []*ibctesting.Endpoint{path.EndpointA, path.EndpointB} {
			endpoint.ChannelConfig.PortID = ibctesting.TransferPort
			endpoint.ChannelConfig.Order = channeltypes.UNORDERED
			endpoint.ChannelConfig.Version = transfertypes.V1
		}

		s.Require().NoError(path.EndpointA.ChanOpenInit())
		apptesting.RunWithDifferentBechPrefix(sdk.Bech32MainPrefix, func() {
			s.Require().NoError(path.EndpointB.ChanOpenTry())
		})
		s.Require().NoError(path.EndpointA.ChanOpenAck())
		apptesting.RunWithDifferentBechPrefix(sdk.Bech32MainPrefix, func() {
			s.Require().NoError(path.EndpointB.ChanOpenConfirm())
		})
		s.Require().NoError(path.EndpointA.UpdateClient())
	}
	s.Ctx = s.StrideChain.GetContext()
}

type sweepTestCase struct {
	operator sdk.AccAddress
	holders  map[string]sdk.AccAddress
	atomIbc  string // single-hop uatom voucher over channel-0
	twoHop   string // stuatom that came back through channel-5 (two hops)
}

// SetupSweep opens channel-0 .. channel-5, funds a base account with an stToken, ustrd and
// a single-hop voucher, and registers the sweep operator
func (s *KeeperTestSuite) SetupSweep() sweepTestCase {
	s.CreateTransferChannel("GAIA")
	s.openExtraTransferChannels(5)
	_, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, transfertypes.PortID, "channel-5")
	s.Require().True(found, "channel-5 should exist after opening five extra channels")

	operator := s.TestAccs[0]
	s.App.AccountKeeper.SetAccount(s.Ctx, authtypes.NewBaseAccountWithAddress(operator))
	types.SweepOperatorAddress = operator.String()
	s.T().Cleanup(func() { types.SweepOperatorAddress = "" })

	kinds := s.setupSweepAccountKinds()
	atomIbc := s.registerVoucher("uatom", transfertypes.NewHop(transfertypes.PortID, "channel-0"))
	twoHop := s.registerVoucher(sweepTestStToken,
		transfertypes.NewHop(transfertypes.PortID, "channel-5"),
		transfertypes.NewHop(transfertypes.PortID, "channel-326"))

	return sweepTestCase{operator: operator, holders: kinds, atomIbc: atomIbc, twoHop: twoHop}
}

func (s *KeeperTestSuite) sweep(tc sweepTestCase, denoms []string, addresses ...sdk.AccAddress) (*types.MsgSweepTokensOffStrideResponse, error) {
	holders := make([]string, 0, len(addresses))
	for _, address := range addresses {
		holders = append(holders, address.String())
	}
	msg := types.NewMsgSweepTokensOffStride(tc.operator.String(), denoms, holders)
	s.Require().NoError(msg.ValidateBasic())
	return s.GetMsgServer().SweepTokensOffStride(s.Ctx, msg)
}

func (s *KeeperTestSuite) escrowBalance(channelId, denom string) sdkmath.Int {
	escrow := transfertypes.GetEscrowAddress(transfertypes.PortID, channelId)
	return s.App.BankKeeper.GetBalance(s.Ctx, escrow, denom).Amount
}

// One base account holding an stToken and ustrd: both leave over channel-5 to the same bytes
// with the osmo prefix, the full balances are escrowed, and only the listed denoms move
func (s *KeeperTestSuite) TestSweepTokensOffStride_NativeDenomsToOsmosis() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 1_000_000))
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStrd, 250_000))
	s.FundAccount(holder, sdk.NewInt64Coin("stuosmo", 9)) // not on the list, must stay

	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")

	resp, err := s.sweep(tc, []string{sweepTestStToken, sweepTestStrd}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), resp.NumTransfers)
	s.Require().Equal(uint64(0), resp.NumSkipped)

	endSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")
	s.Require().Equal(startSequence+2, endSequence, "two packets on channel-5")

	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64())
	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStrd).Amount.Int64())
	s.Require().Equal(int64(9), s.App.BankKeeper.GetBalance(s.Ctx, holder, "stuosmo").Amount.Int64(), "unlisted denom untouched")
	s.Require().Equal(int64(1_000_000), s.escrowBalance("channel-5", sweepTestStToken).Int64(), "stToken escrowed on channel-5")
	s.Require().Equal(int64(250_000), s.escrowBalance("channel-5", sweepTestStrd).Int64(), "ustrd escrowed on channel-5")

	expectedReceiver := sdk.MustBech32ifyAddressBytes("osmo", holder)
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepReceiver, expectedReceiver)
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepChannel, "channel-5")
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepDenom, sweepTestStToken)
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepAmount, "1000000")
	s.CheckEventTypeNotEmitted(types.EventTypeSweepSkipped)

	// The derived receiver decodes to exactly the sender's bytes
	receiverBytes, err := sdk.GetFromBech32(expectedReceiver, "osmo")
	s.Require().NoError(err)
	s.Require().Equal([]byte(holder), receiverBytes)
}

// A single-hop voucher goes back over channel-0 with the cosmos prefix and is burned (Stride
// is not the source of that denom)
func (s *KeeperTestSuite) TestSweepTokensOffStride_SingleHopVoucherUnwinds() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(tc.atomIbc, 40_000))
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-0")

	resp, err := s.sweep(tc, []string{tc.atomIbc}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), resp.NumTransfers)

	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-0"))
	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, tc.atomIbc).Amount.Int64())
	s.Require().Zero(s.App.BankKeeper.GetSupply(s.Ctx, tc.atomIbc).Amount.Int64(), "voucher burned on the way back")
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepReceiver, sdk.MustBech32ifyAddressBytes("cosmos", holder))
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepChannel, "channel-0")
}

// A two-hop voucher (stuatom that came back through Osmosis) unwinds one hop over channel-5
func (s *KeeperTestSuite) TestSweepTokensOffStride_TwoHopVoucherUnwindsOneHop() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(tc.twoHop, 500))
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")

	resp, err := s.sweep(tc, []string{tc.twoHop}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), resp.NumTransfers)
	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5"))
	s.Require().Zero(s.App.BankKeeper.GetSupply(s.Ctx, tc.twoHop).Amount.Int64(), "two-hop voucher burned")
	s.CheckEventValueEmitted(types.EventTypeSweepTransfer, types.AttributeKeySweepReceiver, sdk.MustBech32ifyAddressBytes("osmo", holder))
}

// Every sweepable account type is swept; every non-sweepable one is skipped with its reason
// while the rest of the batch goes through and num_skipped counts it
func (s *KeeperTestSuite) TestSweepTokensOffStride_SkipRulesInOneBatch() {
	tc := s.SetupSweep()
	swept := []string{"base", "continuous_vesting", "delayed_vesting", "periodic_vesting", "stride_periodic_vesting"}
	skipped := map[string]string{
		"thirty_two_bytes":   "address is not 20 bytes",
		"unknown":            "account not found",
		"escrow":             "transfer escrow address",
		"module":             "account type *types.ModuleAccount is not sweepable",
		"interchain_account": "account type *types.InterchainAccount is not sweepable",
	}
	addresses := []sdk.AccAddress{}
	for _, kind := range swept {
		s.FundAccount(tc.holders[kind], sdk.NewInt64Coin(sweepTestStToken, 1_000))
		addresses = append(addresses, tc.holders[kind])
	}
	// The module account and the escrow hold a balance too: skipping must leave it in place
	s.FundModuleAccount(distrtypes.ModuleName, sdk.NewInt64Coin(sweepTestStToken, 777))
	s.FundAccount(tc.holders["escrow"], sdk.NewInt64Coin(sweepTestStToken, 555))
	s.FundAccount(tc.holders["interchain_account"], sdk.NewInt64Coin(sweepTestStToken, 333))
	for kind := range skipped {
		addresses = append(addresses, tc.holders[kind])
	}

	resp, err := s.sweep(tc, []string{sweepTestStToken}, addresses...)
	s.Require().NoError(err)
	s.Require().Equal(uint64(len(swept)), resp.NumTransfers)
	s.Require().Equal(uint64(len(skipped)), resp.NumSkipped)

	for _, kind := range swept {
		s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, tc.holders[kind], sweepTestStToken).Amount.Int64(), kind)
	}
	s.Require().Equal(int64(777), s.App.BankKeeper.GetBalance(s.Ctx, tc.holders["module"], sweepTestStToken).Amount.Int64())
	s.Require().Equal(int64(555), s.App.BankKeeper.GetBalance(s.Ctx, tc.holders["escrow"], sweepTestStToken).Amount.Int64())
	s.Require().Equal(int64(333), s.App.BankKeeper.GetBalance(s.Ctx, tc.holders["interchain_account"], sweepTestStToken).Amount.Int64())

	for kind, reason := range skipped {
		s.CheckEventValueEmitted(types.EventTypeSweepSkipped, types.AttributeKeySweepAddress, tc.holders[kind].String())
		s.CheckEventValueEmitted(types.EventTypeSweepSkipped, types.AttributeKeySweepReason, reason)
	}
	s.Require().Len(s.CheckEventTypeEmitted(types.EventTypeSweepSkipped), len(skipped))
	s.Require().Len(s.CheckEventTypeEmitted(types.EventTypeSweepTransfer), len(swept))
}

// A zero balance is skipped silently: no transfer, no skip event, not counted
func (s *KeeperTestSuite) TestSweepTokensOffStride_ZeroBalanceSilent() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")

	resp, err := s.sweep(tc, []string{sweepTestStToken, sweepTestStrd}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), resp.NumTransfers)
	s.Require().Equal(uint64(0), resp.NumSkipped)
	s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5"))
	s.CheckEventTypeNotEmitted(types.EventTypeSweepSkipped)
	s.CheckEventTypeNotEmitted(types.EventTypeSweepTransfer)
}

// A holder with two of three listed denoms yields exactly two transfers
func (s *KeeperTestSuite) TestSweepTokensOffStride_TwoOfThreeDenoms() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 10))
	s.FundAccount(holder, sdk.NewInt64Coin(tc.atomIbc, 20))

	resp, err := s.sweep(tc, []string{sweepTestStToken, sweepTestStrd, tc.atomIbc}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), resp.NumTransfers)
	s.Require().Len(s.CheckEventTypeEmitted(types.EventTypeSweepTransfer), 2)
}

// A denom with no destination rejects the whole tx before any address is touched
func (s *KeeperTestSuite) TestSweepTokensOffStride_UnwhitelistedVoucherRejectsBatch() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	luna := s.registerVoucher("uluna", transfertypes.NewHop(transfertypes.PortID, "channel-1")) // channel-1 exists, not whitelisted
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 10))
	s.FundAccount(holder, sdk.NewInt64Coin(luna, 10))

	_, err := s.sweep(tc, []string{sweepTestStToken, luna}, holder)
	s.Require().ErrorIs(err, types.ErrSweepDestinationUnavailable)
	s.Require().Equal(int64(10), s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64(), "nothing moved")
	s.CheckEventTypeNotEmitted(types.EventTypeSweepTransfer)
}

// A transfer error (a whitelisted channel that does not exist on chain) rejects the whole tx
// even when earlier holders in the batch were fine, and the failed tx moves nothing.
//
// In production baseapp runs every message in a cache-wrapped context and only writes it on
// success (runMsgs / runTx), so a keeper error discards every state change the message made.
// The keeper suite calls the keeper directly with no baseapp in front, so this test reproduces
// that boundary by hand: it runs the sweep on s.Ctx.CacheContext(), never calls write, and then
// asserts on s.Ctx that nothing of the first holder's transfer survived.
func (s *KeeperTestSuite) TestSweepTokensOffStride_TransferErrorRejectsBatch() {
	tc := s.SetupSweep()
	junoVoucher := s.registerVoucher("ujuno", transfertypes.NewHop(transfertypes.PortID, "channel-24")) // whitelisted, no such channel
	first := tc.holders["base"]
	second := tc.holders["continuous_vesting"]
	s.FundAccount(first, sdk.NewInt64Coin(sweepTestStToken, 10))
	s.FundAccount(second, sdk.NewInt64Coin(junoVoucher, 10))

	sequenceBefore := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")
	escrowBefore := s.escrowBalance("channel-5", sweepTestStToken)

	msg := types.NewMsgSweepTokensOffStride(tc.operator.String(), []string{sweepTestStToken, junoVoucher},
		[]string{first.String(), second.String()})
	s.Require().NoError(msg.ValidateBasic())

	// The first holder's transfer succeeds inside the cache, the second holder's fails
	cacheCtx, _ := s.Ctx.CacheContext()
	_, _, err := s.App.StakeibcKeeper.SweepTokensOffStride(cacheCtx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "channel-24")

	// Nothing written to the cache reaches s.Ctx: balances, the channel sequence, the packet
	// commitment and the events are all as they were before the call
	s.Require().Equal(int64(10), s.App.BankKeeper.GetBalance(s.Ctx, first, sweepTestStToken).Amount.Int64(), "first holder untouched")
	s.Require().Equal(int64(10), s.App.BankKeeper.GetBalance(s.Ctx, second, junoVoucher).Amount.Int64(), "second holder untouched")
	s.Require().Equal(escrowBefore, s.escrowBalance("channel-5", sweepTestStToken), "escrow untouched")
	s.Require().Equal(sequenceBefore, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5"), "no packet sequence consumed")
	s.Require().Empty(s.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(s.Ctx, transfertypes.PortID, "channel-5"), "no packet commitment")
	s.CheckEventTypeNotEmitted(types.EventTypeSweepTransfer)
	s.CheckEventTypeNotEmitted(types.EventTypeSweepSkipped)
}

// The positive counterpart: through the msg server on s.Ctx, a good batch leaves its state
// changes in place (the cache boundary only discards on error)
func (s *KeeperTestSuite) TestSweepTokensOffStride_SuccessfulBatchPersists() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 10))
	sequenceBefore := s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5")

	resp, err := s.sweep(tc, []string{sweepTestStToken}, holder)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), resp.NumTransfers)

	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64(), "holder drained")
	s.Require().Equal(int64(10), s.escrowBalance("channel-5", sweepTestStToken).Int64(), "escrowed")
	s.Require().Equal(sequenceBefore+1, s.MustGetNextSequenceNumber(transfertypes.PortID, "channel-5"), "one packet sent")
	s.Require().Len(s.App.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(s.Ctx, transfertypes.PortID, "channel-5"), 1, "one commitment")
	s.Require().Len(s.CheckEventTypeEmitted(types.EventTypeSweepTransfer), 1)
}

// The ICS-20 timeout refund lands the escrowed balance back on the holder
func (s *KeeperTestSuite) TestSweepTokensOffStride_TimeoutRefundsHolder() {
	tc := s.SetupSweep()
	holder := tc.holders["base"]
	s.FundAccount(holder, sdk.NewInt64Coin(sweepTestStToken, 1_000))

	_, err := s.sweep(tc, []string{sweepTestStToken}, holder)
	s.Require().NoError(err)
	s.Require().Zero(s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64())

	// Replay the timeout through the transfer keeper with the packet the sweep built
	data := transfertypes.NewInternalTransferRepresentation(
		transfertypes.Token{Denom: transfertypes.NewDenom(sweepTestStToken), Amount: "1000"},
		holder.String(),
		sdk.MustBech32ifyAddressBytes("osmo", holder),
		"",
	)
	err = s.App.TransferKeeper.OnTimeoutPacket(s.Ctx, transfertypes.PortID, "channel-5", data)
	s.Require().NoError(err)
	s.Require().Equal(int64(1_000), s.App.BankKeeper.GetBalance(s.Ctx, holder, sweepTestStToken).Amount.Int64(), "refunded")
	s.Require().Zero(s.escrowBalance("channel-5", sweepTestStToken).Int64())
}
```

Test-writing notes for the implementer:
- `TestSweepTokensOffStride_TransferErrorRejectsBatch` calls the keeper on `s.Ctx.CacheContext()` and never writes it, which is exactly what baseapp does around a failing message, so the assertions on `s.Ctx` prove the whole tx rolled back (balances, sequence, commitments, events) rather than only that an error surfaced. `TestSweepTokensOffStride_SuccessfulBatchPersists` is its positive counterpart through the msg server.
- `s.CheckEventTypeEmitted` returns the matching events; `Len` on it counts them. `CheckEventValueEmitted` asserts at least one event of the type has the attribute value.
- Bech32 decoding of the derived receiver uses `sdk.GetFromBech32(addr, "osmo")`, which ignores the SDK config prefix.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/TestSweepTokensOffStride' -v`
Expected: every test FAILS at the msg-server call with the PR 4 stub error `MsgSweepTokensOffStride is delivered in the next PR` (wrapped `ErrNotSupported`).

- [ ] **Step 3: Write the sweep and replace the stub**

Append to `x/stakeibc/keeper/wind_down_sweep.go` (add `"github.com/Stride-Labs/stride/v34/utils"` to the imports):

```go
// SweepTokensOffStride sends every listed denom each listed holder owns to the holder's own
// address bytes on the destination chain (spec §7): Stride-native denoms to Osmosis, vouchers
// back one hop over the channel they arrived on. Destinations are resolved once, before any
// address is read, so a bad denom rejects the batch and a bad holder only skips itself.
// Returns how many transfers were submitted and how many addresses were skipped
func (k Keeper) SweepTokensOffStride(
	ctx sdk.Context,
	msg *types.MsgSweepTokensOffStride,
) (numTransfers uint64, numSkipped uint64, err error) {
	destinations := map[string]sweepDestination{}
	for _, denom := range msg.Denoms {
		destination, err := k.resolveSweepDestination(ctx, denom)
		if err != nil {
			return 0, 0, err
		}
		destinations[denom] = destination
	}

	escrows := k.transferEscrowAddresses(ctx)
	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())

	for _, holderBech32 := range msg.Addresses {
		holder := sdk.MustAccAddressFromBech32(holderBech32) // validated in ValidateBasic
		if reason, skip := k.sweepSkipReason(ctx, holder, escrows); skip {
			emitSweepSkippedEvent(ctx, holderBech32, reason)
			numSkipped++
			continue
		}

		for _, denom := range msg.Denoms {
			balance := k.bankKeeper.GetBalance(ctx, holder, denom)
			if balance.IsZero() {
				continue
			}

			destination := destinations[denom]
			receiver := sdk.MustBech32ifyAddressBytes(destination.Bech32Prefix, holder)
			transfer := transfertypes.MsgTransfer{
				SourcePort:       transfertypes.PortID,
				SourceChannel:    destination.ChannelId,
				Token:            balance,
				Sender:           holderBech32,
				Receiver:         receiver,
				TimeoutTimestamp: timeoutTimestamp,
				Memo:             "",
			}
			// A failed submission (closed channel, send disabled) is a batch problem, not a
			// holder problem: reject the whole tx so ops fix the cause and resubmit
			if _, err := k.RecordsKeeper.TransferKeeper.Transfer(ctx, &transfer); err != nil {
				return 0, 0, errorsmod.Wrapf(err, "unable to sweep %s from %s over %s",
					balance.String(), holderBech32, destination.ChannelId)
			}

			emitSweepTransferEvent(ctx, holderBech32, balance, destination.ChannelId, receiver)
			numTransfers++
		}
	}

	k.Logger(ctx).Info(fmt.Sprintf("Sweep submitted %d transfers for %d denoms across %d addresses (%d skipped)",
		numTransfers, len(msg.Denoms), len(msg.Addresses), numSkipped))
	return numTransfers, numSkipped, nil
}

func emitSweepSkippedEvent(ctx sdk.Context, address, reason string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSweepSkipped,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeySweepAddress, address),
			sdk.NewAttribute(types.AttributeKeySweepReason, reason),
		),
	)
}

func emitSweepTransferEvent(ctx sdk.Context, address string, amount sdk.Coin, channelId, receiver string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSweepTransfer,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeySweepAddress, address),
			sdk.NewAttribute(types.AttributeKeySweepDenom, amount.Denom),
			sdk.NewAttribute(types.AttributeKeySweepAmount, amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeySweepChannel, channelId),
			sdk.NewAttribute(types.AttributeKeySweepReceiver, receiver),
		),
	)
}
```

In `x/stakeibc/keeper/msg_server_wind_down.go`, replace the PR 4 stub body:

```go
// SweepTokensOffStride is the batched token sweep, gated on the sweep operator in ValidateBasic
func (k msgServer) SweepTokensOffStride(goCtx context.Context, msg *types.MsgSweepTokensOffStride) (*types.MsgSweepTokensOffStrideResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	numTransfers, numSkipped, err := k.Keeper.SweepTokensOffStride(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgSweepTokensOffStrideResponse{NumTransfers: numTransfers, NumSkipped: numSkipped}, nil
}
```

Remove the `sdkerrors` import from that file if the stub was its only user.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./x/stakeibc/keeper/... -run 'TestKeeperTestSuite/(TestSweepTokensOffStride|TestResolveSweepDestination|TestSweepSkipReason)' -v`
Expected: PASS, 9 sweep tests plus the two helper tests.

If `openExtraTransferChannels` fails on `ChanOpenTry` with a client-update error, add `s.Require().NoError(path.EndpointB.UpdateClient())` (inside `RunWithDifferentBechPrefix`) before the `ChanOpenTry` and `ChanOpenConfirm` calls; the handshake in `apptesting.CreateTransferChannel` is the reference.

- [ ] **Step 5: Run the whole stakeibc suite**

Run: `go test ./x/stakeibc/... 2>&1 | tail -20`
Expected: `ok` for every package.

- [ ] **Step 6: Commit**

```bash
git add x/stakeibc/keeper/wind_down_sweep.go x/stakeibc/keeper/msg_server_wind_down.go x/stakeibc/keeper/wind_down_sweep_test.go
git commit -m "feat(stakeibc): MsgSweepTokensOffStride keeper and handler with end-to-end tests

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Parallel-safe tasks

Tasks 4 and 5 depend only on Tasks 1-3 (Task 4 on the message constructor; Task 5 on nothing in Go at all) and not on each other.

### Task 4: CLI command

**Files:**
- Modify: `x/stakeibc/client/cli/tx_wind_down.go` (PR 4 created it with the three ICA-side commands)
- Modify: `x/stakeibc/client/cli/tx.go` if PR 4 registered its commands there (`cmd.AddCommand(CmdSweepTokensOffStride())` beside the other three); otherwise register in the same place PR 4 did
- Test: `x/stakeibc/client/cli/tx_wind_down_test.go` (append; create with `package cli_test` if PR 4 did not)

**Interfaces:**
- Consumes: `types.NewMsgSweepTokensOffStride`, `cli_test.ExecuteCLIExpectError` (existing helper in `cli_test.go`).
- Produces: `CmdSweepTokensOffStride() *cobra.Command`, `Use: "sweep-tokens-off-stride [denoms] [addresses-file]"`.
- Depends on: Task 1.
- Review: no.

- [ ] **Step 1: Write the failing CLI tests**

Append to `x/stakeibc/client/cli/tx_wind_down_test.go`:

```go
func TestCmdSweepTokensOffStride(t *testing.T) {
	t.Run("addresses file missing", func(t *testing.T) {
		cmd := cli.CmdSweepTokensOffStride()
		ExecuteCLIExpectError(t, cmd, []string{"stuatom,ustrd", "/nonexistent/addresses.txt"}, "unable to read addresses file")
	})

	t.Run("empty denoms", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "addresses.txt")
		require.NoError(t, os.WriteFile(file, []byte("stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7\n"), 0o600))
		cmd := cli.CmdSweepTokensOffStride()
		ExecuteCLIExpectError(t, cmd, []string{"", file}, "at least one denom is required")
	})

	t.Run("empty addresses file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "addresses.txt")
		require.NoError(t, os.WriteFile(file, []byte("\n\n"), 0o600))
		cmd := cli.CmdSweepTokensOffStride()
		ExecuteCLIExpectError(t, cmd, []string{"stuatom", file}, "addresses file is empty")
	})
}
```

Add `"os"`, `"path/filepath"` and `"github.com/stretchr/testify/require"` to the file's imports if absent.

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./x/stakeibc/client/cli/... -run TestCmdSweepTokensOffStride -v`
Expected: FAIL to compile with `undefined: cli.CmdSweepTokensOffStride`.

- [ ] **Step 3: Write the command**

Append to `x/stakeibc/client/cli/tx_wind_down.go` (add `"bufio"`, `"os"`, `"strings"` and `errorsmod "cosmossdk.io/errors"`, `sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"` to the imports if absent):

```go
// CmdSweepTokensOffStride submits one sweep batch: every listed denom, for every holder in the
// file (one bech32 address per line; blank lines and lines starting with # are ignored).
// The file is produced by scripts/wind-down/build_sweep_batches.py
func CmdSweepTokensOffStride() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sweep-tokens-off-stride [denoms] [addresses-file]",
		Short: "Sweep the listed denoms off Stride for every holder in the file (sweep operator only)",
		Long: `Sends each listed denom that each holder in the file owns to the holder's own address on the
destination chain: stTokens and ustrd to Osmosis over channel-5, IBC vouchers back one hop over the
channel they arrived on (whitelisted channels only). denoms is comma-separated. The file holds one
Stride address per line, at most 100.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			denoms := parseCommaSeparated(args[0])
			if len(denoms) == 0 {
				return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one denom is required")
			}
			addresses, err := readAddressesFile(args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgSweepTokensOffStride(clientCtx.GetFromAddress().String(), denoms, addresses)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseCommaSeparated(raw string) []string {
	values := []string{}
	for _, value := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func readAddressesFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to read addresses file %s: %s", path, err)
	}
	defer file.Close()

	addresses := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		addresses = append(addresses, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to read addresses file %s: %s", path, err)
	}
	if len(addresses) == 0 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "addresses file is empty: %s", path)
	}
	return addresses, nil
}
```

Register it next to PR 4's three commands (in `GetTxCmd` in `tx.go`, or wherever PR 4 put its `AddCommand` lines):

```go
	cmd.AddCommand(CmdSweepTokensOffStride())
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./x/stakeibc/client/cli/... -run TestCmdSweepTokensOffStride -v`
Expected: PASS (3 sub-tests).

- [ ] **Step 5: Build the binary and check the help text**

Run: `go build ./... && go run ./cmd/strided tx stakeibc sweep-tokens-off-stride --help | head -12`
Expected: the `Use` line and the `Long` text above.

- [ ] **Step 6: Commit**

```bash
git add x/stakeibc/client/cli/tx_wind_down.go x/stakeibc/client/cli/tx_wind_down_test.go x/stakeibc/client/cli/tx.go
git commit -m "feat(stakeibc): sweep-tokens-off-stride CLI

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

### Task 5: Batch builder script

**Files:**
- Create: `scripts/wind-down/bech32_ref.py` (vendored BIP-173 reference implementation)
- Create: `scripts/wind-down/build_sweep_batches.py`
- Test: `scripts/wind-down/test_build_sweep_batches.py`

**Interfaces:**
- Consumes: a trimmed `strided export` JSON (`app_state.bank.balances`, `app_state.auth.accounts`, `app_state.ibc.channel_genesis.channels`), a prices JSON `{denom: {"usd_per_token": float, "decimals": int}}`.
- Produces: `<out-dir>/batch-001.txt ...` (one address per line, ≤ 100 lines, the CLI's input) and `<out-dir>/summary.json` (per batch: addresses, denoms, USD swept; the skipped list with reasons). Public functions used by the test: `load_export(path) -> Export`, `classify_holders(export, denoms, prices, floor_usd) -> HolderPlan`, `write_batches(plan, out_dir, batch_size) -> list[pathlib.Path]`, `parse_extra_denom(spec) -> ExtraDenom`, `check_denom_destination(denom, traces) -> None`, `batch_size_arg(text) -> int`; from `bech32_ref`: `encode(hrp, address_bytes) -> str`, `decode(bech) -> tuple[str, bytes]`.
- Depends on: none (mirrors constants; no Go dependency).
- Review: no.

The script mirrors the on-chain rules exactly, so a batch it emits should skip nothing on chain; any `sweep_skipped` event after a submission is a disagreement worth investigating. It is stdlib plus the vendored `bech32_ref.py` (no pip dependency, so it runs in a clean checkout) and follows the repo's Python conventions (module imports, typed signatures, dataclasses, guard clauses). It also mirrors the two on-chain bounds a batch can violate: the address bound (`MAX_SWEEP_ADDRESSES_PER_TX = 100`, the value of `types.MaxSweepAddressesPerTx`) is enforced on `--batch-size`, and the destination rule is enforced on every denom in the final list, whether it came from `--denoms` or `--extra-denom`.

- [ ] **Step 1: Write the failing test**

`scripts/wind-down/test_build_sweep_batches.py`:

```python
"""Unit tests for build_sweep_batches over a synthetic export.

    python3 -m unittest scripts/wind-down/test_build_sweep_batches.py
"""

import argparse
import json
import pathlib
import tempfile
import unittest

import bech32_ref
import build_sweep_batches

STRIDE_BASE = "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"
STRIDE_VESTING = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"
STRIDE_MODULE = "stride1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8d8v4ck"  # any 20-byte address, typed ModuleAccount below
STRIDE_ICA = "stride1d6ntc7s8gs86tpdyn422vsqc6uaz9cejp8nc04"
STRIDE_NO_ACCOUNT = "stride15up3hegy8zuqhy0p9m8luh0c984ptu2gxqy20g"
STRIDE_DUST = "stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd"
ATOM_VOUCHER = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"


def synthetic_export() -> dict:
    escrow = build_sweep_batches.escrow_address("transfer", "channel-5")
    balances = [
        {"address": STRIDE_BASE, "coins": [{"denom": "stuatom", "amount": "10000000"}, {"denom": "ustrd", "amount": "5000000"}]},
        {"address": STRIDE_VESTING, "coins": [{"denom": "stuatom", "amount": "2000000"}]},
        {"address": STRIDE_MODULE, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": STRIDE_ICA, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": STRIDE_NO_ACCOUNT, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": escrow, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": STRIDE_DUST, "coins": [{"denom": "stuatom", "amount": "1000"}]},
    ]
    accounts = [
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": STRIDE_BASE},
        {"@type": "/stride.vesting.StridePeriodicVestingAccount", "base_vesting_account": {"base_account": {"address": STRIDE_VESTING}}},
        {"@type": "/cosmos.auth.v1beta1.ModuleAccount", "base_account": {"address": STRIDE_MODULE}, "name": "distribution"},
        {"@type": "/ibc.applications.interchain_accounts.v1.InterchainAccount", "base_account": {"address": STRIDE_ICA}, "account_owner": "x"},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": STRIDE_DUST},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": escrow},
    ]
    channels = [{"port_id": "transfer", "channel_id": "channel-5", "state": "STATE_OPEN"}]
    return {
        "app_state": {
            "bank": {"balances": balances},
            "auth": {"accounts": accounts},
            "ibc": {"channel_genesis": {"channels": channels}},
        }
    }


PRICES = {
    "stuatom": {"usd_per_token": 10.0, "decimals": 6},
    "ustrd": {"usd_per_token": 0.05, "decimals": 6},
}


class BuildSweepBatchesTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.export_path = pathlib.Path(self.tmp.name) / "export.json"
        self.export_path.write_text(json.dumps(synthetic_export()))
        self.export = build_sweep_batches.load_export(self.export_path)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_classify_applies_skip_rules_and_floor(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom", "ustrd"], prices=PRICES, floor_usd=1.0)

        swept = [holder.address for holder in plan.holders]
        self.assertEqual(swept, [STRIDE_BASE, STRIDE_VESTING], "ordered by USD, base ($100.25) before vesting ($20)")

        skipped = {entry.address: entry.reason for entry in plan.skipped}
        self.assertEqual(skipped[STRIDE_MODULE], "account type /cosmos.auth.v1beta1.ModuleAccount is not sweepable")
        self.assertEqual(skipped[STRIDE_ICA], "account type /ibc.applications.interchain_accounts.v1.InterchainAccount is not sweepable")
        self.assertEqual(skipped[STRIDE_NO_ACCOUNT], "account not found")
        self.assertEqual(skipped[build_sweep_batches.escrow_address("transfer", "channel-5")], "transfer escrow address")
        self.assertEqual(skipped[STRIDE_DUST], "below floor ($0.01 < $1.00)")

    def test_write_batches_splits_at_batch_size(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom", "ustrd"], prices=PRICES, floor_usd=1.0)
        out_dir = pathlib.Path(self.tmp.name) / "out"

        files = build_sweep_batches.write_batches(plan=plan, out_dir=out_dir, batch_size=1)

        self.assertEqual([path.name for path in files], ["batch-001.txt", "batch-002.txt"])
        self.assertEqual(files[0].read_text().strip(), STRIDE_BASE)
        summary = json.loads((out_dir / "summary.json").read_text())
        self.assertEqual(summary["denoms"], ["stuatom", "ustrd"])
        self.assertEqual(summary["batches"][0]["num_addresses"], 1)
        self.assertEqual(len(summary["skipped"]), 5)

    def test_extra_denom_parsing_and_whitelist(self) -> None:
        extra = build_sweep_batches.parse_extra_denom("ustrd=0.05:6")
        self.assertEqual((extra.denom, extra.usd_per_token, extra.decimals), ("ustrd", 0.05, 6))

        with self.assertRaises(ValueError):
            build_sweep_batches.parse_extra_denom("ustrd=0.05")

    def test_ibc_denom_requires_whitelisted_outer_hop_however_it_is_passed(self) -> None:
        traces = {ATOM_VOUCHER: ["transfer/channel-0"], "ibc/AAAA": ["transfer/channel-52"]}
        build_sweep_batches.check_denom_destination(denom="stuatom", traces=traces)
        build_sweep_batches.check_denom_destination(denom=ATOM_VOUCHER, traces=traces)

        # The same check guards a voucher given through --denoms, not only --extra-denom
        with self.assertRaises(ValueError) as raised:
            build_sweep_batches.check_denom_destination(denom="ibc/AAAA", traces=traces)
        self.assertIn("ibc/AAAA", str(raised.exception))
        self.assertIn("channel-52", str(raised.exception))

        with self.assertRaises(ValueError):
            build_sweep_batches.check_denom_destination(denom="ibc/BBBB", traces=traces)  # no trace at all

    def test_batch_size_is_bounded_by_the_chain_maximum(self) -> None:
        self.assertEqual(build_sweep_batches.batch_size_arg("100"), 100)
        self.assertEqual(build_sweep_batches.batch_size_arg("1"), 1)
        with self.assertRaises(argparse.ArgumentTypeError):
            build_sweep_batches.batch_size_arg("101")
        with self.assertRaises(argparse.ArgumentTypeError):
            build_sweep_batches.batch_size_arg("0")

        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom"], prices=PRICES, floor_usd=1.0)
        with self.assertRaises(ValueError):
            build_sweep_batches.write_batches(plan=plan, out_dir=pathlib.Path(self.tmp.name) / "out", batch_size=101)

    def test_bech32_round_trip_matches_the_on_chain_derivation(self) -> None:
        # The stride and osmo forms of the F5 key are the same 20 bytes under two prefixes
        hrp, address = bech32_ref.decode("stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh")
        self.assertEqual(hrp, "stride")
        self.assertEqual(len(address), 20)
        self.assertEqual(bech32_ref.encode("osmo", address), "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af")
        self.assertEqual(bech32_ref.encode("stride", address), "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh")
        with self.assertRaises(ValueError):
            bech32_ref.decode("stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlx")  # bad checksum


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd scripts/wind-down && python3 -m unittest test_build_sweep_batches.py 2>&1 | tail -3`
Expected: `ModuleNotFoundError: No module named 'bech32_ref'`.

- [ ] **Step 3: Vendor the bech32 reference implementation**

There is no bech32 package in the environment (`python3 -c 'import bech32'` fails) and the
script must run in a clean checkout, so the BIP-173 reference implementation is vendored
verbatim with two thin helpers for cosmos addresses (no segwit witness version byte).

`scripts/wind-down/bech32_ref.py`:

```python
"""Bech32 reference implementation (BIP-173), vendored for the wind-down scripts.

Copyright (c) 2017 Pieter Wuille. Licensed under the MIT License
(https://github.com/sipa/bech32/blob/master/ref/python/segwit_addr.py). The functions named
bech32_* and convertbits are the reference code unchanged; encode/decode are thin helpers for
cosmos-style addresses, which carry raw address bytes and no witness version.
"""

CHARSET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"


def bech32_polymod(values: list[int]) -> int:
    """Internal function that computes the Bech32 checksum."""
    generator = [0x3B6A57B2, 0x26508E6D, 0x1EA119FA, 0x3D4233DD, 0x2A1462B3]
    chk = 1
    for value in values:
        top = chk >> 25
        chk = (chk & 0x1FFFFFF) << 5 ^ value
        for i in range(5):
            chk ^= generator[i] if ((top >> i) & 1) else 0
    return chk


def bech32_hrp_expand(hrp: str) -> list[int]:
    """Expand the HRP into values for checksum computation."""
    return [ord(x) >> 5 for x in hrp] + [0] + [ord(x) & 31 for x in hrp]


def bech32_verify_checksum(hrp: str, data: list[int]) -> bool:
    """Verify a checksum given HRP and converted data characters."""
    return bech32_polymod(bech32_hrp_expand(hrp) + data) == 1


def bech32_create_checksum(hrp: str, data: list[int]) -> list[int]:
    """Compute the checksum values given HRP and data."""
    values = bech32_hrp_expand(hrp) + data
    polymod = bech32_polymod(values + [0, 0, 0, 0, 0, 0]) ^ 1
    return [(polymod >> 5 * (5 - i)) & 31 for i in range(6)]


def bech32_encode(hrp: str, data: list[int]) -> str:
    """Compute a Bech32 string given HRP and data values."""
    combined = data + bech32_create_checksum(hrp, data)
    return hrp + "1" + "".join([CHARSET[d] for d in combined])


def bech32_decode(bech: str) -> tuple[str | None, list[int] | None]:
    """Validate a Bech32 string, and determine HRP and data."""
    if (any(ord(x) < 33 or ord(x) > 126 for x in bech)) or (bech.lower() != bech and bech.upper() != bech):
        return (None, None)
    bech = bech.lower()
    pos = bech.rfind("1")
    if pos < 1 or pos + 7 > len(bech) or len(bech) > 90:
        return (None, None)
    if not all(x in CHARSET for x in bech[pos + 1 :]):
        return (None, None)
    hrp = bech[:pos]
    data = [CHARSET.find(x) for x in bech[pos + 1 :]]
    if not bech32_verify_checksum(hrp, data):
        return (None, None)
    return (hrp, data[:-6])


def convertbits(data: list[int] | bytes, frombits: int, tobits: int, pad: bool = True) -> list[int] | None:
    """General power-of-2 base conversion."""
    acc = 0
    bits = 0
    ret = []
    maxv = (1 << tobits) - 1
    max_acc = (1 << (frombits + tobits - 1)) - 1
    for value in data:
        if value < 0 or (value >> frombits):
            return None
        acc = ((acc << frombits) | value) & max_acc
        bits += frombits
        while bits >= tobits:
            bits -= tobits
            ret.append((acc >> bits) & maxv)
    if pad:
        if bits:
            ret.append((acc << (tobits - bits)) & maxv)
    elif bits >= frombits or ((acc << (tobits - bits)) & maxv):
        return None
    return ret


def encode(hrp: str, address: bytes) -> str:
    """Bech32-encode raw cosmos address bytes under the given prefix."""
    data = convertbits(address, 8, 5)
    if data is None:
        raise ValueError(f"cannot convert {len(address)} address bytes to base32")
    return bech32_encode(hrp, data)


def decode(bech: str) -> tuple[str, bytes]:
    """Decode a cosmos bech32 address into (prefix, raw bytes); raises ValueError when invalid."""
    hrp, data = bech32_decode(bech)
    if hrp is None or data is None:
        raise ValueError(f"invalid bech32 string {bech!r}")
    converted = convertbits(data, 5, 8, False)
    if converted is None:
        raise ValueError(f"invalid bech32 payload in {bech!r}")
    return hrp, bytes(converted)
```

The 90-character limit of the reference decoder is fine here: a 32-byte stride address is 65
characters. Run: `cd scripts/wind-down && python3 -m unittest test_build_sweep_batches.py 2>&1 | tail -3`
Expected: `ModuleNotFoundError: No module named 'build_sweep_batches'` (the bech32 import now resolves).

- [ ] **Step 4: Write the script**

`scripts/wind-down/build_sweep_batches.py`:

```python
#!/usr/bin/env python3
"""Build MsgSweepTokensOffStride batches from a strided export.

Mirrors the on-chain skip rules of x/stakeibc/keeper/wind_down_sweep.go (20-byte address, account
type, transfer escrow exclusion) and adds the off-chain USD floor, so a batch this script emits
should skip nothing on chain. Holders are ordered by the USD value of the listed denoms they hold
and split into files of at most --batch-size addresses, one file per tx for
`strided tx stakeibc sweep-tokens-off-stride DENOMS FILE`.

    python3 scripts/wind-down/build_sweep_batches.py \
        --export export.json --prices prices.json --floor-usd 100 \
        --denoms stuatom,stuosmo,stutia,ustrd --out-dir sweep-batches \
        [--extra-denom ibc/27394F...=10.5:6] [--batch-size 100]

prices.json: {"stuatom": {"usd_per_token": 4.2, "decimals": 6}, ...}. An --extra-denom that is an
ibc/ voucher must have its outermost hop in UNWIND_CHANNELS (the export's denom traces are read
from app_state.transfer.denoms when present).
"""

import argparse
import dataclasses
import hashlib
import json
import pathlib
from decimal import Decimal

import bech32_ref

# Mirror of types.MaxSweepAddressesPerTx; a batch above it fails ValidateBasic on chain
MAX_SWEEP_ADDRESSES_PER_TX = 100
BATCH_SIZE_DEFAULT = MAX_SWEEP_ADDRESSES_PER_TX
ADDRESS_LENGTH_BYTES = 20
TRANSFER_PORT = "transfer"
IBC_PREFIX = "ibc/"

# Mirror of types.SweepUnwindChannels; keep in sync with x/stakeibc/types/wind_down.go
UNWIND_CHANNELS = {
    "channel-0": "cosmos",
    "channel-162": "celestia",
    "channel-5": "osmo",
    "channel-24": "juno",
    "channel-150": "somm",
    "channel-213": "saga",
    "channel-160": "dydx",
}

SWEEPABLE_ACCOUNT_TYPES = {
    "/cosmos.auth.v1beta1.BaseAccount",
    "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
    "/cosmos.vesting.v1beta1.DelayedVestingAccount",
    "/cosmos.vesting.v1beta1.PeriodicVestingAccount",
    "/stride.vesting.StridePeriodicVestingAccount",
}


@dataclasses.dataclass
class Export:
    balances: dict[str, dict[str, int]]  # address -> denom -> amount
    account_types: dict[str, str]  # address -> @type
    escrow_addresses: set[str]
    denom_traces: dict[str, list[str]]  # ibc/HASH -> ["transfer/channel-x", ...] outermost first


@dataclasses.dataclass
class Holder:
    address: str
    usd: Decimal
    balances: dict[str, int]


@dataclasses.dataclass
class Skipped:
    address: str
    reason: str
    usd: Decimal


@dataclasses.dataclass
class HolderPlan:
    denoms: list[str]
    holders: list[Holder]
    skipped: list[Skipped]


@dataclasses.dataclass
class ExtraDenom:
    denom: str
    usd_per_token: float
    decimals: int


def main() -> None:
    args = parse_args()
    export = load_export(args.export)
    prices = json.loads(args.prices.read_text())
    denoms = [denom for denom in args.denoms.split(",") if denom]

    for spec in args.extra_denom:
        extra = parse_extra_denom(spec)
        prices[extra.denom] = {"usd_per_token": extra.usd_per_token, "decimals": extra.decimals}
        denoms.append(extra.denom)

    # Every denom on the final list must have an on-chain destination, however it got there:
    # an ibc/ voucher passed through --denoms is checked exactly like an --extra-denom one
    for denom in denoms:
        check_denom_destination(denom=denom, traces=export.denom_traces)

    plan = classify_holders(export=export, denoms=denoms, prices=prices, floor_usd=args.floor_usd)
    files = write_batches(plan=plan, out_dir=args.out_dir, batch_size=args.batch_size)

    total_usd = sum((holder.usd for holder in plan.holders), Decimal(0))
    print(f"{len(plan.holders)} holders in {len(files)} batches, ${total_usd:.2f} swept, {len(plan.skipped)} skipped")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--export", type=pathlib.Path, required=True)
    parser.add_argument("--prices", type=pathlib.Path, required=True)
    parser.add_argument("--denoms", required=True, help="comma-separated denoms to sweep")
    parser.add_argument("--floor-usd", type=float, required=True)
    parser.add_argument("--out-dir", type=pathlib.Path, required=True)
    parser.add_argument("--batch-size", type=batch_size_arg, default=BATCH_SIZE_DEFAULT, help=f"1..{MAX_SWEEP_ADDRESSES_PER_TX}")
    parser.add_argument("--extra-denom", action="append", default=[], help="DENOM=USD_PER_TOKEN:DECIMALS")
    return parser.parse_args()


def batch_size_arg(text: str) -> int:
    """argparse type for --batch-size: the chain rejects a tx with more than MAX_SWEEP_ADDRESSES_PER_TX addresses."""
    value = int(text)
    if value < 1 or value > MAX_SWEEP_ADDRESSES_PER_TX:
        raise argparse.ArgumentTypeError(f"batch size must be between 1 and {MAX_SWEEP_ADDRESSES_PER_TX}, got {value}")
    return value


def load_export(path: pathlib.Path) -> Export:
    app_state = json.loads(path.read_text())["app_state"]

    balances: dict[str, dict[str, int]] = {}
    for entry in app_state["bank"]["balances"]:
        balances[entry["address"]] = {coin["denom"]: int(coin["amount"]) for coin in entry["coins"]}

    account_types = {account_address(account): account["@type"] for account in app_state["auth"]["accounts"]}

    channels = app_state.get("ibc", {}).get("channel_genesis", {}).get("channels", [])
    escrows = {escrow_address(ch["port_id"], ch["channel_id"]) for ch in channels if ch["port_id"] == TRANSFER_PORT}

    traces: dict[str, list[str]] = {}
    for denom in app_state.get("transfer", {}).get("denoms", []):
        hops = [f"{hop['port_id']}/{hop['channel_id']}" for hop in denom.get("trace", [])]
        traces[ibc_denom(denom["base"], hops)] = hops

    return Export(balances=balances, account_types=account_types, escrow_addresses=escrows, denom_traces=traces)


def classify_holders(export: Export, denoms: list[str], prices: dict, floor_usd: float) -> HolderPlan:
    floor = Decimal(str(floor_usd))
    holders: list[Holder] = []
    skipped: list[Skipped] = []

    for address, coins in export.balances.items():
        listed = {denom: amount for denom, amount in coins.items() if denom in denoms and amount > 0}
        if not listed:
            continue
        usd = sum((usd_value(denom, amount, prices) for denom, amount in listed.items()), Decimal(0))

        reason = skip_reason(address=address, export=export)
        if reason is not None:
            skipped.append(Skipped(address=address, reason=reason, usd=usd))
            continue
        if usd < floor:
            skipped.append(Skipped(address=address, reason=f"below floor (${usd:.2f} < ${floor:.2f})", usd=usd))
            continue
        holders.append(Holder(address=address, usd=usd, balances=listed))

    holders.sort(key=lambda holder: holder.usd, reverse=True)
    return HolderPlan(denoms=denoms, holders=holders, skipped=skipped)


def skip_reason(address: str, export: Export) -> str | None:
    """The on-chain rules, in the on-chain order; None means sweepable."""
    if len(address_bytes(address)) != ADDRESS_LENGTH_BYTES:
        return "address is not 20 bytes"
    account_type = export.account_types.get(address)
    if account_type is None:
        return "account not found"
    if address in export.escrow_addresses:
        return "transfer escrow address"
    if account_type not in SWEEPABLE_ACCOUNT_TYPES:
        return f"account type {account_type} is not sweepable"
    return None


def write_batches(plan: HolderPlan, out_dir: pathlib.Path, batch_size: int) -> list[pathlib.Path]:
    # Guarded here too so a caller that bypasses parse_args cannot emit a batch the chain rejects
    if batch_size < 1 or batch_size > MAX_SWEEP_ADDRESSES_PER_TX:
        raise ValueError(f"batch size must be between 1 and {MAX_SWEEP_ADDRESSES_PER_TX}, got {batch_size}")
    out_dir.mkdir(parents=True, exist_ok=True)
    files: list[pathlib.Path] = []
    batches: list[dict] = []

    for index in range(0, len(plan.holders), batch_size):
        batch = plan.holders[index : index + batch_size]
        path = out_dir / f"batch-{len(files) + 1:03d}.txt"
        path.write_text("".join(f"{holder.address}\n" for holder in batch))
        files.append(path)
        batches.append({
            "file": path.name,
            "num_addresses": len(batch),
            "usd": f"{sum((holder.usd for holder in batch), Decimal(0)):.2f}",
        })

    summary = {
        "denoms": plan.denoms,
        "batches": batches,
        "skipped": [{"address": entry.address, "reason": entry.reason, "usd": f"{entry.usd:.2f}"} for entry in plan.skipped],
    }
    (out_dir / "summary.json").write_text(json.dumps(summary, indent=2))
    return files


def parse_extra_denom(spec: str) -> ExtraDenom:
    if "=" not in spec or ":" not in spec.split("=", 1)[1]:
        raise ValueError(f"extra denom must be DENOM=USD_PER_TOKEN:DECIMALS, got {spec!r}")
    denom, rest = spec.split("=", 1)
    usd_per_token, decimals = rest.split(":", 1)
    return ExtraDenom(denom=denom, usd_per_token=float(usd_per_token), decimals=int(decimals))


def check_denom_destination(denom: str, traces: dict[str, list[str]]) -> None:
    """Mirror of resolveSweepDestination: a native denom always has one (channel-5), an ibc/ denom
    only if its outermost hop is a whitelisted unwind channel. Raises ValueError naming the denom."""
    if not denom.startswith(IBC_PREFIX):
        return
    hops = traces.get(denom)
    if not hops:
        raise ValueError(f"{denom} has no denom trace in the export")
    outer_channel = hops[0].split("/")[1]
    if outer_channel not in UNWIND_CHANNELS:
        raise ValueError(f"{denom} arrived over {outer_channel}, which is not in UNWIND_CHANNELS")


def usd_value(denom: str, amount: int, prices: dict) -> Decimal:
    price = prices.get(denom)
    if price is None:
        raise KeyError(f"no price for {denom}")
    tokens = Decimal(amount) / (Decimal(10) ** int(price["decimals"]))
    return tokens * Decimal(str(price["usd_per_token"]))


def account_address(account: dict) -> str:
    if "address" in account:
        return account["address"]
    if "base_account" in account:
        return account["base_account"]["address"]
    return account["base_vesting_account"]["base_account"]["address"]


def address_bytes(address: str) -> bytes:
    """The raw bytes behind a bech32 address, or b"" when the string is not valid bech32 (which
    the 20-byte rule then rejects, matching the on-chain decode failure)."""
    _, data = bech32_ref.bech32_decode(address)
    if data is None:
        return b""
    converted = bech32_ref.convertbits(data, 5, 8, False)
    return bytes(converted) if converted is not None else b""


def escrow_address(port_id: str, channel_id: str) -> str:
    """ibc-go transfertypes.GetEscrowAddress: sha256("ics20-1\\0" + port/channel)[:20], bech32 stride."""
    preimage = b"ics20-1\x00" + f"{port_id}/{channel_id}".encode()
    digest = hashlib.sha256(preimage).digest()[:ADDRESS_LENGTH_BYTES]
    return bech32_ref.encode("stride", digest)


def ibc_denom(base: str, hops: list[str]) -> str:
    full = "/".join(hops + [base]) if hops else base
    return IBC_PREFIX + hashlib.sha256(full.encode()).hexdigest().upper()


if __name__ == "__main__":
    main()
```

Verify `escrow_address("transfer", "channel-5")` against the chain before use: `strided q bank balances $(python3 -c 'import build_sweep_batches as b; print(b.escrow_address("transfer","channel-5"))')` should show the channel-5 stToken escrows from `docs/wind-down/sttoken-locations.md`.

- [ ] **Step 5: Run to verify it passes**

Run: `cd scripts/wind-down && python3 -m unittest test_build_sweep_batches.py -v`
Expected: 7 tests, `OK`. Then, from a clean shell with no site-packages bech32, `python3 -c 'import bech32' ; echo exit=$?` prints a `ModuleNotFoundError` and `exit=1`, while `cd scripts/wind-down && python3 build_sweep_batches.py --help` prints the usage: the script has no dependency outside the checkout.

- [ ] **Step 6: Commit**

```bash
git add scripts/wind-down/bech32_ref.py scripts/wind-down/build_sweep_batches.py scripts/wind-down/test_build_sweep_batches.py
git commit -m "ops: build_sweep_batches.py mirrors the on-chain sweep skip rules and bounds with a USD floor

Vendors the BIP-173 bech32 reference implementation so the script runs in a clean checkout.

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Verification before the PR

- [ ] `go build ./...` and `go vet ./x/stakeibc/...` clean.
- [ ] `go test ./x/stakeibc/... ./app/...` green (the pre-existing `utils` `TestCreateModuleAccount` failure on main is not in this set).
- [ ] `make lint` clean for the touched files.
- [ ] Gas: on localstride, submit one batch of 100 funded base accounts with three denoms each and record the gas used in the PR description; if it exceeds a comfortable fraction of the block gas limit, lower `MaxSweepAddressesPerTx` in a one-line follow-up in `wind_down.go` (a PR 4 file) and say so in the PR (spec §7 allows the bound to move after measurement).
- [ ] PR description lists the skip rules and destination rules verbatim from §7 so the reviewer checks the code against the spec, not against the plan.

## Self-review

- **Spec coverage.** §7 sweep: denom list, address bound, destination resolution per denom before any address (Task 2/3), the four skip reasons and skip-does-not-move (Tasks 2/3), silent zero balance, full-balance `MsgTransfer` with the derived receiver, one-day timeout, empty memo, transfer error rejects the tx, refund on timeout (Task 3 tests). §11 sweep list: every bullet maps to a named test in Task 3 or Task 1/2 (`batch over the bound`, `invalid denom string`, `empty denom list` are `ValidateBasic` tests in Task 1). §13 ops script: `build_sweep_batches.py` with `--extra-denom`, the whitelist refusal applied to every denom on the final list, and the 100-address bound on `--batch-size` (Task 5). The "transfer error rejecting the whole tx" bullet of §11 is proven through a cache boundary with balance, sequence, commitment and event assertions (Task 3). CLI (Task 4).
- **Placeholders.** None: every code step carries its code, including the vendored `bech32_ref.py`, so the script has no dependency outside the checkout.
- **Type consistency.** `sweepDestination{ChannelId, Bech32Prefix}` used identically in Tasks 2 and 3; `SweepTokensOffStride` returns `(uint64, uint64, error)` and the handler maps to `NumTransfers/NumSkipped`, the response field names from the PR 4 proto; error and event names match between `types` and the tests.
- **Review tags.** Tasks 1-3 `Review: yes` (auth gate and user funds), Tasks 4-5 `Review: no`.
- **Out of scope, noted for the parent.** The tests open extra transfer channels rather than overriding `StrideToOsmosisTransferChannelId`; if PR 4 ends up defining that constant as a `var`, the helper can be dropped and the tests can point the constant at `channel-0`, but the plan does not depend on it.
