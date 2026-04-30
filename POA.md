# POA Migration — Context Primer

This document is **background reading** for anyone reviewing the v33 ICS → POA migration plan or its implementation. It does not describe the migration itself; for that, see `docs/superpowers/specs/2026-04-27-ics-to-poa-migration-design.md`.

The goal here is to take a reviewer from "vaguely familiar with Cosmos SDK chains" to "able to evaluate the migration plan with the rigor of someone who wrote each module being touched." It covers:

1. How a default Cosmos SDK validator chain works
2. How Interchain Security (ICS) reshapes that
3. What Stride's setup specifically looks like today
4. What the new POA module is and how it fits
5. SDK upgrade mechanics (VersionMap, store keys, panic conditions)
6. What actually changes in the migration

If you're already deeply familiar with Cosmos SDK, ICS, and POA, skip to §6.

---

## 1. Default Cosmos SDK validator stack

A "normal" Cosmos SDK chain (Cosmos Hub, Osmosis, Juno) runs validators and delegators via a tightly-coupled stack of seven modules. The crucial mental model:

> **CometBFT's validator set is whatever `x/staking`'s `EndBlock` returns.**

Everything else hangs off that.

### x/staking — the validator + delegation system

- **Validators** register with `MsgCreateValidator`, providing a consensus pubkey, an operator address, a self-delegation, and a commission rate.
- **Delegators** bond their tokens to validators with `MsgDelegate`. Delegated tokens determine validator voting power.
- Two module accounts hold all bonded/unbonded tokens: `bonded_tokens_pool` (live stake) and `not_bonded_tokens_pool` (in unbonding period or slashed-but-not-yet-burned).
- Each block, `x/staking.EndBlock` ranks validators by power, applies the active-set cap, and returns `[]abci.ValidatorUpdate` to CometBFT. **This is the only authoritative source of CometBFT validator changes** in a default chain.
- Unbonding delegations sit in the `not_bonded_tokens_pool` for `UnbondingPeriod` (typically 21 days), then are released back to delegators.
- Hooks on validator/delegation lifecycle (e.g., `AfterValidatorBonded`, `BeforeDelegationCreated`) drive integrations with x/distribution and x/slashing.

### x/distribution — reward distribution

- Watches the `fee_collector` module account, which receives:
  - Tx fees from the ante handler (every transaction's fee).
  - Inflation from `x/mint` each epoch/block.
- Each `BeginBlock`, allocates the `fee_collector` balance to validators **proportional to their voting power** during the previous block.
- Splits each validator's allocation: validator's commission → operator account; remainder → distributed to that validator's delegators (claimable via `MsgWithdrawDelegatorReward`).
- Maintains the **community pool** as a fund that any address can deposit to; spending requires governance approval.
- Listens to staking hooks to track per-delegator reward state across power changes.

### x/slashing — punishment for downtime and equivocation

- Tracks per-validator **signing info**: how many of the recent N blocks each validator signed.
- **Downtime jailing**: if a validator misses too many blocks, calls `Slash` (burning some bonded stake) and `Jail` (removing them from the active set until they unjail).
- **Equivocation**: when notified by `x/evidence` of double-sign, slashes a much larger fraction (typically 5%) and jails permanently.
- Implements `BeforeValidatorRemoved`, `AfterValidatorBonded` etc. as hooks on x/staking — that's how it tracks signing info correctly.

### x/evidence — routing equivocation evidence

- CometBFT detects double-signs (a validator signing two different blocks at the same height) and forwards the evidence as part of block headers.
- `x/evidence` routes that evidence to `x/slashing` for actual punishment.
- Has its own KV store but minimal logic.

### x/gov — governance

- Voting power for proposal tallying is computed from **bonded stake**: a delegator with X tokens delegated to a validator gets X voting power.
- Asks `x/staking` for the current bonded set + each validator's delegations to compute tallies.
- Proposal lifecycle: deposit period → voting period → tallied (passed/rejected/failed).

### x/mint — inflation

- Each block (or each epoch, in custom mint modules), mints new tokens at some configured rate.
- Sends them to the `fee_collector` so they flow through `x/distribution` to validators + delegators.
- The "staking inflation rate" knob is what most chains tune.

A useful distinction: **inflation** and **staking rewards** are not synonymous, even on a default chain. Inflation is *everything* `x/mint` produces. Staking rewards are what eventually reaches validators and delegators — which on a default chain happens to be all of inflation plus tx fees, both flowing through `fee_collector`. On a chain with a custom mint that splits inflation across multiple destinations (Stride is one — see §3), only the portion routed to `fee_collector` becomes staking rewards; the rest goes elsewhere (community pools, treasury, etc.).

### How CometBFT uses the validator set

CometBFT keeps an internal table of `(pubkey, power)` tuples — the active validator set. It uses this table for two things:

1. **Verifying incoming precommit signatures.** Each block, every active validator signs a precommit and broadcasts it on the p2p network. CometBFT collects them, verifies each signature against the corresponding pubkey in its validator-set table, and rejects any that don't match a known validator. A block commits when valid precommits totalling **>2/3 of voting power** are collected.
2. **Counting voting power.** A validator's power determines how much weight their precommit carries. If 8 validators have powers `[10, 8, 7, 5, 5, 4, 3, 2]` (total 44), then ≥30 power needs to precommit for a block to commit.

In steady state, **all active validators sign every block.** The >2/3 threshold is a fault-tolerance property, not a sampling design — losing any single validator is fine, losing 1-2 is fine, but losing more than 1/3 of voting power halts the chain. CometBFT tracks per-validator missed-block counts as part of `LastCommitInfo`, which `x/slashing` uses for downtime detection.

### Where consensus keys live

Each validator's signing key is an ed25519 keypair stored in their node's `priv_validator_key.json`. Production validators don't keep this file in plaintext; common arrangements:

- **TMKMS** (Tendermint KMS) — a separate process holds the key and signs over a TCP/Unix socket. Often paired with a hardware backend.
- **Hardware HSM / YubiHSM** — key never leaves a tamper-resistant chip.
- **Remote signer** — the validator node delegates signing to another machine entirely.

Whichever backend, the **private key never moves** — it sits behind the validator operator's signing infrastructure. Public keys are written to chain state via `MsgCreateValidator` (or, on ICS chains, via key-assignment messages on the provider).

### Validator set updates and the one-block ABCI lag

`abci.ValidatorUpdate`s have a **one-block lag**. If `x/staking.EndBlock` returns updates at block N:

- Block N is signed by the *old* validator set.
- Block N+1 is *also* signed by the old set (CometBFT applies updates one block late).
- Block N+2 is the first block signed by the new set.

This lag is critical for the migration. We rely on it to ensure block production doesn't break during the validator-set handoff.

---

## 2. Interchain Security (ICS)

ICS is a Cosmos primitive that lets a "consumer" chain inherit its validator set from a "provider" chain. Stride is a **consumer**; the Cosmos Hub is the **provider**.

### Provider/consumer model

- The provider runs a normal `x/staking` validator set. Validators stake ATOM on the Hub.
- For each consumer chain, the provider sends a stream of **VSC packets** (Validator Set Change packets) over an ordered IBC channel.
- The consumer chain's `x/ccv/consumer` module receives those packets and uses them to drive its own CometBFT validator set.
- **The consumer's `x/staking` does NOT drive CometBFT**, even though it may exist for other purposes.

### What's actually sent over the IBC channel

This is a common point of confusion: **VSC packets carry only the new validator set**, expressed as `(pubkey, power)` tuples. They do not contain any block signatures, attestations, or proofs from Hub validators. The packet is essentially: "consumer chain, here's your new validator-set snapshot."

So Hub and Stride run **two completely independent CometBFT consensus instances**:

- The Hub validates and commits Hub blocks against Hub validators' Hub-side keys.
- Stride validates and commits Stride blocks against the keys in Stride's CometBFT validator-set table — which were populated from VSC packets, so they're the *consumer-side* (assigned) keys.

Each chain's CometBFT independently verifies its own incoming precommits. There is **no cross-chain signature verification** at the consensus layer. The only cross-chain communication is the IBC packet stream:

- Hub → Stride: VSC packets (set updates), reward acknowledgments.
- Stride → Hub: SlashPackets (when Stride detects misbehavior), reward forwarding (15% of fees IBC-transferred).

### How Stride trusts incoming VSC packets

Since the Hub doesn't sign VSC packets directly, how does Stride know they're genuinely from the Hub and not forged? Through **standard IBC light-client verification**, the same trust path every IBC packet uses (token transfers, ICA, etc.):

- Stride runs a Tendermint light client of the Hub. The light client tracks Hub block headers and verifies that each header was signed by >2/3 of the Hub's validator set.
- When a relayer delivers a VSC packet to Stride over IBC, the packet includes a Merkle inclusion proof. Stride's IBC module verifies the proof against a Hub block header that the light client has already accepted.
- Only after that verification does Stride's `ccvconsumer` see the packet contents.

So Stride does verify *something* about Hub-side cryptography — but at the IBC layer (signed Hub headers), not the consensus layer. After v33, no more VSC packets are received, and this whole IBC verification path becomes irrelevant to Stride's consensus.

### How often the validator set changes

Two distinct kinds of "change":

- **Membership changes (which 8 validators are active)**: rare. Stride uses a PSS allowlist that's fixed by Hub governance — adding/removing a validator from Stride's set requires a Hub governance proposal. Likely fewer than 5 changes per year.
- **Power changes among the existing 8**: frequent. Every time someone delegates or undelegates ATOM to one of the 8 validators on the Hub, that validator's power shifts. The provider module batches these changes and emits VSC packets at a configurable cadence (typically every block or two during active periods).

So Stride probably receives multiple VSC packets per hour, but most are just power adjustments — the same 8 validators with shifting weights. For our migration, this means the snapshot we take in the upgrade handler captures whatever set is current at that exact block; after the upgrade, no more VSC packets are processed (the IBC channel goes idle), so the POA validator set is frozen until the multisig changes it.

### What validators actually run

Each PSS validator runs **two separate nodes** today, one for each chain:

- **A Hub node** — `gaiad` binary, syncs Hub state, signs Hub blocks with the Hub-side priv-val key.
- **A Stride node** — `strided` binary, syncs Stride state, signs Stride blocks with the Stride-side priv-val key.

Each node operates as a normal Cosmos validator on its chain. Neither node directly knows about the other or about ICS — the linkage between them is purely on-chain (Hub state knows "validator V has assigned consumer key K_stride for stride-1") and mediated by IBC relayers running between the chains. Validator operators don't relay or coordinate anything manually; they just keep both nodes online and signing.

### x/ccv/consumer — the consumer-side ICS module

- Maintains a `CrossChainValidator` (CCValidator) record per active validator: `{Address (consensus), Pubkey, Power, OptedOut}`.
- Each `EndBlock`:
  - Drains any pending VSC packets that arrived since last block.
  - Returns `[]abci.ValidatorUpdate` to CometBFT — this is what CometBFT picks up.
  - Distributes some portion of `fee_collector` locally (via `cons_redistribute`) and IBC-transfers the rest to the provider's fee pool.
- Maintains the IBC channel to the provider.
- Routes slash packets *back* to the provider when local downtime is detected.

### Key assignment

A subtle but important ICS feature: each provider validator can assign a *different* consensus key to use on each consumer. Done via `MsgAssignConsumerKey` on the provider. This means:

- A validator's Hub-side consensus key can be `K_hub`.
- Their Stride-side consensus key can be `K_stride`.
- VSC packets to Stride carry `K_stride`, not `K_hub`.
- Stride's CometBFT signs blocks with `K_stride`.

What `MsgAssignConsumerKey` actually transmits to the Hub is **only the public key** — never a private key. The flow:

1. Validator runs `strided init ...` on their Stride node, which auto-generates a fresh ed25519 keypair into `priv_validator_key.json`.
2. They extract just the public key with `strided tendermint show-validator`.
3. They submit `gaiad tx provider assign-consensus-key stride-1 <pubkey>` on the Hub. This is a normal Cosmos transaction signed by their Hub operator key. The pubkey gets stored in Hub state as: "this Hub validator will use this pubkey on this consumer chain."
4. From then on, VSC packets to Stride carry `K_stride`'s pubkey for that validator.

The private key never leaves the Stride node. ICS adds no extra exposure relative to running any Cosmos validator.

In practice nearly all ICS validators do use key assignment to keep separate keys on each chain — using the same consensus key on both nodes is bad operational practice (a compromise of one node compromises both, and there's a real double-sign risk if the same priv-val file is accidentally loaded by both binaries).

For our migration, the implication is: **the keys we want to seed POA with are the keys currently in `CrossChainValidator` records on Stride** — those are guaranteed to be the keys CometBFT is actively using, regardless of any Hub-side key assignment.

### PSS (Partial Set Security)

PSS is a *provider-side* knob. It lets the provider opt to send only a subset of its validators to a given consumer:

- `Top_N`: only the top N validators by power are eligible to validate this consumer.
- `Allowlist`: an explicit whitelist of consensus addresses.
- `Denylist`: explicit blocklist.
- `ValidatorSetCap`: maximum number of validators in the consumer's set.

**Stride uses an Allowlist of 8 validators** on the Hub. From Stride's perspective there is no PSS-specific code path — it just receives a smaller validator set in each VSC packet. PSS is invisible to the consumer.

### Slashing across chains: consumer detects, provider punishes

When a validator misbehaves on Stride (downtime or double-sign), the actual bonded stake to slash lives on the Hub — Stride doesn't have local stake to take. ICS handles this by routing punishment back over IBC:

1. Stride's CometBFT detects the infraction (missed blocks → tracked by `LastCommitInfo`; double-sign → routed by `x/evidence`).
2. Stride's `x/slashing` (wired to `ConsumerKeeper`) is notified.
3. `ConsumerKeeper.SlashWithInfractionReason` queues a SlashPacket. `ConsumerKeeper.Jail` is a local no-op.
4. Stride's `ccvconsumer.EndBlock` IBC-sends the SlashPacket to the Hub.
5. Hub's provider module receives it, looks up which Hub validator the consumer key maps to, slashes their actual ATOM stake (~5% for double-sign, smaller for downtime), and jails them.
6. The jailing propagates back to Stride via the next VSC packet (jailed validator gets removed from the set).

This is why ICS is described as providing "shared security" — the consumer borrows the provider's stake for cryptoeconomic security, and misbehavior on the consumer is punishable on the provider side.

Post-POA, this whole pipeline goes away. POA has no slashing integration, `x/slashing` and `x/evidence` are removed from the module manager, and CometBFT-level evidence is detected but routed nowhere. The multisig zero-powers a misbehaving validator manually.

### The "democracy consumer" pattern

ICS provides two specialized wrapper modules for chains that want to run consumer-driven validators *and* keep some local staking-like functionality:

- **`x/ccv/democracy/staking`** (commonly called `ccvstaking`): wraps the standard `x/staking` AppModule. Its only job is to **discard the validator updates** returned by `x/staking.EndBlock`, so they don't fight with `ccvconsumer`'s updates.
- **`x/ccv/democracy/distribution`** (`ccvdistr`): wraps the standard `x/distribution` AppModule, configured to use `cons_redistribute` (the consumer's local share account) as its fee pool instead of `fee_collector` directly.

The idea: a consumer chain can let users delegate locally and earn rewards from the local share of fees, while still having block production driven by the provider's validators. This is the pattern Stride uses.

---

## 3. Stride's actual setup today

### History

- Stride was originally a standalone PoS chain. Validators bonded STRD, signed blocks, earned rewards normally.
- At **v12** (the "changeover" upgrade), Stride became an ICS consumer. The standard ICS changeover handler ran: `ConsumerKeeper.InitGenesis` was called with `PreCCV = true`, and from the next block, CometBFT's validator set was driven by VSC packets from the Hub instead of by Stride's own `x/staking`.
- Crucially, **the v12 changeover did not drain `x/staking`**. The bonded pool, validator records, and delegations all carried over. They just stopped driving CometBFT.
- `app.ConsumerKeeper.SetStandaloneStakingKeeper(app.StakingKeeper)` was set to allow pre-CCV slashing infractions to still route to the legacy staking module.

### Today's architecture (v32)

Stride is an ICS consumer using PSS with 8 whitelisted validators on the Hub. The on-chain layout:

| Module | Role on Stride today |
|---|---|
| `x/ccv/consumer` | Drives CometBFT's validator set from Hub VSC packets |
| `x/staking` | **Active for delegations and govenators**, but its validator updates are suppressed by ccvstaking |
| `x/ccv/democracy/staking` (`ccvstaking`) | Wraps x/staking — discards validator updates so they don't fight ccvconsumer |
| `x/distribution` | Wraps the standard module via `ccvdistr` — fee pool is `cons_redistribute` (the local 85% share) |
| `x/ccv/democracy/distribution` (`ccvdistr`) | Wrapper module for democracy distribution |
| `x/slashing` | Active. Wired to `ConsumerKeeper` (not `StakingKeeper`) — when downtime detected, sends `SlashPacket` back to Hub |
| `x/evidence` | Active. Routes equivocation evidence into slashing |
| `x/gov` | Standard. Tally based on bonded STRD via x/staking |
| `x/mint` | **Custom Stride implementation** with hourly epoch + 4-way split |
| `stakeibc` | Stride's liquid staking product. Mints stTokens, delegates host-zone tokens, collects fees. |
| `x/auction`, `x/strdburner` | Buyback-and-burn pipeline for liquid-staking revenue |

### The "govenator" concept — Stride's most surprising property

A "govenator" is a Stride-specific term for a validator that exists in `x/staking` but doesn't sign blocks. Here's how this works:

- STRD holders can delegate to validators registered in `x/staking` via `MsgDelegate`. Real bonded STRD; real `bonded_tokens_pool`; real delegation records.
- Those validators earn rewards from the local 85% share of inflation/fees, distributed through `ccvdistr` → `x/distribution` → delegator-claimable rewards.
- They don't sign blocks because `ccvstaking` suppresses their validator updates. CometBFT's actual signing set is the 8 ICS validators pushed from the Hub.
- Govenators give STRD holders a way to stake their tokens, earn yield, and (through gov tally) participate in governance — without those validators having to actually run a CometBFT node tied to the Hub.

The two sets are independent on-chain entities:

- **Block-producing set (8 ICS validators):** consensus pubkeys pushed from Hub, signs blocks, gets paid via Hub-side rewards.
- **Govenator set (variable, anyone who registered):** local Stride validators with delegations, gets paid via local share of inflation/fees.

A real-world person/team could be both (running an ICS validator on the Hub side AND a govenator on Stride side), but on-chain they're two separate validator records.

### Stride's custom x/mint — 4-way inflation split

Stride does not use the standard SDK x/mint. Its custom version mints STRD on an **hourly epoch** and splits it 4 ways:

| Portion | % | Recipient |
|---|---|---|
| `Staking` | 27.64% | `fee_collector` (then 85/15 split via ccvconsumer) |
| `StrategicReserve` | 42.05% | Hardcoded address `stride1alnn79kh...` |
| `CommunityPoolGrowth` | 18.60% | Submodule address |
| `CommunityPoolSecurityBudget` | 11.71% | Submodule address |

Only the 27.64% "Staking" portion touches the validator-set machinery. The other ~72% goes to fixed addresses unrelated to staking and is unaffected by the migration.

### Liquid-staking revenue flow (separate from the validator stack)

Stride's core product (liquid staking other chains' tokens) generates its own revenue, completely independent of the validator-set mechanism:

- LS fees from host zones (ATOM, OSMO, TIA, etc.) arrive as IBC denoms in the `RewardCollector` module account.
- `stakeibc.AuctionOffRewardCollectorBalance` runs periodically:
  - 15% (`utils.PoaValPaymentRate`) → liquid-staked → stTokens → split evenly across **8 hardcoded addresses** in `utils.PoaValidatorSet` (these addresses correspond to the validators).
  - 85% → `x/auction` → bidders pay STRD → STRD lands in `strdburner` module account.
- `x/strdburner.EndBlock` burns all STRD it holds each block.

This is the buyback-and-burn flow plus direct validator revenue. **It is independent of `x/staking`, `x/ccv/consumer`, or any consensus module.** Migration doesn't touch it.

A note: `utils.PoaValidatorSet` has 8 entries (which is why we say "8 validators"). The constants are hardcoded with a `DO NOT MODIFY` warning. Long-term this should become a dynamic lookup against POA state, but that's a future cleanup.

### Stride's slashing wiring quirk

In Stride's `app.go`:

```go
app.SlashingKeeper = slashingkeeper.NewKeeper(
    ..., &app.ConsumerKeeper, ...,  // staking keeper interface = ConsumerKeeper, NOT StakingKeeper
)
```

This is unusual. `x/slashing` normally takes the staking keeper as its source of truth for who's a validator. On Stride, it asks `ConsumerKeeper` instead — because the validators that CometBFT is signing with come from ICS, not staking. When slashing detects downtime:

- It calls `ConsumerKeeper.SlashWithInfractionReason`, which queues a `SlashPacket` for the Hub (no local stake to slash).
- It calls `ConsumerKeeper.Jail`, which is a no-op locally (Hub will jail).

For the migration, this matters because we're removing `x/slashing` entirely. POA validators don't get slashed — misbehavior is handled by the multisig zero-powering them.

---

## 4. POA — the destination

POA (Proof of Authority) is a small, opinionated module from `cosmos-sdk/enterprise/poa`. It's the inverse of `x/staking`: deliberately minimal.

### Conceptual differences from staking

| Concept | x/staking | x/poa |
|---|---|---|
| Validator selection | By bonded stake | By admin assignment |
| Delegation | Yes | No |
| Self-bond | Required | None |
| Commission | Per-validator | None |
| Slashing | Downtime + equivocation | None |
| Distribution | Proportional to delegation | Lazy per-validator allocation |
| Admin | Governance via gov module | Configurable admin address |

### State model

The POA keeper stores:

- `Params` — currently just `{Admin: string}`
- `Validators` — indexed map: cons addr → `Validator{PubKey, Power, Metadata{Moniker, Description, OperatorAddress}}`
- `TotalPower` — sum of validator powers
- `TotalAllocatedFees` — sum of allocated fees (DecCoins)
- `QueuedUpdates` — transient store, drained each block as ABCI updates
- `ValidatorAllocatedFees` — per-validator fee balances

That's it. No bonded pools, no commission rates, no unbonding periods, no commission histories.

### Admin model

A single bech32 string controls all mutations. Common choices:

- A multisig address (Stride's choice)
- The gov module address (`authtypes.NewModuleAddress("gov")`) — every change requires a governance proposal
- An EOA — fastest but most centralized

The admin sends three messages:

- `MsgCreateValidator` — adds a validator with pubkey, initial power, metadata
- `MsgUpdateValidators` — batch power/metadata updates (set power to 0 to effectively remove)
- `MsgUpdateParams` — change the admin address

### Validator powers under POA

POA's `Power` field is just an integer voting weight in CometBFT consensus — there's no underlying token bond, no commission share, no economic meaning. Two natural choices for a permissioned chain:

- **Inherit existing powers** from whatever the prior validator-set source (e.g., ICS) had. Useful at upgrade time because a no-diff handoff is the cleanest possible transition — POA simply emits the set CometBFT already has.
- **Equal powers (e.g., all = 1)**. Cleaner story for a permissioned chain: all 8 validators are equally trusted, the >2/3 threshold becomes "any 3 of the 8 colluding compromises consensus," symmetric and easy to reason about.

These aren't mutually exclusive — a chain can preserve current powers at the upgrade and rebalance to equal powers later via the multisig admin. That's the path Stride takes (see §6).

### EndBlock and CometBFT integration

POA's `EndBlock` is two lines:

```go
func (k *Keeper) EndBlocker(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
    if ctx.BlockHeight() == 1 {
        validateFeeRecipient()  // panics if ante handler isn't wired for POA fees
    }
    return k.ReapValidatorUpdates(ctx), nil
}
```

It drains the transient store of pending updates and returns them. No bonding/unbonding logic, no rebalancing, no historical info. Just emit what the admin queued.

### Fee distribution — lazy checkpoint model

POA is opinionated about fees: tx fees go to the POA module account (via a custom ante handler), and they accumulate until either (a) a power change triggers a checkpoint, or (b) a validator calls `MsgWithdrawFees`. At checkpoint, the unallocated balance is split proportional to current power across active validators.

Two consequences:

1. **The ante handler must be rewired.** Default Cosmos chains route tx fees to `fee_collector`. POA expects them to land in its own module account. POA's `EndBlock` at block 1 panics if this isn't done — a deliberate safety check.
2. **Validators see "lumpy" rewards** rather than continuous accrual. They have an "allocated fees" balance that updates at checkpoints, withdrawable any time.

### What POA explicitly does not do

- **No slashing integration.** No downtime detection, no equivocation handling. Misbehavior is handled by admin zero-powering.
- **No `x/distribution` integration.** POA replaces it for validator rewards.
- **No `x/evidence` integration.** Equivocation evidence has nowhere to go.
- **No staking-keeper interface.** It does NOT implement the `StakingKeeper` interface that other modules might expect.

### Source available license

The cosmos-sdk POA module is shipped under the **Cosmos Labs Source Available Evaluation License**, not Apache-2.0. It forbids commercial/production use without a commercial license from Cosmos Labs. Stride has been granted permission, so this is not a blocker for our migration — but reviewers should be aware that the module isn't open source in the usual sense.

---

## 5. SDK upgrade mechanics

A few SDK internals matter a lot for getting upgrade handlers right.

### VersionMap and `RunMigrations`

Each module has a `ConsensusVersion()` (an integer). The `x/upgrade` module stores a **VersionMap** in state: `{module_name: consensus_version}`. On upgrade, the handler typically does:

```go
versionMap, err := mm.RunMigrations(ctx, configurator, vm)
```

`RunMigrations` iterates over modules in the **current `ModuleManager`** (not the VersionMap). For each module:

- If it has a stored version in `vm`, run any registered migration handlers from `vm[name]` → current `ConsensusVersion()`.
- If it has no stored version (i.e., a new module), run `InitGenesis` with default state.

Important consequences:

1. **Modules removed from the manager are silently skipped.** Their entries in `vm` are not iterated. No panic.
2. **Stale entries persist in state.** `SetModuleVersionMap` only writes; it never deletes. So removed modules' VersionMap entries stick around indefinitely. This is harmless — just cosmetic.
3. **Adding a module with no `RegisterMigrations` and no prior state runs InitGenesis.** Make sure new modules expect that behavior.

### Module manager vs store key mounting — orthogonal

These are two completely separate concepts:

- **Module manager** (`mm.NewManager(...)`) controls which modules run BeginBlocker/EndBlocker/InitGenesis and have message/query routes. Pure runtime concern.
- **Store key mounting** (`MountStores(...)`) controls which KV substores exist in BaseApp. Pure storage concern.

They can be mixed in any combination:

- Mounted + in manager = module fully active (normal case).
- Mounted + not in manager = module's data exists, but no blockers run (the v33 pattern for ccvconsumer/slashing/evidence).
- Not mounted + in manager = compile error (module construction would fail).
- Not mounted + not in manager = module fully gone (the v34 target state).

### `StoreUpgrades` — when stores are added or deleted

`UpgradePlan` includes a `StoreUpgrades` field with `Added`, `Deleted`, and `Renamed` lists. These take effect at the upgrade block:

- **`Added`**: a new substore is created, version-aligned with the chain.
- **`Deleted`**: the substore's data is wiped AND it's removed from the Merkle tree.

### The panic to avoid

The most common upgrade footgun:

> **Listing a store key in `StoreUpgrades.Deleted` while leaving its `MountStores(keys[X])` call in `app.go` causes a panic at the next chain restart.**

The error: `"version of store X mismatch root store's version"`. Because `Deleted` set the substore version to 0 but the chain is at some higher version, and the still-mounted store can't reconcile.

The rule: **always pair `Deleted` with removing `MountStores`, `NewKVStoreKeys` entry, and any keeper instantiation referring to that key.**

The inverse footgun (much less common): unmounting a store without listing it in `Deleted`. This causes the substore's data to be orphaned — the substore is no longer in the chain's hash, but its data is still on disk. Mostly harmless, but cosmetically bad.

### Stride's upgrade test pattern

Stride has a well-established pattern for upgrade tests in `app/upgrades/vN/upgrades_test.go`:

```go
type UpgradeTestSuite struct {
    apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
    s.Setup()
}

func (s *UpgradeTestSuite) TestUpgrade() {
    // Pre-upgrade state setup
    s.SetupSomething()
    // Run upgrade handler
    s.ConfirmUpgradeSucceeded(vN.UpgradeName)
    // Assertions
    s.CheckSomethingChanged()
}
```

`apptesting.AppTestHelper` spins up a full `StrideApp` instance (no real CometBFT, no networking) with all keepers wired. Tests can directly invoke any keeper method via `s.App.X` and run the upgrade handler via `s.ConfirmUpgradeSucceeded`. These tests run via `make test-unit`.

This is the test framework v33's tests will use.

---

## 6. What the migration changes

This section is the bridge to the v33 plan. It's a curated diff between today's setup and post-v33.

### Module manager

| Module | Status pre-v33 | Status post-v33 |
|---|---|---|
| `ccvconsumer` | Registered | Removed from manager (keeper still mounted) |
| `ccvdistr` | Registered | Removed from manager |
| `ccvstaking` | Registered | **Kept** — still suppresses x/staking validator updates |
| `staking` | Active (wrapped by ccvstaking) | Active (still wrapped) — unchanged behavior |
| `distribution` | Wrapped by ccvdistr | Used directly — fee pool changes from `cons_redistribute` to `fee_collector` |
| `slashing` | Active, wired to ConsumerKeeper | Removed from manager (keeper still mounted) |
| `evidence` | Active | Removed from manager (keeper still mounted) |
| `poa` | — | **Added** with WithSecp256k1Support if applicable |
| `gov` | Tally based on x/staking | **Unchanged** — still tally based on staking |
| `mint` | Custom Stride mint, 4-way split | Unchanged |
| `stakeibc`, `auction`, `strdburner`, etc. | Active | Unchanged |

### Keeper wiring changes (in app.go)

- `DistrKeeper` constructed with `authtypes.FeeCollectorName` instead of `ccvconsumertypes.ConsumerRedistributeName`.
- `POAKeeper` newly constructed with KV + transient store + account + bank keepers.
- `SlashingKeeper`, `EvidenceKeeper`, `ConsumerKeeper` — keeper construction stays for now (not used at runtime, deleted in v34).
- Module account perms: add `poatypes.ModuleName: nil`.
- Store keys: add `poatypes.StoreKey`, `poatypes.TransientStoreKey`.

### Ante handler

Tx fee recipient changes from `fee_collector` → POA module account. Required by POA's design.

### Upgrade handler (v33)

- Snapshot `consumerKeeper.GetAllCCValidator(ctx)` → 8 POA validators with their *current* ICS-assigned powers.
- Call `poaKeeper.InitGenesis(ctx.WithBlockHeight(0), ...)` to seed POA state with that set + admin multisig.
- Sweep residual `cons_to_send_to_provider` and `cons_redistribute` balances to community pool.
- That's it. No staking drain, no proposal failures, no IBC channel close.

The handler preserves current ICS-assigned powers for the cleanest possible CometBFT handoff (no validator-update diff emitted at the upgrade block). Long-term, equal powers across all 8 validators is preferable for a permissioned chain — that rebalance can happen either as a few extra lines in the same v33 handler or via a multisig `MsgUpdateValidators` post-upgrade. Either is safe; the choice is operational preference.

### What stays the same

- Govenators and their delegations (untouched).
- All inflation flows except the 27.64% staking portion (which now goes 100% to govenators rather than 85% local + 15% Hub).
- LS revenue flow (stakeibc → 15%/85% → auction → strdburner).
- Strategic reserve, community pool growth, security budget allocations.
- Governance and tally function.
- All Stride-specific modules.

### Why two binaries (v33 + v34)

v33 is the high-risk migration. v34 is housekeeping (delete keepers, drop ICS dep). They're split because:

- v33 stays focused on the consensus-layer change.
- Real-world post-v33 feedback informs v34's exact shape.
- Rollback path is preserved during v33's stabilization period.

See `docs/superpowers/specs/2026-04-27-ics-cleanup-design.md` for v34 details. **Implementers of v33 should not consult that document.**

---

## 7. Quick reference: the validators

For clarity since this comes up a lot:

| "Validator" type | Lives in | Purpose | Who can change it |
|---|---|---|---|
| **ICS validators (today)** | `ccvconsumer.CrossChainValidator` records, pushed from Hub | Block production on Stride | Hub governance + Stride's PSS allowlist |
| **POA validators (post-v33)** | `poa.Validator` records | Block production on Stride | Stride multisig admin |
| **Govenators** | `x/staking.Validator` records | Local STRD staking + governance tally weight | Anyone (open `MsgCreateValidator`) |
| **`utils.PoaValidatorSet`** | Hardcoded constants | LS revenue recipients (15% of stakeibc fees) | Code change + upgrade |

These are independent on-chain. They may overlap in real-world identity (the same operator might run an ICS/POA validator AND a govenator AND be in the LS-revenue list) but on-chain they're distinct records.

---

## 8. Where to look in the code

For reviewers wanting to verify any of the above:

- **Stride's app.go**: `app/app.go` — top-level wiring of every module above.
- **v32 upgrade handler** (most recent example of Stride's upgrade style): `app/upgrades/v32/upgrades.go`.
- **Custom x/mint**: `x/mint/keeper/keeper.go` (`DistributeMintedCoin`) and `x/mint/types/params.go` (default proportions).
- **LS revenue flow**: `x/stakeibc/keeper/reward_allocation.go` (`AuctionOffRewardCollectorBalance`).
- **Hardcoded validator constants**: `utils/poa.go`.
- **ICS consumer module**: `interchain-security/v7/x/ccv/consumer/` (vendored via go.mod).
- **POA module**: `cosmos-sdk/enterprise/poa/x/poa/` (in cosmos-sdk repo).

---

## What this document is not

- Not a step-by-step implementation guide. See the design specs for that.
- Not exhaustive on Cosmos SDK internals. Just enough to evaluate the migration.
- Not a substitute for reading the actual code where details matter.

If after reading this you're still unclear on something specific, that's a signal the doc needs to be expanded — open a PR.
