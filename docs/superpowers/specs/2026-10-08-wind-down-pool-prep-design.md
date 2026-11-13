# Wind-down dashboard: pool prep (planned pools, route seeding, pool creation)

Date: 2026-10-08. Extends `2026-10-07-wind-down-pools-tab-and-transfer-sets-design.md`. The dashboard
lives in `scripts/wind-down/dashboard/`.

## Why

The pools get created in the voting week (2026-10-06..10), but the Pools tab only reads pools that
already exist and the Multisig tab has no creation set, so the three preparatory steps — seed every
foreign-route stToken denom on Osmosis, confirm the canonical denoms have supply, instantiate 39
pools with the right denoms and factors — have no data and no commands behind them. This adds:

1. **Planned pools** per zone on the Pools tab: one canonical + one per policy route, each with the
   denom on the holder chain, the denom on Osmosis, whether it is seeded, the factors it will be
   created with, a copyable seed-transfer command, and a link to its creation tx. A planned row is
   replaced by the live pool row once a pool holding that denom exists.
2. **`pool-creation`** multisig set, first on the Multisig tab: one `create-pool` tx per planned pool
   that does not exist yet, with the exact instantiate message.
3. Live checks on the three voting-week steps.

Facts (2026-10-08): poolmanager `pool_creation_fee` is `20000000 factory/osmo147h5x9pcj7lm0cttlaefx6sqq5vdfnmwfcqxkmjd7exqm9gc7grqhr75m0/alloyed/allUSDC`
per pool (gamm's 100 OSMO param does not apply to cosmwasm pools); the vault holds none. Policy
channels resolve on Stride to: channel-0 cosmoshub-4 (counterparty channel-391), channel-6 injective-1
(channel-89), channel-11 axelar-dojo-1 (channel-33), channel-13 phoenix-1 (channel-25), channel-24
juno-1 (channel-139), channel-40 secret-4 (channel-37), channel-52 phoenix-1 (channel-46), channel-69
axelar-dojo-1 (channel-64), channel-123 neutron-1 (channel-8), channel-148 agoric-3 (channel-59),
channel-160 dydx-mainnet-1 (channel-1), channel-162 celestia (channel-4), channel-197 dymension_1100-1
(channel-0), channel-213 ssc-1 (channel-0), channel-240 haqq_11235-1 (channel-7), channel-258
laozi-mainnet (channel-161). These are resolved live, not hard-coded; the list is for tests.

## 1. Planned pools (`pools.py`)

For each zone and each Stride channel in `config.REQUIRED_ROUTES[st_denom]`, plus the canonical pool:

```
PlannedPool: kind (canonical|route), stride_channel (route only), holder_chain_id, holder_name, holder_binary, holder_node,
  counterparty_channel (X → Stride), osmosis_channel (Osmosis → X), holder_to_osmosis_channel (X → Osmosis),
  denom_on_holder (ibc/… of transfer/<X→Stride>/<st_denom>; null for canonical),
  denom_on_osmosis (ibc/… of transfer/<Osmosis→X>/transfer/<X→Stride>/<st_denom>, or transfer/channel-326/<st_denom> for canonical),
  seeded (bool|null: Osmosis bank supply of denom_on_osmosis > 0), supply_on_osmosis (str|null),
  st_factor, native_factor, alloyed_subdenom, instantiate_msg (dict), seed_command (str|null), live_contract (str|null),
  error (str|null: why it could not be resolved)
```

Resolution per route, cached per process (nothing here changes): Stride channel end (`counterparty.channel_id`)
and its client's chain id (`chain.ibc_client_chain_id` via the connection); then the Osmosis → X channel
= the `config.HOLDER_ROUTES` entry whose channel's client tracks X (Osmosis channel end → connection →
client chain id); X → Osmosis channel = that end's counterparty. A route whose X is not in
`config.HOLDER_CHAINS` or whose Osmosis channel cannot be found gets `error` and no commands.

`config.HOLDER_CHAINS: dict[str, HolderChain(name, binary, node)]` keyed by chain id, for the sixteen
chains above (short names as the chain registry: cosmoshub, injective, axelar, terra, juno, secret,
neutron, agoric, dydx, celestia, dymension, saga, haqq, band, osmosis; binaries gaiad, injectived,
axelard, terrad, junod, secretd, neutrond, agd, dydxprotocold, celestia-appd, dymd, sagad, haqqd, bandd;
nodes `https://<name>-strd-rpc.polkachu.com:443` where a private endpoint exists in `config.ZONES`'s
endpoint names, else the public `https://<name>-rpc.polkachu.com:443`).

Factors: six-decimal zones `st_factor = 10**18`, `native_factor = int(rate × 10**18)`; 18-decimal zones
`st_factor = 10**6`, `native_factor = floor(rate × 10**6)` (`config.ZONES[].decimals`; the rate is the
zone's live `stride_rate`). `alloyed_subdenom`: `<stSymbol>` for canonical (e.g. `stATOM`),
`<stSymbol>.<holder_name>` for a route (`stATOM.cosmoshub`; two routes to one chain, Axelar's channel-11
and channel-69 and Terra's channel-13/52, get `<stSymbol>.<holder_name>.<stride_channel>` with the dash
dropped, e.g. `stATOM.axelar.channel11`). Instantiate message exactly as `docs/wind-down/transmuter.md`:

```json
{"pool_asset_configs": [{"denom": "<denom_on_osmosis>", "normalization_factor": "<st_factor>"},
                        {"denom": "<native osmosis denom>", "normalization_factor": "<native_factor>"}],
 "alloyed_asset_subdenom": "<alloyed_subdenom>", "alloyed_asset_normalization_factor": "<native_factor>",
 "admin": "<vault>", "moderator": "<vault>"}
```

Seed command (route only; a plain single-signer tx on X, no multisig):
`<binary> tx ibc-transfer transfer transfer <holder_to_osmosis_channel> <vault> <amount><denom_on_holder> --from <KEY_ON_<NAME>> --chain-id <holder_chain_id> --node <holder_node> --gas auto --gas-adjustment 1.5 --fees <FEES>`
with `amount = 10**(decimals-2)` (0.01 stToken; any non-zero amount gives the denom supply, and the vault
can later join it into that route's pool). Canonical seeding is the same over Stride's channel-5 with
`strided` from any Stride key, rendered only when the canonical denom is not seeded.

A planned pool's `live_contract` is the existing pool whose stToken denom equals `denom_on_osmosis`.

Zone payload additions: `planned: [PlannedPool]`, `routes_seeded` (bool|null: every route planned pool
seeded), `canonical_seeded` (bool|null), `pools_created` (bool|null: every planned pool has `live_contract`),
`creation_fee` ({denom, amount} from `/osmosis/poolmanager/v1beta1/params` — the REST path on Polkachu
answers "Not Implemented" for the lower-case `params`; try `Params` and fall back to null),
`vault_fee_balance` (str|null: the vault's balance of the fee denom), `creation_fee_short` (bool|null:
vault balance < fee × pools still to create).

## 2. Multisig set `pool-creation` (`multisig.py`)

First set on the tab (step `vote-pools-create`). Per zone, per planned pool with no `live_contract`, one tx:
`osmosisd tx cosmwasmpool create-pool 996 '<instantiate_msg as compact JSON>' --from <vault> --generate-only --chain-id osmosis-1 --node https://osmosis-strd-rpc.polkachu.com:443 --gas 2000000 --fees 20000uosmo > /tmp/wind-down/create-<chain_id>-<alloyed_subdenom lower, dots to dashes>.unsigned.json`
then the Osmosis sign ×3 / multisign / broadcast as the funding set. `ready=False` with a reason when the
planned pool has an `error`, is not seeded ("<denom> has no supply on Osmosis: seed it first (Pools tab)"),
or the vault's fee balance is short. Title: `<chain_id> · create <alloyed_subdenom> (<route chain · stride channel> | canonical)`.
Set description: creation pays the poolmanager `pool_creation_fee` per pool (live value shown) from the
vault, which must hold it in that denom; six-decimal zones use 1e18 factors, the three 18-decimal zones 1e6
(overflow), the rate is read live at render time so a pool created tomorrow carries tomorrow's rate.

## 3. Pools tab (`static/pools.js`)

Each zone panel shows, above the live pool rows, a **Planned** table: pool (canonical / `<holder_name> · <stride_channel>`),
denom on holder chain (mono, copy on click), denom on Osmosis (copy), seeded (✓ with supply / ✗ / n/a),
factors (`st 1e18 · native 2013525450106978250`), status (`created → <pool id>` linking to the live row,
or `planned` with a "→ create" link to `#multisig/pool-creation/<chain_id>`), and a **seed** copy button
for an unseeded route. A planned row with `live_contract` is shown collapsed to one line ("created") so
the table shrinks as pools appear. The zone header gains the fee line: `creation fee 20 allUSDC × N to
create · vault holds M` (red when short). A planned pool with `error` shows it in place of the denoms.

## 4. Plan

- `vote-seed-routes`: text ends "The Pools tab lists every route denom with its seed command; live below:
  every route denom has supply on Osmosis." `zones` = all eleven, `auto: {tab: pools, path: routes_seeded}`.
- `vote-canonical-supply`: likewise with `canonical_seeded`.
- `vote-pools-create`: `multisig: "pool-creation"` (one link per zone), text adds "Commands per pool, with the
  exact denoms and factors, are on the Multisig tab; the vault must hold the poolmanager creation fee (20
  allUSDC per pool today) — top it up first." Keep `auto: pools_ready`.

## Testing

- `test_pools.py`: planned pools from fakes: route resolution (chain id, counterparty, Osmosis channel via
  HOLDER_ROUTES, X → Osmosis channel), both denoms' hashes (assert `chain.ibc_denom` of the expected paths),
  factors per decimals (6 → 1e18 / int, 18 → 1e6 / floor; a rate whose 1e6 floor differs from round),
  subdenoms (canonical, route, two routes to one chain), seeded / supply, `live_contract` matching,
  `routes_seeded` / `canonical_seeded` / `pools_created` tri-state, fee and shortfall, an unknown holder
  chain → `error`; `config.HOLDER_CHAINS` covers every chain id a policy channel resolves to (fixture
  from the list above).
- `test_multisig.py`: full-string equality for one canonical and one route creation tx; not-ready reasons
  (error, unseeded, fee short, already created).
- `test_ops.py`: the three steps' fields.
- Frontend: headless Chrome screenshot of the Pools tab (planned tables for every zone; most routes
  unseeded today) and the Multisig tab's first set.

## Build plan

**Chunk 1 — backend** (`pools.py`, `config.py` (`HOLDER_CHAINS`), `multisig.py`, `server.py` if needed,
tests, README). Produces §1 payload and §2 set. Depends on: none.

**Chunk 2 — frontend and plan** (`static/pools.js`, `static/style.css`, `ops/plan.json`, `test_ops.py`,
README's Pools section wording). Builds against the §1/§2 shapes with fixtures. Depends on: none
(merge after chunk 1).
