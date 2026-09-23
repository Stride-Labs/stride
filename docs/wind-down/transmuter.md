# Transmuter reference (Osmosis alloyed pools for the wind-down)

Everything here was read from source or queried from mainnet on 2026-09-23. Re-verify
anything marked *live* before relying on it; channel and client state in particular changes.

Companion documents: the wind-down design (`docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md`,
§3 and §7) and the mainnet test plan (`docs/superpowers/plans/2026-09-23-transmuter-mainnet-test.md`).

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
| Asset count | 1 to 20 denoms; `add_new_assets` is bounded by the same cap |

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

### Fees and routing around the pool

- Swap fee inside the contract: zero, hard-coded.
- Poolmanager taker fee: charged on every routed swap, on top of the contract. Default
  0.1%; the stATOM/ATOM pair has an override of **0.02%** (`trading-pair-taker-fee`, live).
  Join-then-exit through `MsgExecuteContract` bypasses the poolmanager and pays no taker fee.
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
| Comdex | channel-49 | – | – | not in scope for the test | 1,420 (0.11%) |

### stATOM arriving on Osmosis from those chains (two-hop denoms)

`transfer/<Osmosis channel to X>/transfer/<X channel to Stride>/stuatom`, hashed with
sha256 and upper-cased. Supply is live on Osmosis.

| From | Osmosis channel | X's channel to Osmosis | Trace on Osmosis | IBC denom on Osmosis | Supply |
|---|---|---|---|---|---|
| Cosmos Hub | channel-0 | channel-141 | `transfer/channel-0/transfer/channel-391/stuatom` | `ibc/7451074F46885686D3B47B12A6BF74F6D36847ED1891AC612FCFAEB7FB551E14` | 0.176527 |
| Injective | channel-122 | channel-8 | `transfer/channel-122/transfer/channel-89/stuatom` | `ibc/F65724D2AE4A14F5BC149FC12C984D53D2307D95EC57BBA1CE52F94EB670EF60` | 0 |
| Secret (channel-88 pair) | channel-88 | channel-1 | `transfer/channel-88/transfer/channel-37/stuatom` | `ibc/8AEB813EE960508AEDC1C2EB605788CE6A32F4E632583336D840BE4B8EC24CC7` | 0 |
| Secret (channel-476 pair) | channel-476 | channel-44 | `transfer/channel-476/transfer/channel-37/stuatom` | `ibc/857DD753120BBE0D39C15CD0A4DB8396B4C80EF756A9B8F89986E3A533F37B07` | 0 |
| Penumbra | channel-79703 | channel-4 | `transfer/channel-79703/transfer/channel-8/stuatom` | `ibc/B66737925072CEF58C5E9990038D5B869D778DC42C7C2F1F5CEE8665D907AE8B` | 0 |
| Kujira | channel-259 | channel-3 | `transfer/channel-259/transfer/channel-32/stuatom` | `ibc/DED75871F78AF8FC9BCFE75BEA82D66A2B2366204E210FD8E4C77A2AAEA1B1E3` | 0 |
| Agoric | channel-320 | channel-1 | `transfer/channel-320/transfer/channel-59/stuatom` | `ibc/C86C2FA56D954AB05960450215E63605528CB3481694ABEA87CE4DB0EF17D265` | 0.782288 |

Both Secret pairs are tagged preferred in the chain registry; a Secret user's wallet picks
one, so both denoms need adding. The Hub and Agoric denoms already have supply, so
`add_new_assets` can take them without seeding.

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
