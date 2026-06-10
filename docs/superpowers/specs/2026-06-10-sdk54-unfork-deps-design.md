# SDK 54 — Un-fork 08-wasm, ibc-hooks, and PFM

## §1. Overview and objective

The `sdk54` branch (PR #1495, base of the v33 PoA migration stack) currently
points three dependencies at Stride-Labs forks / pre-release tags because their
official upstream releases did not exist when the branch was cut. Three of those
are now resolvable:

- **08-wasm** — official `modules/light-clients/08-wasm/v11.1.0` exists and uses
  `wasmvm/v3 v3.0.5`, the exact wasmvm v2→v3 bump the fork
  (`v11.0.0-wasmvm-v3-rc0`) was created for.
- **ibc-hooks** — official `modules/ibc-hooks/v11.0.0` exists.
- **packet-forward-middleware (PFM)** — moved *into core ibc-go* as
  `github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware`. It is
  part of the main `ibc-go/v11` module (no separate `go.mod`), so it comes free
  with the ibc-go core bump — the separate `ibc-apps` require/replace is dropped
  entirely.

**Objective:** drop those three fork `replace` directives on `sdk54`, bump
`ibc-go/v11` core to `v11.1.0`, migrate the PFM wiring to the core API, and get
the branch building + green. Nothing else changes.

**Scope (explicitly): just the dependency surgery on `sdk54`.** Propagating the
change up the PR stack (`poa-migration` → scripts → rehearsal → ig-tests) is a
mechanical rebase handled separately, not part of this work.

## §2. What stays forked

Three `replace` directives are **not** touched:

- **rate-limiting** (`Stride-Labs/ibc-apps/.../rate-limiting/v11
  v11.0.0-stride-rc0`) — upstream `ibc-apps` has no v11 tag yet (latest is
  `rate-limiting/v10.1.0`). Expected in the ibc-apps **v11.2** release. This is
  the lone remaining un-fork; we will bump again when it ships.
- **ics23** (`Stride-Labs/ics23/go ...-non-membership-icq-rc4`) — permanent
  Stride patch for empty-value-leaf ICQ non-membership proofs (tracks
  cosmos/ics23#134). Not release-gated.
- **vesting** (`Stride-Labs/vesting ...`) — permanent Stride fork.

The big TODO comment block in `go.mod` is trimmed to describe only the
rate-limiting fork and its v11.2 unblock condition.

## §3. go.mod changes

**Require block:**

| Module | From | To |
|---|---|---|
| `github.com/cosmos/ibc-go/v11` | `v11.0.0` | `v11.1.0` |
| `github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v11` | `v11.0.0` | `v11.1.0` |
| `github.com/cosmos/ibc-apps/modules/ibc-hooks/v11` | `v11.0.0` | `v11.0.0` (unchanged; now resolves to official tag) |
| `github.com/CosmWasm/wasmvm/v3` | `v3.0.4` | `v3.0.5` (pulled by 08-wasm v11.1.0; `go mod tidy` settles) |
| `github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v11` | `v11.0.0` | **removed** (now in core ibc-go) |

**Replace block — drop three, keep three:**

- Remove: 08-wasm fork, ibc-hooks fork, PFM fork.
- Keep: rate-limiting fork, ics23 fork, vesting fork (see §2).

## §4. Import path changes (PFM only)

08-wasm and ibc-hooks import paths are *already* the canonical `cosmos/...`
paths — they were only being `replace`d — so removing the replaces requires
**no code change** for those two.

Only PFM changes repos:

```
github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v11/packetforward         → github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware
github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v11/packetforward/keeper  → github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware/keeper
github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v11/packetforward/types   → github.com/cosmos/ibc-go/v11/modules/apps/packet-forward-middleware/types
```

Known affected files: `app/app.go` (3 imports), `app/upgrades.go` (1 import).
A full `grep` sweep for `packet-forward-middleware` across `*.go` (including
tests) is part of implementation; the package alias `packetforward` /
`packetforwardkeeper` / `packetforwardtypes` and exported names are unchanged,
so only the import lines move.

## §5. PFM wiring migration (`app/app.go`)

Core ibc-go v11's PFM has a different API than the `ibc-apps` fork. The
governing constraint: **Stride's transfer-middleware stack behavior must be
unchanged.** Stride has many more middlewares than ibc-go's reference simapp, so
we deliberately do *not* adopt simapp's `.Next()`/router construction — we keep
the existing manual middleware wrap and change only the PFM node.

### §5.1 Keeper construction — flip order, new signature

Today (fork API), the PFM keeper is built *before* the transfer keeper with a
`nil` transfer keeper resolved later:

```go
app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
    appCodec, storeService,
    nil,                  // transfer keeper set later
    channelKeeper, bankKeeper,
    app.HooksICS4Wrapper, // ICS4Wrapper
    authority,
)
// ... TransferKeeper created ...
app.TransferKeeper.WithICS4Wrapper(app.PacketForwardKeeper)
app.PacketForwardKeeper.SetTransferKeeper(app.TransferKeeper)  // deferred resolve
```

Core ibc-go's `NewKeeper` takes the transfer keeper at construction and has no
`SetTransferKeeper`. So the order **flips** — `TransferKeeper` is built first,
then the PFM keeper:

```go
// TransferKeeper constructed first (ICS4Wrapper defaults to ChannelKeeper)
app.TransferKeeper = ibctransferkeeper.NewKeeper(...)

app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
    appCodec,
    app.AccountKeeper.AddressCodec(),  // NEW arg
    runtime.NewKVStoreService(keys[packetforwardtypes.StoreKey]),
    app.TransferKeeper,                // now passed at construction
    app.IBCKeeper.ChannelKeeper,
    app.BankKeeper,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
// Preserve old behavior: PFM's outbound ICS4Wrapper was HooksICS4Wrapper.
app.PacketForwardKeeper.WithICS4Wrapper(app.HooksICS4Wrapper)
// Outbound transfer sends still route up through PFM (unchanged).
app.TransferKeeper.WithICS4Wrapper(app.PacketForwardKeeper)
```

Deletions: the `nil` placeholder, the `SetTransferKeeper(...)` call, and the
"PFM Keeper must be initialized before the Transfer Keeper" comment (now
inverted). The `TransferKeeper.WithICS4Wrapper(PacketForwardKeeper)` override is
kept verbatim — the `pfm → ibchooks → ratelimit → core IBC` outbound chain is
unchanged.

**Circular-dependency note:** TransferKeeper↔PFM is not actually circular under
the new API. `ibctransferkeeper.NewKeeper` accepts a default ICS4Wrapper
(ChannelKeeper) at construction and is overridden to PFM afterward, so it does
not need the PFM keeper at construction time — only PFM needs the transfer
keeper, which now exists first.

### §5.2 Middleware stack — preserve order, swap PFM node only

The stack at `app/app.go` (`transferStack = ...`) is preserved in exact order:
`pfm → ibchooks → ratelimit → staketia → stakedym → records → autopilot`. Only
the PFM node changes from the fork's 4-arg wrap to core's
keeper-only constructor plus `SetUnderlyingApplication`:

```go
var transferStack porttypes.IBCModule = transfer.NewIBCModule(app.TransferKeeper)

pfm := packetforward.NewIBCMiddleware(
    app.PacketForwardKeeper,
    0,                                                                // retries on timeout (unchanged)
    packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp, // forward timeout (unchanged)
)
pfm.SetUnderlyingApplication(transferStack)
transferStack = pfm

transferStack = ibchooks.NewIBCMiddleware(transferStack, &app.HooksICS4Wrapper)
transferStack = ratelimit.NewIBCMiddleware(&app.RatelimitKeeper, transferStack)
// ... staketia, stakedym, records, autopilot unchanged ...
```

`SetUnderlyingApplication` is the `porttypes.Middleware` (v11) method that slots
`transferStack` beneath PFM — exactly the position the fork's first constructor
arg occupied. Net effect on the stack is identity.

### §5.3 AppModule + params subspace — match the ratelimit precedent

Core's `packetforward.NewAppModule(keeper)` drops the legacy params subspace
arg. Per existing convention in this codebase — `ratelimit.NewAppModule(appCodec,
&keeper)` takes no subspace yet `paramsKeeper.Subspace(ratelimittypes.ModuleName)`
remains registered — PFM matches:

- Change `packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(...))`
  → `packetforward.NewAppModule(app.PacketForwardKeeper)`.
- **Keep** `paramsKeeper.Subspace(packetforwardtypes.ModuleName)` registered.

No params migration behavior changes — the subspace registration is retained
identically to the other modules awaiting the scheduled x/params removal.

## §6. Verification

Local bar (matches the branch's existing CI):

1. `go mod tidy` — clean, no unexpected churn beyond the four require lines and
   the dropped PFM require.
2. `go build ./...` — compiles.
3. `go test ./app/...` — full app suite green (this is where any PFM wiring
   regression surfaces, since the change is behavioral only in the transfer
   stack).

Then push and let existing CI (Go Lint, Gosec, Unit Tests, cross-arch builds,
`Verify poa.go` / `Verify admins.go`) confirm. End-to-end PFM forwarding is an
integration concern covered downstream by the rehearsal / ig-test branches and
is intentionally **not** a gate on this dependency bump.

## §7. Non-goals

- Un-forking rate-limiting (blocked on ibc-apps v11.2).
- Touching the ics23 or vesting forks (permanent).
- Propagating the change up the PR stack (separate mechanical rebase).
- Bumping the `stride/v32` → `v33` module path (out of scope; tracked
  separately).
- Any change to the transfer-middleware stack behavior or ordering.

## §8. Follow-up

When `ibc-apps` publishes rate-limiting v11 (expected v11.2), drop the last fork
`replace`, bump the require, and re-run §6. That is the only remaining
dependency un-fork after this work.
