# Transmuter reference (Osmosis alloyed pools for the wind-down)

Everything here was read from source or queried from mainnet on 2026-09-23. Re-verify
anything marked *live* before relying on it; channel and client state in particular changes.

Companion documents: the wind-down design (`docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md`,
§3 and §8) and the mainnet test plan (`docs/superpowers/plans/2026-09-23-transmuter-mainnet-test.md`).

## 1. The contract

| Item | Value |
|---|---|
| Source | github.com/osmosis-labs/transmuter, tag `v3.2.0`, `contracts/transmuter/src` |
| Code id on osmosis-1 | `996` (cw2 `contract_info` of every pool on it reads `crates.io:transmuter 3.2.0`) |
| Other whitelisted ids | `641`, `842`, `867`, `885` (cosmwasmpool `code_id_whitelist`); `867` and `814` are older alloyed versions |
| Instantiate permission on 996 | `AnyOfAddresses [osmo1rxjakgd8yhks2j7hc7pt6a22z3zd64grexpyf7]` = the `cosmwasmpool` module account. Nobody instantiates it directly; `MsgCreateCosmWasmPool` does, with the module account as both creator and wasm admin |
| Wasm admin of every pool | the cosmwasmpool module account. Code migration is therefore Osmosis governance only (`pool_migration_limit` = 20 pools per proposal). A future gov proposal that migrates "all alloyed pools" to a newer transmuter would take ours with it |
| Pools on 996 today (live) | 37 of 261 cosmwasm pools; 3,588 pools in total on the chain |
| Pool creation | `MsgCreateCosmWasmPool { code_id, instantiate_msg, sender }`, permissionless. The poolmanager charges `pool_creation_fee` to the community pool and skips it only for the concentrated-liquidity creator whitelist |
| Pool creation fee (live) | `20000000 factory/osmo147h5x9pcj7lm0cttlaefx6sqq5vdfnmwfcqxkmjd7exqm9gc7grqhr75m0/alloyed/allUSDC` = 20 allUSDC. The creator must hold allUSDC (swap USDC for it on Osmosis first) |
| Uniqueness | none. Nothing stops two pools with the same assets or the same subdenom; the alloyed denom is `factory/<pool contract address>/alloyed/<subdenom>`, namespaced by contract |
| Asset count | 1 to 20 denoms; the wind-down pools use exactly 2 (one stToken route + native) |

### Normalization factors: read this before writing an instantiate message

The contract converts amounts as

```
out = in × out_factor / in_factor        (exact-in rounds down, exact-out rounds up)
spot_price(base, quote) = quote_factor / base_factor
```

so a **larger factor makes a unit of that token worth less** (the factor is "base units per
normalized unit", the same way an 18-decimal token gets factor `1e12` next to a 6-decimal one).
To price 1 stATOM at `RR` ATOM:

```
stATOM  normalization_factor = 1e18            = 1000000000000000000
ATOM    normalization_factor = RR × 1e18       = 2000174393052495819   (RR 2.000174393052495819 on 2026-09-23)
alloyed normalization_factor = ATOM's factor   → 1 alloyed unit = 1 uatom of pool value
```

Check: 1,000,000 ustatom in → 1,000,000 × 2000174393052495819 / 1e18 = 2,000,174 uatom out.
The design spec originally had the two factors swapped; that was corrected on 2026-09-23.

Rounding and size limits:

- `convert_amount` multiplies in `Uint256` and the quotient must fit `Uint128`.
- `weights()` runs on every swap, join and exit. It takes the lcm of all factors and
  converts every balance to it. With factors `1e18` and `2000174393052495819` (coprime) the
  lcm is ~2e36 and a balance of 1.3M stATOM (1.3e12 base units) normalizes to ~2.6e30, well
  under `Uint128::MAX` (3.4e38). Re-check this arithmetic whenever a pool gets a third factor
  (`add_new_assets` with a different factor raises the lcm).
- **18-decimal zones overflow at the 1e18 scale** (found 2026-10-07): an 18-decimal balance is
  already ~1e22-1e26 base units, and with coprime factors the lcm is ~1e36, so `weights()`
  normalizes haqq's 1.07e26 aISLM to 1.07e44, dYdX's 4.0e23 adydx to 4.0e41 and Injective's
  2.1e22 inj to 2.1e40 — all above `Uint128::MAX`, so the funding join itself fails (a 1-token
  test join passes, 1e36). For haqq, dYdX and Injective instantiate with factors scaled to 1e6
  instead: stToken `1000000`, native `floor(RR × 1e6)` (rounded down, so the pool never prices above RR), alloyed = native. The ratio still prices
  the pool; the rate truncates at 6 decimals (~1e-7 relative, inside the accepted staleness) and
  the balances normalize to ≤1.1e32. The dashboard's Pools tab headroom check computes this bound
  from each zone's needed amount before anything is joined.
- `rescale_normalization_factor { numerator, denominator }` multiplies every factor
  (including the alloyed one) by the same ratio and rejects any factor the ratio does not
  divide exactly. It cannot change relative prices.

### Messages (sylvia snake_case; all verified against the live allUSDC pool)

Instantiate (goes inside `MsgCreateCosmWasmPool.instantiate_msg`; funds must be empty):

```json
{
  "pool_asset_configs": [
    {"denom": "<stATOM ibc denom>", "normalization_factor": "1000000000000000000"},
    {"denom": "<ATOM ibc denom>",   "normalization_factor": "2000174393052495819"}
  ],
  "alloyed_asset_subdenom": "stATOMtest",
  "alloyed_asset_normalization_factor": "2000174393052495819",
  "admin": "<osmo address>",
  "moderator": "<osmo address>"
}
```

Execute (via `MsgExecuteContract` on the pool's contract address):

| Message | Who | Notes |
|---|---|---|
| `{"join_pool":{}}` + `--amount` | anyone | any subset of pool assets; mints alloyed to sender, rounds down |
| `{"exit_pool":{"tokens_out":[{"denom":..,"amount":..}]}}` | anyone with alloyed | burns from sender's balance, rounds the burn up; no funds attached |
| `{"add_new_assets":{"asset_configs":[{"denom":..,"normalization_factor":..}]}}` | admin | each denom must have non-zero bank supply on Osmosis (`DenomHasNoSupply` otherwise); no funds |
| `{"rescale_normalization_factor":{"numerator":"..","denominator":".."}}` | admin | uniform only |
| `{"register_limiter":{"denom":..,"label":..,"limiter_params":{"static_limiter":{"upper_limit":"0.9"}}}}` | admin | or `{"change_limiter":{"window_config":{"window_size":..,"division_count":..},"boundary_offset":".."}}`; none registered by default |
| `{"deregister_limiter":{"denom":..,"label":..}}` | admin | |
| `{"set_alloyed_denom_metadata":{"metadata":{...}}}` | admin | bank `Metadata` for the alloyed denom |
| `{"mark_corrupted_assets":{"denoms":[..]}}` / `{"unmark_corrupted_assets":{"denoms":[..]}}` | moderator | a corrupted asset's amount and weight may never increase; it is dropped from the pool once drained to zero |
| `{"set_active_status":{"active":false}}` | moderator | every other execute and every sudo swap fails with `InactivePool` while inactive; setting the same status twice errors |
| `{"transfer_admin":{"candidate":..}}` then `{"claim_admin":{}}` from the candidate | admin / candidate | `cancel_admin_transfer` (admin) and `reject_admin_transfer` (candidate) undo it |
| `{"assign_moderator":{"address":..}}` | admin | |

Swaps do not have an execute entry point. They come through the poolmanager (`sudo`
`SwapExactAmountIn` / `SwapExactAmountOut`) with `swap_fee` that must equal the hard-coded
zero, or through join then exit. Token-to-token, token-to-alloyed and alloyed-to-token all
work through the poolmanager.

Queries (`osmosisd q wasm contract-state smart <addr> '<json>'`):
`list_asset_configs`, `list_limiters`, `get_shares {address}`, `get_share_denom`,
`get_swap_fee`, `is_active`, `get_total_shares`, `get_total_pool_liquidity`,
`spot_price {base_asset_denom, quote_asset_denom}`,
`calc_out_amt_given_in {token_in, token_out_denom, swap_fee:"0"}`,
`calc_in_amt_given_out {token_out, token_in_denom, swap_fee:"0"}`,
`get_corrupted_denoms`, `get_admin`, `get_admin_candidate`, `get_moderator`.

### Limiters (read from `src/limiter/limiters.rs`)

What a limiter bounds is a denom's **weight**: that denom's normalized value divided by the
pool's total normalized value, recomputed after every join, exit and swap (`weights()`).
Two kinds, both registered per `(denom, label)` by the admin, at most 10 per denom:

- `static_limiter { upper_limit }` (a `Decimal` in `(0, 1]`): the action fails with
  `UpperLimitExceeded { denom, upper_limit, value }` if the denom's post-action weight
  would be above the limit **and the weight is going up**. Actions that lower or hold the
  weight are never blocked, so a capped denom can always be taken out of the pool.
- `change_limiter { window_config { window_size, division_count }, boundary_offset }`: the
  same rule against a moving average of the weight over a time window (≤ 10 divisions),
  plus `boundary_offset`. Its history resets on `add_new_assets`.

Rules that matter for us:

- A denom's **last limiter cannot be deregistered** (`EmptyLimiterNotAllowed`); once a denom
  has one it always has one. It can be widened to `upper_limit: "1"`, which disables it in
  effect. Plan any limiter as permanent.
- Limiters apply to the admin's own `join_pool` / `exit_pool` too. An exit that takes only
  ATOM out raises every stToken denom's weight and is checked against their caps.
- Under swaps the pool's total normalized value is constant (stToken in at RR, native out
  at RR), so a cap on a denom is effectively a cap on how much of that denom the pool will
  ever absorb, measured in native value, as long as the vault does not exit native.
- A weight is per denom, so the two Secret routes count separately.

What a cap can and cannot protect against here: every stATOM route is the same claim, so
there is no "depeg" to bound. The one thing a cap does bound is a *counterfeit* route: if
a source chain (or its light client on Osmosis) were compromised and minted two-hop
stATOM that never existed on Stride, the attacker could drain ATOM up to that route's cap
and no further. Stride's per-channel escrow is a hard upper bound on the genuine amount
each route can ever deliver, so caps at `escrow share + margin` cost honest users nothing.
Canonical stATOM needs no cap: after the Stride halt no more of it can be minted on
Osmosis.

Decision 2026-09-24: the real deployment uses **one pool per stToken route** (canonical denom + native,
and a separate two-asset pool for each foreign-route denom), each route pool funded with exactly its
escrow share. That gives per-route isolation by construction, so **no limiters are registered**; the
mechanics above stay documented because the contract offers them and the test exercised them.
`add_new_assets` is likewise not part of normal operation.

### Fees and routing around the pool

- Swap fee inside the contract: zero, hard-coded.
- Poolmanager taker fee: charged on every routed swap, on top of the contract. Default
  0.1%; the stATOM/ATOM pair has an override of **0.02%** (`trading-pair-taker-fee`, live).
  Join-then-exit through `MsgExecuteContract` bypasses the poolmanager and pays no taker fee.
- Where the override comes from: a per-pair entry in poolmanager state, written by
  `MsgSetDenomPairTakerFee`, which only the `taker_fee_params.admin_addresses` may send
  (live: `osmo162wk8qc3w5s9hfs8dm76wrqnk6fjmsez2t4kk6zyugmrlzgds8sqfesmlm` and
  `osmo10d07y265gmmuvt4z0w9aw880jnsr700jjeq4qp`, the Osmosis team admin that also
  administers the allUSDC pool). The 0.02% is Osmosis's standing policy for LST/underlying
  pairs: stOSMO/OSMO is also 0.02%, while ATOM/OSMO is 0.2% and stATOM/OSMO the 0.1%
  default. Nothing about our pool triggers it; it is keyed on the two denoms. So the
  two-hop stATOM denoms pay the **0.1% default** against ATOM (verified live for the Hub and
  Agoric denoms) unless the Osmosis team adds entries for them, which is a request worth
  making once the denom list is final.
- SQS (the router behind app.osmosis.zone) lists `996` in `AlloyedTransmuterCodeIDs`, so it
  understands the pool type. Its dynamic liquidity filter decides whether a pool is a route
  candidate for a given trade size: trades of at least $1 / $1k / $10k / $250k / $1M consider
  pools holding at least $1 / $10 / $1k / $15k / $40k. Price feeds ignore pools under $1k.
- Today's stATOM→ATOM quote (live, 2026-09-23): 10 stATOM → 19.806 ATOM through pool 1283,
  about 1% below the redemption rate. A transmuter at the redemption rate is therefore the
  best stATOM→ATOM route the moment it holds ATOM, and arbitrageurs will use it. That is
  the intended end state; for a small test pool it means the ATOM side will be taken
  within minutes of funding, by strangers, at a loss to us of liquidity × (RR − market)/RR.

## 2. Denoms

Redemption rate (live): `2.000174393052495819` (`stakeibc/host_zone/cosmoshub-4`).

### stATOM by chain (first hop from Stride)

Balances are the escrow map of 2026-09-22 (`/Users/sampocs/Downloads/sttoken_escrow.html`);
supplies were read from each chain on 2026-09-23 and agree with the escrows.

| Chain | Stride channel | Chain's channel | Denom trace on that chain | IBC denom on that chain | Held there |
|---|---|---|---|---|---|
| Stride | – | – | `stuatom` | `stuatom` | 719,062 (55.5%) |
| Osmosis | channel-5 | channel-326 | `transfer/channel-326/stuatom` | `ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901` | 456,310 (35.2%) |
| Cosmos Hub | channel-0 | channel-391 | `transfer/channel-391/stuatom` | `ibc/B05539B66B72E2739B986B86391E5D08F12B8D5D2C2A7F8F8CF9ADF674DFA231` | 62,487 (4.8%) |
| Injective | channel-6 | channel-89 | `transfer/channel-89/stuatom` | `ibc/A8F39212ED30B6A8C2AC736665835720D3D7BE4A1D18D68566525EC25ECF1C9B` | 31,062 (2.4%) |
| Secret | channel-40 | channel-37 | `transfer/channel-37/stuatom` | `ibc/A0E80E59956C754F1D9CB37234D13E0CF2949E7254896359F284512FA8428E18` | 8,385 (0.65%) |
| Penumbra | channel-307 | channel-8 | `transfer/channel-8/stuatom` | Penumbra does not use the `ibc/` hash form; the trace is what matters | 6,237 (0.48%) |
| Kujira | channel-8 | channel-32 | `transfer/channel-32/stuatom` | `ibc/0306D6B66EAA2EDBB7EAD23C0EC9DDFC69BB43E80B398035E90FBCFEF3FD1A87` | 5,788 (0.45%) |
| Agoric | channel-148 | channel-59 | `transfer/channel-59/stuatom` | `ibc/B1E6288B5A0224565D915D1F66716486F16D8A44BF33A9EC323DD6BA30764C35` | 2,425 (0.19%) |
| Neutron | channel-123 | channel-8 | `transfer/channel-8/stuatom` | `ibc/B7864B03E1B9FD4F049243E92ABD691586F682137037A9F3FCA5222815620B3C` | 1,094 (0.08%) |
| Carbon | channel-47 | channel-8 | `transfer/channel-8/stuatom` | `ibc/B7864B03E1B9FD4F049243E92ABD691586F682137037A9F3FCA5222815620B3C` (same hash as Neutron's: both chains use channel-8 to Stride) | 545 (0.04%) |
| Axelar | channel-69 | channel-64 | `transfer/channel-64/stuatom` | `ibc/CFB9D610B7BB02E8E6FDB0AF87FCADDF97687A3F13E76AC624C500E54335ADAA` | 375 (0.03%) |
| Comdex | channel-49 | – | – | chain stopped, ignored | 1,420 (0.11%) |

### stATOM arriving on Osmosis from those chains (two-hop denoms)

`transfer/<Osmosis channel to X>/transfer/<X channel to Stride>/stuatom`, hashed with
sha256 and upper-cased. Supply is live on Osmosis.

| From | Osmosis channel | X's channel to Osmosis | Trace on Osmosis | IBC denom on Osmosis | Supply |
|---|---|---|---|---|---|
| Cosmos Hub | channel-0 | channel-141 | `transfer/channel-0/transfer/channel-391/stuatom` | `ibc/7451074F46885686D3B47B12A6BF74F6D36847ED1891AC612FCFAEB7FB551E14` | 0.176527 |
| Injective | channel-122 | channel-8 | `transfer/channel-122/transfer/channel-89/stuatom` | `ibc/F65724D2AE4A14F5BC149FC12C984D53D2307D95EC57BBA1CE52F94EB670EF60` | 0, and **cannot be created**: Osmosis's rate limiter rejects every such packet (prefix bug, `channel-8` vs `channel-89`; see the test log) |
| Secret (channel-88 pair) | channel-88 | channel-1 | `transfer/channel-88/transfer/channel-37/stuatom` | `ibc/8AEB813EE960508AEDC1C2EB605788CE6A32F4E632583336D840BE4B8EC24CC7` | 0 |
| Secret (channel-476 pair) | channel-476 | `wasm.secret1tqmms5…` channel-44 | not a transfer channel: Osmosis channel-476's counterparty is Secret's SNIP-20 IBC bridge contract port, verified 2026-09-23 (`channel not found` on Secret's transfer port). Plain stATOM holders on Secret cannot use it | – | – |
| Penumbra | channel-79703 | channel-4 | `transfer/channel-79703/transfer/channel-8/stuatom` | `ibc/B66737925072CEF58C5E9990038D5B869D778DC42C7C2F1F5CEE8665D907AE8B` | 0 |
| Kujira | channel-259 | channel-3 | `transfer/channel-259/transfer/channel-32/stuatom` | `ibc/DED75871F78AF8FC9BCFE75BEA82D66A2B2366204E210FD8E4C77A2AAEA1B1E3` | 0 |
| Agoric | channel-320 | channel-1 | `transfer/channel-320/transfer/channel-59/stuatom` | `ibc/C86C2FA56D954AB05960450215E63605528CB3481694ABEA87CE4DB0EF17D265` | 0.782288 |
| Neutron | channel-874 | channel-10 | `transfer/channel-874/transfer/channel-8/stuatom` | `ibc/8FCFAF3AE6BA4C5BDFF85B41449FBACE547E2BAC23895E839230404FB0EC3837` | 7.841929 |
| Carbon | channel-188 | channel-0 | `transfer/channel-188/transfer/channel-8/stuatom` | `ibc/A1FC8CB6B2E965DEDC6F57749F04CCE3D7C15DD10FC2F7BBEEC19E16D0F82397` | 0 |
| Axelar | channel-208 | channel-3 | `transfer/channel-208/transfer/channel-64/stuatom` | `ibc/7FA89E771D836CC136CEDD28AD88DD5F2A1083883FA4681FDBBFFD1D78E04FCB` | 0 |

The chain registry tags both Secret pairs preferred, but only channel-1 is a `transfer` channel;
channel-44 belongs to Secret's private-token bridge contract, so bank-held stATOM on Secret has
exactly one route, channel-1 → channel-88. The Hub, Agoric and Neutron denoms already have supply, so `add_new_assets` can take them
without seeding. None of the Neutron, Carbon or Axelar pairs collide with the rate-limiter prefix
bug (channel-10 vs channel-8, channel-0 vs channel-8, channel-3 vs channel-64).

### Other denoms on Osmosis

| Token | Denom |
|---|---|
| ATOM | `ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2` (`transfer/channel-0/uatom`) |
| allUSDC (pool creation fee) | `factory/osmo147h5x9pcj7lm0cttlaefx6sqq5vdfnmwfcqxkmjd7exqm9gc7grqhr75m0/alloyed/allUSDC` |
| OSMO (gas) | `uosmo` |
| A pool's alloyed asset | `factory/<pool contract address>/alloyed/<subdenom>` |

## 3. Client health per route (live, 2026-09-23)

A transfer needs the *destination's* client of the source to be un-expired.

| Route | Client checked | Status | Pending packets | Meaning |
|---|---|---|---|---|
| Stride → Osmosis (channel-5) | Stride's `07-tendermint-1` of osmosis-1 | Active | 0 | fine |
| Stride → Hub (channel-0) | Stride's `07-tendermint-0` of cosmoshub-4 | Active | 1,742 | backlog is stale timed-out packets from sequence 98k to 227k that nobody cleared; new packets near 227.5k still land. Clearing may be needed |
| Hub → Stride (channel-391) | Hub's `07-tendermint-913` of stride-1 | Active | 0 | fine |
| Stride → Injective (channel-6) | Stride's `07-tendermint-2` | Active | 0 | fine |
| Injective → Stride (channel-89) | Injective's `07-tendermint-131` | Active | 0 | fine |
| Stride → Secret (channel-40) | Stride's `07-tendermint-37` | Active | 3 | fine |
| Secret → Stride (channel-37) | Secret's `07-tendermint-75` | Active | 12 | fine |
| Stride → Agoric (channel-148) | Stride's `07-tendermint-129` | **Expired** | 6 | cannot seed Agoric from Stride |
| Agoric → Stride (channel-59) | Agoric's `07-tendermint-74` | **Expired** | 1 | |
| Stride ↔ Penumbra (channel-307) | Stride's `07-tendermint-153` | **Expired** | 0 | |
| Stride ↔ Kujira (channel-8) | Stride's `07-tendermint-5` (2-day trusting period) | **Expired** | 0 | |
| Hub → Osmosis (channel-141 → channel-0) | Osmosis's `07-tendermint-1` | Active | | public relayers |
| Injective → Osmosis (channel-8 → channel-122) | Osmosis's `07-tendermint-1703` | Active | | public relayers |
| Secret → Osmosis (channel-1 → channel-88, channel-44 → channel-476) | Osmosis's `07-tendermint-1588` | Active | | public relayers |
| Agoric → Osmosis (channel-1 → channel-320) | Osmosis's `07-tendermint-2109` | Active | | public relayers; two-hop supply exists |
| Penumbra → Osmosis (channel-4 → channel-79703) | Osmosis's `07-tendermint-3242` | **Expired** | | Penumbra holders cannot reach Osmosis until someone recovers the client (Osmosis gov `MsgRecoverClient`) or opens a new channel |
| Kujira → Osmosis (channel-3 → channel-259) | Osmosis's `07-tendermint-2017` | **Expired** | | same; kaiyo-1 REST endpoints are also down |
| Stride → Neutron (channel-123 → channel-8) | Stride's client of neutron-1 | Active | | stTIA holders on Neutron; audited 2026-09-23 |
| Neutron → Osmosis (channel-10 → channel-874) | Osmosis's `07-tendermint-2823` | Active, last header 35 h old | | no relayer keeping it fresh; we may need to relay |
| Stride ↔ Carbon (channel-47 ↔ channel-8) | both clients | Active but stale (Stride's 297 h, Carbon's 87 h old on 2026-09-24) | | Stride leg not needed after the upgrade; Osmosis leg: we relay it (decided 2026-10-08, see sttoken-locations.md) |
| Carbon → Osmosis (channel-0 → channel-188) | Osmosis's client of carbon-1 | Active, last header 41 h old | | same; on 2026-10-07 the last packet was 12 days old and the client near its 14-day trusting period |
| Stride ↔ Axelar (channel-69 ↔ channel-64) | both clients | **Expired** | | Axelar holders cannot redeem through Stride |
| Axelar → Osmosis (channel-3 → channel-208) | Osmosis's client of axelar-dojo-1 | Active | | public relayers; Axelar holders are fine after the halt |
| Stride ↔ Dymension (channel-197 ↔ channel-0), Dymension → Osmosis (channel-2 → channel-19774) | all clients | Active | | stTIA holders on Dymension |

Channel audit 2026-09-23: every channel id, counterparty id, client destination and `ibc/` hash in this
file and in `sttoken-locations.md` was re-derived from live Stride and Osmosis state and the chain registry
by a separate pass; no mismatches. Where a chain pair has more than one channel, the one used here carries
the traffic (Osmosis↔Secret: channel-88 has sent 509,459 packets, channel-476 13,284; Stride's unused sibling
channels 36/75/18/10/110 hold none of any stToken's escrow).

Consequence for the migration: Penumbra (6.2k stATOM) and Kujira (5.8k stATOM), together
about $42k at today's price, are stranded on both hops. That is a design question for the
spec, not something the test can fix.

## 4. Commands

Set once per shell:

```bash
OSMO_NODE=https://osmosis-rpc.polkachu.com:443
STRIDE_NODE=https://stride-rpc.polkachu.com:443
OSMO_TX="--node $OSMO_NODE --chain-id osmosis-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.0025uosmo -y -o json"
STATOM=ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901
ATOM=ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
KEY=<osmosis key name>        # the test admin/moderator/LP
```

Create the pool (returns `pool_id` in the `pool_created` event; the contract address comes
from `contract-info`):

```bash
osmosisd tx cosmwasmpool create-pool 996 "$(cat instantiate.json)" --from $KEY $OSMO_TX
osmosisd q cosmwasmpool contract-info <pool_id> --node $OSMO_NODE
POOL=<contract address>
```

_Retired 2026-10-07: `check_transmuter_pool.py` is replaced by the dashboard's Pools tab (`scripts/wind-down/dashboard`), which runs the same checks live on every pool the vault administers; the two commands below are kept for the record._

Pre-funding check, run after `create-pool` and again after `add_new_assets` and the limiters. No
arguments: the vault and moderator addresses and one entry per pool (host zone, pool id, the rate the
factors were created with, the per-route caps) live in the CONSTANTS block at the top of the script.
One line per check, exit 1 on any failure:

```bash
python3 scripts/wind-down/check_transmuter_pool.py
```

After the funding join and the one-way mark (`mark_corrupted_assets` on the native token), run it
with `--funded`: the corrupted set must then be exactly the native token instead of empty.

```bash
python3 scripts/wind-down/check_transmuter_pool.py --funded
```

Read state:

```bash
osmosisd q wasm contract-state smart $POOL '{"list_asset_configs":{}}' --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL '{"get_total_pool_liquidity":{}}' --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL "{\"spot_price\":{\"base_asset_denom\":\"$STATOM\",\"quote_asset_denom\":\"$ATOM\"}}" --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL "{\"calc_out_amt_given_in\":{\"token_in\":{\"denom\":\"$STATOM\",\"amount\":\"1000000\"},\"token_out_denom\":\"$ATOM\",\"swap_fee\":\"0\"}}" --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL '{"get_shares":{"address":"<osmo addr>"}}' --node $OSMO_NODE
osmosisd q poolmanager pool <pool_id> --node $OSMO_NODE
osmosisd q poolmanager trading-pair-taker-fee $STATOM $ATOM --node $OSMO_NODE
```

Join, exit, swap:

```bash
osmosisd tx wasm execute $POOL '{"join_pool":{}}' --amount 10000000$ATOM --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL "{\"exit_pool\":{\"tokens_out\":[{\"denom\":\"$ATOM\",\"amount\":\"1000000\"}]}}" --from $KEY $OSMO_TX
# exact-in through the router (pays the 0.02% taker fee):
osmosisd tx poolmanager swap-exact-amount-in 1000000$STATOM 1 --swap-route-pool-ids <pool_id> --swap-route-denoms $ATOM --from $KEY $OSMO_TX
# exact-out through the router:
osmosisd tx poolmanager swap-exact-amount-out 2000000$ATOM 1100000 --swap-route-pool-ids <pool_id> --swap-route-denoms $STATOM --from $KEY $OSMO_TX
```

Admin and moderator:

```bash
osmosisd tx wasm execute $POOL '{"add_new_assets":{"asset_configs":[{"denom":"<two-hop denom>","normalization_factor":"1000000000000000000"}]}}' --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"set_active_status":{"active":false}}' --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"mark_corrupted_assets":{"denoms":["<denom>"]}}' --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"transfer_admin":{"candidate":"<osmo addr>"}}' --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"claim_admin":{}}' --from <candidate key> $OSMO_TX
osmosisd tx wasm execute $POOL '{"assign_moderator":{"address":"<osmo addr>"}}' --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"rescale_normalization_factor":{"numerator":"1","denominator":"1000000"}}' --from $KEY $OSMO_TX
```

IBC transfers (the CLI is the same shape on every SDK chain; use the source chain's binary):

```bash
strided  tx ibc-transfer transfer transfer channel-0   <cosmos addr> 1000000stuatom --from <stride key> --node $STRIDE_NODE --chain-id stride-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.005ustrd -y
strided  tx ibc-transfer transfer transfer channel-6   <inj addr>    1000000stuatom ...
strided  tx ibc-transfer transfer transfer channel-40  <secret addr> 1000000stuatom ...
gaiad    tx ibc-transfer transfer transfer channel-141 <osmo addr>   1000000ibc/B05539B66B72E2739B986B86391E5D08F12B8D5D2C2A7F8F8CF9ADF674DFA231 ...
injectived tx ibc-transfer transfer transfer channel-8 <osmo addr>   1000000ibc/A8F39212ED30B6A8C2AC736665835720D3D7BE4A1D18D68566525EC25ECF1C9B ...
secretd  tx ibc-transfer transfer transfer channel-1   <osmo addr>   1000000ibc/A0E80E59956C754F1D9CB37234D13E0CF2949E7254896359F284512FA8428E18 ...
```

Relaying by hand (hermes 1.13.2 is at `/opt/hermes`; `~/.hermes/config.toml` currently has
stride-1 and haqq only, but `~/.hermes/keys` has keys for stride-1, cosmoshub-4, osmosis-1,
injective-1 among others; rly 2.5.2 has paths stride-cosmos, stride-injective,
stride-osmosis). Hermes one-shot commands relay everything pending on a channel and exit,
no daemon; the chains involved must be in the config and the signing key on the
destination chain must hold gas:

```bash
# receive on the destination everything Stride sent over channel-6, plus timeouts back on Stride
hermes tx packet-recv --dst-chain injective-1 --src-chain stride-1 --src-port transfer --src-channel channel-6
# then acks back to Stride
hermes tx packet-ack  --dst-chain stride-1 --src-chain injective-1 --src-port transfer --src-channel channel-89
# or both directions of one channel in one go
hermes clear packets --chain stride-1 --port transfer --channel channel-6
# refresh a client that is near expiry (not one that has expired)
hermes update client --host-chain injective-1 --client 07-tendermint-131
# check what is pending first
hermes query packet pending --chain stride-1 --port transfer --channel channel-6
```

rly's equivalent is `rly transact flush stride-injective`. Neither tool can revive an
expired client; that needs `MsgRecoverClient` through governance on the chain holding it.

## 5. The test pools

**Day 1–2 multi-denom pool (superseded design):** pool **3590**, contract
`osmo13wcvdtkcu459zjuqdh9f7jshlg64sps0xsduxdsn9r6yez0t270qeg6z6e`, canonical stATOM + ATOM + Hub, Agoric and
Secret two-hop denoms. Drained back to the keys on 2026-09-25 and frozen (dust left).

**Day 3 per-route pools (current design), created 2026-09-25 at RR `2.002036647211047463`, all frozen, admin
and moderator `osmo1v0694qqq6ztzxvzl807dgq7h3e857hdxvpmdlc`, ATOM marked corrupted in the three funded pools:**

| Pool | Contract | stToken asset |
|---|---|---|
| 3595 | `osmo1yjlnqxxa92kpjfl9tx0pf6gru3t4z9d2x8rkxtfpl3yc58hgh69s0lzman` | canonical stATOM |
| 3596 | `osmo1jnyv2nppaes6sasyr9h5d9zm8kxs6j0swng6cqz3twkuyqrh8tnquvy768` | Hub-hop stATOM |
| 3597 | `osmo1n36rynafmm7eucl6zz4ct8nxj0pchpakzpwjgn0phzyfwcuhhlysxpjdr6` | Secret-hop stATOM |
| 3598 | `osmo147uka66qctyk25up83vgcmd52g3ydlt35rw5srdwjf8eu5gpnf9sqv3u9x` | canonical stATOM with **inverted** factors (negative control, never funded) |

None of these are the real pools; do not add them to any list. Full record: `transmuter-test-log.md`.
