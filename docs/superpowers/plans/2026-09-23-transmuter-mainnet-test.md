# Transmuter Mainnet Test (stATOM) — Test Plan

> **For agentic workers:** this plan is executed by a human operator on osmosis-1 with real
> funds, with Claude driving queries and drafting txs. It is NOT for
> subagent-driven-development: nothing here is repo code. Every tx step is signed by the
> operator; Claude never holds keys. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove on Osmosis mainnet, with a few dollars of real stATOM and ATOM, that a
transmuter pool created from code id 996 does what the wind-down design (§7) needs: a
fixed stToken→native rate, funding by the vault only, permissionless redemption at that
rate from every chain stATOM lives on, admin/moderator controls that work, and no hole
that lets value leak or lets a stranger break the pool.

**Architecture:** One throwaway pool `stATOMtest` (stATOM + ATOM, rate frozen at the
redemption rate read at creation), funded with ~10 ATOM. Tests run in four layers:
contract math and state; the two user paths (router swap, join-then-exit); foreign-route
denoms added with `add_new_assets` and redeemed after real IBC hops from the Hub,
Injective, Secret and Agoric; admin, moderator and adversarial probes. The pool is then
frozen. Findings go into `docs/wind-down/transmuter.md` and the spec. Tasks 1–11 were run on
that multi-denom pool (3590); after the design moved to one pool per route, Task 12 re-ran the
core tests on per-route pools.

**Reference:** `docs/wind-down/transmuter.md` (denoms, channels, message formats, every
command used here). Spec: `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` §3, §7.

## Global Constraints

- Pool assets at creation: canonical stATOM `ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901`
  and ATOM `ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2`, in that order.
- Factors: stATOM `1000000000000000000`; ATOM and the alloyed asset `RR × 1e18` with `RR`
  read from `stakeibc/host_zone/cosmoshub-4` `redemption_rate` immediately before Task 2
  and written as an 18-decimal integer (today `2000174393052495819`). Larger factor = less
  valuable unit; see the reference doc before changing anything here.
- Subdenom `stATOMtest`. The real pools will not use a `test` suffix; nothing on chain
  enforces uniqueness, so this cannot collide with or block the real pool.
- Admin and moderator at creation: the operator's key (`$KEY`). A second key (`$KEY2`) is
  needed for the admin-transfer test and as the "stranger".
- Liquidity ceiling: never more than 10 ATOM and 5 stATOM in the pool at once. The ATOM
  side will be arbitraged by third parties as soon as it is funded (the pool pays RR, the
  market pays ~1% less); the expected loss is under $0.20 per 10 ATOM. Re-fund as needed
  and record every stranger swap: it is evidence.
- Every tx: `--gas auto --gas-adjustment 1.5 --gas-prices 0.0025uosmo -o json`, and record
  `txhash`, height and the relevant events in `docs/wind-down/transmuter-test-log.md`
  (create it in Task 1). Query with `--node https://osmosis-rpc.polkachu.com:443`.
- Never run a tx whose expected outcome is not written down first (amount out, error
  string, or state change). Compare, then tick the box.
- Do not push. Branch `wind-down-design`; commit the log and doc updates as you go.
- Penumbra and Kujira are out of reach (clients expired on both hops, see reference §3).
  Their two-hop denoms are still added in Task 7 so the pool shape is rehearsed; they are
  recorded as untestable, not skipped silently.

## Prerequisites (operator)

| Where | Needs | Why |
|---|---|---|
| Osmosis `$KEY` | ≥ 25 allUSDC, ≥ 5 OSMO, 10 ATOM, 5 stATOM | pool fee (20 allUSDC), gas, funding, swaps |
| Osmosis `$KEY2` | ≥ 1 OSMO, 1 stATOM, 1 ATOM | stranger and admin-candidate tests |
| Stride `$SKEY` | 4 stATOM + STRD gas | seeding Hub, Injective, Secret (1 stATOM each, 1 spare) |
| Hub key | 0.5 ATOM gas | forwarding Hub-hop stATOM to Osmosis |
| Injective key | ~0.01 INJ gas | forwarding Injective-hop stATOM |
| Secret key | ~1 SCRT gas | forwarding Secret-hop stATOM |
| hermes | chains stride-1, cosmoshub-4, injective-1, secret-4 in `~/.hermes/config.toml`, funded relayer keys on each | only if a hop is not picked up by public relayers within ~30 min |

Address book: Osmosis addresses for `$KEY`/`$KEY2`, Hub/Injective/Secret addresses derived
from the same mnemonic or separate keys; write them all into the log header.

---

### Task 1: Snapshot and log

**Files:**
- Create: `docs/wind-down/transmuter-test-log.md`

- [ ] **Step 1: Record the live inputs**

```bash
UA="User-Agent: curl/8.0"
curl -s -H "$UA" https://stride-api.polkachu.com/Stride-Labs/stride/stakeibc/host_zone/cosmoshub-4 | python3 -c "import json,sys; h=json.load(sys.stdin)['host_zone']; print(h['redemption_rate'])"
osmosisd q poolmanager trading-pair-taker-fee $STATOM $ATOM --node $OSMO_NODE
osmosisd q poolmanager params --node $OSMO_NODE | grep -A3 pool_creation_fee
osmosisd q bank balances <KEY addr> --node $OSMO_NODE
curl -s "https://sqs.osmosis.zone/router/quote?tokenIn=10000000$STATOM&tokenOutDenom=$ATOM" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['amount_out'], [[p['id'] for p in r['pools']] for r in d['route']])"
```

Expected: RR ≈ `2.0001…`, taker fee `0.0002`, fee `20000000 …/alloyed/allUSDC`, balances
cover the prerequisites, the market quote is below RR. Write all five into the log with
the timestamp.

- [ ] **Step 2: Off-chain arithmetic check for the real pool**

```bash
python3 - <<'EOF'
from math import gcd
rr = 2000174393052495819; st = 10**18
lcm = rr*st//gcd(rr,st)
worst_st = 1_300_000 * 10**6          # all stATOM in supply
worst_atom = int(1_300_000 * 2.0002 * 10**6)
norm = worst_st*lcm//st + worst_atom*lcm//rr
print("lcm", lcm, "<= u128", lcm < 2**128)
print("normalized total", norm, "<= u128", norm < 2**128)
print("alloyed mint for all stATOM", worst_st*rr//st, "<= u128", worst_st*rr//st < 2**128)
EOF
```

Expected: all three `True`. Record the numbers; this is the overflow argument for the
real pool.

- [ ] **Step 3: Commit the log skeleton**

```bash
git add docs/wind-down/transmuter-test-log.md && git commit -m "docs(wind-down): transmuter mainnet test log, snapshot"
```

### Task 2: Create the pool

- [ ] **Step 1: Write `instantiate.json`** (in the scratchpad, not the repo) with the RR
  from Task 1:

```json
{"pool_asset_configs":[{"denom":"ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901","normalization_factor":"1000000000000000000"},{"denom":"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2","normalization_factor":"<RR×1e18>"}],"alloyed_asset_subdenom":"stATOMtest","alloyed_asset_normalization_factor":"<RR×1e18>","admin":"<KEY addr>","moderator":"<KEY addr>"}
```

- [ ] **Step 2: Dry-run, then create**

```bash
osmosisd tx cosmwasmpool create-pool 996 "$(cat instantiate.json)" --from $KEY $OSMO_TX --dry-run
osmosisd tx cosmwasmpool create-pool 996 "$(cat instantiate.json)" --from $KEY $OSMO_TX
osmosisd q tx <txhash> --node $OSMO_NODE -o json | python3 -c "import json,sys; d=json.load(sys.stdin); print([a for e in d['events'] if e['type'] in ('pool_created','instantiate') for a in e['attributes']])"
```

Expected: tx code 0; `pool_created` event with `pool_id`; `instantiate` event with the
contract address; 20 allUSDC left `$KEY`. Record `POOL_ID` and `POOL`.

- [ ] **Step 3: Verify initial state**

```bash
for q in '{"list_asset_configs":{}}' '{"get_admin":{}}' '{"get_moderator":{}}' '{"is_active":{}}' '{"get_share_denom":{}}' '{"get_total_pool_liquidity":{}}' '{"list_limiters":{}}' '{"get_swap_fee":{}}'; do osmosisd q wasm contract-state smart $POOL "$q" --node $OSMO_NODE; done
osmosisd q wasm contract-state smart $POOL "{\"spot_price\":{\"base_asset_denom\":\"$STATOM\",\"quote_asset_denom\":\"$ATOM\"}}" --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL "{\"spot_price\":{\"base_asset_denom\":\"$ATOM\",\"quote_asset_denom\":\"$STATOM\"}}" --node $OSMO_NODE
osmosisd q wasm contract $POOL --node $OSMO_NODE
```

Expected: three asset configs (stATOM `1e18`, ATOM `RR×1e18`, alloyed
`factory/$POOL/alloyed/stATOMtest` `RR×1e18`); admin and moderator `$KEY`; active `true`;
liquidity two zero coins; no limiters; swap fee `0`; stATOM/ATOM spot price `= RR` and the
inverse `= 1/RR`; wasm `admin` is `osmo1rxjakgd8yhks2j7hc7pt6a22z3zd64grexpyf7` (the
module). If the spot price comes out as `1/RR` the factors are inverted: stop, the pool is
unusable, create another (a second 20 allUSDC).

- [ ] **Step 4: Quote math before any liquidity**

```bash
osmosisd q wasm contract-state smart $POOL "{\"calc_out_amt_given_in\":{\"token_in\":{\"denom\":\"$STATOM\",\"amount\":\"1000000\"},\"token_out_denom\":\"$ATOM\",\"swap_fee\":\"0\"}}" --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL "{\"calc_in_amt_given_out\":{\"token_out\":{\"denom\":\"$ATOM\",\"amount\":\"2000000\"},\"token_in_denom\":\"$STATOM\",\"swap_fee\":\"0\"}}" --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL "{\"calc_out_amt_given_in\":{\"token_in\":{\"denom\":\"$STATOM\",\"amount\":\"1000000\"},\"token_out_denom\":\"$ATOM\",\"swap_fee\":\"0.001\"}}" --node $OSMO_NODE
```

Expected: the `calc_*` queries simulate the swap against pool state, so with an empty
pool the first two fail with `InsufficientPoolAsset` (required vs available `0`); the
third fails with `InvalidSwapFee` regardless of liquidity (only zero is accepted, which
is what the router passes). Re-run the first two after Task 3, when they must return
`floor(1e6 × RR)` = `2000174` uatom and `ceil(2e6 / RR)` = `999913` ustatom (recompute
both if RR differs from `2.000174393052495819`).

- [ ] **Step 5: Log and commit.**

### Task 3: Fund like the vault will

- [ ] **Step 1: Join with ATOM only**

```bash
osmosisd tx wasm execute $POOL '{"join_pool":{}}' --amount 10000000$ATOM --from $KEY $OSMO_TX
osmosisd q wasm contract-state smart $POOL '{"get_total_pool_liquidity":{}}' --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL '{"get_shares":{"address":"<KEY addr>"}}' --node $OSMO_NODE
osmosisd q bank balances <KEY addr> --node $OSMO_NODE
```

Expected: liquidity ATOM `10000000`, stATOM `0`; shares (alloyed) `10000000` exactly
(alloyed factor equals ATOM's); `$KEY` holds `10000000 factory/$POOL/alloyed/stATOMtest`.

- [ ] **Step 2: Re-run the Task 2 Step 4 quotes.** Expected: `2000174` and `999913`.

- [ ] **Step 3: Router sees the pool**

```bash
curl -s "https://sqs.osmosis.zone/router/routes?tokenIn=$STATOM&tokenOutDenom=$ATOM" | python3 -c "import json,sys; print(sorted(json.load(sys.stdin)['UniquePoolIDs']))"
curl -s "https://sqs.osmosis.zone/router/quote?tokenIn=1000000$STATOM&tokenOutDenom=$ATOM" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['amount_out'], [[p['id'] for p in r['pools']] for r in d['route']])"
```

Expected: `$POOL_ID` appears among the candidates within a few minutes (SQS ingests
each block); the 1 stATOM quote routes through `$POOL_ID` for ~`1999774` (= 2000174 minus
the 0.02% taker fee on the way in, rounding per poolmanager; record the exact figure).
If SQS quotes a different pool at a better price, the arb has not happened yet and the
pool is still cheaper than market: note it. Then open app.osmosis.zone, set up a
1 stATOM → ATOM swap with `$KEY2`'s wallet and screenshot the route it proposes. Do not
execute from the app yet.

- [ ] **Step 4: Watch for strangers for 15 minutes**

```bash
osmosisd q wasm contract-state smart $POOL '{"get_total_pool_liquidity":{}}' --node $OSMO_NODE
osmosisd q txs --events "wasm._contract_address=$POOL" --node $OSMO_NODE --limit 20
```

Expected: possibly a third-party swap draining ATOM at RR. Either outcome is fine;
record it. If ATOM is gone, `join_pool` another 5 ATOM before Task 4.

### Task 4: The two user paths

- [ ] **Step 1: Router exact-in**

```bash
osmosisd tx poolmanager swap-exact-amount-in 1000000$STATOM 1990000 --swap-route-pool-ids $POOL_ID --swap-route-denoms $ATOM --from $KEY2 $OSMO_TX
```

Expected (write before sending): `token_swapped` event with `tokens_in 1000000…stATOM`,
`tokens_out` = `floor((1000000 − taker_fee) × RR)`; a `taker_fee` attribute of ~`200`
ustatom; ATOM balance of `$KEY2` up by that amount; pool stATOM `= 1000000 − 200`,
pool ATOM down by the out amount.

- [ ] **Step 2: Router exact-out**

```bash
osmosisd tx poolmanager swap-exact-amount-out 2000000$ATOM 1010000 --swap-route-pool-ids $POOL_ID --swap-route-denoms $STATOM --from $KEY2 $OSMO_TX
```

Expected: `$KEY2` receives exactly `2000000` uatom; pays `ceil(2000000 / RR)` = `999913`
plus the taker fee on top (`token_in_max` 1010000 is enough).

- [ ] **Step 3: Direct path, no taker fee** (join with stATOM, exit with ATOM)

```bash
osmosisd tx wasm execute $POOL '{"join_pool":{}}' --amount 1000000$STATOM --from $KEY2 $OSMO_TX
osmosisd q wasm contract-state smart $POOL '{"get_shares":{"address":"<KEY2 addr>"}}' --node $OSMO_NODE
osmosisd tx wasm execute $POOL "{\"exit_pool\":{\"tokens_out\":[{\"denom\":\"$ATOM\",\"amount\":\"2000174\"}]}}" --from $KEY2 $OSMO_TX
```

Expected: shares minted `floor(1e6 × RR)` = `2000174`; exit of `2000174` uatom burns
exactly `2000174` alloyed and leaves `$KEY2` with zero alloyed; net 1 stATOM → 2.000174
ATOM with no fee at all. This is the path a script or a power user would take.

- [ ] **Step 4: Rounding edges**

```bash
osmosisd q wasm contract-state smart $POOL "{\"calc_out_amt_given_in\":{\"token_in\":{\"denom\":\"$STATOM\",\"amount\":\"1\"},\"token_out_denom\":\"$ATOM\",\"swap_fee\":\"0\"}}" --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL "{\"calc_out_amt_given_in\":{\"token_in\":{\"denom\":\"$ATOM\",\"amount\":\"1\"},\"token_out_denom\":\"$STATOM\",\"swap_fee\":\"0\"}}" --node $OSMO_NODE
osmosisd tx poolmanager swap-exact-amount-in 1$ATOM 0 --swap-route-pool-ids $POOL_ID --swap-route-denoms $STATOM --from $KEY2 $OSMO_TX
```

Expected: `1` ustatom → `2` uatom; `1` uatom → `0` ustatom in the query; the tx with
1 uatom in fails (`ZeroValueOperation` from the contract, or the poolmanager's own
zero-out check). No dust can be extracted for free in either direction.

- [ ] **Step 5: Reverse direction is allowed** (ATOM → stATOM at RR)

```bash
osmosisd tx poolmanager swap-exact-amount-in 2000174$ATOM 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $STATOM --from $KEY2 $OSMO_TX
```

Expected: succeeds, ~`999…` ustatom out. Note for the spec: the pool is two-way; nobody
rational buys stATOM above market, but nothing forbids it. Task 8 tests the lever that
forbids it if we want one.

- [ ] **Step 6: Larger than the pool**

```bash
osmosisd tx poolmanager swap-exact-amount-in 4000000$STATOM 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $ATOM --from $KEY $OSMO_TX
```

(with less than 8 ATOM in the pool.) Expected: fails with `InsufficientPoolAsset
{required, available}`; nothing moves. Record the exact error text: it is what a user
sees when a pool is under-funded.

- [ ] **Step 7: Log and commit.**

### Task 5: Seed the foreign hops from Stride

Only Hub, Injective and Secret can be seeded (reference §3). Agoric already has a
two-hop supply on Osmosis but we hold none; Task 7 adds its denom without swapping it.

- [ ] **Step 1: Send 1 stATOM from Stride to each**

```bash
strided tx ibc-transfer transfer transfer channel-0  <hub addr>    1000000stuatom --from $SKEY --node $STRIDE_NODE --chain-id stride-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.005ustrd -y
strided tx ibc-transfer transfer transfer channel-6  <inj addr>    1000000stuatom --from $SKEY ...
strided tx ibc-transfer transfer transfer channel-40 <secret addr> 1000000stuatom --from $SKEY ...
```

Record each `send_packet` sequence.

- [ ] **Step 2: Confirm arrival within 30 minutes**

```bash
gaiad q bank balances <hub addr> --node https://cosmos-rpc.polkachu.com:443 | grep -A1 B05539
injectived q bank balances <inj addr> --node https://injective-rpc.polkachu.com:443 | grep -A1 A8F392
secretd q bank balances <secret addr> --node <secret rpc> | grep -A1 A0E80E
```

Expected: `1000000` of the chain's one-hop denom (reference §2). If one has not arrived:

```bash
hermes query packet pending --chain stride-1 --port transfer --channel <stride channel>
hermes clear packets --chain stride-1 --port transfer --channel <stride channel>
```

after adding that chain to `~/.hermes/config.toml` (copy the stride-1 block, change id,
rpc/grpc, `key_name`, gas denom; `hermes keys add --chain <id> --mnemonic-file …`). If a
packet times out it refunds on Stride; resend.

- [ ] **Step 3: Log which routes needed manual relaying.** That answer feeds the spec
  (§7 says "ops seed each one with a small transfer").

### Task 6: Forward each to Osmosis as a user would

- [ ] **Step 1: One transfer per chain**

```bash
gaiad      tx ibc-transfer transfer transfer channel-141 <KEY addr> 1000000ibc/B05539B66B72E2739B986B86391E5D08F12B8D5D2C2A7F8F8CF9ADF674DFA231 --from <hub key> ...
injectived tx ibc-transfer transfer transfer channel-8   <KEY addr> 1000000ibc/A8F39212ED30B6A8C2AC736665835720D3D7BE4A1D18D68566525EC25ECF1C9B --from <inj key> ...
secretd    tx ibc-transfer transfer transfer channel-1   <KEY addr> 500000ibc/A0E80E59956C754F1D9CB37234D13E0CF2949E7254896359F284512FA8428E18 --from <secret key> ...
secretd    tx ibc-transfer transfer transfer channel-44  <KEY addr> 500000ibc/A0E80E59956C754F1D9CB37234D13E0CF2949E7254896359F284512FA8428E18 --from <secret key> ...
```

- [ ] **Step 2: Confirm the denoms that land**

```bash
osmosisd q bank balances <KEY addr> --node $OSMO_NODE
osmosisd q ibc-transfer denom-trace <hash> --node $OSMO_NODE   # for each new ibc/ denom
```

Expected: exactly the four two-hop denoms in reference §2 (`ibc/7451…` Hub, `ibc/F657…`
Injective, `ibc/8AEB…` Secret via channel-88, `ibc/857D…` Secret via channel-476), traces
`transfer/<osmosis ch>/transfer/<X ch>/stuatom`. Any other hash means a wallet or relayer
used a non-preferred channel: record it, it is a denom the real pool must also add.

- [ ] **Step 3: Log and commit.**

### Task 7: Add the foreign-route denoms

- [ ] **Step 1: Negative test first, on a denom with zero supply**

```bash
osmosisd tx wasm execute $POOL '{"add_new_assets":{"asset_configs":[{"denom":"ibc/B66737925072CEF58C5E9990038D5B869D778DC42C7C2F1F5CEE8665D907AE8B","normalization_factor":"1000000000000000000"}]}}' --from $KEY $OSMO_TX
```

(the Penumbra two-hop denom, supply 0). Expected: fails `DenomHasNoSupply`. Then the
same from `$KEY2`: expected `Unauthorized`. Then with `--amount 1uosmo` from `$KEY`:
expected the nonpayable error.

- [ ] **Step 2: Add the four that have supply, all with stATOM's factor**

```bash
osmosisd tx wasm execute $POOL '{"add_new_assets":{"asset_configs":[
 {"denom":"ibc/7451074F46885686D3B47B12A6BF74F6D36847ED1891AC612FCFAEB7FB551E14","normalization_factor":"1000000000000000000"},
 {"denom":"ibc/F65724D2AE4A14F5BC149FC12C984D53D2307D95EC57BBA1CE52F94EB670EF60","normalization_factor":"1000000000000000000"},
 {"denom":"ibc/8AEB813EE960508AEDC1C2EB605788CE6A32F4E632583336D840BE4B8EC24CC7","normalization_factor":"1000000000000000000"},
 {"denom":"ibc/857DD753120BBE0D39C15CD0A4DB8396B4C80EF756A9B8F89986E3A533F37B07","normalization_factor":"1000000000000000000"},
 {"denom":"ibc/C86C2FA56D954AB05960450215E63605528CB3481694ABEA87CE4DB0EF17D265","normalization_factor":"1000000000000000000"}]}}' --from $KEY $OSMO_TX
osmosisd q wasm contract-state smart $POOL '{"list_asset_configs":{}}' --node $OSMO_NODE
```

Expected: success; seven asset configs plus the alloyed one; the spot price of each new
denom against ATOM equals RR; existing liquidity and shares unchanged. (Agoric's
`ibc/C86C…` is included because supply exists.)

- [ ] **Step 3: Redeem each two-hop denom both ways**

For each of Hub, Injective, Secret-88, Secret-476 denoms `$D` held by `$KEY`:

```bash
osmosisd tx poolmanager swap-exact-amount-in 500000$D 990000 --swap-route-pool-ids $POOL_ID --swap-route-denoms $ATOM --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"join_pool":{}}' --amount 500000$D --from $KEY $OSMO_TX
```

Expected: router path pays `floor((500000 − taker fee) × RR)`; note that the taker fee for
a two-hop denom/ATOM pair is the **default 0.1%**, not the 0.02% override (verify with
`trading-pair-taker-fee $D $ATOM` and record). Join mints `floor(500000 × RR)` alloyed.
Then a two-hop → canonical stATOM swap through the pool (`--swap-route-denoms $STATOM`):
expected 1:1 minus taker fee, which is a free "de-hop" service the pool provides; note it.

- [ ] **Step 4: Router discovery of the new denoms**

```bash
curl -s "https://sqs.osmosis.zone/router/quote?tokenIn=100000ibc/7451074F46885686D3B47B12A6BF74F6D36847ED1891AC612FCFAEB7FB551E14&tokenOutDenom=$ATOM"
```

Expected: either a route through `$POOL_ID` or an error that the token is unknown. The
Osmosis asset list does not know two-hop stATOM, so the app will most likely show it as
an unlisted or unknown asset. Record what the app shows for `$KEY`'s balance of `ibc/7451…`
and whether it can be swapped from the UI. This decides whether the spec needs a user
guide (direct contract execution or a hosted page) for foreign-route holders.

- [ ] **Step 5: Log and commit.**

### Task 8: Admin, moderator and adversarial probes

- [ ] **Step 1: Stranger cannot administer** (`$KEY2` on each; expected `Unauthorized`):
  `set_active_status`, `mark_corrupted_assets`, `rescale_normalization_factor`,
  `register_limiter`, `assign_moderator`, `transfer_admin`, `claim_admin` (no candidate:
  `Unauthorized`).

- [ ] **Step 2: Freeze**

```bash
osmosisd tx wasm execute $POOL '{"set_active_status":{"active":false}}' --from $KEY $OSMO_TX
osmosisd tx poolmanager swap-exact-amount-in 100000$STATOM 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $ATOM --from $KEY2 $OSMO_TX
osmosisd tx wasm execute $POOL '{"join_pool":{}}' --amount 100000$STATOM --from $KEY2 $OSMO_TX
osmosisd tx wasm execute $POOL "{\"exit_pool\":{\"tokens_out\":[{\"denom\":\"$ATOM\",\"amount\":\"1000\"}]}}" --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"set_active_status":{"active":false}}' --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"set_active_status":{"active":true}}' --from $KEY $OSMO_TX
```

Expected: swap, join and the admin's own exit all fail `InactivePool`; re-setting
`false` fails `UnchangedActiveStatus`; `true` restores everything. Also check SQS drops
the pool from routes while inactive.

- [ ] **Step 3: Corrupted-asset lever** (candidate one-way switch for the real pool)

```bash
osmosisd tx wasm execute $POOL "{\"mark_corrupted_assets\":{\"denoms\":[\"$ATOM\"]}}" --from $KEY $OSMO_TX
osmosisd tx poolmanager swap-exact-amount-in 100000$STATOM 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $ATOM --from $KEY2 $OSMO_TX      # stATOM in, ATOM out
osmosisd tx poolmanager swap-exact-amount-in 100000$ATOM 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $STATOM --from $KEY2 $OSMO_TX     # ATOM in, stATOM out
osmosisd tx wasm execute $POOL '{"join_pool":{}}' --amount 100000$ATOM --from $KEY $OSMO_TX                                                   # vault tops up ATOM
osmosisd tx wasm execute $POOL "{\"exit_pool\":{\"tokens_out\":[{\"denom\":\"$STATOM\",\"amount\":\"100000\"}]}}" --from $KEY $OSMO_TX         # vault takes stATOM only
osmosisd tx wasm execute $POOL "{\"unmark_corrupted_assets\":{\"denoms\":[\"$ATOM\"]}}" --from $KEY $OSMO_TX
```

Expected, with ATOM marked corrupted: stATOM→ATOM succeeds (ATOM amount and weight
fall); ATOM→stATOM fails `CorruptedAssetRelativelyIncreased`; the vault's ATOM top-up
also fails (amount increase) — so this lever blocks *funding* too, and can only be
flipped after the pool is fully funded; exiting stATOM only fails (ATOM weight rises).
Unmark restores all four. Write the conclusion into the spec's §7 as an optional
hardening with that ordering constraint.

- [ ] **Step 4: Rescale is uniform**

```bash
osmosisd tx wasm execute $POOL '{"rescale_normalization_factor":{"numerator":"1","denominator":"7"}}' --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"rescale_normalization_factor":{"numerator":"1","denominator":"1000000000000"}}' --from $KEY $OSMO_TX
osmosisd q wasm contract-state smart $POOL '{"list_asset_configs":{}}' --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL "{\"spot_price\":{\"base_asset_denom\":\"$STATOM\",\"quote_asset_denom\":\"$ATOM\"}}" --node $OSMO_NODE
```

Expected: `/7` fails `NonDivisibleRescaleError` (RR×1e18 is not divisible by 7 — check
with python first; pick a prime that does not divide it); `/1e12` succeeds only if every
factor is divisible by 1e12, which `2000174393052495819` is not, so expect the same
error. Then `{"numerator":"2","denominator":"1"}`: succeeds, all factors doubled, spot
price unchanged. Rescale back with `1/2`.

- [ ] **Step 5: Per-route static cap** (the candidate safety lever for the real pool,
  see reference §1 "Limiters"). Register it on a two-hop denom, never on canonical stATOM
  or ATOM. The last limiter on a denom can never be deregistered, only widened, so this
  step leaves a permanent limiter on the test pool's Hub denom; that is fine for a
  throwaway pool.

```bash
HUB2=ibc/7451074F46885686D3B47B12A6BF74F6D36847ED1891AC612FCFAEB7FB551E14
# pool holds ~10 ATOM of value; cap the Hub route at 5% of pool value = ~0.5 ATOM = ~0.25 stATOM
osmosisd tx wasm execute $POOL "{\"register_limiter\":{\"denom\":\"$HUB2\",\"label\":\"route-cap\",\"limiter_params\":{\"static_limiter\":{\"upper_limit\":\"0.05\"}}}}" --from $KEY $OSMO_TX
osmosisd q wasm contract-state smart $POOL '{"list_limiters":{}}' --node $OSMO_NODE
osmosisd tx poolmanager swap-exact-amount-in 100000$HUB2 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $ATOM --from $KEY $OSMO_TX     # 0.1 stATOM: under the cap
osmosisd tx poolmanager swap-exact-amount-in 300000$HUB2 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $ATOM --from $KEY $OSMO_TX     # would take the route past 5%
osmosisd tx poolmanager swap-exact-amount-in 50000$ATOM 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $HUB2 --from $KEY $OSMO_TX      # takes the route OUT: always allowed
osmosisd tx wasm execute $POOL "{\"deregister_limiter\":{\"denom\":\"$HUB2\",\"label\":\"route-cap\"}}" --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL "{\"set_static_limiter_upper_limit\":{\"denom\":\"$HUB2\",\"label\":\"route-cap\",\"upper_limit\":\"1\"}}" --from $KEY $OSMO_TX
osmosisd tx poolmanager swap-exact-amount-in 300000$HUB2 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $ATOM --from $KEY $OSMO_TX     # now passes
```

Expected, in order: limiter listed as `(denom, "route-cap") StaticLimiter{0.05}`; the
0.1 stATOM swap passes; the 0.3 stATOM swap fails `UpperLimitExceeded {denom, upper_limit
0.05, value …}` and moves nothing; the ATOM→Hub-stATOM swap passes because the Hub
denom's weight falls; `deregister_limiter` fails `EmptyLimiterNotAllowed` (it is the
denom's only limiter); widening to `1` succeeds; the 0.3 stATOM swap then passes. Also
confirm the cap counts *normalized value*: 0.25 stATOM at RR ≈ 0.5 ATOM ≈ 5% of a 10 ATOM
pool. Then test that the cap does not bite the vault: `join_pool` with ATOM and
`exit_pool` with ATOM both pass with the limiter in place (ATOM has no limiter, and the
Hub denom's weight only falls or stays).

- [ ] **Step 6: Admin hand-over, two-step**

```bash
osmosisd tx wasm execute $POOL '{"transfer_admin":{"candidate":"<KEY2 addr>"}}' --from $KEY $OSMO_TX
osmosisd q wasm contract-state smart $POOL '{"get_admin_candidate":{}}' --node $OSMO_NODE
osmosisd tx wasm execute $POOL '{"add_new_assets":{"asset_configs":[{"denom":"ibc/B66737925072CEF58C5E9990038D5B869D778DC42C7C2F1F5CEE8665D907AE8B","normalization_factor":"1000000000000000000"}]}}' --from $KEY2 $OSMO_TX     # candidate is not admin yet: Unauthorized (an empty list would fail the non-empty check before the admin check)
osmosisd tx wasm execute $POOL '{"claim_admin":{}}' --from $KEY2 $OSMO_TX
osmosisd q wasm contract-state smart $POOL '{"get_admin":{}}' --node $OSMO_NODE
osmosisd tx wasm execute $POOL '{"assign_moderator":{"address":"<KEY2 addr>"}}' --from $KEY2 $OSMO_TX
osmosisd tx wasm execute $POOL '{"transfer_admin":{"candidate":"<KEY addr>"}}' --from $KEY2 $OSMO_TX
osmosisd tx wasm execute $POOL '{"claim_admin":{}}' --from $KEY $OSMO_TX
osmosisd tx wasm execute $POOL '{"assign_moderator":{"address":"<KEY addr>"}}' --from $KEY $OSMO_TX
```

Expected: candidate visible; `$KEY2` unauthorized until claim; after claim `$KEY2` is
admin and can set the moderator; hand back the same way. This is the rehearsal for
moving both roles to the Osmosis vault multisig. Also test `cancel_admin_transfer` from
the admin and `reject_admin_transfer` from the candidate once each.

- [ ] **Step 7: Alloyed asset held by a stranger**

```bash
osmosisd tx bank send $KEY <KEY2 addr> 1000000factory/$POOL/alloyed/stATOMtest $OSMO_TX
osmosisd tx wasm execute $POOL "{\"exit_pool\":{\"tokens_out\":[{\"denom\":\"$ATOM\",\"amount\":\"1000000\"}]}}" --from $KEY2 $OSMO_TX
osmosisd tx wasm execute $POOL "{\"exit_pool\":{\"tokens_out\":[{\"denom\":\"$ATOM\",\"amount\":\"1\"}]}}" --from $KEY2 $OSMO_TX
osmosisd tx poolmanager swap-exact-amount-in 500000factory/$POOL/alloyed/stATOMtest 1 --swap-route-pool-ids $POOL_ID --swap-route-denoms $STATOM --from $KEY2 $OSMO_TX
```

Expected: the alloyed token is a bearer claim: `$KEY2` exits 1 ATOM; the second exit
fails `InsufficientShares`; the alloyed→stATOM router swap works too. Conclusion for the
spec: the vault's alloyed balance is the withdrawal key and must be held like funds.

- [ ] **Step 8: Log and commit.**

### Task 9: Teardown

- [ ] **Step 1: Empty the pool**

```bash
osmosisd q wasm contract-state smart $POOL '{"get_total_pool_liquidity":{}}' --node $OSMO_NODE
osmosisd q wasm contract-state smart $POOL '{"get_total_shares":{}}' --node $OSMO_NODE
# exit_pool with every non-zero coin listed, from each holder of alloyed, until liquidity is all zeros
```

Expected: total shares burn to zero when liquidity reaches zero (they are the same
value by construction); if a stranger holds alloyed, a matching residue stays in the
pool — record it.

- [ ] **Step 2: Freeze it for good** (`set_active_status false`) and confirm SQS drops
  the pool from `router/routes`. Record `POOL_ID`, `$POOL` and the alloyed denom in the
  reference doc under a "test pool" heading so nobody mistakes it for the real one.

- [ ] **Step 3: Reconcile** every balance against the log: total ATOM and stATOM spent
  = gas + taker fees + arb loss + dust rounding. Any unexplained delta is a finding.

### Task 10: Write-up

- [ ] **Step 1:** Update `docs/wind-down/transmuter.md` with every measured number that
  differed from the prediction, the exact error strings, the app/SQS behaviour for
  two-hop denoms, and which hops needed manual relaying.
- [ ] **Step 2:** Update the spec: §3 transmuter facts (factor orientation is already
  fixed; add taker fee, creation fee in allUSDC, gov-migration exposure, the
  corrupted-asset ordering constraint), §7 (funding order, `add_new_assets` list per
  token now known from the escrow map, user guide need for two-hop holders), §11 (Penumbra
  and Kujira stranded holders; whether to pursue client recovery).
- [ ] **Step 3:** Commit; do not push.

## Parallel-safe tasks

None: this is a single mainnet pool driven by one operator, and each task's state is the
input to the next. Tasks 5 and 6 can overlap with Task 4 in wall-clock time (IBC hops take
minutes) but are sequenced here for a single log.

### Task 11: Adversarial pass (added 2026-09-24)

Goal: try to take value out of the pool that was not put in, and try to break redemption for others.
Pool 3590 is unfrozen for this and frozen again at the end. Every step states its prediction first.

- [ ] **11.1 Rounding harvest.** Off-chain: for every ordered pair of the pool's assets plus the alloyed
  asset, 200 random amounts each through `calc_out_amt_given_in` and `calc_in_amt_given_out`; assert
  `out ≤ exact` and `in ≥ exact` (exact = amount × out_factor / in_factor as a rational). Then execute 10
  random swaps on chain and assert the pool's normalized value never decreases.
- [ ] **11.2 Multi-leg routes.** Router route `[3590 stATOM→alloyed, 3590 alloyed→ATOM]` and
  `[3590 Hub-stATOM→stATOM, 3590 stATOM→ATOM]` versus the direct swap of the same amount. Expected: never
  more out than direct.
- [ ] **11.3 Multi-asset join and exit.** `join_pool` with stATOM + Hub-stATOM + ATOM in one call
  (expected mint = sum of floors); `exit_pool` with `[1 uatom, 1 ustatom, 1 Hub-stATOM]` (expected burn
  = 1 + 3 + 3 = 7 alloyed, value received 5.0003).
- [ ] **11.4 Wrong inputs.** Swap OSMO→ATOM through 3590 (`InvalidTransmuteDenom`); stATOM→stATOM
  (`SameDenomNotAllowed`); `join_pool` with alloyed (`InvalidJoinPoolDenom`); exact-out of alloyed via
  the router (works, equals join); exact-in of `340282366920938463463374607431768211455` ustatom
  (bank `insufficient funds`, no state change).
- [ ] **11.5 Bank send to the contract.** Send 1,000 uatom directly to the contract address; expected
  `get_total_pool_liquidity` unchanged, and the contract's bank balance exceeds its liquidity by 1,000.
  No message can withdraw it (exit is bounded by liquidity state).
- [ ] **11.6 Front-run the funding.** Create nothing new: model it on 3590 by the stranger joining
  100,000 ustatom while ATOM is low, then the vault adding ATOM; expected the stranger's alloyed is exactly
  200,017 and exits at exactly that.
- [ ] **11.7 Admin fat-finger.** `add_new_assets` `uosmo` with factor `1000000000000000000000` (1e21,
  i.e. 1 uosmo priced at 0.002 uatom); query `calc_out_amt_given_in` 1,000,000 uosmo → ATOM to show the
  damage a wrong factor would allow; before any swap, `mark_corrupted_assets [uosmo]`; expected: swap
  uosmo→ATOM fails `CorruptedAssetRelativelyIncreased` (amount would rise from 0), join with uosmo fails
  the same way, and after the next unrelated swap the asset is gone from `list_asset_configs`
  (`clean_up_drained_corrupted_assets` removes a corrupted asset with zero balance).
- [ ] **11.8 Limiter bypass.** With the Hub cap re-tightened to current weight + 0.005: `join_pool`
  Hub-stATOM over the cap (expected `UpperLimitExceeded`); `exit_pool` ATOM large enough to push Hub over
  the cap (expected `UpperLimitExceeded`, the vault is bound too); router swap Hub→ATOM under the cap
  (ok). Widen back to 1.
- [ ] **11.9 Agoric route.** Router ATOM→Agoric-stATOM fails (pool holds 0); seed by... not possible
  without holdings; record as "added, not swappable until someone brings some". If the Agoric two-hop
  balance on Osmosis (0.78 held by a third party) cannot be obtained, leave it.
- [ ] **11.10 Freeze again; reconcile value.**

### Task 12: Per-route retest (added and run 2026-09-25)

Goal: re-prove the core behaviour on the current design, one two-asset pool per stToken route.

- [x] **12.1** Drain pool 3590 back to the keys and freeze it.
- [x] **12.2** Create pools: canonical stATOM + ATOM, Hub-hop + ATOM, Secret-hop + ATOM, and one with the
  factors inverted as a negative control. Factors: stToken `1e18`, ATOM and alloyed `RR × 1e18`.
- [x] **12.3** Run `scripts/wind-down/check_transmuter_pool.py` before funding. Expected: three pools pass,
  the inverted pool fails the ratio, orientation, spot-price and quote checks. Freeze it unfunded.
- [x] **12.4** Fund each route pool with exactly its route holdings × RR, the canonical pool with the rest.
- [x] **12.5** Isolation: each flavour through every other pool, join with a foreign denom, ask a route
  pool for canonical stATOM. Expected: all `Unable to transmute token with denom`.
- [x] **12.6** Per pool: router exact-in, exact-out, reverse, dust; a two-pool de-hop route.
- [x] **12.7** One-transaction join + exit from a key with no shares: exact exit succeeds fee-free, an exit
  one unit over reverts the whole tx.
- [x] **12.8** Mark ATOM corrupted on every funded pool. Expected: every cross-pool route, every ATOM → stToken
  swap, joins with ATOM and stToken-only exits refused; redemptions and join+exit still work; vault top-up
  refused until unmarked.
- [x] **12.9** Drain one pool's ATOM to zero with ATOM marked: ATOM is deleted from the pool; recover with
  `add_new_assets` at the same factor plus a join, and confirm the rate.
- [x] **12.10** Stranger authority probes and per-pool freeze isolation; rounding fuzz on each pool; freeze all.

Results: `docs/wind-down/transmuter-test-log.md` Task 12, summarised in `transmuter-test-summary.md`.
